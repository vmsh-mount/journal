---
title: Building Stable Systems, Yet So Fast
date: 2024-04-25
image:
tags: engineering-philosophy
summary: 
---

These are notes I took from a talk by the CEO of one of the companies I worked at.
The session happened right after a major production issue caused by one of the internal teams. 
We all expected him to be furious. Instead, he walked in calmly, pulled up an old presentation, and delivered it with complete composure. 
The whole room sat in silence and listened.

That calmness — especially in moments when everything is on fire — is one of the qualities I’ve admired in him the most.

I may have missed a few points, and I might have added some of my own, but these notes capture the essence.

---
## Development

### Build in Small, Safe Steps
- Develop things in small chunks.
- We should know **how to test** anything we develop — straightforward or hacks.
- Not testing and not taking care of **backward compatibility** is a sin. 
- We should have a **backup mechanism** to fix data or other damage that might occur in case of a deployment catastrophe.

### Use Feature Flags Wisely
- It’s good to have a feature flag to toggle a feature in production.
- A feature flag need not be just binary; it can be a value between 0 and 1. _For example, 0.5 indicates enabling the feature for 50% of traffic._
- **We should always be in control** and be aware of potential checkpoints in the code that might fail if damage occurs.
- Maintaining versioning is important.

### Core & Non-core components
- We should double-check when making changes to core components such as request/response models, business logic, decorators, etc.
- It’s okay to take risks (not 100% sure) on non-core components like UI, analytics dashboards, logging, etc.

### Code Quality
- Following good coding practices is necessary.
- Static analysis tools help.
- When your code is not self-explanatory, write simple one-line documentation.

### Strategy
- It’s okay to revisit design modules or create an alternate version to support more use cases.
- Don't be afraid to make decisions even with 80% clarity.

> “Code quality is important but not fast” <br>
> “Elegance is important but not fast” <br>
> “Fast is not so important!”

--- 
## Deployment Thumb Rules
- Always **ship one gamble at a time** to production; combining multiple critical features increases the risk factor.
- When an issue is found during deployment, **first rollback** and then look into the issue, since we know the previous patch is stable.
- Logging into production and fixing things works, but it’s not sustainable — so maintain a safe distance from production environments.
- There is absolutely no guarantee that code tested in a sacrosanct sandbox will behave the same in production, so **be watchful during deployment**.
- When things don’t go according to plan, **staying mentally calm is important**; otherwise, it can lead to cascading failures.

---
## Communication
- Try to drop an email/Slack message whenever a deployment is done. It creates a **psychological impact** to be more careful when you know others may look at or question the deployment. 
- **Accepting mistakes and informing the team is necessary**. It’s okay if it takes a small hit on your credibility — staying accountable is important.
- When someone is wearing headphones, don’t disturb them. That’s a basic social norm.

---
## Closing Thoughts
These notes aren't rules — just reminders from someone who had seen systems break, heal and grow.

What stayed with me most was not the checklist, but the mindset: **build carefully**, **deploy responsibly**, **stay calm under pressure**, and **always remain accountable**.

Stable systems aren't just result of good engineering practices, but of the attitude we bring to the work. 