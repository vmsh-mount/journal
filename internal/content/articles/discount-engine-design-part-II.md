---
title: Why Discounts Are Hard (Part II) - Designing a Scalable Discount Engine for E-commerce
date: 2026-01-30
image: 
tags: distributed-systems, system-design, ecommerce, scalability
summary: Once discounts become redeemable, correctness matters more than speed. This part explores redemption limits, atomicity, idempotency, and how control-plane decisions avoid dangerous fan-out.
---

## Introducing Redemption Limits — When Reads Turn into Writes
Displaying discounts is easy to scale. Redeeming them is not.

The moment we introduce a rule like:
> "Each user can redeem this discount at most N times"

the discount engine stops being a pure read system.

Every "apply discount" action now has consequences:
- _State must be checked_
- _State must be updated_
- _Races must be prevented_

And all of this must happen while traffic is peaking.

### Why This Is a Qualitatively Harder Problem
Two users viewing the same product can safely see the same discount.

Two concurrent requests from the **same user** trying to redeem it cannot both succeed.

That single distinction forces us to answer uncomfortable questions:
- Where is redemption state stored?
- How do we prevent double-spend?
- What consistency guarantees do we actually need?

The product page read path is latency-critical, cache-heavy and fan-out at massive scale. Pulling **user-specific redemption state** into this path would:
- Break cacheability
- Add a write-optimised dependency
- Increase tail latency

That defeats everything we carefully designed earlier. So we should **not** do strict redemption checks during discount display.

The first design decision is to keep **display** and **redemption** paths separate.
- Product page APIs remain read-only
- Redemption happens via a dedicated API

The system only enforces limit when a user explicitly tries to apply a discount.

### Modeling Redemption State
At minimum, the system needs to answer one question efficiently:
> "How many times has user U redeemed discount D?"

This naturally leads to a small, write-optimised records:
```java
class RedemptionRecord {
    String userId;
    String discountId;
    int redemptionCount;
}
```
This data is tiny, but extremely sensitive to concurrency.

### The Concurrency Problem
Consider two near-simultaneous requests:
- Same user
- Same discount
- Last available redemption

A naive read-increment-write flow will fail under race conditions. So the system must enforce redemption limits **atomically**.

This allows us to enforce limits using **single-key atomic operations**, rather than distributed locks.

```java
boolean tryRedeem(String userId, String discountId, int maxLimit) {
    return atomicIncrementIfBelowLimit(userId, discountId, maxLimit);
}
```
Whether this lives in a database, cache or specialised store is a choice we'll see next — but the shape of the operation is fixed.

### Consistency Over Availability (Here)
Unlike display, redemption is not something we can approximate. If the system is unsure whether a redemption succeeded, the safest answer is "no".

This is one place where we consciously trade availability for correctness. A failed redemption is better than an over-redeemed discount.

#### What the User Sees on the Product Page?
Instead of binary "can/cannot redeem", the UI can show:
- All eligible discounts
- A hint if a discount has a redemption limit

Example: "10% off - up to 3 uses per user"

No guarantee is implied at display time. This matches real-world expectations and avoids false promises.

If redemption fails, the user gets a clear, immediate message:
> "you've reached the maximum number of redemptions for this offer"

### Choosing the Right Storage for Redemption State
Redemption storage must support:
- Single-key atomic updates (no read-modify-write races)
- Low latency under contention
- High availability
- Durability (this is money, not analytics)


#### Why a Traditional Relational DB Struggles
A relational database can model this cleanly:
```shell
UPDATE `redemptions` 
SET count = count + 1
WHERE user_id = ? AND discount_id = ?
AND count < max_limit;
```

But under peak load:
- Hot rows become contention points
- Lock waits spike tail latency
- Scaling writes becomes painful

We can shard aggressively, but we're still paying the cost of heavyweight transactions.

This works — until it doesn’t.

#### A Better Fit: Atomic Counters in KV Store
Redemption state is a perfect match for a key-value store that supports atomic operations.

The **key** is simple:
```markdown
redemption:{userId}:{discountId}
```

The **operation** is simpler still: either increment or fail if limit exceeded.
```java
boolean tryRedeem(String key, int maxLimit) {
    long newValue = kvStore.increment(key);
    return newValue <= maxLimit;
}
```
No Locks. No multi-row transactions. No cross-node coordination.

This keeps latency predictable even when the same discount is being redeemed heavily.

#### Durability and Correctness
Pure in-memory storage is not enough. Redemption counts must survive: process restarts, cache eviction, partial outrages.

The counter store must be durable, or backed by durable storage via **write-through** semantics.

Redemption APIs must be **idempotent**. A retry should never double-count. This is achieved by:
- Attaching an idempotency key (_could be orderId or checkoutId_) to each redemption attempt 
- Ensuring the counter increment is applied at most once per key

```java
boolean tryRedeem(String userId, String discountId, String idempotencyKey) {
    if (idempotencyStore.exists(idempotencyKey)) {
        return idempotencyStore.previousResult(idempotencyKey);
    }

    boolean success = incrementIfBelowLimit(userId, discountId);
    
    idempotencyStore.save(idempotencyKey, success);
    return success;
}
```

If downstream systems fail after redemption succeeds, retries remain safe.

---

_Now, let's talk about few scenarios and see whether the approach we took fits or not._

## Scenario #1 — A Discount for 1 Million Products
Assume ops wants to create a discount that applies to ~1 million product IDs. 

At this scale, the question is no longer _can we support it?_ <br>
It becomes _how do we accept the input and process it safely without destabilising the system?_

### Ops API Contract
The first design decision is that 1 million product IDs cannot be sent inline. The ops API accepts references, not raw lists.

Product IDs are uploaded separately (CSV/Parquet)

A typical request looks like this:
```json
{
  "discountType": "PERCENTAGE",
  "value": 10,
  "target": {
    "type": "PRODUCT_SET",
    "source": "S3",
    "location": "s3://ops-bucket/discounts/diwali_sale_products.csv"
  },
  "validFrom": "2026-10-01T00:00:00Z",
  "validTill": "2026-10-05T23:59:59Z"
}
```

### Validating and Persisting the Input
Before accepting the discount:
- The system validates the file format
- Verifies product IDs exist
- Deduplicate IDs
- Computes the approximate cardinality

Once validated, the discount is persisted in canonical form.

```java 
class Discount {
    String discountId;
    DiscountType type;
    int value;

    TargetType target;        // PRODUCT_SET
    String targetReference;   // file location / dataset id

    Instant validFrom;
    Instant validTill;

    Status status; // CREATED | EXPANDING | ACTIVE
}
```
No lookup entries are created yet.

### Processing the Discount (Fan-out Pipeline)
A background worker picks up the discount and transitions it to `EXPANDING`.

Processing happens in fixed-size batches:
1. Read next N product IDs from the dataset (e.g. 2000) (streaming)
2. Generate lookup entries
3. Write them to the read model
4. Checkpoint progress
5. Repeat

Conceptually:
```java
while (hasMoreProducts()) {
    List<String> batch = readNextBatch(2000);
    upsertLookupEntries(batch, discount);
    saveCheckpoint();
}
```

Why this matters:
- Batching controls write pressure
- Checkpoints allow safe restarts
- Failures are localised to a batch

### Read Path Behaviour During Expansion
While expansion is in progress:
- Products already processed show the disount
- Others don't

This is acceptable because:
- There is no partial state per product
- The read path logic remains unchanged
- Consistency is preserved at the product level

Once expansion completes, the discount becomes fully active.

---
## Scenario #2 — 100M Products, 10 Discounts Each, ~100ms SLA
Let's restate the problem in concrete terms.
- 100 million products
- ~10 active discounts per product (on average)
- Product page API must return results in ~100ms
- This API sits on a hot path — every browse, search, recommendation click

At this scale, the question is not _how do we compute discounts?_
It is _how do we guarantee predictable latency for the slowest 1% of requests?_

### Shape the Read Path First (Non-Negotiable)
To engineer latency, the read path must collapse into a **single bounded operation**:
```markdown
hash(productId)
 -> single shard
 -> single KV read
 -> bounded payload
```
Anything unbounded introduces tail latency. <br>
That means:
- No DB joins. 
- No scans.
- No dynamic rule evaluation. 
- No per-request fan-out.
- No synchronous cross-service calls.

This forces the read model to be **product-addressable**:
```markdown
key: product:{productId}
value: [discountEntry1, discountEntry2, ...]
```
Each product resolves to one key and one lookup (bounded list). No branching, no discovery.

### Payload & Bounds (This is the Latency Lever)
Latency at scale is dominated by **network + deserialisation**, not CPU

We impose a hard bounds:
- Max discounts per product: 20 <br> (_Excess discounts are resolved earlier through priority rules or ops validation_)
- Discounts entry size (tight DTO): ~80 bytes

Worst-case payload:
```markdown
20 * 80 bytes ~ 1.6KB (~2 KB on the wire)
```
This keeps:
- The response fits in a single TCP packet
- Deserialisation stays sub-millisecond
- Cache behaviour predictable

No unbounded lists -> no tail spikes

### Data Store Choice (Drive by p99, Not Mean)
This data cannot live in
- RDBMS (joins + disk I/O)
- Search Indexes (Query planning)
- Disk-backed Document stores

To keep p99 stable, reads mut hit **memory first**, not disk.

A **_memory-first KV_** means:
- The key index lives entirely in RAM
- Reads never block on disk I/O
- Disk exists only for durability and recovery

This gives us:
```markdown
network hop -> memory lookup -> response
```

Systems in this class include **Redis**, **Aerospike**, **Dynamo-style** KV stores.

From the product APIs perspective, this KV is **read-only**.

### Cache Hierarchy (Why L1 + L2 Exists)
Product traffic is highly skewed:
- A small % of products get a large % of views
- Flash sales amplify this skew

Relying only on a network KV makes p99 vulnerable to network jitter.

So we use two layers:
- L1 (in-process cache) <br>
  Very small, extremely fast, absorbs hot keys and bursts
- L2 (distributed KV) <br>
  Holds all 100M products, still memory-first, one network hop

This isn't about maximizing cache hit rate — it's about **protecting tail latency** when traffic spikes.

### Shard Sizing and Load Math (Back-of-Envelope)
Let's size this conservatively.

Assume peak traffic:
- 200k product views/second 

Assume:
- 70% served by L1 + warm L2 cache
- 30% hit the primary KV

That 30% give:
```markdown
~60k KV reads/sec
```
Now shard it.

If we run:
- 60 KV nodes (60 shards)
- Uniform hash on productId

Each shard handles:
```markdown
~1k reads/sec
```
Why this matters:
- Memory-first KV nodes comfortably handle **50k+ reads/sec**
- 1k reads/sec is ~1-2% of safe capacity
- Even a 10x spike stays well below saturation

This distance from saturation is what keeps **p99 stable**.

### How This Gets Us to ~100ms End-to-End
We don't aim for 100ms at the discount layer. We aim for **<=30ms p99**.

Rough budget:
- L1 hit: ~1ms
- L2 KV read (network + memory): 8-15ms
- Deserialize + filter: 2-4ms
- Safety buffer: ~5ms

Total:
```markdown
 <= 25-20ms p99
```
That leaves enough headroom for the rest of the product API to stay within ~100ms

### The Core Idea
We don't optimise discounts at runtime. Instead, we:
- Pay the cost during write/expansion
- Bound everything in the read path
- Keep shards far from saturation
- Design for the slowest requests, not the average.

That's how this system stays fast when traffic is highest  — which is the only time it matters.

---
## Scenario #3 — Disabling a Discount (What Actually Happens?)
At first glance, "disable a discount" sounds trivial:
> Flip a flag and move on

At scale, it's anything but.

A single discount might be:
- Applied to millions of products
- Cached across thousands of servers
- Actively being viewed by users right now

So the real question becomes:
> How do we remove a discount safely without breaking latency or consistency?

### What "Disable" Must Mean Semantically
Disabling a discount has two guarantees:
1. **No new user should see it**
2. **No user should be able to apply it**

We do not require:
- Instant global removal
- Strong consistency across all product pages

This distinction is critical.

### Step 1 — Disable at the Source of Truth (Immediate)
The ops action first updates the canonical discount record:
```java
discount.status = Status.DISABLED;
discount.disabledAt = now();
```

This write is fast, strongly consistent and immediately authoritative.

From this point:
- Any **apply/redeem** request must fail
- Even if some product pages still display the discount

This protects money first.

### Step 2 — Cut Visibility on the Read Path
A `DiscountEntry` is deliberately **not fully authoritative**.

It contains:
```java
class DiscountEntry {
    String lookupKey;    
    String discountId;
    DiscountType type;
    int value;
    Instant validFrom;
    Instant validTill;
}
```

Notice what's missing:
- No mutable status
- No enable/disable flag that needs mass updates

This is intentional.

On the read path, discount evaluation becomes:
```markdown
DiscountEntry
-> check validity window
-> check discountId against active-discount registry
-> include or skip
```

**Where does that registry live?**
- A small **in-memory map/cache**:
    ```markdown
    discountId -> status/version
    ```
- Loaded from the canonical store
- Refreshed frequently (or event-driven)

This registry is:
- Tiny (number of active discounts is small)
- Cheap to keep hot in memory
- Shared across requests

So even if a `DiscountEntry` exists:
- if it's `discountId` is `DISABLED` -> it's ignored
No product-level writes required.

#### Why This Works at Scale
Disabling a discount now costs:
- 1 write
- 0 fan-out
- 0 cache invalidation storms

And affects:
- All products immediately (logically)
- All users consistently (for apply path)

Visibility might lag slightly (depending on cache refresh), but correctness does not.

#### What About Physical Cleanup?
Physical deletion of `DiscountEntry`'s is optional **asynchronous**. But correctness never depends on cleanup.

This is the crucial inversion:
> Entries are allowed to be stale; decisions are not.

To conclude this with a One-Line Design Principle:
> **Never propagate control changes by rewriting duplicated data on the hot path.**
> **Gate them centrally, clean up lazily.**

--- 
### Conclusion

> _"Once money is involved, correctness beats availability — and control must stay centralized."_


**A Final Note** <br>
_This write-up is based on a system design interview discussion I was part of. <br>
The problem, constraints, and trade-offs were explored collaboratively during that conversation._