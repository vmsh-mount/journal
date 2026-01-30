---
title: Why Discounts Are Hard (Part I) - Designing a Scalable Discount Engine for E-commerce
date: 2026-01-23
image: 
tags: distributed-systems, system-design, ecommerce, scalability
summary: Discounts look like arithmetic, but at scale they become a distributed read problem. This part builds the mental model — who the system serves, why reads dominate, and how discounts must be reshaped to stay fast under extreme traffic.
---

## The Problem That Looks Simple
Discount looks trivial. 

- _A product costs 500_ 
- _Apply a 10% discount_
- _Show 450_

But this mental model breaks down quickly at scale.

During major sales events, millions of users load product pages simultaneously. Every page load needs to answer a deceptively complex question:
> "What discounts apply to this product, for this user, right now?"

At that point, discounts stop being arithmetic and start becoming a **distributed system problem**.

This post focuses on the design of a scalable discount engine built for a high-traffic e-commerce platform — focusing not just on what we build, but why certain trade-offs are unavoidable.

## One System, Three Perspectives
Rather than starting with APIs or databases, it helps to view the problem through the eyes of the people who depend on the system.

### 1. Ops Teams
Ops teams need to define and manage discounts continuously without engineering support — often during live sale. 

They should be able to:
- _Create product-level discounts (percentage or flat)_
- _Target discounts by product, brand or supplier_
- _Activate, edit, or disable offers safely_

Most importantly, these changes should:
- Take effect quickly
- Require no code deployments
- Never compromise site stability during peak traffic

For ops teams, flexibility matters — but safety matters more.

### 2. End Users
From a customer's point of view, discounts must feel obvious and dependable.

When viewing a product, users expect:
- **Visibility**: All eligible offers clearly displayed
- **Predictability**: The same discounts across sessions and devices
- **Speed**: No perceptible delay in page load

Users need to see eligible discounts, apply offers, subject to per-user redemption limits.

### 3. The Platform
The discount engine sits directly on the critical path of product discovery.

During flash sales or holidays:
- Traffic spikes are extreme and unpredictable
- Read traffic dwarfs writes
- Latency budgets shrink to single-digit milliseconds

The system must be:
- **Low latency**: _Discount evaluation must complete within ~10ms as part of the product fetch API_
- **Highly available**: _The system cannot become a single point of failure during peak traffic_
- **Durability**: _Applied discounts and redemption events must be persisted reliably_
- **Consistency**: _Users should see and redeem discounts consistently across devices_
- Horizontally scalable
- Resilient to partial failures

If the discount engine slows down or fails, it doesn't just break discounts — it degrades the entire shopping experience.

## The Core Abstraction — What is a "Discount"?
The biggest mistake teams make with discounts is treating them as a price calculation problem.

They are not.

In a large e-commerce system, a discount is fundamentally a **decision**.

For every product page view, the system must decide:
> _Should this offer apply here, and if yes, how?_

That decision has two very different phases.

### Phase 1 — Finding What Might Apply
When a user opens a product page, the system cannot afford to "check everything". 
Instead, it needs a fast way to narrow the universe of discounts to a small, relevant set. 
This immediately forces a design choice: discounts must be **described in terms of what they target**.

Some discounts target a single product.<br>
Others target a brand or a supplier.

This targeting is not about price at all — it exists purely to make the system searchable at scale.

If discounts cannot be filtered quickly, latency is already lost.

### Phase 2 — Deciding What Actually Applies
Once we have a small set of candidate discounts, a second decision happens.

Some discounts simply change the price:
- _10% off_
- _10 INR off_

Others introduce conditions:
- _Active only during a sale window_
- _Limited redemptions per user_

This is where discounts stop being stateless.

The moment we introduce limits or usage history, the system must reason about **time**, **state**, and **consistency** — especially across devices and concurrent requests.

This two-phase nature of discounts explains most of the architecture that follows.
- The first phase must be **fast**, **cacheable**, and **read-only**
- The second phase must be **correct**, **durable**, and **concurrency-safe**

#### Note: A Deliberate Simplification
At this stage, discounts are intentionally limited to product-level rules. There are no cart-wide interactions, stacking rules, or personalised pricing.

## Read-Heavy by Design
One thing becomes clear once you think of discounts as a decision made on every product view:
> **_Reads will dominate writes — by several orders of magnitude_**

Ops teams may update discounts a few times an hour.<br>
Product pages are read millions of times in the same window.

### The Key Design Decision
The core architectural choice is simple:
> Ops writes should never be on the critical read path.

This means:
- Ops-facing APIs can afford validation, normalisation, and safety checks.
- User-facing reads must operate on a format optimised for fast lookup and evaluation.

In practice, this implies that discounts are **transformed** as they move from the write path to the read path.

#### I. Write Path — Flexible, Human Friendly
This data is stored in a canonical form — easy to edit, audit, and reason about — but not necessarily fast to query at scale.

That's acceptable. Ops traffic is small.

````java
class Discount {
    String discountId;
    
    Target target; // PRODUCT, BRAND, SUPPLIER
    String targetId;
    
    DiscountType type; // PERCENTAGE, FLAT
    int value;         // 10 (%) or 10 (INR)
    
    Instant validFrom;
    Instant validTill;
    
    Status status;  // ACTIVE, INACTIVE
}
````

#### II. Read Path — Flat, Predictable, Fast
The read path serves the product page. At this point, the system needs answers, not abstractions.

For a given product, it should be able to:
- _Fetch a small set of candidate discounts_
- _Evaluate eligibility deterministically_
- _Return results with minimal computation_

Anything that requires joins, scans or rule interpretation at runtime becomes a liability under load.

## Modeling Discounts for Fast Eligibility Lookup
Imagine a product page request asking:
> Give me all discounts that apply to product P.

If answering that requires:
- _Scanning discounts_
- _Evaluating rule expressions_
- _Resolving joins across entities_

...then the latency budget is already gone.

For a high-traffic system, discounts must be **addressable**, not discoverable.

### Turning Rules into Lookups
The key idea is to pre-compute where a discount should appear.

Instead of asking:
> "Does this discount apply to this product?"

We flip the question:
> "Which discounts should I even consider for this product?"

This naturally leads to modelling discounts by their **targeting dimension**.

```java
class DiscountLookupEntry {
    String lookupKey;   // productId | brandId | supplierId
    String discountId;
    
    DiscountType type;
    int value;
    
    Instant validFrom;
    Instant validTo;
}
```

At runtime, the system simply gathers _discounts directly attached to the `product:{productId}` or `brand:{brandId}` or `supplier:{supplierId}`_.

No rule interpretation. No joins. Just lookups.

When an ops user creates a discount:
- _A brand-level discount is expanded to all products under that brand._
- _A supplier-level discount is expanded similarly._
- _Product level discounts map one-to-one_

This expansion happens outside the critical read path. The read path only ever consumes precomputed entries.

### Why Duplication Is Acceptable Here
This model deliberately duplicates data. The same discount may appear under thousands of products (via brand or supplier)

That's not a flaw — it's the cost of predictable latency. Optimising for reads, even at the cost of duplication, is the only viable choice at scale.

#### The Payoff
With this structure:
- Discount lookup becomes O(1) per dimension
- Results are easy to cache
- The read path remains stateless and fast

But this introduces a new responsibility — **Keeping the read model in sync with ops change**s.

## From Writes to Reads — Propagating Discount Changes Safely
A tempting approach is to update the lookup entries synchronously when an ops user creates or edits a discount.

This fails under real conditions.

Brand or supplier level discounts may fan out to thousands of products. Doing this inline:
- _increases write latency_
- _risks timeouts_
- _Couples ops actions to read-path health_

Worst of all, a partial failure can leave the system in an inconsitent state.

### Asynchronous by Default
The safer approach is to **decouple intent from materialisation**. 

When an ops user creates or updates a discount:
1. The canonical discount record is persisted
2. A change event is emitted
3. The read model is updated asynchronously

The ops API returns success once intent is stored — not when fan-out completes.

```java
class DiscountChangedEvent {
    String discountId;
    ChangeType type;  // CREATED, UPDATED, DEACTIVATED
    Instant eventTime;
}
```

A background worker consumes these events and performs expansion:

```java
void handle(DiscountChangedEvent event) {
    Discount discount = discountStore.get(event.discountId);
    List<String> productIds = resolveTargets(discount);
    for (String productId: productIds) {
        upsertLookupEntry(productId, discount);
    }
}
```

This approach introduces **eventual consistency** — deliberately. A discount change may take a short time to reflect on product pages. That delay is acceptable.

What is not acceptable:
- Partial visibility
- Mixed states within a single product view

To prevent this, lookup updates for a discount are applied in an **idempotent** and **replace-based** manner, not incremental mutations.

The read path always sees a complete, coherent snapshot.

### What We've Avoided (So Far)
Notice what we haven't introduced yet:
- Distributed locks
- Cross-entity transactions
- User-level state

All of that comes later — and only where absolutely necessary.

By treating discounts as a read-time decision and pushing all complexity into the write path, we’ve kept the hot path:
- Stateless
- Cacheable
- Horizontally scalable
- Bounded in both computation and data size

This is what allows the system to survive flash sales without collapsing under its own logic.

But this simplicity has a limit.

So far, discounts are **informational**. <br>
They don’t change system state. <br>
They don’t involve money moving hands. <br>

The moment we introduce **redemption limits**, discounts stop being read-only hints and become stateful commitments. 
Concurrency, correctness, and failure handling can no longer be postponed.

In Part II, we cross that line.
