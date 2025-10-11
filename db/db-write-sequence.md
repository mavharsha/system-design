# Database Write Operation: Detailed Sequence

## Overview

When a database processes a write operation (INSERT, UPDATE, or DELETE), it must balance three competing concerns:
- **Durability**: Data must survive crashes and power failures
- **Performance**: Minimize latency and maximize throughput
- **Consistency**: Maintain data integrity and index correctness

This document explains the complete sequence of operations from receiving a write request to ensuring durable storage.

---

## The Complete Write Sequence

### Phase 1: Query Planning & Validation

#### 1.1 Parse and Validate
```
Client → SQL Parser → Query Validator → Planner
```

**Steps:**
1. **Parse SQL**: Convert SQL text into an Abstract Syntax Tree (AST)
2. **Semantic Analysis**: Verify table exists, columns are valid, types match
3. **Permission Check**: Ensure user has write privileges
4. **Constraint Validation**: Check PRIMARY KEY, UNIQUE, FOREIGN KEY, CHECK constraints

**Example:**
```sql
UPDATE users SET balance = balance + 100 WHERE user_id = 42;
```

**Validation checks:**
- Does `users` table exist?
- Does `balance` column exist and support arithmetic?
- Does `user_id = 42` exist?
- Will the new balance violate any CHECK constraints?

#### 1.2 Acquire Locks
```
Lock Manager → Row-level or Table-level Locks
```

**Purpose**: Prevent concurrent modifications that could violate consistency

**Lock Types:**
- **Pessimistic Locking**: Acquire locks before reading (2PL - Two-Phase Locking)
  - Shared locks for reads
  - Exclusive locks for writes
- **Optimistic Locking**: Check for conflicts at commit time (MVCC systems)

**Example with 2PL:**
```
1. Acquire SHARED lock on user_id = 42 (to read current balance)
2. Upgrade to EXCLUSIVE lock (to write new balance)
3. Hold locks until transaction commits
```

**Deadlock Prevention**: Lock acquisition follows a consistent order (e.g., sorted by primary key)

#### 1.3 Create Execution Plan
```
Query Optimizer → Execution Plan
```

**For writes, the plan includes:**
- Which index to use for locating rows
- Order of operations (table update before/after index updates)
- Which indexes need updating (primary, secondary, covering indexes)

---

### Phase 2: Write-Ahead Log (WAL)

#### 2.1 WAL Purpose & Guarantee

**WAL Principle**: "Write the log before writing the data pages"

**Why WAL?**
- **Atomicity**: All-or-nothing transaction semantics
- **Durability**: Survive crashes by replaying logs
- **Performance**: Sequential writes are much faster than random writes

#### 2.2 Generate Log Records

**REDO Log Record Structure:**
```
[LSN | Transaction ID | Operation Type | Table ID | Page ID | Offset | Old Value | New Value | Checksum]
```

**Example Log Entries:**
```
LSN=1001 | TXN=42 | BEGIN
LSN=1002 | TXN=42 | UPDATE | Table=users | Page=5 | Offset=200 | Old=balance:100 | New=balance:200
LSN=1003 | TXN=42 | UPDATE | Index=idx_balance | Page=12 | Offset=50 | Old=[100→Page5] | New=[200→Page5]
LSN=1004 | TXN=42 | COMMIT
```

**Key Properties:**
- **LSN (Log Sequence Number)**: Monotonically increasing, globally ordered
- **Idempotent**: Replaying the same log entry multiple times has the same effect
- **Physical vs Logical Logs**:
  - Physical: "Change byte X at page Y to value Z"
  - Logical: "Update users SET balance = 200 WHERE user_id = 42"
  - Modern DBs use physiological logs (physical page + logical within page)

#### 2.3 Write to WAL Buffer

**In-Memory Buffer:**
```
Application → WAL Buffer (memory) → WAL File (disk)
```

**Buffer Management:**
- Each transaction appends to the WAL buffer
- Buffer is typically 1-16 MB
- Avoids individual disk writes for every operation

#### 2.4 WAL Flush Strategy

**When to flush WAL to disk?**

**Option 1: Commit-time Flush (Default)**
```sql
BEGIN TRANSACTION;
UPDATE users SET balance = 200 WHERE user_id = 42;  -- Log written to buffer
COMMIT;  -- fsync(WAL) here! Ensures durability
```

**Option 2: Group Commit**
- Wait for N transactions or T milliseconds
- Flush multiple transactions' logs in one fsync
- Reduces disk I/O at cost of slight latency increase

**Option 3: Async Commit (PostgreSQL, MySQL)**
```sql
SET synchronous_commit = OFF;
COMMIT;  -- Returns immediately, WAL flushed in background
```
- **Tradeoff**: Better latency, but risk losing committed transactions in crash

---

### Phase 3: fsync - Ensuring Durability

#### 3.1 The Durability Problem

**Storage Hierarchy:**
```
Application Write
    ↓
OS Page Cache (RAM)
    ↓
Disk Controller Cache (battery-backed or not)
    ↓
Disk Platters / SSD NAND Flash
```

**Problem**: `write()` system call only writes to OS cache, not durable storage!

#### 3.2 fsync System Call

**What fsync does:**
```c
fd = open("wal.log", O_WRONLY | O_APPEND);
write(fd, log_record, size);
fsync(fd);  // Blocks until data is on persistent storage
```

**Guarantees:**
- Forces OS page cache to disk controller
- Forces disk controller to flush its cache
- Blocks until all data is on non-volatile storage

**Performance Cost:**
- **HDD**: 5-10ms per fsync (rotational latency + seek time)
- **SSD**: 0.1-1ms per fsync (NAND write + FTL overhead)
- **NVMe SSD**: 0.01-0.1ms

**WAL fsync order:**
```
1. Write WAL record to buffer
2. Write buffer to OS cache
3. fsync(WAL)  ← Transaction cannot commit until this completes
4. Return success to client
```

#### 3.3 Optimizations

**Group Commit:**
```
Transaction A commits → Wait 1ms
Transaction B commits → Wait 1ms
Transaction C commits → Wait 1ms
→ Single fsync for all three → 3x throughput improvement!
```

**Direct I/O (O_DIRECT):**
- Bypass OS page cache entirely
- Application manages its own buffer cache
- Reduces double-caching overhead
- Used by InnoDB, PostgreSQL (with wal_sync_method = open_datasync)

**Battery-Backed Write Cache:**
- Enterprise storage controllers have battery-backed cache
- DB can treat write to cache as durable
- Controller flushes to disk asynchronously

---

### Phase 4: Modify In-Memory Buffer Pool

#### 4.1 Buffer Pool Structure

**Buffer Pool**: In-memory cache of disk pages

```
Buffer Pool (RAM)
├── Data Pages (table rows)
├── Index Pages (B+ tree nodes)
├── Lock Table
└── Hash Table (Page ID → Buffer Frame)
```

**Page Structure:**
```
[Page Header | Row Directory | Free Space | Row Data | Page Footer]
```

#### 4.2 Update the Page

**Steps:**
1. **Lookup**: Check if page is in buffer pool (hash table lookup)
2. **Load if needed**: If not in memory, read from disk
3. **Pin the page**: Prevent eviction while modifying
4. **Mark as dirty**: Set dirty bit to indicate unsaved changes
5. **Update the data**: Modify the row in-place
6. **Set LSN**: Record the LSN of the log record that modified this page
7. **Unpin the page**: Allow eviction (but keep dirty bit set)

**Example:**
```
Page 5 (users table):
Before: | user_id=42 | balance=100 | ...
After:  | user_id=42 | balance=200 | ...
         LSN=1002 (stamped on page)
```

**Why stamp LSN on page?**
- During recovery, DB can determine if a log record was already applied
- If page LSN ≥ log record LSN, skip the redo operation

---

### Phase 5: Update Indexes (B+ Tree Operations)

#### 5.1 Why Update Indexes?

Every secondary index must be updated to reflect the new data.

**Example:**
```sql
CREATE INDEX idx_balance ON users(balance);
UPDATE users SET balance = 200 WHERE user_id = 42;
```

**Index changes needed:**
- Remove entry: `balance=100 → user_id=42`
- Add entry: `balance=200 → user_id=42`

#### 5.2 B+ Tree Structure Recap

```
                    [Root: 100|200]
                   /       |        \
          [10|50|75]   [100|125|150]   [200|250|300]
             ↓              ↓                ↓
        [Leaf Pages]   [Leaf Pages]    [Leaf Pages]
```

**Properties:**
- Balanced: All leaf nodes at same depth
- Sorted: Keys in ascending order
- High fanout: 100-500 children per node (minimizes depth)

#### 5.3 B+ Tree Insertion

**Case 1: Leaf has space**
```
[100|125|150| _ | _ ]  ← Insert 130
[100|125|130|150| _ ]  ← Simple insertion, no rebalancing
```
**Cost**: 1 page write

**Case 2: Leaf is full → Split**
```
[100|125|150|175|200]  ← Insert 160 (no space!)

Split into:
[100|125|150]  [160|175|200]
       ↓               ↓
Push 160 up to parent
```

**Parent update:**
```
[100|200]  →  [100|160|200]
```

**Cost**: 3 page writes (2 leaf pages + 1 parent)

**Case 3: Cascading splits**
```
Parent is also full → Split parent
Grandparent is also full → Split grandparent
...
Root is full → Split root (tree height increases!)
```

**Worst-case cost**: O(log N) page writes

#### 5.4 B+ Tree Deletion

**Case 1: Leaf has extra keys (above minimum)**
```
[100|125|150|175]  ← Delete 125
[100|150|175]      ← Simple deletion
```
**Cost**: 1 page write

**Case 2: Leaf below minimum → Merge or redistribute**
```
Leaf A: [100]  ← Below minimum (suppose min=2)
Leaf B: [150|175|200]

Option 1: Merge
[100|150|175|200]  ← Merge A into B

Option 2: Redistribute
Leaf A: [100|150]
Leaf B: [175|200]
```

**Cost**: 2-3 page writes

**Case 3: Cascading merges**
- Parent becomes underfull → merge parent
- Can propagate up to root
- If root has only 1 child → make child new root (tree height decreases)

#### 5.5 Write Amplification in B+ Trees

**Problem**: A single row update can trigger multiple page writes

**Example:**
```
UPDATE users SET balance = balance + 1 WHERE user_id = 42;
```

**Worst case writes:**
1. Data page: 1 write
2. Primary index leaf: 1 write
3. Primary index leaf split: +1 write
4. Primary index parent update: +1 write
5. Secondary index (idx_balance) delete: 1 write
6. Secondary index insert: 1 write
7. Secondary index split: +1 write

**Total: 7 page writes for 1 row update!**

**Mitigation strategies:**
- **Fill factor**: Leave pages 70% full during bulk loads (space for future inserts)
- **Delayed splits**: Use overflow pages temporarily
- **Bulk updates**: Sort updates by key order, minimize random writes

---

### Phase 6: Page Flushing to Disk

#### 6.1 Dirty Pages

**Dirty Page**: A page in buffer pool that has been modified but not written to disk

**Why not flush immediately?**
- **Performance**: fsync is expensive (5-10ms for HDD)
- **Write coalescing**: Multiple updates to same page → one disk write
- **WAL already provides durability**: Can reconstruct from logs

#### 6.2 Flushing Strategies

**Strategy 1: Lazy Flushing (Background Writer)**
```
Checkpoint Process (runs every N seconds):
1. Scan buffer pool for dirty pages
2. Write dirty pages to disk in batches
3. fsync data files
4. Advance checkpoint LSN in control file
```

**PostgreSQL:**
```
bgwriter process:
- Writes dirty pages in background
- Limits: bgwriter_delay=200ms, bgwriter_lru_maxpages=100
```

**MySQL InnoDB:**
```
Page cleaner threads:
- Flush dirty pages based on innodb_io_capacity
- Adaptive flushing: speed up when redo log space is low
```

**Strategy 2: Eager Flushing (Write-Through Cache)**
- Write page to disk immediately after modification
- Simple but slow
- Not used in modern DBs

**Strategy 3: Eviction-Triggered Flushing**
```
Buffer pool is full → Need to evict a page
If page is dirty → Flush before eviction
If page is clean → Evict immediately
```

#### 6.3 Checkpoint Process

**Purpose**: Limit recovery time after crash

**What is a checkpoint?**
```
Checkpoint = "All dirty pages with LSN ≤ X have been flushed to disk"
```

**Checkpoint steps:**
1. Record checkpoint start LSN
2. Flush all dirty pages with LSN ≤ checkpoint LSN
3. fsync data files
4. Write checkpoint record to WAL
5. fsync WAL
6. Update control file with checkpoint LSN

**Recovery after crash:**
```
1. Read last checkpoint LSN from control file
2. Replay WAL from checkpoint LSN to end
3. Dirty pages since checkpoint will be reconstructed
```

**Checkpoint frequency tradeoff:**
- **Too frequent**: High I/O overhead, fsync storms
- **Too infrequent**: Slow recovery (must replay long WAL)

**Typical settings:**
- PostgreSQL: checkpoint_timeout=5min, checkpoint_completion_target=0.5
- MySQL: innodb_flush_log_at_trx_commit=1, innodb_flush_method=O_DIRECT

---

## Complete Timeline Example

Let's trace a single UPDATE operation through all phases:

```sql
UPDATE users SET balance = 200 WHERE user_id = 42;
```

### Timeline:

**T=0ms: Planning**
- Parse SQL
- Validate table, columns, constraints
- Acquire EXCLUSIVE lock on user_id=42
- Create execution plan: Use primary key index to find row

**T=1ms: WAL**
- Generate log record: LSN=1002, UPDATE, old=100, new=200
- Append to WAL buffer
- Write WAL buffer to OS cache
- **fsync(WAL)** ← Blocks here for 0.5ms (SSD)

**T=1.5ms: Modify Data Page**
- Lookup page 5 in buffer pool (cache hit)
- Pin page 5
- Update row: balance 100 → 200
- Mark page as dirty, stamp LSN=1002
- Unpin page 5

**T=1.6ms: Update Primary Index**
- Primary key index (user_id) → No change needed (user_id unchanged)

**T=1.7ms: Update Secondary Index**
- Remove old entry: balance=100 → user_id=42
  - Traverse B+ tree to find leaf with key 100
  - Remove entry from leaf
  - Mark page 12 as dirty
- Insert new entry: balance=200 → user_id=42
  - Traverse B+ tree to find leaf for key 200
  - Leaf has space, insert entry
  - Mark page 15 as dirty

**T=2ms: Commit**
- Write COMMIT log record (LSN=1004) to WAL
- fsync(WAL) ← Already flushed at T=1.5ms, or flush now if grouped
- Release lock on user_id=42
- Return success to client

**T=5min: Background Checkpoint**
- Checkpoint process wakes up
- Flush dirty pages: page 5, page 12, page 15
- fsync data files
- Write checkpoint record to WAL
- Update control file: last checkpoint LSN=1004

---

## Ensuring Durability and Consistency

### Durability Guarantee

**WAL + fsync = ACID Durability**

**Invariant:** "If a transaction commits, its changes survive any crash"

**How it's enforced:**
1. **Write WAL before data pages**: Log records are flushed before dirty pages
2. **fsync at commit time**: Transaction cannot commit until WAL is on disk
3. **Recovery replay**: After crash, replay WAL to reconstruct lost dirty pages

**Crash scenarios:**

**Scenario 1: Crash after WAL flush, before data page flush**
```
T=0: Write log record LSN=1002
T=1: fsync(WAL)  ← Log is durable
T=2: COMMIT succeeds
T=3: Modify data page (mark dirty)
T=4: CRASH!  ← Data page not flushed
```
**Recovery:**
- Replay WAL from last checkpoint
- Redo LSN=1002: Update page 5, balance 100 → 200
- Data is reconstructed ✓

**Scenario 2: Crash before WAL flush**
```
T=0: Write log record to WAL buffer
T=1: CRASH!  ← Log not flushed
```
**Recovery:**
- Log record is lost
- Transaction never committed (client received no success)
- Consistent state ✓

### Consistency Guarantee

**ACID Consistency:** "Database transitions from one valid state to another"

**Mechanisms:**

**1. Constraint Checking**
- Before commit: Validate PRIMARY KEY, FOREIGN KEY, CHECK, UNIQUE
- If violation: ROLLBACK transaction

**2. Transaction Atomicity**
- UNDO logs: Record old values
- If transaction aborts: Replay UNDO logs to revert changes

**3. Isolation Levels**
- **Read Uncommitted**: Dirty reads allowed
- **Read Committed**: Only see committed data
- **Repeatable Read**: Snapshot isolation (MVCC)
- **Serializable**: Equivalent to serial execution

**4. Lock-Based Concurrency Control**
- Two-Phase Locking (2PL): Acquire all locks before releasing any
- Prevents write-write conflicts

**Example conflict:**
```
Transaction A: UPDATE users SET balance = balance + 100 WHERE user_id = 42;
Transaction B: UPDATE users SET balance = balance - 50 WHERE user_id = 42;

Without locking:
- A reads balance=100
- B reads balance=100
- A writes balance=200
- B writes balance=50
- Final balance: 50 (incorrect! should be 150)

With locking:
- A acquires EXCLUSIVE lock
- B waits
- A writes balance=200, commits, releases lock
- B acquires lock, reads balance=200
- B writes balance=150, commits
- Final balance: 150 ✓
```

---

## Minimizing Write Amplification

**Write Amplification**: Ratio of bytes written to storage vs bytes changed by application

```
Write Amplification = Total Bytes Written / Logical Bytes Changed
```

**Example:**
- Update 1 row (100 bytes)
- Writes 3 pages (12 KB) + WAL (500 bytes) = 12.5 KB
- Amplification = 12,500 / 100 = 125x

### Strategies to Reduce Write Amplification

#### 1. Larger Page Size
- **Tradeoff**: Amortizes overhead over more data, but increases space wastage
- PostgreSQL: 8 KB pages
- MySQL InnoDB: 16 KB pages
- SQL Server: 8 KB pages

#### 2. Log-Structured Merge Trees (LSM-Trees)
- Alternative to B+ trees
- Append-only writes → No random writes
- Compaction merges sorted runs
- Used by: LevelDB, RocksDB, Cassandra, HBase

**Write amplification:**
- B+ tree: 10-100x
- LSM-tree: 10-50x (during compaction)

#### 3. Batching & Bulk Operations
```sql
-- Bad: 1000 individual updates
UPDATE users SET balance = balance + 1 WHERE user_id = 1;
UPDATE users SET balance = balance + 1 WHERE user_id = 2;
...

-- Good: Bulk update
UPDATE users SET balance = balance + 1 WHERE user_id IN (1, 2, ..., 1000);
```

#### 4. Fill Factor
- Leave pages partially empty (e.g., 70% full)
- Space for future inserts without splits
```sql
CREATE INDEX idx_balance ON users(balance) WITH (FILLFACTOR=70);
```

#### 5. Partitioning
- Split large tables into smaller partitions
- Smaller indexes → faster rebalancing
```sql
CREATE TABLE users (
  user_id INT,
  balance INT,
  created_at DATE
) PARTITION BY RANGE (created_at) (
  PARTITION p2023 VALUES LESS THAN ('2024-01-01'),
  PARTITION p2024 VALUES LESS THAN ('2025-01-01')
);
```

#### 6. Compression
- Compress pages before writing to disk
- Reduces I/O at cost of CPU
```sql
ALTER TABLE users ROW_FORMAT=COMPRESSED;
```

---

## Minimizing Latency

**Latency Components:**
```
Total Latency = Network + Planning + Locking + WAL + fsync + Index Update + Commit
```

### Strategies to Reduce Latency

#### 1. Group Commit (Batching fsync)
```
Wait for multiple transactions → Single fsync
```
**Impact**: Reduces fsync from 10ms × 100 = 1000ms to 10ms × 1 = 10ms (100x improvement)

**Tradeoff**: Adds 1-5ms latency per transaction

#### 2. Async Commit
```sql
SET synchronous_commit = OFF;
COMMIT;  -- Returns before fsync
```
**Impact**: Reduces latency from 10ms to 0.1ms

**Tradeoff**: Risk losing last 200ms of commits in crash

#### 3. SSD/NVMe Storage
- HDD fsync: 5-10ms
- SATA SSD fsync: 0.1-1ms
- NVMe SSD fsync: 0.01-0.1ms

#### 4. Battery-Backed Write Cache
- Storage controller with battery backup
- DB treats write to cache as durable
- Reduces fsync latency by 10-100x

#### 5. Read Replicas (Read Offloading)
```
Primary (writes) → Replicas (reads)
```
- Offload read traffic to replicas
- Primary handles only writes → less lock contention

#### 6. Prepared Statements
```sql
PREPARE stmt FROM 'UPDATE users SET balance = ? WHERE user_id = ?';
EXECUTE stmt USING 200, 42;
```
**Impact**: Skip parsing and planning → save 0.5-2ms per query

#### 7. Connection Pooling
- Reuse connections instead of establishing new ones
- Avoids TCP handshake + authentication overhead

#### 8. Index Optimization
- Use covering indexes to avoid table lookups
```sql
CREATE INDEX idx_covering ON users(user_id, balance);
-- Query can be satisfied from index alone
```

#### 9. Reduce Lock Contention
- **MVCC (Multi-Version Concurrency Control)**: Readers don't block writers
- **Optimistic Locking**: Check conflicts at commit time
- **Fine-grained Locking**: Row-level instead of table-level

---

## Summary: The Complete Write Path

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Application                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
                    [Parse & Validate]
                            ↓
                      [Acquire Locks]
                            ↓
                   [Create Execution Plan]
                            ↓
                ┌───────────────────────┐
                │    Write-Ahead Log    │
                │  (Sequential Write)   │
                └───────────────────────┘
                            ↓
                      [fsync WAL]  ← Durability checkpoint
                            ↓
                ┌───────────────────────┐
                │   Buffer Pool (RAM)   │
                │  - Data Page (dirty)  │
                │  - Index Page (dirty) │
                └───────────────────────┘
                            ↓
                    [B+ Tree Update]
                    (split if needed)
                            ↓
                    [COMMIT Success]
                            ↓
                    [Release Locks]
                            ↓
                ┌───────────────────────┐
                │   Background Flush    │
                │  (Checkpoint Process) │
                └───────────────────────┘
                            ↓
                      [fsync Data Files]
                            ↓
                ┌───────────────────────┐
                │   Persistent Storage  │
                │   (SSD/HDD)           │
                └───────────────────────┘
```

### Key Takeaways

1. **WAL + fsync** guarantees durability (ACID D)
2. **Locking + MVCC** guarantees consistency (ACID C)
3. **Two-Phase Locking** guarantees isolation (ACID I)
4. **UNDO logs** guarantee atomicity (ACID A)
5. **Background checkpoints** limit recovery time
6. **B+ tree rebalancing** maintains query performance
7. **Group commit** reduces fsync overhead
8. **Fill factor & partitioning** reduce write amplification
9. **SSD + async commit** reduce latency
10. **Trade-offs everywhere**: Durability vs Performance, Consistency vs Availability

---

## Further Reading

- **PostgreSQL WAL**: https://www.postgresql.org/docs/current/wal-intro.html
- **MySQL InnoDB Architecture**: https://dev.mysql.com/doc/refman/8.0/en/innodb-architecture.html
- **ARIES Recovery Algorithm**: https://cs.stanford.edu/people/chrismre/cs345/rl/aries.pdf
- **B+ Tree Visualization**: https://www.cs.usfca.edu/~galles/visualization/BPlusTree.html
- **Write-Ahead Logging (Jim Gray)**: "Transaction Processing: Concepts and Techniques"

---

*Document created: October 2025*

