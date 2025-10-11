# ACID: Durability

## Definition

**Durability** ensures that once a transaction is committed, its changes are permanent and will survive any subsequent system failures, crashes, or power outages. The data is safely stored on non-volatile storage and can be recovered even after catastrophic failures.

> **Key Principle**: "Changes Are Permanent"

---

## Core Concepts

### Persistence Guarantee

Once a transaction commits, the data persists:

```sql
BEGIN TRANSACTION;
    UPDATE accounts SET balance = 500 WHERE id = 'A';
COMMIT;  -- ✅ Changes guaranteed to persist

-- System crashes here
-- Power outage occurs
-- Hardware fails

-- After recovery:
-- balance is still 500 (durability maintained)
```

**Key Properties:**
- Committed data survives crashes
- Data is written to non-volatile storage (disk, SSD)
- Recovery mechanisms restore committed transactions
- Uncommitted transactions are rolled back

---

## How Durability Works

### Write-Ahead Logging (WAL)

The primary mechanism for ensuring durability:

```
1. Write change to transaction log (on disk)
2. Apply change to in-memory buffer
3. Mark transaction as committed in log
4. Eventually flush buffer to database files

Log Entry Example:
[LSN: 1001] [TXN: T1] [BEGIN]
[LSN: 1002] [TXN: T1] [UPDATE accounts SET balance=500 WHERE id='A']
[LSN: 1003] [TXN: T1] [COMMIT]
```

**Why WAL Works:**
- Log is sequential (fast writes)
- Log is written before data pages (hence "write-ahead")
- On crash, replay log to recover committed transactions
- Redo committed but not flushed, undo uncommitted

### ACID Flow with Durability

```
Application                 Database                 Storage
    |                          |                         |
    |-- BEGIN TRANSACTION ---→|                         |
    |                          |                         |
    |-- UPDATE statement ----→|                         |
    |                          |-- Write to WAL -------→| (Disk)
    |                          |-- Update buffer pool   |
    |                          |                         |
    |-- COMMIT -------------→|                         |
    |                          |-- Write COMMIT to WAL →| (Disk) ✅ Durability!
    |←- Success ---------------|                         |
    |                          |                         |
    |                          |-- Background flush ---→| (Eventually)
```

---

## Implementation Mechanisms

### 1. Transaction Logs

Records all database modifications:

```sql
-- Transaction T1
BEGIN;
    INSERT INTO orders VALUES (1, 'customer1', 100.00);
    UPDATE inventory SET stock = stock - 1 WHERE product_id = 5;
COMMIT;

-- Log entries (simplified):
<T1, START>
<T1, INSERT, orders, (1, 'customer1', 100.00)>
<T1, UPDATE, inventory, product_id=5, old_stock=10, new_stock=9>
<T1, COMMIT>
```

**Types of Logging:**

#### Undo Logging
Records old values for rollback:
```
<T1, START>
<T1, accounts.id=1, old_balance=1000>  -- Store old value
<T1, COMMIT>
```

#### Redo Logging
Records new values for recovery:
```
<T1, START>
<T1, accounts.id=1, new_balance=500>  -- Store new value
<T1, COMMIT>
```

#### Undo/Redo Logging
Records both old and new values:
```
<T1, START>
<T1, accounts.id=1, old=1000, new=500>
<T1, COMMIT>
```

### 2. Checkpointing

Periodically flush in-memory changes to disk:

```
Timeline:
|--- Checkpoint 1 ---|--- Normal Operation ---|--- Checkpoint 2 ---|--- Crash ---|

Recovery process:
1. Start from last checkpoint (Checkpoint 2)
2. Replay log entries after Checkpoint 2
3. Much faster than replaying entire log
```

**Checkpoint Process:**

```
1. Flush all dirty pages (modified in memory) to disk
2. Write checkpoint record to log
3. Record which transactions are active

Checkpoint Record:
<CHECKPOINT, [T5, T7, T8]>  -- Active transactions
```

### 3. Buffer Pool Management

In-memory cache of database pages:

```
RAM (Buffer Pool)          Disk (Database Files)
┌─────────────┐           ┌─────────────┐
│ Page 1      │           │ Page 1      │
│ Page 2      │ ←flush→   │ Page 2      │
│ Page 3      │           │ Page 3      │
│   (dirty)   │           │   ...       │
└─────────────┘           └─────────────┘
```

**Write Strategies:**

#### Write-Through
- Write immediately to disk on every change
- Slower but simpler
- Rarely used in modern databases

#### Write-Back (most common)
- Write to memory, flush later
- Much faster
- Requires WAL for durability

### 4. Force vs No-Force Policy

**Force:** Flush all changes to disk before commit
- Slower commits
- Simpler recovery

**No-Force:** Don't flush, rely on WAL
- Faster commits
- More complex recovery
- Used by most modern databases

### 5. Steal vs No-Steal Policy

**Steal:** Allow uncommitted changes to be written to disk
- Better memory management
- Requires undo logging

**No-Steal:** Only write committed changes to disk
- Simpler but requires more memory
- No undo needed

---

## Database Recovery

### Recovery Process

After a crash, database must restore consistent state:

```
1. Analysis Phase
   - Read transaction log
   - Identify committed vs uncommitted transactions
   - Determine which pages need recovery

2. Redo Phase
   - Replay all committed transactions from log
   - Restore committed changes that weren't flushed

3. Undo Phase
   - Rollback all uncommitted transactions
   - Restore database to consistent state
```

### ARIES Algorithm

**A**lgorithms for **R**ecovery and **I**solation **E**xploiting **S**emantics

Most databases use ARIES or variants:

```
Log Sequence Number (LSN):
- Monotonically increasing ID for each log record
- Each page has LSN of last update (pageLSN)
- Each log record has prevLSN (forms linked list)

Recovery:
1. Analysis: Scan log forward from checkpoint
2. Redo: Replay log from earliest dirty page
3. Undo: Rollback uncommitted transactions
```

### Example Recovery Scenario

```sql
-- System state before crash:

-- Transaction T1 (committed)
BEGIN;
    UPDATE accounts SET balance = 500 WHERE id = 'A';  -- [LSN: 100]
COMMIT;  -- [LSN: 101] ✅ Logged

-- Transaction T2 (uncommitted)
BEGIN;
    UPDATE accounts SET balance = 800 WHERE id = 'B';  -- [LSN: 102]
-- CRASH HERE (no COMMIT logged)

-- Recovery Process:

1. Analysis:
   - Find last checkpoint
   - Identify T1 as committed, T2 as uncommitted

2. Redo:
   - Replay LSN 100: Update account A to 500
   - Replay LSN 102: Update account B to 800

3. Undo:
   - Rollback LSN 102: Restore account B to original value

-- Final State:
-- Account A: 500 (T1's change persisted - DURABILITY!)
-- Account B: original value (T2 rolled back)
```

---

## Real-World Examples

### Example 1: Bank Transaction

```sql
-- Scenario: Transfer $100 between accounts

BEGIN TRANSACTION;
    UPDATE accounts SET balance = balance - 100 WHERE id = 'A';
    -- Log: <T1, accounts.A, balance, old=1000, new=900>
    
    UPDATE accounts SET balance = balance + 100 WHERE id = 'B';
    -- Log: <T1, accounts.B, balance, old=500, new=600>
    
COMMIT;
-- Log: <T1, COMMIT> ← Durability guaranteed!

-- Crash scenarios:

-- Scenario 1: Crash before COMMIT
-- Result: Both updates rolled back (atomicity)

-- Scenario 2: Crash after COMMIT
-- Result: Both updates persist (durability)

-- Scenario 3: Crash after COMMIT, before flush to disk
-- Result: Changes recovered from log (durability)
```

### Example 2: E-commerce Order

```sql
-- Customer places order

BEGIN TRANSACTION;
    -- Create order
    INSERT INTO orders (id, customer_id, total) VALUES (1001, 42, 299.99);
    -- Log written to disk
    
    -- Reserve inventory
    UPDATE products SET stock = stock - 1 WHERE id = 567;
    -- Log written to disk
    
    -- Record payment
    INSERT INTO payments (order_id, amount) VALUES (1001, 299.99);
    -- Log written to disk
    
COMMIT;
-- Commit record written to disk ← Durability point

-- Even if system crashes immediately after COMMIT:
-- - Order 1001 exists
-- - Inventory is reserved
-- - Payment is recorded
-- All because transaction log is on disk!
```

### Example 3: Social Media Post

```sql
-- User publishes a post

BEGIN TRANSACTION;
    INSERT INTO posts (id, user_id, content, created_at)
    VALUES (9999, 123, 'Hello World!', NOW());
    
    UPDATE users SET post_count = post_count + 1 WHERE id = 123;
    
COMMIT;  -- Post is now durable

-- After commit, even if:
-- - Server crashes
-- - Datacenter loses power
-- - Disk controller fails
-- 
-- The post will still exist after recovery
-- (assuming disk hardware is functional)
```

---

## Durability Guarantees

### fsync() System Call

Forces OS to flush data to disk:

```c
// Pseudocode for database commit

void commit_transaction(Transaction txn) {
    // 1. Write commit record to log buffer
    write_to_log_buffer(txn.commit_record);
    
    // 2. Force log to disk (durability!)
    int fd = open("/db/wal.log");
    write(fd, log_buffer);
    fsync(fd);  // ← Block until data is on disk
    close(fd);
    
    // 3. Return success to application
    return SUCCESS;
}
```

**Operating System Guarantees:**
- `write()`: May stay in OS buffer
- `fsync()`: Guarantees data is on disk
- Battery-backed cache: Treated as durable

### Storage Hierarchy

```
CPU Registers (volatile)
    ↓
L1/L2/L3 Cache (volatile)
    ↓
RAM (volatile) ← Data lost on crash
    ↓
─────────────────────────────────
    ↓
SSD/HDD (non-volatile) ← Durable storage
    ↓
RAID Arrays (redundant)
    ↓
Offsite Backups (disaster recovery)
```

---

## Durability in Distributed Systems

### Replication

Data copied to multiple servers:

```
Write request
    ↓
[Primary Server] --replicate→ [Replica 1]
                 --replicate→ [Replica 2]
                 --replicate→ [Replica 3]
```

**Synchronous Replication:**
```sql
-- Transaction not committed until replicas acknowledge
BEGIN;
    UPDATE accounts SET balance = 500 WHERE id = 'A';
COMMIT;  -- Waits for replicas to confirm

-- Pros: Strong durability
-- Cons: Higher latency
```

**Asynchronous Replication:**
```sql
-- Transaction commits immediately
BEGIN;
    UPDATE accounts SET balance = 500 WHERE id = 'A';
COMMIT;  -- Returns immediately

-- Replicas updated later
-- Pros: Low latency
-- Cons: Potential data loss if primary fails
```

### Quorum Writes

Require N/2 + 1 servers to acknowledge:

```
Configuration: 3 replicas, quorum = 2

Write request
    ↓
Server 1: ✅ ACK (written to disk)
Server 2: ✅ ACK (written to disk) ← Quorum reached!
Server 3: ⏳ (slow, still writing)

Client receives success (durable)
```

### Multi-Datacenter Durability

```
Datacenter 1 (US-East)    Datacenter 2 (US-West)    Datacenter 3 (EU)
[Primary]                 [Sync Replica]            [Async Replica]
    |                            |                         |
    |-------- Write ----------→  |                         |
    |←------- ACK ---------------|                         |
    |                            |                         |
    |(commit successful)         |-------- Async -------→  |
```

---

## Durability vs Performance

### Trade-offs

| Configuration | Durability | Performance | Use Case |
|--------------|------------|-------------|----------|
| fsync() every commit | Highest | Lowest | Financial systems |
| fsync() every N commits | Medium | Medium | OLTP applications |
| Async commits | Lowest | Highest | Analytics, logging |

### Group Commit

Batch multiple transactions into single fsync:

```
Transaction T1 commits → Add to group
Transaction T2 commits → Add to group
Transaction T3 commits → Add to group
    ↓
Single fsync() for all three
    ↓
All three return success

-- Benefits:
-- - Fewer disk I/O operations
-- - Higher throughput
-- - Slight increase in latency per transaction
```

### Tuning Durability

#### PostgreSQL

```sql
-- Full durability (default)
synchronous_commit = on

-- Faster, but ~100ms data loss window
synchronous_commit = off

-- Group commit batch size
wal_writer_delay = 200ms

-- Force fsync method
wal_sync_method = fsync
```

#### MySQL

```sql
-- Full durability
innodb_flush_log_at_trx_commit = 1  -- fsync every commit

-- Better performance, less durable
innodb_flush_log_at_trx_commit = 2  -- fsync every second

-- Fastest, least durable
innodb_flush_log_at_trx_commit = 0  -- fsync controlled by OS
```

---

## Testing Durability

### Kill Test

```python
import psycopg2
import subprocess
import time

def test_durability():
    conn = psycopg2.connect("dbname=test")
    cur = conn.cursor()
    
    # Insert data
    cur.execute("BEGIN")
    cur.execute("INSERT INTO test_table VALUES (1, 'test data')")
    cur.execute("COMMIT")
    
    # Verify insert
    cur.execute("SELECT * FROM test_table WHERE id = 1")
    assert cur.fetchone() is not None
    
    # Kill database process abruptly
    subprocess.run(["killall", "-9", "postgres"])
    time.sleep(2)
    
    # Restart database
    subprocess.run(["pg_ctl", "start"])
    time.sleep(5)
    
    # Reconnect and verify data persisted
    conn = psycopg2.connect("dbname=test")
    cur = conn.cursor()
    cur.execute("SELECT * FROM test_table WHERE id = 1")
    
    assert cur.fetchone() is not None  # Data should still exist!
    print("Durability test PASSED")
```

### Power Failure Simulation

```bash
# Using Linux device-mapper to simulate disk failures

# 1. Insert data
psql -c "BEGIN; INSERT INTO test VALUES (1); COMMIT;"

# 2. Simulate power failure
echo 0 > /sys/block/sda/device/power/control  # Immediate shutdown

# 3. Restart system
# (reboot)

# 4. Verify data
psql -c "SELECT * FROM test WHERE id = 1;"
# Should return the inserted row
```

---

## Programming Examples

### Python: Explicit Durability

```python
import psycopg2

def durable_insert(data):
    conn = psycopg2.connect("dbname=production")
    cur = conn.cursor()
    
    try:
        cur.execute("BEGIN")
        cur.execute("INSERT INTO critical_data VALUES (%s)", (data,))
        
        # Explicitly ensure durability
        cur.execute("COMMIT")
        
        # After commit returns, data is guaranteed durable
        print("Data committed and durable")
        
    except Exception as e:
        cur.execute("ROLLBACK")
        raise e
    
    finally:
        cur.close()
        conn.close()

# After this function returns successfully,
# data will survive any system crash
durable_insert("important data")
```

### Java: Transaction with Durability

```java
import java.sql.*;

public class DurableTransaction {
    public void saveOrder(Order order) throws SQLException {
        Connection conn = DriverManager.getConnection(
            "jdbc:postgresql://localhost/orders");
        
        try {
            conn.setAutoCommit(false);
            
            // Insert order
            PreparedStatement ps = conn.prepareStatement(
                "INSERT INTO orders (id, customer, total) VALUES (?, ?, ?)");
            ps.setInt(1, order.getId());
            ps.setString(2, order.getCustomer());
            ps.setDouble(3, order.getTotal());
            ps.executeUpdate();
            
            // Commit - data is now durable
            conn.commit();
            
            // After commit() returns, the order is guaranteed to persist
            System.out.println("Order saved durably");
            
        } catch (SQLException e) {
            conn.rollback();
            throw e;
        } finally {
            conn.close();
        }
    }
}
```

### Node.js: Async Durability

```javascript
const { Pool } = require('pg');
const pool = new Pool({ database: 'production' });

async function durableUpdate(userId, newData) {
    const client = await pool.connect();
    
    try {
        await client.query('BEGIN');
        
        // Update user data
        await client.query(
            'UPDATE users SET data = $1, updated_at = NOW() WHERE id = $2',
            [newData, userId]
        );
        
        // Commit transaction
        await client.query('COMMIT');
        
        // After this point, even if Node.js process crashes,
        // the update is durable in the database
        console.log('Update committed and durable');
        
    } catch (error) {
        await client.query('ROLLBACK');
        throw error;
    } finally {
        client.release();
    }
}
```

---

## Durability Failures

### What Can Cause Data Loss?

#### 1. Disk Corruption

```
Physical damage to storage media
→ Data may be unrecoverable
→ Solution: RAID, backups, replication
```

#### 2. Silent Data Corruption (Bit Rot)

```
Bits flip without detection
→ Corrupt data returned as valid
→ Solution: Checksums, ZFS/Btrfs, error-correcting codes
```

#### 3. Firmware/Controller Bugs

```
SSD controller lies about fsync completion
→ Data in volatile cache presented as durable
→ Solution: Battery-backed cache, enterprise-grade storage
```

#### 4. Filesystem Issues

```
Filesystem bugs or corruption
→ Committed data may be lost
→ Solution: Use tested filesystems (ext4, XFS, ZFS)
```

#### 5. Backup Failures

```
Backups not tested or corrupt
→ Cannot recover from disaster
→ Solution: Regular backup testing, multiple backup locations
```

---

## Best Practices

### 1. Use Transactions Properly

```sql
-- GOOD: Explicit transaction
BEGIN;
    INSERT INTO orders VALUES (...);
    UPDATE inventory SET stock = stock - 1;
COMMIT;  -- Data is durable after this

-- BAD: Auto-commit mode (less control)
INSERT INTO orders VALUES (...);
-- Implicitly committed, but harder to group operations
```

### 2. Verify Commit Success

```python
# BAD: Ignore commit result
try:
    conn.execute("INSERT INTO ...")
    conn.commit()  # Might fail!
except:
    pass  # Silently fail

# GOOD: Check commit result
try:
    conn.execute("INSERT INTO ...")
    conn.commit()
    # Only now is data durable
    return SUCCESS
except Exception as e:
    conn.rollback()
    log.error(f"Commit failed: {e}")
    return ERROR
```

### 3. Use Replication for Critical Data

```
Primary Database (writes)
    ↓ (synchronous replication)
Standby Database (ready for failover)
    ↓ (asynchronous replication)
DR Database (disaster recovery)
```

### 4. Regular Backups

```bash
# Full backup daily
pg_dump mydb > /backups/mydb_$(date +%Y%m%d).sql

# Point-in-time recovery with WAL archiving
archive_command = 'cp %p /wal_archive/%f'

# Test restore regularly!
pg_restore /backups/mydb_20231010.sql
```

### 5. Monitor Storage Health

```bash
# Check disk health
smartctl -a /dev/sda

# Monitor filesystem errors
dmesg | grep -i error

# Check RAID status
cat /proc/mdstat
```

---

## Durability in Different Storage Types

### Hard Disk Drive (HDD)

```
Characteristics:
- Mechanical (spinning platters)
- ~100 MB/s sequential write
- fsync() takes 5-10ms (platter rotation)
- Prone to mechanical failure

Durability:
- Generally reliable for fsync()
- Write cache can be problematic
- Use with RAID for redundancy
```

### Solid State Drive (SSD)

```
Characteristics:
- Electronic (flash memory)
- ~500 MB/s sequential write
- fsync() takes 0.1-1ms
- Limited write endurance

Durability:
- Faster than HDD
- May cache writes (check specs)
- Power loss can cause issues
- Enterprise SSDs have power-loss protection
```

### NVMe SSD

```
Characteristics:
- PCIe interface
- ~3000 MB/s sequential write
- fsync() takes 0.05-0.5ms
- Highest performance

Durability:
- Excellent performance
- Same caching concerns as SATA SSD
- Enterprise models with supercapacitors
```

### Battery-Backed Cache

```
Characteristics:
- RAM cache + battery backup
- Ultra-fast writes (nanoseconds)
- Battery keeps data alive during power loss
- Flushes to disk when power restored

Durability:
- Can treat battery-backed cache as durable
- Excellent performance
- Enterprise RAID controllers
```

---

## Summary

| Aspect | Description |
|--------|-------------|
| **Definition** | Committed changes persist after crashes |
| **Key Mechanism** | Write-Ahead Logging (WAL) |
| **Storage** | Non-volatile media (disk, SSD) |
| **Recovery** | ARIES algorithm, redo/undo |
| **Trade-offs** | Performance vs durability guarantees |
| **Best Practice** | Use transactions, replication, backups |

### Durability Guarantees

✅ **Guaranteed Durable After:**
- Transaction COMMIT returns successfully
- fsync() completes
- Replication quorum acknowledges

❌ **Not Durable:**
- Before COMMIT
- In application buffer
- In OS buffer (before fsync)
- On single server (without replication)

---

## Related Topics

- [Atomicity](./db-acid-atomicity.md) - All-or-nothing transactions
- [Consistency](./db-acid-consistency.md) - Valid database states
- [Isolation](./db-acid-isolation.md) - Concurrent transaction handling
- [Database Write Sequence](./db-write-sequence.md) - How writes are processed
- [Replication and Backup Strategies](../week1/2-db.md)

