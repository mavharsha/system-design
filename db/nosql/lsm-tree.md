# LSM Tree (Log-Structured Merge Tree)

## Quick Summary

**What it does:** LSM Tree is a write-optimized data structure that buffers writes in memory and periodically flushes them to disk in sorted, immutable files, then merges these files in the background to maintain read performance.

**Primary use case:** Powers modern NoSQL databases that need extremely high write throughput (millions of writes/sec) while maintaining acceptable read performance. Ideal for write-heavy workloads like time-series data, logging, messaging systems, and event streams.

**Key features:**
- **Write optimization**: O(1) append-only writes with in-memory buffering
- **Immutable storage**: All disk files are immutable (no in-place updates)
- **Background compaction**: Automatically merges files to reduce read amplification
- **Sequential I/O**: Converts random writes into sequential disk writes

## Setup (if applicable)

### Popular Databases Using LSM Trees

| Database | Language | Primary Use Case |
|----------|----------|------------------|
| **Cassandra** | Java | Wide-column, distributed |
| **ScyllaDB** | C++ | High-performance Cassandra clone |
| **RocksDB** | C++ | Embedded key-value store |
| **LevelDB** | C++ | Embedded key-value store |
| **HBase** | Java | Hadoop-based wide-column |
| **BigTable** | C++ | Google's distributed database |

### Quick Start - RocksDB Embedded Example

```java
import org.rocksdb.*;

public class LSMExample {
    public static void main(String[] args) {
        // RocksDB is an embeddable LSM-based key-value store
        RocksDB.loadLibrary();
        
        try (Options options = new Options().setCreateIfMissing(true);
             RocksDB db = RocksDB.open(options, "/tmp/rocksdb_example")) {
            
            // Write operations (goes to MemTable first)
            db.put("user:1001".getBytes(), "Alice".getBytes());
            db.put("user:1002".getBytes(), "Bob".getBytes());
            db.put("user:1003".getBytes(), "Charlie".getBytes());
            
            // Read operation (checks MemTable, then SSTables)
            byte[] value = db.get("user:1001".getBytes());
            System.out.println(new String(value)); // Output: Alice
            
            // Behind the scenes:
            // - Writes accumulated in MemTable (in-memory)
            // - When MemTable full → flush to SSTable (Level 0)
            // - Background compaction merges SSTables
            
        } catch (RocksDBException e) {
            e.printStackTrace();
        }
    }
}
```

### Quick Start - Cassandra CQL

```java
import com.datastax.driver.core.*;

public class CassandraLSMExample {
    public static void main(String[] args) {
        Cluster cluster = Cluster.builder()
            .addContactPoint("127.0.0.1")
            .build();
        Session session = cluster.connect();
        
        // Create keyspace and table
        session.execute(
            "CREATE KEYSPACE IF NOT EXISTS app " +
            "WITH replication = {'class':'SimpleStrategy', 'replication_factor':1}"
        );
        
        session.execute(
            "CREATE TABLE IF NOT EXISTS app.events (" +
            "    event_id uuid PRIMARY KEY," +
            "    timestamp timestamp," +
            "    event_type text," +
            "    payload text" +
            ")"
        );
        
        // High-volume writes (LSM shines here!)
        PreparedStatement stmt = session.prepare(
            "INSERT INTO app.events (event_id, timestamp, event_type, payload) " +
            "VALUES (?, ?, ?, ?)"
        );
        
        for (int i = 0; i < 100000; i++) {
            session.execute(stmt.bind(
                java.util.UUID.randomUUID(),
                new java.util.Date(),
                "click",
                "{\"user_id\": " + i + "}"
            ));
        }
        
        // LSM tree handles this efficiently:
        // - All writes to MemTable (memory)
        // - Periodic flush to SSTables (disk)
        // - Background compaction keeps reads fast
        
        session.close();
        cluster.close();
    }
}
```

## Core Concepts

### Main Components

**1. MemTable (In-Memory Buffer)**
- Sorted tree structure in memory (Red-Black Tree, Skip List, or B-tree)
- Accumulates recent writes
- Provides fast O(log n) writes and reads
- Size-limited (typically 64-512 MB)

**2. WAL/CommitLog (Durability Layer)**
- Append-only log on disk
- Ensures durability before MemTable write
- Used for crash recovery
- Can be replayed to rebuild MemTable

**3. SSTable (Sorted String Table)**
- Immutable sorted file on disk
- Written sequentially when MemTable flushes
- Contains sorted key-value pairs
- Includes index and bloom filter for fast lookups

**4. Compaction (Background Merge)**
- Merges multiple SSTables into fewer, larger ones
- Removes deleted/obsolete data (garbage collection)
- Reduces read amplification
- Multiple strategies: Size-Tiered, Leveled, Time-Window

### The LSM Tree Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    WRITE PATH                            │
└─────────────────────────────────────────────────────────┘

Write Request
     ↓
┌────────────┐
│ CommitLog  │ ← 1. Append to WAL (durability)
│ (disk)     │    Sequential write, O(1)
└────────────┘
     ↓
┌────────────┐
│  MemTable  │ ← 2. Insert into sorted tree
│ (memory)   │    O(log n), fast
└──────┬─────┘
       │ When full (size/time threshold)
       ↓
┌────────────┐
│  Flush     │ ← 3. Write MemTable → SSTable
│  (async)   │    Sequential disk write
└──────┬─────┘
       ↓
┌────────────┐
│ SSTable L0 │ ← 4. Immutable sorted file
│ (disk)     │    New file, no overwrites
└──────┬─────┘
       │
       │ Background compaction
       ↓
┌──────────────────────┐
│ SSTables L1...Ln     │ ← 5. Compacted levels
│ (disk, sorted)       │    Fewer, larger files
└──────────────────────┘


┌─────────────────────────────────────────────────────────┐
│                    READ PATH                             │
└─────────────────────────────────────────────────────────┘

Read Request
     ↓
┌────────────┐
│  MemTable  │ ← 1. Check memory first
│ (memory)   │    O(log n), < 1ms
└──────┬─────┘
       │ Not found?
       ↓
┌────────────┐
│ Bloom      │ ← 2. Check bloom filters
│ Filters    │    Quickly skip SSTables
└──────┬─────┘
       │ Might exist?
       ↓
┌────────────┐
│  SSTable   │ ← 3. Binary search in SSTables
│  Index     │    Find data blocks
└──────┬─────┘
       ↓
┌────────────┐
│  SSTable   │ ← 4. Read data blocks
│  Data      │    Merge results from all levels
└──────┬─────┘
       ↓
Return merged result (latest timestamp wins)
```

### Key Parameters and Trade-offs

| Parameter | Effect | Trade-off |
|-----------|--------|-----------|
| **MemTable Size** | Larger = fewer flushes | Larger = more memory, longer recovery time |
| **Compaction Frequency** | More frequent = better reads | More frequent = higher I/O and CPU cost |
| **Level Count** | More levels = less space | More levels = slower reads (more files to check) |
| **Bloom Filter Size** | Larger = fewer false positives | Larger = more memory |

### Important Relationships

**Write Amplification:**
```
Write Amplification = Total Bytes Written to Disk / Bytes Written by User

Example with 3 compaction levels:
User writes: 100 MB
- MemTable flush: 100 MB written (Level 0)
- L0 → L1 compaction: 100 MB written
- L1 → L2 compaction: 100 MB written
Total: 300 MB → Write amplification = 3x
```

**Read Amplification:**
```
Read Amplification = Number of SSTables Checked per Read

Without compaction: 100 SSTables = 100 files to check
After compaction: 10 SSTables = 10 files to check
Bloom filters reduce false positives by ~99%
```

## Usage Examples

### Example 1: High-Throughput Time-Series Data

```java
import org.rocksdb.*;

public class TimeSeriesLSM {
    private RocksDB db;
    
    public void initialize() throws RocksDBException {
        Options options = new Options()
            .setCreateIfMissing(true)
            .setWriteBufferSize(128 * 1024 * 1024)  // 128MB MemTable
            .setMaxWriteBufferNumber(3)              // 3 MemTables
            .setMinWriteBufferNumberToMerge(2)       // Merge when 2 full
            .setCompressionType(CompressionType.LZ4_COMPRESSION);
        
        db = RocksDB.open(options, "/data/timeseries");
    }
    
    public void ingestMetrics() throws RocksDBException {
        // Simulate IoT sensor data ingestion
        WriteBatch batch = new WriteBatch();
        
        for (int sensor = 0; sensor < 10000; sensor++) {
            for (int i = 0; i < 1000; i++) {
                String key = String.format("sensor:%d:ts:%d", 
                    sensor, System.currentTimeMillis());
                String value = String.format("{\"temp\":%.2f,\"humidity\":%.2f}", 
                    20.0 + Math.random() * 10, 
                    40.0 + Math.random() * 20);
                
                batch.put(key.getBytes(), value.getBytes());
            }
        }
        
        // Batch write (10M operations)
        WriteOptions writeOpts = new WriteOptions()
            .setDisableWAL(false)  // Keep WAL enabled for durability
            .setSync(false);       // Async for performance
        
        db.write(writeOpts, batch);
        
        // Behind the scenes:
        // - All writes buffered in MemTable
        // - WAL ensures durability
        // - Periodic flush to SSTables (every ~128MB)
        // - Background compaction merges SSTables
        // - Achieves 100K+ writes/sec on commodity hardware
    }
}
```

**Expected behavior:**
- Write throughput: 100K-500K ops/sec
- Write latency: < 1ms (in-memory MemTable)
- Disk writes: Sequential, batched (efficient)
- Space: Old data compacted away periodically

### Example 2: Read-Modify-Write Pattern

```java
public class CounterService {
    private RocksDB db;
    
    public void incrementCounter(String counterId) throws RocksDBException {
        // Read current value
        byte[] key = counterId.getBytes();
        byte[] valueBytes = db.get(key);
        
        long currentValue = 0;
        if (valueBytes != null) {
            currentValue = Long.parseLong(new String(valueBytes));
        }
        
        // Increment and write back
        long newValue = currentValue + 1;
        db.put(key, String.valueOf(newValue).getBytes());
        
        // LSM handling:
        // - Read checks: MemTable → SSTables (latest value)
        // - Write appends new version (doesn't update in-place)
        // - Old versions remain until compaction
        // - Compaction merges and keeps latest version only
    }
    
    public long getCounter(String counterId) throws RocksDBException {
        byte[] value = db.get(counterId.getBytes());
        
        // Read path:
        // 1. Check MemTable (if recently written)
        // 2. Check bloom filters (skip SSTables that don't have key)
        // 3. Binary search in SSTable indexes
        // 4. Read data block from disk
        // 5. Return latest version
        
        return value != null ? Long.parseLong(new String(value)) : 0;
    }
}
```

**Expected behavior:**
- Reads may hit multiple SSTables (read amplification)
- Most recent write always in MemTable (fast)
- Compaction reduces number of SSTables over time
- Bloom filters prevent unnecessary disk reads (~99% effective)

### Example 3: Configuring Compaction Strategies

```java
public class CompactionTuning {
    
    // Size-Tiered Compaction (Default - Write-Heavy)
    public RocksDB openSizeTiered() throws RocksDBException {
        Options options = new Options()
            .setCreateIfMissing(true)
            .setCompactionStyle(CompactionStyle.LEVEL)  // Leveled compaction
            .setNumLevels(7)
            .setLevelCompactionDynamicLevelBytes(true)
            .setMaxBytesForLevelBase(512 * 1024 * 1024)  // 512MB
            .setMaxBytesForLevelMultiplier(10);           // Each level 10x larger
        
        return RocksDB.open(options, "/data/write-heavy");
        
        // Behavior:
        // - Fast writes (fewer compactions)
        // - Slower reads (more SSTables in L0)
        // - Higher space amplification
        // - Best for: Logging, events, time-series
    }
    
    // Leveled Compaction (Read-Optimized)
    public RocksDB openLeveled() throws RocksDBException {
        Options options = new Options()
            .setCreateIfMissing(true)
            .setCompactionStyle(CompactionStyle.LEVEL)
            .setNumLevels(7)
            .setLevelCompactionDynamicLevelBytes(false)
            .setTargetFileSizeBase(64 * 1024 * 1024)     // 64MB files
            .setMaxBytesForLevelBase(256 * 1024 * 1024)  // 256MB
            .setMaxBytesForLevelMultiplier(10);
        
        return RocksDB.open(options, "/data/read-heavy");
        
        // Behavior:
        // - More frequent compaction
        // - Fewer SSTables per level (faster reads)
        // - Higher write amplification
        // - Best for: User profiles, product catalogs, caches
    }
    
    // Universal Compaction (Balanced)
    public RocksDB openUniversal() throws RocksDBException {
        Options options = new Options()
            .setCreateIfMissing(true)
            .setCompactionStyle(CompactionStyle.UNIVERSAL)
            .setMaxSizeAmplificationPercent(200)  // Max space overhead: 2x
            .setCompressionType(CompressionType.LZ4_COMPRESSION);
        
        return RocksDB.open(options, "/data/balanced");
        
        // Behavior:
        // - Compacts when total size exceeds threshold
        // - Good balance of read/write performance
        // - Lower space amplification
        // - Best for: General-purpose workloads
    }
}
```

**Performance comparison:**

| Strategy | Write Speed | Read Speed | Space Usage | Best For |
|----------|-------------|------------|-------------|----------|
| Size-Tiered | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ | Logs, events |
| Leveled | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | User data |
| Universal | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | General-purpose |

## Key Points to Remember

### Important Gotchas

⚠️ **Write amplification is real** - Data gets rewritten multiple times during compaction. A 1GB write might result in 3-10GB of actual disk writes.

⚠️ **Read amplification without compaction** - Without proper compaction, reads must check dozens of SSTables, killing performance.

⚠️ **Tombstones aren't immediate deletes** - Deletes write tombstone markers. Data isn't removed until compaction, wasting space temporarily.

⚠️ **Bloom filters have false positives** - A bloom filter might say "key exists" when it doesn't. Always verify with actual disk read.

⚠️ **Range scans are slower than point reads** - LSM trees optimize for point reads/writes. Range scans must merge data from multiple SSTables.

### Performance Considerations

**Why LSM Trees Are Fast for Writes:**
```
Traditional B-Tree:
Write → Search tree → Update in-place → Random disk I/O
100 writes = 100 random disk seeks = SLOW (10-100ms each)

LSM Tree:
Write → Append to WAL → Insert in MemTable → Done
100 writes = 2 sequential writes (WAL + eventual flush) = FAST (< 1ms each)
```

**Write Performance:**
- ✅ **Sequential writes:** All disk writes are sequential (fast on HDD and SSD)
- ✅ **Batching:** Multiple writes flushed together
- ✅ **In-memory:** Writes served from RAM
- ⚠️ **Compaction overhead:** Background I/O can impact performance

**Read Performance:**
- ✅ **Recent data fast:** MemTable serves hot data from memory
- ✅ **Bloom filters:** Skip 99% of unnecessary SSTable reads
- ✅ **Caching:** Block cache speeds up repeated reads
- ⚠️ **Cold reads slow:** Must check multiple SSTables
- ⚠️ **Range scans expensive:** Must merge multiple sorted files

**Space Efficiency:**
```
Space Amplification = Total Disk Space / Logical Data Size

Factors affecting space:
1. Multiple versions of same key (until compaction)
2. Tombstones (deleted data marked, not removed)
3. Fragmentation across SSTables
4. Compression effectiveness

Typical space amplification: 1.5x - 3x
```

### Common Mistakes to Avoid

❌ **Mistake 1: Disabling WAL for "performance"**
```java
// BAD: No durability!
WriteOptions opts = new WriteOptions().setDisableWAL(true);
db.put(opts, key, value);
// If crash happens → data loss
```

✅ **Good: Async WAL sync for throughput + durability**
```java
WriteOptions opts = new WriteOptions()
    .setDisableWAL(false)  // WAL enabled
    .setSync(false);       // Async fsync (OS buffers)
db.put(opts, key, value);

// Periodically sync to disk
db.syncWal();  // Explicit sync every N seconds
```

❌ **Mistake 2: Ignoring compaction lag**
```java
// Writing faster than compaction can keep up
for (int i = 0; i < 100_000_000; i++) {
    db.put(("key" + i).getBytes(), value);
}
// Result: Hundreds of SSTables, read performance dies
```

✅ **Good: Monitor and throttle**
```java
// Check pending compaction bytes
long pendingBytes = db.getLongProperty("rocksdb.estimate-pending-compaction-bytes");
if (pendingBytes > 10_000_000_000L) {  // > 10GB
    Thread.sleep(100);  // Backpressure
}
```

❌ **Mistake 3: Small MemTables with high write volume**
```java
// BAD: MemTable too small for write load
Options options = new Options()
    .setWriteBufferSize(4 * 1024 * 1024);  // Only 4MB!
// Result: Constant flushing, too many L0 files
```

✅ **Good: Size MemTable appropriately**
```java
Options options = new Options()
    .setWriteBufferSize(128 * 1024 * 1024)  // 128MB
    .setMaxWriteBufferNumber(4)              // Allow 4 MemTables
    .setMinWriteBufferNumberToMerge(2);      // Merge 2 before flush
```

❌ **Mistake 4: Not using bloom filters**
```java
// BAD: Every read checks every SSTable
Options options = new Options()
    .setCreateIfMissing(true);
// No bloom filters = slow negative lookups
```

✅ **Good: Enable bloom filters**
```java
BlockBasedTableConfig tableConfig = new BlockBasedTableConfig()
    .setFilterPolicy(new BloomFilter(10, false));  // 10 bits per key
Options options = new Options()
    .setCreateIfMissing(true)
    .setTableFormatConfig(tableConfig);
// 99% reduction in false positive reads
```

### Edge Cases

**1. Hot Key Problem**
```
Same key updated frequently:
key:user:123 → value1
key:user:123 → value2
key:user:123 → value3
...
key:user:123 → value1000

Without compaction:
- 1000 versions across SSTables
- Read must merge all versions
- Space wasted on old versions

With compaction:
- Merges to single latest version
- Fast reads
- Space reclaimed
```

**2. Time-Series Data Expiration**
```java
// Time-based compaction for old data
Options options = new Options()
    .setCompactionStyle(CompactionStyle.FIFO)
    .setMaxTableFilesSizeFIFO(1024 * 1024 * 1024)  // 1GB total
    .setTtl(86400);  // 24 hours TTL

// Automatically drops old SSTables
// Perfect for logs, metrics, events
```

**3. Crash Recovery**
```
Node crashes mid-write:
1. MemTable lost (volatile)
2. Unflushed data lost... OR IS IT?

Recovery process:
1. Read CommitLog/WAL from disk
2. Replay all operations
3. Rebuild MemTable
4. Resume normal operation

Result: Zero data loss (if WAL enabled)
```

## Quick Reference

### When to Use LSM Trees

✅ **Perfect for:**
- High write throughput (logs, events, metrics, time-series)
- Append-heavy workloads (messaging, streaming)
- Write-to-read ratio > 1:1
- Sequential access patterns
- Cloud storage (S3, etc.) - sequential writes are efficient

❌ **Not ideal for:**
- Read-heavy workloads (consider B-trees)
- Frequent updates to same keys (read-modify-write)
- Range scans as primary operation
- Strong consistency with immediate reads
- Storage-constrained environments (space amplification)

### Key Metrics to Monitor

```java
// RocksDB monitoring
public class LSMMetrics {
    public void printStats(RocksDB db) throws RocksDBException {
        // Write metrics
        long memtableSize = db.getLongProperty("rocksdb.cur-size-all-mem-tables");
        long pendingFlushes = db.getLongProperty("rocksdb.mem-table-flush-pending");
        
        // Compaction metrics
        long pendingCompaction = db.getLongProperty("rocksdb.estimate-pending-compaction-bytes");
        long numRunningCompactions = db.getLongProperty("rocksdb.num-running-compactions");
        
        // Read metrics
        long numFilesLevel0 = db.getLongProperty("rocksdb.num-files-at-level0");
        long blockCacheUsage = db.getLongProperty("rocksdb.block-cache-usage");
        
        // Performance indicators
        System.out.println("MemTable Size: " + memtableSize / (1024*1024) + " MB");
        System.out.println("Pending Compaction: " + pendingCompaction / (1024*1024) + " MB");
        System.out.println("L0 Files: " + numFilesLevel0);  // Alert if > 10
        
        // Alert conditions:
        if (numFilesLevel0 > 20) {
            System.err.println("WARNING: Too many L0 files! Writes will stall!");
        }
        if (pendingCompaction > 50 * 1024 * 1024 * 1024L) {  // > 50GB
            System.err.println("WARNING: Compaction falling behind!");
        }
    }
}
```

**Alert Thresholds:**
- L0 files > 10: Compaction struggling
- L0 files > 20: Write stalls imminent
- Pending compaction > 50GB: Compaction far behind
- MemTable flush pending > 2: Memory pressure

### Compaction Strategies Quick Guide

```
┌─────────────────────────────────────────────────────────┐
│         SIZE-TIERED COMPACTION (STCS)                   │
├─────────────────────────────────────────────────────────┤
│  L0: [10MB][10MB][10MB][10MB]                          │
│       ↓ Compact when 4 similar-sized files              │
│  L1: [────────── 40MB ──────────]                      │
│       ↓ Compact when 4 similar-sized files              │
│  L2: [────────────── 160MB ──────────────]             │
│                                                          │
│  ✅ Fast writes (fewer compactions)                     │
│  ❌ Slower reads (more files per level)                 │
│  ✅ Good for: Write-heavy, time-series, logs            │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│         LEVELED COMPACTION (LCS)                        │
├─────────────────────────────────────────────────────────┤
│  L0: [10MB][10MB][10MB][10MB]                          │
│       ↓ Flush to L1                                     │
│  L1: [10MB][10MB][10MB][10MB]  (non-overlapping)       │
│       ↓ Compact to L2                                   │
│  L2: [10MB][10MB]...[10MB]  (10x more files)           │
│       ↓ Compact to L3                                   │
│  L3: [10MB]...[many 10MB files]  (100x more)           │
│                                                          │
│  ❌ Slower writes (more compactions)                    │
│  ✅ Faster reads (fewer files to check per level)       │
│  ✅ Good for: Read-heavy, user data, point reads        │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│         UNIVERSAL COMPACTION                             │
├─────────────────────────────────────────────────────────┤
│  [50MB][40MB][30MB][20MB][10MB]                        │
│   ↓ When total exceeds threshold, compact all           │
│  [────────── 150MB ──────────]                         │
│                                                          │
│  ✅ Balanced reads/writes                               │
│  ✅ Lower space amplification                           │
│  ✅ Good for: General-purpose, balanced workloads       │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│         TIME-WINDOW COMPACTION (TWCS)                   │
├─────────────────────────────────────────────────────────┤
│  Window 1 (Hour 1): [10MB][10MB] → Compact             │
│  Window 2 (Hour 2): [10MB][10MB] → Compact             │
│  Window 3 (Hour 3): [10MB][10MB] → Keep writing        │
│       ↓ After TTL                                       │
│  Window 1: DELETE entire window                         │
│                                                          │
│  ✅ Perfect for time-series with TTL                    │
│  ✅ Fast deletes (drop whole files)                     │
│  ✅ Good for: Metrics, logs, events with expiration     │
└─────────────────────────────────────────────────────────┘
```

### Configuration Cheat Sheet

```java
// Write-Heavy Configuration (Maximum throughput)
Options writeHeavy = new Options()
    .setWriteBufferSize(256 * 1024 * 1024)       // Large MemTable: 256MB
    .setMaxWriteBufferNumber(6)                   // Many MemTables
    .setMinWriteBufferNumberToMerge(2)
    .setLevel0FileNumCompactionTrigger(10)        // Delay compaction
    .setMaxBackgroundCompactions(4)               // More compaction threads
    .setCompactionStyle(CompactionStyle.UNIVERSAL);

// Read-Heavy Configuration (Minimum latency)
Options readHeavy = new Options()
    .setWriteBufferSize(64 * 1024 * 1024)        // Smaller MemTable: 64MB
    .setLevel0FileNumCompactionTrigger(2)         // Aggressive compaction
    .setMaxBackgroundCompactions(8)               // Many compaction threads
    .setCompactionStyle(CompactionStyle.LEVEL)
    .setTableFormatConfig(
        new BlockBasedTableConfig()
            .setBlockSize(16 * 1024)              // Smaller blocks
            .setFilterPolicy(new BloomFilter(10)) // Bloom filters
            .setBlockCache(new LRUCache(2 * 1024 * 1024 * 1024L))  // 2GB cache
    );

// Balanced Configuration (General purpose)
Options balanced = new Options()
    .setWriteBufferSize(128 * 1024 * 1024)       // 128MB MemTable
    .setMaxWriteBufferNumber(4)
    .setLevel0FileNumCompactionTrigger(4)
    .setMaxBackgroundCompactions(4)
    .setCompressionType(CompressionType.LZ4_COMPRESSION);
```

## Related Topics

### Storage Engine Comparisons

**LSM Tree vs B-Tree:**

| Aspect | LSM Tree | B-Tree |
|--------|----------|--------|
| **Write Performance** | ⭐⭐⭐⭐⭐ Sequential | ⭐⭐⭐ Random |
| **Read Performance** | ⭐⭐⭐ Multiple files | ⭐⭐⭐⭐⭐ Single tree |
| **Space Efficiency** | ⭐⭐⭐ Amplification | ⭐⭐⭐⭐ In-place |
| **Write Amplification** | ⭐⭐ High (3-10x) | ⭐⭐⭐⭐ Low (1-2x) |
| **Compaction** | Required | Not needed |

**Databases using B-trees:** MySQL (InnoDB), PostgreSQL, MongoDB

**Databases using LSM trees:** Cassandra, RocksDB, LevelDB, HBase, ScyllaDB

### Related Concepts

- **MemTable**: In-memory component of LSM tree ([memtable.md](./memtable.md))
- **SSTable**: Sorted String Table on disk ([db/nosql/cassandara.md](./cassandara.md))
- **Write-Ahead Log (WAL)**: Durability mechanism ([db/db-wal.md](../db-wal.md))
- **Bloom Filters**: Probabilistic data structure for fast negative lookups
- **Compaction**: Background merge process ([db/db-storage-engines.md](../db-storage-engines.md))

### Databases Deep Dives

- **Cassandra LSM Implementation**: [cassandara.md](./cassandara.md)
- **RocksDB**: Facebook's embedded LSM database
- **ScyllaDB**: High-performance Cassandra alternative
- **Apache HBase**: Hadoop-based LSM database

### Further Reading

- [Database Storage Engines](../db-storage-engines.md)
- [Database Indexes](../db-indexes.md)
- [Write-Ahead Log](../db-wal.md)
- [Cache Design](../../cache/0-designing-single-node-cache%20copy.md)

---

## Summary

**LSM trees trade write performance for read complexity:**

✅ **Strengths:**
- Extremely fast writes (100K-1M ops/sec)
- Efficient sequential I/O (SSD and HDD friendly)
- Excellent for write-heavy workloads
- Scales well with cloud storage

⚠️ **Challenges:**
- Write amplification (data rewritten 3-10x)
- Read amplification (multiple files to check)
- Space amplification (old versions kept until compaction)
- Compaction overhead (background I/O)

**Bottom Line:** If you're building a system that needs to handle millions of writes per second (logging, messaging, time-series, events), LSM trees are your best friend. If you're building a system with mostly reads and occasional writes (user profiles, product catalogs), consider B-trees instead.

**Key Insight:** LSM trees convert expensive random writes into cheap sequential writes, at the cost of more complex reads. This trade-off is perfect for modern distributed systems where writes are cheap but reads can be cached or optimized with clever data structures (bloom filters, indexes).

