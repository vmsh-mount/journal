---
title: Synchronising Time in Distributed Systems
date: 2024-10-15
image: https://images.unsplash.com/photo-1506686098657-5eeb165b4353?q=80&w=2940&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D
tags: distributed-systems, time
summary: Computers don’t actually “know” the real time — they just count from whatever they’re told, which leads to drift and mismatches across machines. In distributed systems, these tiny timing errors snowball into big coordination problems.
---

## Time in Chaos
We rarely think about how the systems we rely on actually work. Most of the time, that’s perfectly fine — we don’t need to understand the complexities behind every piece of technology we use. One such complexity is **Time** — something we often take for granted.

Most modern devices have a built-in sense of time, from your smartphone to your smart fridge. In the realm of the internet, where billions of devices and servers interact daily, maintaining an accurate sense of time is a big challenge. What happens when a server in Tokyo thinks a transaction happened before the server in New York processed it? Or when two machines disagree on the exact sequence of events in a stock trade?

## How Computers Tick
Imagine time as a universal rhythm that everyone follows. In the physical world, we use clocks to track this rhythm — but how do computers know what time it is? 
<br><br>The answer is simpler than you might think — **they also use clocks!**.

### Real-Time Clock (RTC)
Each computer has a small but crucial hardware component called a **Real-Time Clock (RTC)** — a tiny integrated circuit embedded on the motherboard. It’s powered by a small **battery** so it keeps ticking even when the computer is off.

The RTC keeps track of time using a **crystal oscillator**, which vibrates at a constant rate. Imagine a quartz crystal vibrating thousands of times per second — this vibration is the heartbeat of the clock. Each oscillation is counted, and these ticks are stored in a binary counter circuit.

When the computer is off, the RTC continues running using its battery and oscillator. When the system boots up, the **system clock** (a software-based clock) is initialised from the RTC.

### System Clock
The system clock is a software-based clock that operates only when the computer is on, relying on the main power supply. 
It retrieves the current time from the RTC at startup and runs independently during operation. 
Since the system clock can drift over time, it periodically synchronises with the RTC to correct any discrepancies. 
The machine’s operations follow the time provided by the system clock, regardless of external sources.


All machine operations use the **system clock**, not the RTC. Essentially:
- **RTC** keeps accurate time when the system is off
- **System clock** keeps time when the system is running

Both use oscillators but function independently.

## Time-Setting Knob
An interesting quirk about system time is that it always begins counting from whatever value the system clock is set to. If you set a new computer’s clock to be **12 minutes slow** or **80 days fast**, the system begins counting from that value — whether it’s correct or not.

> "Essentially, Computers don’t know the “real” time; they simply count from the starting point."

To avoid confusion, operating systems follow a standardised starting point. For example, Unix-based systems use **Unix time**, which starts at: `January 1, 1970 — 00:00:00 UT`

When you set up a new system, the RTC provides the initial timestamp (which may be inaccurate). To correct this, systems use the **Network Time Protocol (NTP)** to synchronise time with reliable external servers.

## Consistency in Time
Each machine has its own notion of time. But if two machines have different ideas of time, which one is correct?

You might assume digital clocks are perfectly accurate, but even highly engineered clocks experience small inaccuracies. Even GPS satellites, which use **atomic clocks**, encounter tiny deviations due to gravity, motion, and relativity.

Regular computers rely on quartz oscillators — far less accurate. Over days or weeks, they drift, often by a few seconds.


> "Maybe it’s only a second or two, but these small errors gradually add up."

## Clock Drift and Clock Skew

So, why does this happen? A lot of factors can affect how accurately a clock ticks. Things like temperature changes, the clock’s location, its power source or even how well it was built can throw off its timing.

This gradual shift in how a clock measure seconds, causing it to fall behind or race ahead, is known as **Clock Drift**.

**Clock Skew** takes this one step further. While drift refers to the gradual change in a single clock’s accuracy over time, skew describes the difference between two clocks at any given moment. So, if your phone thinks it’s 3:02:15 PM and your laptop thinks it’s 3:02:20 PM, that’s clock skew in action.

_In short:_
- **Drift** → one clock losing/gaining time
- **Skew** → the gap between two different clocks

These small inconsistencies usually don’t affect daily use, but they cause major issues in **distributed systems**.

## The Impact
A distributed system is:
> “Multiple machines communicating with one another, often scattered across different locations, each performing their own tasks.”

Every machine keeps its own time, and they all keep track of time independently. So far, so good, right? Well, here’s where things gets tricky — these clocks are rarely in perfect sync.

Example issue:
- Machine A: event happened at **12:01:05**
- Machine B: event happened at **12:01:02**

If they need to process events in order, this mismatch creates ambiguity:
- Did Event A happen before Event B?
- Or vice versa?

The system might end up processing things out of order, causing issues like inconsistent results, duplicate tasks, or worse.

With hundreds or thousands of nodes, these timing mismatches can become massive problems.

To solve this, distributed systems employ techniques to ensure that clocks are as synchronised as possible:
- Synchronisation protocols (NTP, PTP)
- Logical clocks (Lamport clocks, vector clocks)
- Event ordering techniques

Because **perfect clock synchronisation is impossible**.

## Conclusion
_Time plays a critical role in distributed systems. Clock drift and skew may seem small, but in a network of many machines, they can cause major inconsistencies._

_Ultimately, it’s not just about keeping clocks in sync — it’s about ensuring smooth coordination across the entire system._
