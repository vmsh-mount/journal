---
title: When Kafka Retries Go Wrong
date: 2025-11-21
image: 
tags: distributed-systems, kafka, engineering, debugging, architecture
summary: A routine Kafka alert turned into a large-scale incident when a slow database insert caused a consumer to miss its poll window, get evicted from the group, and repeatedly reprocess the same message.
---

On Nov 21, our Kafka-based communication pipeline started behaving a little strangely. Nothing dramatic at first — just a periodic alert that a consumer had been removed from its group.
These alerts pop up occasionally in any distributed system, so nobody jumped immediately. But by evening, the pattern changed: the same Kafka messaged appeared to be processed again... and again... and again.
That's when we knew something deeper was off, because Kafka doesn't redeliver unless it believes a consumer never acknowledged progress.

## What Actually Happened
The consumer did process the message correctly. The business logic executed fully. The communication was sent to the downstream service. The provider acknowledged success. 
And the consumer even started the post-processing logic that writes customer-level communication tracker tables.

But it never made it to the final step: **committing the offset**.

Right when the consumer tried to commit, Kafka responded with the familiar but dreaded message:

> Offset commit cannot be completed as the consumer is not part of an active group. It is likely that the consumer was kicked out of the group.

A healthy consumer never hears this unless something violated Kafka's expectations of time. 
From Kafka's perspective, our service simply took too long and stopped responding. And in Kafka's world, "stop responding" is indistinguishable from "dead".

Once removed from the group, the consumer lost its partitions, its session was invalidated, and the offset commit request was no longer meaningful.

And because the offset never committed, Kafka did exactly what Kafka is designed to do — it retied the same message. Every hour. For nearly six hours.

## Why the Consumer Was Kicked Out?
Kafka doesn't remove consumers for fun. There are only a few reasons a consumer is expelled from a group: <br>
- The poll loop took longer than `max.poll.interval.ms`: _how long Kafka allows a consumer to go between polls before considering it “stalled”_
- Heartbeats paused long enough go exceed `session.timeout.ms`: _how long Kafka waits without receiving heartbeats before removing a consumer_
- The consumer thread blocked on something (_DB call, IO call, downstream service call_)

In our case, everything pointed to a single culprit:
**Database insert operations that grew slower over time, eventually exceeding Kafka's poll interval**

When those inserts took too long, the consumer couldn't send heartbeats. Without heartbeats, the coordinator removed it. And once removed, the commit attempt became meaningless.

Kafka's logic is brutally simple:
>  "If you don’t poll in X minutes, I assume you’re dead."
>> To Kafka, a slow consumer is indistinguishable from a dead consumer. 
>> The logic is intentionally simple:
>> - If heartbeats don't arrive within `session.timeout.ms`, the consumer must be down. This detects death of consumer.
>> - If no poll happens within `max.poll.interval.ms`, the consumer must be stuck. This detects slowness/stall. <br>
>>
>> Kafka doesn't try to differentiate between "the consumer is busy" and "the consumer has crashed". <br>
>> Both cases look the same — a long silence.

## Why It Turned Into a Large-Scale Incident
The actual business failure was not the Kafka rebalance. Rebalances happen. The real issue was **lack of file-level idempotency**.

Every time Kafka redelivered the same message:
- _The business logic executed fully_
- _The communication was sent to the downstream service_
- _The provider acknowledged success_
- _The consumer even started the post-processing logic — inserted new log entries_
- _But it never made it to the final step: **committing the offset**_

And because the communication files for these nudges were large and high-volume, each retry amplified the impact.

This repeated cycle continued until the DB became overloaded, connection pools were exhausted, and the system began throwing consistent exceptions. 
With the actual exception thrown, Kafka finally moved the message to the dead-letter queue (DLQ), which brought the loop to an end.

By then, over 15 million duplicate EMAIL communications had been sent.

## Rebalance: What Actually Happens Under the Hood
Rebalances often get misunderstood, so it's worth describing this one carefully.

When a consumer misses heartbeats, Kafka removes it from the group. But that doesn't mean the process dies.
The JVM continues running. Our code continues executing. But the consumer instance is no longer authorised to commit offsets or own partitions.

That leads to two confusing situations: 

**1. The consumer continues processing a message, unaware it has been expelled** <br>
After expulsion, the consumer thread doesn’t crash. It doesn’t restart. It doesn’t even know that it’s no longer in the group.<br>
Processing continues normally. But any offset commit attempt fails with the "not part of group" error.

**2. After the rebalance, the same consumer process may rejoin the group**<br>
Kafka is fine with this. It will let the same instance re-register.<br>
But because the last commit failed, the consumer starts from the last committed offset — which, in this case, was the same message it just processed.<br>
So the cycle repeats.

This explains the repeated hourly retries we observed.

### How Would This Behave With One Consumer? With Multiple Consumers?
It's also worth explaining how this kind of failure behaves depending on the number of consumers.

#### Single-consumer Group (Only One Consumer Instance)
A single consumer still gets kicked out if it violates `max.poll.interval`.
Kafka will simply retry delivery to the same instance when it re-joins.
Duplicate processing still happens.

#### Multi-consumers in the Group
With multiple consumers, the behavior is slightly different:
- When one consumer is expelled, another consumer immediately takes over its partitions.
- If the slow-down is isolated to one instance, duplicates still happen, but faster and more deterministically.
- If all consumers are slow (e.g., shared DB bottleneck), they start dropping out one after another, causing constant churn and unpredictable ownership.

**In short:** <br>
_"A slow poll loop causes problems regardless of number of consumers."_

## Preventing Recurrence: The Design Fixes
### 1. File-level Idempotency - The Real Cure
Kafka’s retry behavior is entirely correct.
The absence of a guard on our side is what allowed retries to turn into duplicates.

The simplest and most robust fix is to treat the input file (or more precisely the file content or URL) as the natural idempotency boundary. A communication file represents one attempt to notify a specific population. No matter how many times the pipeline sees the message, the business outcome should happen only once.

To enforce that, we introduced: `input_file_hash` in the communication_log  table.
```sql
ALTER TABLE `communication_log`
ADD COLUMN `input_file_hash` VARCHAR(255) NOT NULL;
```

Before processing any Kafka message, the consumer now computes a hash of the incoming file (or the file URL, depending on the workflow). The hash becomes a unique, stable identifier for that communication event.

The workflow becomes:
1. Compute hash
2. Look up existing communication_log entry
3. If found: _Skip processing entirely. Commit the offset._ <br>
   If not found: _Continue processing and insert a new log entry._

#### Unique DB Constraint for Atomic Guarantees
To make this idempotency mechanism race-safe, we added a unique constraint on the hash column. Even if multiple consumers attempt to process the same file concurrently, exactly one insert will succeed, and the rest will fail immediately—ensuring clean, deterministic behavior.
```sql
ALTER TABLE `communication_log`
ADD UNIQUE INDEX `idx_input_file_hash_unique` (`input_file_hash`);
```

With this in place, even if the consumer is expelled a dozen times, the worst thing that will happen is that Kafka retries the message — and we’ll skip it gracefully each time.

### 2. Reducing Processing Latency Through Database Partitioning
Idempotency protects us from duplicates. <br>
But preventing the slowdowns that triggered the rebalance in the first place is equally important.

Over the last few quarters, the communication_tracker tables have grown large. The customer-level granularity means that every communication blast results in millions of new rows. Older entries accumulate, and even with proper indexing, these tables naturally slow down over time.

The long insert operations during the incident were not a one-off anomaly — they were a sign of the table’s age and size.

The fix here is about data hygiene and active partitioning.

#### Weekly Partitioning for Customer-Level Tracker Tables
We implemented weekly partitions, which allow us to:
- keep recent partitions small and fast
- move older data into separate partitions
- avoid the performance drag caused by giant monolithic tables
- keep inserts fast by writing to the newest, smallest partition

Partitioning isn’t a cosmetic change — it directly reduces the time spent in the poll loop.

### 3. Strengthening the Consumer's Tuning Envelope
Kafka comes with sensible defaults, but defaults assume that your consumer finishes processing quickly. In real production pipelines — where consumers talk to databases, file systems, external services — these defaults often need to be tuned.

After the incident, we revisited the following:
- **max.poll.interval.ms**: Extended moderately to accommodate bursts where processing may legitimately take longer.
- **session.timeout.ms/heartbeat.interval.ms**: Tuned to ensure heartbeats continue flowing even under moderate load.

We don't want to hide underlying performance issues behind larger timeouts.

### 4. Observability Around Poll Loop Latency
One final part of the fix is observability.<br>
Slow consumers rarely become slow all at once — they drift.

We introduced simple, low-friction metrics and alerts around:
- time spent between `poll()` calls
- time spend in business logic
- commit latency
- DB insert latency
- number of rebalances per hour

These are lightweight but incredibly useful early indicators.
Rebalances should be rare. If they start clustering, something deeper is happening.

## Lessons That Stick
Every incident leaves behind a few lessons that stay with the team long after the issue itself is resolved. 
This one, in particular, was a reminder of how distributed systems behave when things slow down—not when things fail outright, but when they simply take a bit longer than the system expects.

The first takeaway is that **Kafka is brutally honest about time** — _Kafka doesn’t examine your thread dumps, check your CPU metrics, or try to reason whether your DB is under load. It just sees silence, and silence is enough._

The second takeaway is that **idempotency is not an optional luxury** — _You don’t add idempotency because your system is fragile. You add it because Kafka, by design, will retry messages—and it will do so eagerly whenever it thinks a consumer failed._

A third lesson is about the **environments we test in**. — _Lower environments are great for validating functionality, but they rarely behave like production in terms of data volume or write pressure. This is why performance characteristics—especially for stateful operations—must always be monitored over time, not just in pre-production_

Finally, this incident reinforced a broader architectural truth: <br>
**distributed systems recover by retrying.** <br>
Retries are not edge cases—they’re part of the normal execution path. And as long as retries exist, deterministic idempotency must exist too.