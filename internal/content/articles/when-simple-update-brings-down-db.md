---
title: When a "Simple UPDATE" Can Collapse a Database 
date: 2025-11-17
image: 
tags: mysql, scaling, operations
summary: A large UPDATE on a big MySQL table can generate massive binlog volume and exhaust disk space long before the query itself becomes slow. 
---
Running an **UPDATE** on a table with millions of rows seems simple, and the query itself may not even be expensive in terms of CPU or memory.

But when **binary logging enabled**, MySQL has to record every row change in the binlog so replication and recovery remain consistent.
This means a big update doesn't just modify data — it also produces a large amount of binlog data in the background.
That's where the real pressure comes from.

## How MySQL Actually Executes a Bulk UPDATE
When you run a large UPDATE, MySQL doesn't treat it as one big operation. It walks through the table row-by-row and processes each change in a predictable sequence.
At a high level, it goes through the following steps:

1. **Find the rows that match the condition.** <br>
   MySQL scans the index or the table, depending on the query.
2. **Lock each row as it is processed.** <br>
   This prevents other sessions from modifying the same rows while the update is happening.
3. **Modify the row in InnoDB's buffer pool.** <br>
   MySQL updates the in-memory page that contains the row.
4. **Record the change in the redo/undo logs.** <br>
   These logs ensure the update can be committed safely or rolled back if needed.
5. **Write the change to the binlog.** <br>
   This is the key part for replication and recovery. <br>
   In **row-based logging**, every row produces a separate event.
6. **Move on to the next row.** <br>
   The above steps repeat until all matching rows are processed.

Event though the SQL statement is single, the internal work scales with the number of rows touched.
For a table with millions of rows, MySQL performs these steps millions of times.

## Why the Binlog Grows So Quickly During a Large UPDATE
The binlog exists to capture every change that MySQL applies. For small updates, this overhead is barely noticeable.
But during a large update, the amount of log data grows in direct proportion to the number of rows affected.

The key point is this: **MySQL logs the change for every row, not just the statement**.

Most production systems use **ROW** or **MIXED** binlog formats. 
In ROW mode, MySQL writes a separate binlog event for each updated row.
That event contains the row's old values, new values, and metadata needed to replay the change accurately on replicas.

For illustration,
```bash 
  {old_row_snapshot} + {new_row_snapshot}
```

So if the update touches five million rows, MySQL generates roughly five million binlog events. 
These events accumulate continuously as the update runs.

There are a few characteristics of this process that amplify the effect:
- The binlog is written sequentially and must remain consistent.
- MySQL cannot rotate or purge binlog files until the entire transaction completes.
  - _Rotation normally happens when a file reaches `max_binlog_size`, but MySQL waits for a transaction to finish before starting a new file._
  - _Binlog purge depends on retention settings (`binlog_expire_logs_seconds` or `expire_logs_days`), but these rules apply only after the transaction commits._
- Large transactions keep binlog files open for a long time, allowing them to grow unchecked.
  - _Because the update is one continuous transaction, MySQL must log everything into the same file without rotating or purging._
- Row-based logging includes more data than **statement-based logging**.

As a result, the binlog size grows steadily throughout the update. If the server has limited disk space for logs, this growth can outpace available storage.
Once MySQL can't write to the binlog safely, it stops processing writes to protect replication and crash recovery guarantees.

The update may appear simple from the outside, but the combined cost of logging each change is what drives the system into trouble.

### Understanding Binlog Rotation and Purge
The binlog rotation and binlog purge often mixed together, but they serve different purposes and have different effects on disk usage.

#### Binlog Rotation — Creating a New File
Rotation simply means MySQL stops writing to the current binlog file and starts a new one. <br>
This usually happens when the file reaches the configured `max_binlog_size`.
Important details:
- Rotation **does not reduce disk usage**.
- No files are removed; MySQL only switches to writing into a new one.
- Rotation **cannot happen during a long-running transaction**, because MySQL waits for the transaction to finish before it cuts over to the next file.
  Rotation is helpful only because it creates well-defined boundaries that purging can act on later.

#### Binlog Purge — Removing Old Files
Purge is the process that actually frees disk space.
MySQL deletes old binlog files when they satisfy the retention policy:
- `binlog_expire_logs_seconds` (preferred, modern)
- or `expire_logs_days` (legacy)

Purge is only allowed when:
- The transaction that generated those logs has committed. 
- Replicas (if any) have already consumed those logs. 
- The files exceed the configured retention window.

If any of these conditions are not met, MySQL keeps the binlogs.

#### Why These Two Behaviors Matter During a Large UPDATE
A large UPDATE is one long-running transaction. This causes two side effects:
- **Rotation cannot occur** → the current binlog file grows continuously.
- **Purge cannot occur** → nothing is considered “old,” so no files are removed.

This combination forces all binlog data into the same growing file with no relief mechanism. Disk usage rises until the update finishes—or until storage is full.


## Practical Ways to Handle Large Updates Safely
Once we understand that the pressure comes from the binlog rather than the query itself, the solution becomes more about **controlling the size of each transaction and managing how MySQL logs changes**.

The goal is not to reduce the total amount of work, but to avoid doing it all in one unbounded chunk.

### 1. Break the UPDATE into Smaller Batches
This is the most universally safe method because it keeps each transaction small. Instead of updating millions of rows at once, we **update them in fixed-size slices**,
for example,
```sql
UPDATE table
SET status = 1
WHERE id > X and id <= X + 5000;
```
_Then we  increment X in a loop until the entire table is covered._

This pattern works well because MySQL handles small, short-lived transactions very efficiently.

What batching actually changes:
- Binlog usage stays manageable.
  - Each batch generates binlog data, but MySQL can rotate and purge binlog files after each batch commits.
  - So even though the total amount of logged data across all batches is large, it doesn’t accumulate all at once.
- Undo/redo logs stay small.
  - Large single transactions force MySQL to retain undo/redo data for a long time. 
  - With batching, logs are cleared at every commit, preventing internal buildup.
- Disk usage rises and falls naturally.
  - A batch runs → binlog grows. 
  - Batch commits → MySQL rotates/purges binlogs → disk usage drops.
  - This cycle keeps storage stable instead of trending toward exhaustion.

The total amount of work is the same, but the pressure on the system is far lower. <br>
Without batching, MySQL never reaches a commit point, so neither rotation nor purge can help.

### 2. Use Statement-Based Logging Temporarily (When Possible)
If the update is deterministic (meaning it produces the same result everywhere), we can switch the session to **STATEMENT** mode:
```sql
SET SESSION binlog_format = STATEMENT;
UPDATE ...;
SET SESSION binlog_format = ROW;
```
In this mode, MySQL logs the **SQL statement itself**, not every row change. This dramatically reduces binlog volume.

**Important notes:**
- The change applies only to the current session. 
  - Other connections continue using the global binlog_format. Nothing about their transactions changes.
- MySQL will still fall back to row-based logging if the statement involves **non-deterministic** behaviour (NOW(), UUID(), lack of stable ordering, etc.)
  - If the statement isn’t deterministic, MySQL protects consistency by logging row events anyway.
- Safe only if you understand how the update behaves on replicas.
This option is useful but requires careful consideration of replication correctness.

### 3. Disable Binlogging for the Session (Standalone Servers Only)
If the server has no replicas and we don't need point-in-time recovery for this window, we can disable logging entirely:
```sql
SET SESSION sql_log_bin = 0;
UPDATE ...;
SET SESSION sql_log_bin = 1;
```
This removes binlog overhead completely.

It should be only used when:
- The server is not part of a replication topology.
- We accept that the updates will not be recoverable through binlogs.

It can be very effective, but it's not suitable for most distributed setups.

## Operational guardrails for bulk updates
Bulk updates are common in production systems, but they need to be handled with more structure than small day-to-day modifications.
The goal is to avoid long-running transactions, control binlog growth, and prevent replication lag or storage pressure.

### 1. Always Measure How Many Rows Will Be Updated
Before running a large update, check the expected row count.
If the count is in the millions, batching should be assumed, not optional.

### 2. Use an Ordered, Bounced Chunking Strategy
Chunking by `LIMIT` alone is unreliable unless the table has a stable ordering. <br>
The safe pattern is:
- Chunk by a monotonically increasing key (usually the primary key).
- Track progress using the last processed key.
```sql
UPDATE table SET flag = 1
WHERE id > last_id
ORDER BY id
LIMIT 5000;
```
This ensures predictable slices and avoids reprocessing or skipping rows.

### 3. Keep Each Transaction Small
Short transactions help: 
- Binlog rotation and purge can occur more frequently.
- Undo/redo logs are cleared at each commit, preventing internal buildup.
- Disk usage rises and falls naturally.
- Reduce lock contention
- Limits replication lag

A practical target is typically **1k-10k rows per batch**, depending on row size and storage speed.

### 4. Ensure Enough Free Disk Space Before Starting
A safe rule-of-thumb is to ensure free space of at least **3x the target binlog growth of one full batch**. <br>
For large tables or wide rows, this margin should be higher.

## Conclusion
_Large updates in MySQL look simple, but they stress the system through logging, not just computation. 
The binlog grows with each row change, and long transactions prevent rotation and purge, which quickly leads to disk pressure. 
Breaking the work into small, predictable batches — and understanding how MySQL handles binlogs — turns a risky operation into a controlled one._