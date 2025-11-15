# MemTable

```
**General Write Path:**

Client Write → CommitLog (disk, durability) → MemTable (memory, speed) → SSTable (disk, persistent, sorted)
```
## Quick Summary

**What it does:** MemTable is an in-memory write buffer used in LSM-tree databases (Cassandra, ScyllaDB) that temporarily stores recent writes before flushing them to disk as immutable SSTables.

**Primary use case:** Enables extremely fast writes by buffering data in memory and batching disk writes, critical for write-heavy distributed databases.

**Key features:**
- In-memory sorted data structure (typically Red-Black Tree or Skip List)
- Holds most recent writes with timestamps
- Auto-flushes to SSTable when size/time threshold reached
- Works in tandem with CommitLog for durability

## Setup (if applicable)

### Cassandra Configuration (cassandra.yaml)

```yaml
# MemTable threshold settings
memtable_heap_space_in_mb: 2048
memtable_offheap_space_in_mb: 2048
memtable_cleanup_threshold: 0.5
memtable_flush_writers: 2

# When to flush MemTable
commitlog_total_space_in_mb: 8192
```

### ScyllaDB Configuration (scylla.yaml)

```yaml
# ScyllaDB auto-tunes memtable sizes based on available memory
# Manual override (rarely needed):
memtable_total_space_in_mb: 4096
```

### Quick Start - Understanding the Flow

```java
// Conceptual flow - not actual Cassandra/Scylla API
// Write Request Flow:
// 1. Write to CommitLog (durability)
// 2. Write to MemTable (performance)
// 3. Return success to client

Session session = cluster.connect();
session.execute(
    "INSERT INTO users (id, name, email) VALUES (?, ?, ?)",
    UUID.randomUUID(), "John Doe", "john@example.com"
);
// Behind the scenes:
// - Data written to CommitLog immediately
// - Data stored in MemTable (in-memory)
// - When MemTable full → flush to SSTable
```

## Core Concepts

### Main Components

**1. MemTable Structure**
- **Sorted in-memory tree** (Red-Black Tree in Cassandra, B-tree in ScyllaDB)
- Organized by partition key, then clustering key
- Each column family (table) has its own MemTable
- Typically 2 MemTables per table: active + flushing

**2. Write Buffer**
- Accumulates writes in memory
- No disk I/O during write operation
- Provides read-your-write consistency
- Latest version always in MemTable

**3. Flush Mechanism**
- Converts MemTable → immutable SSTable
- Triggered by size, time, or CommitLog thresholds
- New MemTable created immediately for incoming writes
- Old MemTable flushed to disk asynchronously

### Key Parameters

| Parameter | Purpose | Typical Value |
|-----------|---------|---------------|
| **Size threshold** | Max memory before flush | 64-512 MB per table |
| **Flush writers** | Concurrent flush threads | 2-4 |
| **CommitLog threshold** | Total CommitLog size before flush all MemTables | 8 GB |
| **Cleanup threshold** | % of total memory before aggressive cleanup | 0.5 (50%) |

### Important Relationships

```
Write Request
    ↓
CommitLog (disk) ← Durability guarantee
    ↓
MemTable (memory) ← Performance optimization
    ↓ (when full)
SSTable (disk) ← Permanent storage
    ↓
Compaction ← Merge multiple SSTables
```

**Dependency Chain:**
- **MemTable depends on CommitLog** for crash recovery
- **SSTable depends on MemTable** for data source
- **Read path checks MemTable first** for latest data

## Usage Examples

### Example 1: High-Frequency Writes (Time-Series Data)

```java
// Writing sensor data at high frequency
PreparedStatement stmt = session.prepare(
    "INSERT INTO sensor_data (sensor_id, timestamp, temperature, humidity) " +
    "VALUES (?, ?, ?, ?)"
);

// These writes accumulate in MemTable
for (int i = 0; i < 100000; i++) {
    session.execute(stmt.bind(
        "sensor-123",
        Instant.now(),
        25.5 + Math.random(),
        60.0 + Math.random()
    ));
}

// Behind the scenes:
// - All 100K writes go to MemTable in memory (fast!)
// - No disk I/O per write
// - When MemTable reaches ~64-128MB → flush to SSTable
// - Total flushes might be 5-10 instead of 100K disk writes
```

**Expected behavior:**
- Write latency: < 1ms per operation
- MemTable flushes: Every ~10-20K writes (depending on size)
- CommitLog ensures durability even before flush

### Example 2: Reading Recent Writes

```java
// Write data
session.execute(
    "INSERT INTO users (id, name, last_login) VALUES (?, ?, ?)",
    UUID.fromString("123e4567-e89b-12d3-a456-426614174000"),
    "Alice",
    Instant.now()
);

// Immediately read back (read-your-write consistency)
ResultSet rs = session.execute(
    "SELECT * FROM users WHERE id = ?",
    UUID.fromString("123e4567-e89b-12d3-a456-426614174000")
);

Row row = rs.one();
System.out.println(row.getString("name")); // Output: Alice

// Read path checks:
// 1. MemTable first ← Found! (most recent data)
// 2. Skip SSTable checks since found in MemTable
// Result: Fast read, no disk I/O needed
```

**Expected behavior:**
- Read finds data in MemTable (memory) - microsecond latency
- No SSTable or disk access needed for recent writes
- Guaranteed to see your own writes immediately

### Example 3: Handling MemTable Flush

```java
// Monitor and control flushes programmatically

// Force flush for a specific table (rare, for maintenance)
session.execute("NODETOOL flush keyspace_name table_name");

// Query system metrics
ResultSet flushes = session.execute(
    "SELECT * FROM system_views.sstable_tasks " +
    "WHERE keyspace_name = 'myapp' AND kind = 'flush'"
);

for (Row flush : flushes) {
    System.out.println("Flushing: " + flush.getString("table_name"));
    System.out.println("Progress: " + flush.getDouble("progress"));
}

// In production, flushes happen automatically:
// - Size threshold: MemTable reaches 64MB (configurable)
// - CommitLog threshold: Total CommitLog > 8GB
// - Time threshold: Max age of data in MemTable
// - Shutdown: All MemTables flushed on graceful shutdown
```

**Expected behavior:**
- Automatic flushes every few minutes under write load
- Brief pause in writes to that table during flush (milliseconds)
- New MemTable immediately available for writes
- Old data now persisted in SSTable

## Key Points to Remember

### Important Gotchas

⚠️ **MemTable is per-table** - Each table has its own MemTable. High write load on one table doesn't affect others.

⚠️ **MemTable is NOT persistent** - Data only durable after CommitLog write. MemTable alone won't survive crashes.

⚠️ **Size limits are critical** - Too large MemTable = memory pressure, too small = excessive flushes and SSTables.

⚠️ **Read-your-write only per session** - Different sessions might see slightly stale data depending on consistency level.

### Performance Considerations

**Write Performance:**
- ✅ **Fast:** O(log n) insertion into sorted tree
- ✅ **No disk I/O** during write path
- ✅ **Batching:** Many writes buffered before single SSTable flush
- ⚠️ **Memory bound:** Available memory limits write throughput

**Read Performance:**
- ✅ **Recent data fast:** Latest writes in memory
- ⚠️ **Fragmentation:** Must check MemTable + multiple SSTables for older data
- ⚠️ **Compaction needed:** More SSTables = slower reads

**Memory Considerations:**
```
Total Memory = MemTables + Block Cache + JVM Heap + OS Cache
```
- Cassandra: ~25% heap for MemTables
- ScyllaDB: Auto-tunes based on available memory (uses C++ allocation)

### Common Mistakes to Avoid

❌ **Mistake 1: Ignoring heap pressure**
```java
// BAD: Large writes without monitoring heap
for (int i = 0; i < 10_000_000; i++) {
    session.execute("INSERT INTO large_table ...", largeBlob);
}
// Can cause heap pressure and trigger emergency flushes
```

✅ **Good: Rate limiting and monitoring**
```java
// Monitor heap usage
double heapUsed = ManagementFactory.getMemoryMXBean()
    .getHeapMemoryUsage().getUsed() / (1024.0 * 1024 * 1024);

if (heapUsed > 8.0) { // > 8GB
    Thread.sleep(100); // Backpressure
}
```

❌ **Mistake 2: Relying only on MemTable for durability**
- MemTable is volatile; CommitLog provides durability
- Never disable CommitLog in production

❌ **Mistake 3: Ignoring flush metrics**
- Too many flushes = too many SSTables = slow reads
- Monitor `PendingFlushes` and `CompactionBytesWritten` metrics

❌ **Mistake 4: Assuming instant flush**
- Flush is asynchronous
- Data might be in MemTable for seconds/minutes before SSTable

### Edge Cases

**1. Crash Recovery**
```
Node crashes with data in MemTable (not flushed)
    ↓
On restart: Replay CommitLog
    ↓
Rebuild MemTable from CommitLog
    ↓
No data loss!
```

**2. Concurrent Reads During Flush**
- Reads still work during flush
- Old MemTable marked read-only during flush
- New MemTable accepts new writes immediately

**3. Multiple MemTable Versions**
- Update to same partition key → new version in MemTable
- Old version remains in SSTable
- Read merges all versions by timestamp (last-write-wins)

## Quick Reference

### Most Important Operations

| Operation | Behavior | Performance |
|-----------|----------|-------------|
| **Write** | Insert into sorted tree + CommitLog | O(log n), < 1ms |
| **Read** | Check MemTable first, then SSTables | O(log n) for MemTable |
| **Update** | New version added (old in SSTable) | Same as write |
| **Delete** | Tombstone marker written | Same as write |
| **Flush** | MemTable → SSTable (immutable) | Seconds, async |

### Configuration Tuning

```yaml
# For write-heavy workloads
memtable_heap_space_in_mb: 4096        # Larger buffer
memtable_flush_writers: 4              # More parallelism
commitlog_total_space_in_mb: 16384     # Allow more uncommitted data

# For read-heavy workloads
memtable_heap_space_in_mb: 1024        # Smaller, flush faster
file_cache_size_in_mb: 8192            # More disk cache
```

### Monitoring Commands

```bash
# Check MemTable statistics
nodetool tablestats keyspace_name.table_name

# Force flush (rare, for maintenance)
nodetool flush keyspace_name table_name

# Monitor pending flushes
nodetool tpstats | grep MemtableFlushWriter

# Watch heap memory
nodetool info | grep "Heap Memory"
```

### Key Metrics to Monitor

```java
// Using JMX or metrics endpoint
- org.apache.cassandra.metrics.MemtablePool.Size
- org.apache.cassandra.metrics.MemtablePool.PendingFlushes  
- org.apache.cassandra.metrics.Table.MemtableOnHeapSize
- org.apache.cassandra.metrics.Table.MemtableOffHeapSize
- org.apache.cassandra.metrics.Table.MemtableColumnsCount
```

**Alert thresholds:**
- PendingFlushes > 5: Disk I/O bottleneck
- MemtableOnHeapSize > 75% configured: Memory pressure
- Flush latency > 5s: Potential compaction issues

### Architecture Decision Tree

```
Is your workload write-heavy?
    ↓ Yes
    ✅ MemTable shines here
        - Batches disk writes
        - Reduces I/O amplification
        - Enables high throughput
    
    ↓ No (read-heavy)
    ⚠️ Consider:
        - Smaller MemTables (faster flush)
        - More aggressive compaction
        - Larger block cache
        - Read replicas
```

## Related Topics

### Within LSM Tree Architecture
- **CommitLog**: Durability layer that works with MemTable ([db/db-wal.md](../db-wal.md))
- **SSTable**: Immutable on-disk format that MemTable flushes to
- **Compaction**: Merges multiple SSTables to reduce read amplification
- **LSM Tree**: Overall architecture ([week6/0-lsm-tree.md](../../week6/0-lsm-tree.md))

### Cassandra/ScyllaDB Concepts
- **Write Path**: CommitLog → MemTable → SSTable ([db/nosql/cassandara.md](./cassandara.md#writes-in-cassandra))
- **Read Path**: MemTable → Bloom Filters → SSTables ([db/nosql/cassandara.md](./cassandara.md#reads-in-cassandra))
- **Bloom Filters**: Skip unnecessary SSTable reads
- **Compaction Strategies**: STCS, LCS, TWCS

### Performance Related
- **Write Amplification**: MemTable helps reduce by batching
- **Read Amplification**: More SSTables = more reads (compaction helps)
- **Space Amplification**: Old versions in SSTables until compaction

### Further Reading
- [Cassandra Architecture](./cassandara.md)
- [Database Storage Engines](../db-storage-engines.md)
- [Write-Ahead Log (WAL/CommitLog)](../db-wal.md)
- [Database Indexes](../db-indexes.md)

---

## Summary Diagram

```
WRITE PATH
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ INSERT
       ↓
┌─────────────────┐
│   Coordinator   │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ↓         ↓
┌─────────┐ ┌──────────┐
│CommitLog│ │ MemTable │ ← In-Memory, Sorted
│(disk)   │ │(memory)  │ ← Fast writes O(log n)
└─────────┘ └────┬─────┘ ← Auto-flush on threshold
                 │
        When Full/Threshold
                 ↓
           ┌──────────┐
           │ SSTable  │ ← Immutable on disk
           │(disk)    │ ← Sequential write
           └────┬─────┘
                │
         Periodic Compaction
                ↓
       ┌──────────────────┐
       │ Merged SSTables  │ ← Fewer, optimized files
       └──────────────────┘

READ PATH (Recent Data)
┌────────┐
│ Client │
└───┬────┘
    │ SELECT
    ↓
┌────────────┐
│ MemTable   │ ← Check here FIRST
│(in memory) │ ← Most recent data
└────┬───────┘ ← O(log n) lookup
     │
     ├── Found? → Return (FAST! < 1ms)
     │
     └── Not Found? → Check SSTables (slower)
```

**Key Insight:** MemTable is the secret sauce for LSM-tree write performance. It transforms random writes into sequential batch writes, making Cassandra and ScyllaDB incredibly write-efficient.

