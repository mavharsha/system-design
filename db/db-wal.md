# Write-Ahead Logging (WAL)

## Overview

Write-Ahead Logging (WAL) is a fundamental technique used in database systems to ensure data integrity and enable crash recovery. The core principle is simple: **all modifications must be written to a log before they are applied to the database**.

## How WAL Works

### Basic Mechanism

1. **Transaction begins**: A transaction starts modifying data
2. **Log records created**: Before any data page is modified, a log record is written to the WAL
3. **Log flushed to disk**: The log records are flushed to persistent storage (disk)
4. **Data modified**: Only after the log is safely on disk, the actual data pages are modified
5. **Transaction commits**: Once all log records are written, the transaction can commit

### Key Components

```
WAL Buffer (Memory)
    ↓
WAL File (Disk)
    ↓
Checkpoint Process
    ↓
Data Files (Disk)
```

### Log Record Structure

A typical WAL record contains:
- **LSN (Log Sequence Number)**: Unique identifier for the log record
- **Transaction ID**: Which transaction made the change
- **Page ID**: Which data page was affected
- **Before Image**: Original value (for UNDO)
- **After Image**: New value (for REDO)
- **Operation Type**: INSERT, UPDATE, DELETE, etc.

## How WAL Supports ACID

### 1. Atomicity (All or Nothing)

WAL enables atomicity through the **UNDO** mechanism:

- **During normal operation**: All changes are logged before being applied
- **On crash/rollback**: The system can use the "before images" in the log to undo uncommitted transactions
- **Guarantee**: Either all operations of a transaction are applied, or none are

**Example:**
```
Transaction T1:
  - UPDATE account SET balance = 900 WHERE id = 1;  [Log: T1, Page 5, Before: 1000, After: 900]
  - UPDATE account SET balance = 1100 WHERE id = 2; [Log: T1, Page 7, Before: 1000, After: 1100]
  - [CRASH before commit]
  
On Recovery:
  - Read log, find T1 uncommitted
  - UNDO: Restore balance = 1000 for account 1
  - UNDO: Restore balance = 1000 for account 2
  - Result: Transaction completely rolled back (atomicity preserved)
```

### 2. Consistency (Valid State)

WAL maintains consistency by:

- **Ordered logging**: Log records are written in the exact order operations occur
- **Constraint enforcement**: Application-level constraints can be validated before logging
- **Referential integrity**: Multi-step operations are logged together, ensuring relationships remain valid

**Key Rule**: No data page can be written to disk before its corresponding log records are safely persisted.

**Example:**
```
Business Rule: Total balance in system must be constant

Transaction T1: Transfer $100 from A to B
  - Log[1]: T1, UPDATE account A, Before: 1000, After: 900
  - Log[2]: T1, UPDATE account B, Before: 500, After: 600
  - Log[3]: T1, COMMIT
  
If crash occurs:
  - After Log[1]: Recovery will UNDO, restoring A to 1000
  - After Log[2]: Recovery will UNDO both, restoring valid state
  - After Log[3]: Recovery will REDO all, completing the transfer
  
At no point is the constraint violated after recovery.
```

### 3. Isolation (Concurrent Execution)

While WAL primarily supports recovery, it aids isolation through:

- **Versioning**: Log records can be used to maintain multiple versions of data (MVCC)
- **Lock information**: WAL can log lock acquisitions and releases
- **Checkpoint coordination**: Ensures concurrent transactions don't interfere during recovery

**MVCC with WAL:**
```
Time  | Transaction T1        | Transaction T2
------|----------------------|------------------
t1    | BEGIN                | 
t2    | UPDATE row X = 10    | BEGIN
t3    | [Logged LSN 100]     | SELECT row X
t4    |                      | [Reads old version from log/buffer]
t5    | COMMIT               |
t6    |                      | [Still sees old version until T2 commits]
```

### 4. Durability (Persistence)

**This is where WAL shines most brightly.**

WAL guarantees durability through:

#### a) Write-Ahead Rule (WAL Protocol)
- Log records must be flushed to disk **before** the transaction commits
- Once a transaction receives a "commit successful" response, its changes are guaranteed to survive any crash

#### b) Force-Log-at-Commit
```
Transaction commits when:
  1. All log records are in the WAL buffer
  2. fsync() called on WAL file (forces OS to flush to physical disk)
  3. Only then return "commit successful" to application
```

#### c) Recovery Process
```
On System Restart:
  1. Read WAL from last checkpoint
  2. REDO all committed transactions (using "after images")
  3. UNDO all uncommitted transactions (using "before images")
  4. System is now in a consistent state with all committed data restored
```

**Example:**
```
Scenario: Database crashes immediately after transaction commit

Timeline:
  t1: Transaction T1 modifies 100 rows
  t2: All changes logged to WAL (LSN 1000-1099)
  t3: WAL flushed to disk (fsync completed)
  t4: Application receives "COMMIT SUCCESS"
  t5: [POWER FAILURE - data pages not yet written to disk]
  
On Recovery:
  - WAL is intact on disk
  - System reads log entries LSN 1000-1099
  - REDO all 100 changes from log
  - Result: All changes are restored (durability guaranteed)
  
Key Insight: We didn't need the data pages to be written! 
            The WAL was sufficient to restore everything.
```

## WAL Benefits

### 1. Performance
- **Sequential writes**: WAL is append-only, enabling fast sequential I/O
- **Delayed data writes**: Data pages can be written to disk lazily (batched, optimized)
- **Reduced I/O**: Multiple changes to same page only need one eventual write

### 2. Crash Recovery
- **Fast recovery**: Only need to replay log from last checkpoint
- **Guaranteed consistency**: REDO/UNDO ensure correct state
- **No data loss**: Committed transactions always recoverable

### 3. Additional Features Enabled

#### Point-in-Time Recovery (PITR)
```
Archive old WAL segments → Can restore database to any previous moment
```

#### Replication
```
Primary Server → WAL Stream → Standby Server (applies same changes)
```

#### Incremental Backups
```
Backup WAL files instead of entire database
```

## WAL Implementation Details

### Checkpointing

Checkpoints reduce recovery time by:
1. Writing all dirty pages to disk
2. Recording the checkpoint LSN in WAL
3. During recovery, only replay from last checkpoint (not from beginning of time)

```
Timeline:
  LSN 1-1000: Various transactions
  LSN 1001: CHECKPOINT (all dirty pages written)
  LSN 1002-2000: More transactions
  [CRASH]
  
Recovery:
  - Start from LSN 1001 (not LSN 1)
  - Only replay 1000 log records instead of 2000
```

### WAL Segment Rotation

```
WAL Files:
  wal_000001 (16 MB)
  wal_000002 (16 MB)
  wal_000003 (16 MB) ← currently writing
  wal_000004 (16 MB)
  
Once wal_000003 is full → switch to wal_000004
Old segments can be archived or deleted (after checkpoint)
```

### Group Commit

Optimization: Multiple transactions commit together with one fsync

```
Transaction T1 wants to commit → Add to commit queue
Transaction T2 wants to commit → Add to commit queue (wait a few microseconds)
Transaction T3 wants to commit → Add to commit queue
→ One fsync() covers all three transactions
→ Reduces disk I/O overhead
```

## Common WAL Configurations

### PostgreSQL
- `wal_level`: minimal, replica, logical
- `fsync`: on/off (should always be ON in production)
- `wal_buffers`: Size of WAL buffer in memory
- `checkpoint_timeout`: Time between automatic checkpoints

### MySQL (InnoDB)
- `innodb_flush_log_at_trx_commit`:
  - 0: Write and flush every second (fast, less durable)
  - 1: Write and flush at each commit (slow, fully durable) ← ACID compliant
  - 2: Write at commit, flush every second (medium)

## WAL vs Other Approaches

### Shadow Paging
- **Concept**: Write modified pages to new locations, then atomically switch pointers
- **vs WAL**: WAL is more efficient, enables better recovery and replication

### No-Force/Steal Policy
WAL enables databases to use these policies:
- **No-Force**: Don't force dirty pages to disk at commit (WAL provides durability)
- **Steal**: Allow uncommitted pages to be written to disk (WAL provides atomicity via UNDO)

This combination provides optimal performance while maintaining ACID.

## Real-World Example: PostgreSQL WAL

```sql
-- View current WAL position
SELECT pg_current_wal_lsn();
-- Output: 0/1A2B3C4D

-- View WAL settings
SHOW wal_level;
SHOW fsync;

-- Manually trigger checkpoint
CHECKPOINT;

-- View WAL file location
SHOW data_directory;
-- WAL files are in: data_directory/pg_wal/
```

## Summary: WAL's Critical Role in ACID

| ACID Property | How WAL Supports It |
|---------------|---------------------|
| **Atomicity** | UNDO log records restore original state for rollback |
| **Consistency** | Ordered logging ensures state transitions are valid |
| **Isolation** | Enables versioning and helps coordinate concurrent transactions |
| **Durability** | Force-log-at-commit + REDO ensures committed data survives crashes |

**Bottom Line**: WAL is the foundational mechanism that makes ACID properties practical and performant in modern database systems. Without WAL, databases would either be slow (force all writes at commit) or unsafe (risk data loss on crash).

## References

- PostgreSQL WAL Documentation
- MySQL InnoDB Redo Log
- Database Internals by Alex Petrov
- Transaction Processing: Concepts and Techniques by Gray & Reuter

-----
WAL are buffered.
Even with WAL there is potential for data loss.
