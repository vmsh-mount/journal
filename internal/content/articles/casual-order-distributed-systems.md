---
title: Casual Ordering in Distributed Systems
date: 2024-10-26
image: https://images.unsplash.com/photo-1658834047199-c92d1557fbb8?q=80&w=2340&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D
tags: distributed-systems, casual-ordering, coordination, time
summary: Distributed systems don’t rely on perfectly synced clocks — they rely on understanding how events influence one another. Instead of asking when something happened, causal ordering helps us understand why it happened in that sequence. 
---

## Ordering Events

In distributed computing, ordering events correctly is like ensuring the flow of electricity through a complex circuit — if one pulse comes too early or too late, it can disrupt the entire system. 
Whether it’s processing user requests, running financial transactions, or coordinating micro-services, maintaining event order is crucial to avoid data corruption and ensure consistency.

But in systems where multiple nodes are scattered across different regions, each with its own local clocks and processing times, how do we establish a reliable order? 
How can we trust that one event didn’t sneak ahead of another when thousands of messages are flying through the network at any given moment?

## The Illusion of Time
Time is our tool for organising the world around us. We use it to track when a message is sent, when data is updated, or when a process starts and finishes. 
Whether it’s waking up on time for work or tracking a package, we trust time to help us understand the sequence of events.

In an ideal world, we (*except Christopher Nolan*) believe **time is linear**: Event A happens, then Event B, and finally Event C. We rely on time as a shared global truth. 
Clocks across the globe, from New York to Tokyo, are synchronised, leaving no ambiguity about when something occurs.

## The Reality of Time
Unlike our everyday world, distributed systems don’t follow the straightforward understanding of time. Here’s why:

### Events Happen Concurrently
- There’s no global clock.  
- Each machine has its own clock, ticking in isolation. Events that happen on different machines can occur simultaneously, but without a single source of truth, it’s impossible to know which happened first.

### Network Delays
- A message sent from one node to another can arrive late, out-of-order, or even be lost in transit. 
- The receiving node can’t always be sure when the message was sent, and the sending node can’t know when it was received.

### Partial vs. Total Order
- On a single node, events are relatively easy to order (total order) — the machine controls its own processes.  
- Across multiple nodes, events can only be partially ordered. Each node knows the order of its own events, but it can’t always determine how they relate to events on another node.

## Reframing Time
So, how do we determine the correct sequence of events when we can’t rely on time? <br>
This raises a bigger question: `Why do we care so much about time in the first place?`

Perhaps the question we should ask is: `Do we need time at all?`

Perhaps what truly matters isn’t the precise timestamp of each event but the **order** in which they occur.

When we say “Event A happened before Event B”, we aren’t really concerned about the timestamp. 
What we care about is the relationship — _Did Event A cause Event B? Did Event C happen after Event B?_

These **cause-and-effect** relationships matter more than exact timing.  
Instead of aligning clocks, distributed systems focus on ordering events based on how they influence one another.

> “It’s not time we care about — it’s the order of events that time represents.”

This reframing leads us to **Causal Ordering**.

## Causal Ordering
In distributed systems, causal ordering helps us track how events are related — who triggered whom — without relying on exact timestamps.

### The Cause and Effect Principle
At its core, casual ordering is about understanding the cause-and-effect relationships between events. Imagine you’re sending a series of messages between nodes in a distributed system. 
Event A casually precedes Event B if:

#### 1. Same Machine Rule
If two events occur on the same machine and one happens before the other, the first event causally precedes the second.  
_Example: A user clicks a button (Event A) before receiving a message (Event B)._

#### 2. Message Passing Rule
If a machine sends a message and another machine receives it, the sending event causally precedes the receiving event.  
_Example: Node X sends a message to Node Y → send event precedes receive event._

#### 3. Transitive Rule
If Event A causes Event B, and Event B causes Event C, then Event A causally precedes Event C.

We can infer that events are ordered not by time, but by which events cause others. This way, even without synchronised clocks, we can infer a logical order of events.

> “Causal ordering is about determining the order of events based on their relationships, not their timestamps.”

## Partial Order
One key feature of causal ordering is that it allows **partial ordering**, not total ordering.

In a total order, every event is comparable — A before B or B before A.

But distributed systems don’t have perfect shared knowledge.

With partial ordering:
- Events are ordered only when they are causally related.
- If two events occur independently on separate nodes with no communication, they are **concurrent**.
- Neither happens “before” or “after” the other.

This is a key insight: **We don’t need a strict global order — only causal order when needed.**

## Communication Defines Causality
Machines influence each other only through communication.  
If there’s no communication, there’s no causal relationship.

> “If not A → B, then A cannot have caused B.”

Without a causal link, events are independent or concurrent.

In essence, causal ordering shifts our thinking **from tracking time to tracking influence**. It’s not about when things happened, but about understanding why they happened in that particular order.

## The Building Block
Causal ordering is foundational in distributed systems. It influences:

- **Vector Clocks** — track causal relationships
- **Lamport Timestamps** — order events logically
- **Eventual Consistency** — replicas converge despite delays
- **CRDTs** — allow safe, conflict-free updates
- **Consensus algorithms (Paxos, Raft)** — maintain agreement across nodes
- **CAP theorem** trade-offs — consistency vs. availability vs. partition tolerance

Understanding causal ordering helps us understand distributed coordination at its core.

## Conclusion
_In distributed systems, understanding event order through causal relationships rather than relying on synchronised time is crucial.  
It’s not the time that matters — it’s the **order** that time represents._

---
## References
- [Ordering Distributed Events](https://medium.com/baseds/ordering-distributed-events-29c1dd9d1eff)
- [Causal Ordering](https://www.scattered-thoughts.net/writing/causal-ordering/?source=post_page-----84a9d1b4a2e8---------------------------------------)
