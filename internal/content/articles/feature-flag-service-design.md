---
title: Feature Flags Are Easy — Until They Aren’t
date: 2026-01-31
image: 
tags: distributed-systems, system-design, scalability
summary: Feature flags look trivial at small scale, but once they sit on the hot path they become control-plane infrastructure.
---

Feature flags feel like plumbing.

A boolean switch. <br>
If it's ON, run the new code. <br>
If it's OFF, run the old code.

Most teams start with a config file, a database row, or an environment variable — and for a while, it works.

At scale, feature flags stop being a convenience and quietly turn into one of the **most critical runtime dependencies** in our system.

Every request now asks a new question:
> "_Is this feature enabled for this user, in this context, right now?_"

That question is evaluated on the hot path — on every API call, across dozens (sometimes hundreds) of services — often while the system is already under stress: during an incident, a migration, or a rollback.

At that point, feature flags are no longer configuration. They are **control-plane** infrastructure.

This post looks at feature flags as a systems problem: something that must be fast enough to run per-request, reliable enough to function during failures, 
and predictable enough that flipping a flag never becomes the outage itself.

### Why This Problem Is Subtle
A broken payment system is obvious. <br>
A broken feature-flag system is insidious.

If flag evaluation is **slow**, every request slows down.

If flags are **inconsistent**,  different services — or even different instances of the same service — make conflicting decisions.

If flags are **unavailable**, we lose our kill switch exactly when we need it most.

## Reframing Feature Flags as a Systems Problem
Once feature flags sit on the hot path, the problem stops being **"how do we store flags"** and becomes **"what guarantees do we owe the callers?"**

Every consumer implicitly assumes answers to questions like:
- _How fast is a flag read?_
- _How fresh is the result?_
- _What happens if the flag service is unreachable?_
- _Can two instances disagree?_

These questions matter more than where the flag lives.

### The Read Path Is the System
Writes are rare. Reads happen on every request.

_If a service does 5k RPS and evaluates 6 flags per request, that's 30k flag evaluations per second from a single consumer.
Multiply that across services, regions and environments, and the flag system very quickly becomes one of the highest-QPS dependencies in the stack._

This immediately rules out designs that:
- adds network hops on the critical path
- blocks on a database read
- fan-outs to multiple dependencies

Anything that shows up in p99 latency is paid by every downstream request.

A feature flag system that's "fast on average" but slow at p99 is not usable in production — because the cost is paid by every downstream service.

## Core Abstraction: Deterministic Flag Evaluation
For every request, the system must answer:
> Should this feature be active for **this** request, in **this** service, for **this** user, under **these** conditions?

A flag is not just a stored value. It's a **rule applied over context**:
- user identity
- environment
- request attributes
- rollout state

The system's primary responsibility is to apply that rule consistently.

### Evaluation Is a Pure Function
At the abstraction level that matters, flag evaluation should behave like a pure function:
```java
result = evaluate(flagDefinition, context);
```
Given the same inputs, the output must be the same:
- across services
- across instances
- across time (within a bounded freshness window)

This requirement immediately gives us two important properties:
1. Evaluation must be **side-effect free**
2. Evaluation must be **deterministic**

Anything that violates either of these — time-based randomness, per-instance state, hidden dependencies — will eventually surface as inconsistency.

## A Concrete Model for Flag Evaluation
Once we say "flag evaluation is a deterministic function", we should be able to write it down as code.

### Flag Definition
A flag definition describes rules, not outcomes.
```java
class FlagDefinition {
    String flagKey;
    FlagType type; // INTEGER, BOOLEAN, STRING
    
    Value defaultValue;
    
    List<Rule> rules; // ordered
    Rollout rollout;
}
```
Key properties:
- There is **exactly one default value**
- Rules are **ordered**, not a set
- Rollout is explicit and isolated

### Evaluation Context
The context is everything dynamic and request-specific:
```java
class EvaluationContext {
    String userId;
    String accountId;
    String environment; // prod, staging, etc
    Map<String, String> attributes;
}
```
Two constraints worth calling out explicitly:
1. Context must be **fully provided by the caller**
2. Evaluation never fetches missing context

If evaluation needs to call another service, it no longer belongs on the hot path.

### Rule Model
Rules answer a single question:
> _Does this rule apply to this context?_

```java
interface Rule {
    boolean matches(EvaluationContext ctc);
    Value value();
}
```

A simple rule implementation might look like:

````java
class AttributeEqualsRule implements Rule {
    private final String attribute;
    private final String expected;
    private final Value value;

    public boolean matches(EvaluationContext ctx) {
        String actual = ctx.attributes.get(attribute);
        return expected.equals(actual);
    }

    public Value getValue() {
        return value;
    }
}

````

Rules **do not**:
- mutate state
- depend on time
- call external systems

If a rule can't be evaluated purely from the context, it doesn't belong here.

### Rollout Model
Rollouts are where many systems quietly lose determinism, usually by introducing randomness.

We make it explicit and hash-based.
```java
class Rollout {
    int percentage; // 0-100
    String seed; // stable per flag
}
```
Rollout evaluation must be:
- stable for the same user
- evenly distributed
- independent of instance count

Which leads directly to the evaluation logic.

### The Evaluation Algorithm (Pseudo-code)
This is the entire system, logically speaking.

```java
class FlagEvaluator {

    public EvaluationResult evaluate(FlagDefinition flag, EvaluationContext ct) {
        // 1. Apply rules in order
        for (Rule rule : flag.rules) {
            if (rule.matches(ctx)) {
                return EvaluationResult.of(rule.value(), Source.RULE);
            }
        }

        // 2. Apply rollout if present
        if (flag.rollout != null && ctx.userId != null) {
            if (isInRollout(flag.rollout, ctx.userId)) {
                return EvaluationResult.of(flag.defaultValue, Source.ROLLOUT);
            }
        }

        // 3. Fall back to default
        return EvaluationResult.of(flag.defaultValue, Source.DEFAULT);
    }

    private boolean isInRollout(Rollout rollout, String userId) {
        int bucket = stableHash(rollout.seed + ":" + userId) % 100;
        return bucket < rollout.percentage;
    }
}
```

## Separating the Read Path from the Write Path
Reads are on the hot path. Writes are not.

The **read path** must be **local** and **memory-first**. What I mean by that is **flag evaluation never performs I/O**.

When `isEnabled()` is called:
- no network calls are made
- no database queries are executed
- no locks are acquired

All flag definitions already live in memory inside the process.
_I mean, each consuming service runs a local flag runtime inside its own process.
That runtime holds an in-memory snapshot of flag definitions and evaluates flags synchronously._

From the evaluator's perspective, reading a flag is no different from reading a local map or array.

This is not an optimisation — it is the baseline requirement.

For the **write path**, we care about:
- validation (is this change well-formed?)
- ordering (what happens if two changes race?)
- propagation (how long until readers see it?)

Mixing these two concerns is how feature-flag systems becomes fragile.

### The Only Acceptable Contract Between the Two
Once we separate the paths, the contract becomes clear:
- **Write path** produces immutable flag definitions over time.
- **Read path** consumes a snapshot of definitions and evaluates locally.

The snapshot does not have to be perfectly fresh — but it must be:
- internally consistent
- versioned
- safe to evaluate without coordination

### Why "Just Add a Cache" Is the Wrong Mental Model
We generally, tend to start with a remote flag service and "add caching later."

This usually fails because caching is treated as an optimisation, not as the primary mode of operation.

In a real feature flag system:
- cache misses must be impossible, not just rare
- cache invalidation must be deterministic
- stale reads must be known, bounded condition

If the cache is optional, someone will eventually ship a code path that bypasses it — and we'll only discover that during an incident.

## Non-Negotiable Constraints
Before choosing storage, propagation mechanisms, or APIs, we need to lock down the constraints the system must satisfy. These are not optimisations. They are the shape of the system.

### Read Latency Budget
Assume a typical backend request:
- p99 end-to-end budget: `200 ms`
- business logic + downstream calls: `~170 ms`

That leaves `~30 ms` for everything else — including serialisation, logging, metrics, and feature flags.

A reasonable, conservative allocation for flag evaluation is:
> p99 <= 1 ms per flag evaluation

Not per request.
Per flag.

This is why:
- remote reads are off the table
- blocking I/O is unacceptable
- evaluation must be CPU-bound and predictable

"Fast on average" is meaningless here.

### Throughput Expectations
Assume:
- 50 services
- each doing 2k RPS
- evaluating ~5 flags per request

That's:
```markdown
    50 * 2000 * 5 = 500,000 evaluations / second
```
And that's not a peak scenario — just a normal weekday.

### Freshness vs Correctness
Flag changes are control-plane operations. They don't need to be globally instantaneous, but they must converge quickly and predictably.

This means, we are choosing:
- **Eventual consistency** for reads
- **Strong validation** on writes

The system should never promise "immediately everywhere" semantics it cannot enforce.

In this system, safety is enforced structurally, not by operational discipline
- **safe** means a bad write cannot crash readers
- **safe** means evaluation continues during control-plane failures
- **safe** means behavior changes only when definitions change

## Propagating Flag Definitions to Readers
Once reads are local and memory-first, the only remaining infrastructure problem is propagation.

Readers need reasonably fresh flag definitions, but evaluation must never block on updates. That rules out incremental mutation and forces a snapshot-based model.

Readers evaluate against a single immutable snapshot of all flag definitions:
```java
final class FlagSnapshot {
    final long version;
    final Map<String, FlagDefinition> flags;
    
    FlagSnapshot(long version, Map<String, FlagDefinition> flags) {
        this.version = version;
        this.flags = Map.copyOf(flags);
    }
}
```
Evaluation always happens against exactly one snapshot. There is no per-flag state and no partial visibility of updates.

Updating flags means building a new snapshot off the hot path and swapping it in atomically:
```java
class SnapshotStore {
    private volatile FlagSnapshot current;
    
    FlagSnapshot get() {
        return current;
    }
    
    void swap(FlagSnapshot next) {
        this.current = next;
    }
}
```
**How Updates Work:**
- A writer publishes a new flag version (e.g., version 42)
- A background thread (_polling_) fetches version 42 and builds a new `FlagSnapshot`
- Once fully constructed, invoking `swap()` replaces the old snapshot atomically.

This single design choice eliminates most concurrency problems. Evaluation threads never block, and a failed update cannot corrupt in-memory state.

Requests in flight continue using the old snapshot. New requests see the new one. No locks, no partial visibility, no coordination.

If propagation fails, evaluation continues using the last known good snapshot. Freshness degrades — availability does not.

## Consuming Feature Flags
From a consumer's perspective, feature flags should feel boring.

One call. No retries. No async callbacks. No "initialisation phase" that can fail the service.

### The Only API That Matters
At runtime, consumers need exactly one operation:
```java
boolean isEnabled(String flagKey, EvaluationContext ctx);
```
Or for typed flags:
```java
<T> T getValue(String flagKey, EvaluationContext ctx, T defaultValue);
```
This call must be:
- synchronous
- non-blocking
- safe to invoke on every request

Anything more complex becomes friction, and friction becomes misuse.

If a snapshot is available, evaluation proceeds. <br>
If updates are delayed, evaluation proceeds. <br>
If the flag system is “down,” evaluation still proceeds.

The worst-case behavior is stale evaluation — and that is a deliberate choice.

#### Typed Flags Example
Typed flags remove string parsing and makes misuse harder.

```markdown
FlagDefinition maxItems = {
        key: "max_items",
        type: "INTEGER",
        defaultValue: 10,
        rules: [
            if plan == "enterprise" -> 100
            if plan == "pro"        -> 50
        ]
}
```

### Context Is Explicit, Not Inferred
Consumers provide all context required for evaluation:
```java
EvaluationContext ctx = EvaluationContext.builder()
        .userId(userId)
        .accountId(accountId)
        .environment("prod")
        .attribute("plan", "enterprise")
        .build();
```

Also, consumers must supply a default when reading typed flags.
```java
int maxItems = flags.getValue(
        "max_items",
        ctx,
        10 // safe default
);
```

If a flag is missing, malformed, or removed, the service still behaves predictably.
The platform does not invent behavior on behalf of consumers.

For boolean flags, the default is usually “off.” For typed flags, the default encodes safe behavior.

---

#### Closing Thoughts
Get the evaluation mode right, keep reads local and boring, and push complexity out of the request path.

If flipping a flag is ever scarier than deploying code, the system is already telling us something is wrong.