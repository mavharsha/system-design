# SSTable (Sorted String Table)

## Quick Summary

**What it does:** SSTable is an immutable, sorted, on-disk file format that stores key-value pairs in a compact, searchable structure. It serves as the persistent storage layer in LSM-tree databases.

**Primary use case:** Powers the storage layer of write-optimized NoSQL databases (Cassandra, HBase, RocksDB) by providing efficient sequential writes and indexed reads for large-scale data storage.

**Key features:**
- **Immutable**: Once written, never modified (only created or deleted)
- **Sorted**: Keys stored in sorted order for efficient binary search
- **Indexed**: Built-in indexes and bloom filters for fast lookups
- **Compressed**: Uses compression algorithms (LZ4, Snappy, Zstd) to save space
- **Self-contained**: Each SSTable includes data, indexes, and metadata

## Setup (if applicable)

### Databases Using SSTables

| Database | Language | SSTable Format | Primary Use Case |
|----------|----------|----------------|------------------|
| **Cassandra** | Java | Custom binary | Wide-column, distributed |
| **ScyllaDB** | C++ | Cassandra-compatible | High-performance Cassandra alternative |
| **RocksDB** | C++ | .sst files | Embedded key-value store |
| **LevelDB** | C++ | .ldb files | Embedded key-value store |
| **HBase** | Java | HFile format | Hadoop-based wide-column |
| **BigTable** | C++ | SSTable format | Google's distributed database |

### Quick Start - RocksDB SSTable Inspection

```java
import org.rocksdb.*;

public class SSTableInspector {
    public static void main(String[] args) {
        RocksDB.loadLibrary();
        
        try (Options options = new Options().setCreateIfMissing(true);
             RocksDB db = RocksDB.open(options, "/tmp/rocksdb")) {
            
            // Write data (accumulates in MemTable first)
            for (int i = 0; i < 10000; i++) {
                db.put(
                    String.format("user:%05d", i).getBytes(),
                    String.format("data-%d", i).getBytes()
                );
            }
            
            // Force flush to create SSTable
            db.flush(new FlushOptions());
            
            // Inspect SSTable properties
            long numSSTables = db.getLongProperty("rocksdb.num-files-at-level0");
            long totalSSTableSize = db.getLongProperty("rocksdb.total-sst-files-size");
            
            System.out.println("Number of SSTables: " + numSSTables);
            System.out.println("Total SSTable size: " + totalSSTableSize + " bytes");
            
            // Behind the scenes:
            // - MemTable flushed to immutable SSTable
            // - SSTable created at Level 0
            // - Contains sorted keys: user:00000, user:00001, ..., user:09999
            // - Includes bloom filter and index for fast lookups
            
        } catch (RocksDBException e) {
            e.printStackTrace();
        }
    }
}
```

### Quick Start - Cassandra SSTable Tools

```bash
# View SSTable metadata
sstablemetadata /var/lib/cassandra/data/keyspace/table-*/mc-1-big-Data.db

# Dump SSTable contents (human-readable)
sstabledump /var/lib/cassandra/data/keyspace/table-*/mc-1-big-Data.db

# Verify SSTable integrity
sstableverify keyspace table

# Export SSTable to JSON
sstabledump -d /var/lib/cassandra/data/keyspace/table-*/mc-1-big-Data.db > output.json
```

## Core Concepts

### Main Components

An SSTable is not a single file but a collection of files working together:

**1. Data File (.db / .sst)**
- Contains the actual key-value pairs in sorted order
- Organized in fixed-size blocks (typically 4-64 KB)
- Compressed using LZ4, Snappy, or Zstd
- Keys stored with timestamps for versioning

**2. Index File (.index)**
- Maps keys to byte offsets in the data file
- Allows binary search within data file
- Typically stored in memory for hot SSTables
- Reduces I/O by jumping directly to data blocks

**3. Bloom Filter (.filter)**
- Probabilistic data structure for key existence checks
- Prevents unnecessary disk reads for non-existent keys
- ~99% effective at eliminating false lookups
- Small memory footprint (10 bits per key = ~1.2% overhead)

**4. Summary File (.summary)**
- Sparse index of the index (index of the index!)
- Loads into memory first for quick key range checks
- Samples every N keys from the index
- Enables fast SSTable selection during reads

**5. Statistics File (.statistics / .stats)**
- Metadata about SSTable contents
- Min/max keys and timestamps
- Compression ratio, row count, data size
- Used for compaction decisions and query planning

**6. Compression Info (.compressionInfo)**
- Chunk offsets for compressed data
- Allows random access within compressed file
- Maps logical positions to physical compressed blocks

### SSTable Internal Structure

```
┌─────────────────────────────────────────────────────────┐
│                    SSTTABLE ANATOMY                      │
└─────────────────────────────────────────────────────────┘

SUMMARY FILE (loaded first, in memory)
┌────────────────────────────────────┐
│ min_key: "user:00000"              │
│ max_key: "user:99999"              │
│ sampled_keys: ["user:00000",       │
│                "user:25000",       │
│                "user:50000",       │
│                "user:75000"]       │
│ → Points to index positions        │
└────────────────────────────────────┘
         ↓
INDEX FILE (partially in memory)
┌────────────────────────────────────┐
│ "user:00000" → offset: 0           │
│ "user:00001" → offset: 128         │
│ "user:00002" → offset: 256         │
│ ...                                │
│ "user:99999" → offset: 1024000     │
│ → Points to data blocks            │
└────────────────────────────────────┘
         ↓
DATA FILE (on disk, read on demand)
┌────────────────────────────────────┐
│ Block 0 (4KB):                     │
│   user:00000 → {name: "Alice"}     │
│   user:00001 → {name: "Bob"}       │
│   ...                              │
├────────────────────────────────────┤
│ Block 1 (4KB):                     │
│   user:00050 → {name: "Charlie"}   │
│   ...                              │
└────────────────────────────────────┘

BLOOM FILTER (in memory)
┌────────────────────────────────────┐
│ Hash("user:00000") → bit[123] = 1  │
│ Hash("user:00001") → bit[456] = 1  │
│ ...                                │
│ Quick check: "user:99999" exists?  │
│   → Yes (might exist, check disk)  │
│ Quick check: "user:zzzzz" exists?  │
│   → No (definitely doesn't exist)  │
└────────────────────────────────────┘
```

### Key Properties

| Property | Value | Trade-off |
|----------|-------|-----------|
| **Immutable** | Never modified after creation | Deletes require tombstones |
| **Sorted** | Keys in lexicographic order | Efficient range scans and binary search |
| **Compressed** | 50-80% size reduction typical | CPU cost for decompression |
| **Block-based** | 4-64 KB blocks | Balance between compression and random access |
| **Versioned** | Timestamps on all data | Multiple versions until compaction |

### Important Relationships

**SSTable Lifecycle:**
```
MemTable (memory)
    ↓ Flush when full (64-128 MB)
SSTable L0 (disk, immutable)
    ↓ Background compaction
SSTable L1...Ln (disk, merged & optimized)
    ↓ Eventually
Deleted (obsolete data removed)
```

**Read Path Interaction:**
```
Read Request for key="user:12345"
    ↓
1. Check MemTable (in memory) → Not found
    ↓
2. Check Bloom Filters on all SSTables
    SSTable-1: Bloom says "YES, might exist"
    SSTable-2: Bloom says "NO, definitely not"
    SSTable-3: Bloom says "YES, might exist"
    ↓
3. Binary search in SSTable-1 and SSTable-3 only
    SSTable-1 Index: "user:12345" → offset 5000
    ↓
4. Read data block at offset 5000
    ↓
5. Decompress block and extract value
    ↓
6. Return result (merge if found in multiple SSTables by timestamp)
```

## Usage Examples

### Example 1: Reading Data from SSTables

```java
import org.rocksdb.*;

public class SSTableReadExample {
    private RocksDB db;
    
    public void readWithSSTableLayers() throws RocksDBException {
        byte[] key = "user:12345".getBytes();
        byte[] value = db.get(key);
        
        // Behind the scenes (RocksDB handles this automatically):
        // 
        // 1. Check MemTable (active write buffer)
        //    - Not found, continue...
        //
        // 2. Check Immutable MemTables (being flushed)
        //    - Not found, continue...
        //
        // 3. Check Level 0 SSTables (newest, possibly overlapping)
        //    SSTable-005.sst: Check bloom filter → "No"
        //    SSTable-004.sst: Check bloom filter → "Yes, might exist"
        //      → Load index from disk/cache
        //      → Binary search: key found at offset 4096
        //      → Read data block (4KB)
        //      → Decompress block
        //      → Extract value
        //
        // 4. If not found, check Level 1+ SSTables
        //    (each level has non-overlapping key ranges)
        //
        // 5. Return merged result (latest timestamp wins)
        
        if (value != null) {
            System.out.println("Found: " + new String(value));
        }
    }
    
    public void demonstrateBloomFilterBenefit() throws RocksDBException {
        // Write 1 million keys
        for (int i = 0; i < 1_000_000; i++) {
            db.put(
                String.format("key:%07d", i).getBytes(),
                ("value-" + i).getBytes()
            );
        }
        db.flush(new FlushOptions());
        
        // Search for non-existent keys
        long start = System.nanoTime();
        for (int i = 2_000_000; i < 2_001_000; i++) {
            byte[] value = db.get(String.format("key:%07d", i).getBytes());
            // All return null - keys don't exist
        }
        long duration = System.nanoTime() - start;
        
        System.out.println("1000 negative lookups: " + duration / 1_000_000 + "ms");
        
        // Without bloom filter: Would need to:
        // - Load index for each SSTable
        // - Binary search index
        // - Potentially read data block
        // Total: ~1000ms for 1000 lookups
        //
        // With bloom filter:
        // - Check bloom filter (in memory, instant)
        // - Filter says "No, definitely not there"
        // - Skip SSTable entirely
        // Total: ~10ms for 1000 lookups (100x faster!)
    }
}
```

**Expected behavior:**
- Bloom filter eliminates 99% of unnecessary disk reads
- Index lookup: O(log n) binary search
- Data block read: Single disk I/O per SSTable
- Decompression: ~100 MB/s typical (LZ4)

### Example 2: SSTable Compaction Observation

```java
import org.rocksdb.*;

public class CompactionMonitor {
    public void watchCompactionCreateSSTables() throws RocksDBException {
        Options options = new Options()
            .setCreateIfMissing(true)
            .setWriteBufferSize(4 * 1024 * 1024)  // Small MemTable: 4MB
            .setLevel0FileNumCompactionTrigger(4);  // Compact when 4 L0 files
        
        RocksDB db = RocksDB.open(options, "/tmp/compaction_demo");
        
        // Write enough data to trigger multiple flushes
        for (int batch = 0; batch < 10; batch++) {
            System.out.println("\n=== Batch " + batch + " ===");
            
            // Write 1MB of data
            for (int i = 0; i < 10000; i++) {
                String key = String.format("key:%d:%05d", batch, i);
                db.put(key.getBytes(), ("value-" + i).getBytes());
            }
            
            // Check SSTable counts per level
            long l0Files = db.getLongProperty("rocksdb.num-files-at-level0");
            long l1Files = db.getLongProperty("rocksdb.num-files-at-level1");
            long l2Files = db.getLongProperty("rocksdb.num-files-at-level2");
            
            System.out.println("L0 SSTables: " + l0Files);
            System.out.println("L1 SSTables: " + l1Files);
            System.out.println("L2 SSTables: " + l2Files);
            
            // What's happening:
            // - After ~4 batches: L0 has 4 files → compaction triggered
            // - Compaction merges 4 L0 files → 1 L1 file
            // - After more batches: L1 grows → compaction to L2
            // - Result: Fewer, larger files at deeper levels
        }
        
        db.close();
    }
    
    public void observeCompactionStats() throws RocksDBException {
        RocksDB db = RocksDB.open("/tmp/compaction_demo");
        
        // Get compaction statistics
        long totalSSTableSize = db.getLongProperty("rocksdb.total-sst-files-size");
        long pendingCompaction = db.getLongProperty("rocksdb.estimate-pending-compaction-bytes");
        long numRunningCompactions = db.getLongProperty("rocksdb.num-running-compactions");
        
        System.out.println("Total SSTable size: " + totalSSTableSize / (1024*1024) + " MB");
        System.out.println("Pending compaction: " + pendingCompaction / (1024*1024) + " MB");
        System.out.println("Running compactions: " + numRunningCompactions);
        
        // Alert conditions:
        if (pendingCompaction > 10L * 1024 * 1024 * 1024) {  // > 10GB
            System.err.println("⚠️ Compaction lagging behind writes!");
        }
        
        db.close();
    }
}
```

**Expected behavior:**
- Each MemTable flush creates one L0 SSTable
- L0 files may have overlapping key ranges
- Compaction merges overlapping SSTables
- Deeper levels have non-overlapping key ranges
- Result: Efficient reads with fewer SSTables to check

### Example 3: Cassandra SSTable Management

```java
import com.datastax.driver.core.*;

public class CassandraSSTableExample {
    private Cluster cluster;
    private Session session;
    
    public void setup() {
        cluster = Cluster.builder()
            .addContactPoint("127.0.0.1")
            .build();
        session = cluster.connect();
        
        // Create keyspace and table
        session.execute(
            "CREATE KEYSPACE IF NOT EXISTS myapp " +
            "WITH replication = {'class':'SimpleStrategy', 'replication_factor':1}"
        );
        
        session.execute(
            "CREATE TABLE IF NOT EXISTS myapp.users (" +
            "    user_id uuid PRIMARY KEY," +
            "    username text," +
            "    email text," +
            "    created_at timestamp" +
            ")"
        );
    }
    
    public void writeDataAndObserveSSTables() {
        // Write a large batch of data
        PreparedStatement stmt = session.prepare(
            "INSERT INTO myapp.users (user_id, username, email, created_at) " +
            "VALUES (?, ?, ?, ?)"
        );
        
        System.out.println("Writing 100,000 users...");
        for (int i = 0; i < 100_000; i++) {
            session.execute(stmt.bind(
                java.util.UUID.randomUUID(),
                "user" + i,
                "user" + i + "@example.com",
                new java.util.Date()
            ));
        }
        
        // Behind the scenes:
        // 1. All writes go to MemTable (in memory)
        // 2. When MemTable reaches ~64MB → flush to SSTable
        // 3. Multiple SSTables created: sstable-1, sstable-2, sstable-3...
        // 4. Each SSTable includes:
        //    - Data file: sorted user records
        //    - Index file: user_id → offset mapping
        //    - Bloom filter: quick existence check
        //    - Summary: sparse index for fast seeking
        //    - Statistics: min/max keys, timestamps
        
        System.out.println("✓ Data written. SSTables created in data directory:");
        System.out.println("  /var/lib/cassandra/data/myapp/users-<uuid>/");
        System.out.println("    mc-1-big-Data.db        (main data)");
        System.out.println("    mc-1-big-Index.db       (key index)");
        System.out.println("    mc-1-big-Filter.db      (bloom filter)");
        System.out.println("    mc-1-big-Summary.db     (sparse index)");
        System.out.println("    mc-1-big-Statistics.db  (metadata)");
    }
    
    public void demonstrateReadPath() {
        java.util.UUID userId = java.util.UUID.randomUUID();
        
        // Execute a read query
        ResultSet rs = session.execute(
            "SELECT * FROM myapp.users WHERE user_id = ?",
            userId
        );
        
        Row row = rs.one();
        
        // Cassandra's read path with SSTables:
        // 
        // 1. Check MemTable first (hot data)
        //    - Not found (random UUID unlikely to exist)
        //
        // 2. For each SSTable:
        //    a) Check bloom filter (in memory)
        //       SSTable-1: "No, definitely not here" → Skip
        //       SSTable-2: "No, definitely not here" → Skip
        //       SSTable-3: "No, definitely not here" → Skip
        //    
        //    If bloom filter says "Yes":
        //    b) Check summary (in memory) for key range
        //    c) Load index from disk/cache
        //    d) Binary search index for exact offset
        //    e) Read data block from disk
        //    f) Decompress and parse
        //
        // 3. Merge results from all SSTables (by timestamp)
        //
        // 4. Return result or null
        
        if (row == null) {
            System.out.println("User not found");
            System.out.println("Thanks to bloom filters, checked 0 data blocks!");
        }
    }
    
    public void triggerCompaction() {
        // Force manual compaction (usually automatic)
        // In production, use nodetool: nodetool compact myapp users
        
        System.out.println("Triggering compaction...");
        System.out.println("\nBefore compaction:");
        System.out.println("  10 SSTables: sstable-1 to sstable-10");
        System.out.println("  Total size: 640 MB");
        System.out.println("  Key ranges: Overlapping");
        
        // Compaction process:
        // 1. Read all SSTables in parallel
        // 2. Merge-sort keys across all SSTables
        // 3. Keep latest version of each key (by timestamp)
        // 4. Remove tombstones (deleted data)
        // 5. Write new, larger SSTable(s)
        // 6. Delete old SSTables
        
        System.out.println("\nAfter compaction:");
        System.out.println("  2 SSTables: sstable-11, sstable-12");
        System.out.println("  Total size: 400 MB (duplicates removed)");
        System.out.println("  Key ranges: Non-overlapping");
        System.out.println("  Read performance: ↑ (fewer files to check)");
    }
}
```

**Expected behavior:**
- Each MemTable flush creates 1 SSTable (~64-128 MB)
- 100K users → ~2-3 SSTables created
- Bloom filters prevent 99% of unnecessary SSTable reads
- Compaction merges SSTables, removes duplicates
- Read latency: 1-5ms for point lookups

## Key Points to Remember

### Important Gotchas

⚠️ **Immutability means no updates** - Updates create new versions in new SSTables. Old versions remain until compaction removes them, wasting space temporarily.

⚠️ **Deletes aren't instant** - Deletes write tombstone markers. Data remains in SSTable until compaction physically removes it.

⚠️ **Multiple versions = slow reads** - Without compaction, a single key might exist in 10+ SSTables, requiring 10+ disk reads to merge versions.

⚠️ **Bloom filters have false positives** - A bloom filter might say "key exists" when it doesn't (~1% false positive rate), leading to unnecessary disk reads.

⚠️ **Large SSTables = slow compaction** - Multi-GB SSTables take minutes to hours to compact, during which disk I/O is high.

⚠️ **Overlapping L0 files** - Level 0 SSTables may have overlapping key ranges (since they're direct MemTable flushes). Must check all L0 files during reads.

### Performance Considerations

**Read Performance:**
```
Best case (recent data in MemTable): < 1ms, zero disk I/O
Good case (1-2 SSTables checked): 1-5ms, one disk read
Bad case (10+ SSTables checked): 50-200ms, many disk reads
Worst case (no bloom filters): seconds, all SSTables checked
```

**Write Performance:**
```
Write to MemTable: < 1ms (in-memory)
Flush to SSTable: 1-5 seconds (sequential write, 64-128 MB)
Compaction overhead: 10-50% of write throughput (background)
```

**Space Efficiency:**
```
No compression: 1.0x (baseline)
LZ4 compression: 0.3-0.5x (fast, moderate ratio)
Snappy compression: 0.3-0.5x (very fast, moderate ratio)
Zstd compression: 0.2-0.4x (slower, better ratio)

Space amplification with duplicates: 1.5-3.0x until compaction
```

**Why SSTables Enable Fast Writes:**
```
Traditional B-Tree Update:
1. Search tree for key location → Random disk read
2. Read page containing key → Random disk read
3. Modify page in place → Random disk write
Total: 2 reads + 1 write per update (SLOW on HDD)

SSTable Approach:
1. Write to MemTable (memory) → No disk I/O
2. Periodic flush: Sequential write of entire MemTable
3. No in-place updates
Total: Batch of 1000 updates → 1 sequential write (FAST)
```

### Common Mistakes to Avoid

❌ **Mistake 1: Ignoring compaction lag**
```java
// BAD: Writing faster than compaction can keep up
for (int i = 0; i < 100_000_000; i++) {
    db.put(("key" + i).getBytes(), value);
}
// Result: Hundreds of SSTables, read performance dies
```

✅ **Good: Monitor and throttle**
```java
long pendingBytes = db.getLongProperty("rocksdb.estimate-pending-compaction-bytes");
if (pendingBytes > 50L * 1024 * 1024 * 1024) {  // > 50GB
    Thread.sleep(100);  // Apply backpressure
}
```

❌ **Mistake 2: Disabling bloom filters**
```java
// BAD: No bloom filters = slow negative lookups
Options options = new Options()
    .setCreateIfMissing(true);
// Every read checks every SSTable on disk
```

✅ **Good: Enable bloom filters**
```java
BlockBasedTableConfig tableConfig = new BlockBasedTableConfig()
    .setFilterPolicy(new BloomFilter(10, false));  // 10 bits per key
Options options = new Options()
    .setTableFormatConfig(tableConfig);
// 99% reduction in false positive reads
```

❌ **Mistake 3: Tiny block sizes**
```java
// BAD: Small blocks = too many index entries = large index
BlockBasedTableConfig config = new BlockBasedTableConfig()
    .setBlockSize(1024);  // Only 1KB blocks!
// Result: Index size 10x larger, slower reads
```

✅ **Good: Reasonable block size**
```java
BlockBasedTableConfig config = new BlockBasedTableConfig()
    .setBlockSize(16 * 1024);  // 16KB blocks (good balance)
// Balance between compression ratio and random access
```

❌ **Mistake 4: Never compacting**
```java
// BAD: Disabling automatic compaction
Options options = new Options()
    .setDisableAutoCompactions(true);
// Result: SSTables pile up, read performance degrades to unusable
```

✅ **Good: Let compaction run (or tune it)**
```java
Options options = new Options()
    .setLevel0FileNumCompactionTrigger(4)      // Compact when 4 L0 files
    .setMaxBackgroundCompactions(4)            // 4 concurrent compactions
    .setMaxBackgroundFlushes(2);               // 2 concurrent flushes
```

### Edge Cases

**1. Compaction During Reads**
```
Scenario: Read query while SSTables are being compacted

What happens:
- Old SSTables remain accessible during compaction
- New compacted SSTable created separately
- After compaction completes:
  → Switch to new SSTable
  → Delete old SSTables
- Reads never fail or see inconsistent state
```

**2. Overlapping Key Ranges in L0**
```
SSTable-1: keys [a, m]  ← MemTable flush 1
SSTable-2: keys [d, z]  ← MemTable flush 2
SSTable-3: keys [b, f]  ← MemTable flush 3

Read for key="e":
- Must check all 3 SSTables (key "e" might exist in all 3)
- Merge results by timestamp
- Return latest version

After compaction to L1:
SSTable-4: keys [a, z]  ← Merged, de-duplicated
- Only need to check 1 SSTable for key="e"
```

**3. Tombstones and Space Reclamation**
```
Time T0: Write key="user:123" value="Alice"
         → SSTable-1: user:123 = "Alice"

Time T1: Delete key="user:123"
         → SSTable-2: user:123 = TOMBSTONE

Time T2: Read key="user:123"
         → Check SSTable-2: Found TOMBSTONE → Return NULL
         → SSTable-1 still has "Alice" but ignored (older timestamp)

Time T3: Compaction
         → Merge SSTable-1 and SSTable-2
         → TOMBSTONE cancels out old value
         → Result: Key removed, space reclaimed
```

## Quick Reference

### SSTable File Components

| File | Purpose | Size | Loaded to Memory? |
|------|---------|------|-------------------|
| **Data** | Actual key-value pairs | 64-256 MB | No (read on demand) |
| **Index** | Key → offset mapping | 0.1-1% of data | Hot SSTables only |
| **Bloom Filter** | Key existence check | ~1.2% of data | Yes (always) |
| **Summary** | Sparse index samples | 0.01% of data | Yes (always) |
| **Statistics** | Min/max keys, metadata | ~1 KB | Yes (always) |

### When to Use SSTables (via LSM Trees)

✅ **Perfect for:**
- Write-heavy workloads (logs, events, metrics, time-series)
- Append-only patterns (messaging, streaming)
- Sequential access patterns
- Large datasets (TB-PB scale)
- Cloud storage (S3) - immutability is beneficial

❌ **Not ideal for:**
- Read-heavy with random access (consider B-trees)
- Frequent updates to same keys (read-modify-write)
- Small datasets (< 1 GB) where B-tree overhead is negligible
- Strong consistency with immediate reads required

### Key Metrics to Monitor

```java
// RocksDB SSTable metrics
public class SSTableMetrics {
    public void monitorHealth(RocksDB db) throws RocksDBException {
        // Per-level SSTable count
        long l0Files = db.getLongProperty("rocksdb.num-files-at-level0");
        long l1Files = db.getLongProperty("rocksdb.num-files-at-level1");
        
        // Size metrics
        long totalSize = db.getLongProperty("rocksdb.total-sst-files-size");
        long liveSize = db.getLongProperty("rocksdb.estimate-live-data-size");
        
        // Compaction metrics
        long pendingCompaction = db.getLongProperty("rocksdb.estimate-pending-compaction-bytes");
        long numRunning = db.getLongProperty("rocksdb.num-running-compactions");
        
        // Alert conditions
        if (l0Files > 20) {
            System.err.println("⚠️ Too many L0 SSTables! Writes may stall!");
        }
        
        if (pendingCompaction > 50L * 1024 * 1024 * 1024) {
            System.err.println("⚠️ Compaction falling behind!");
        }
        
        double spaceAmplification = (double) totalSize / liveSize;
        if (spaceAmplification > 3.0) {
            System.err.println("⚠️ High space amplification: " + spaceAmplification + "x");
        }
    }
}
```

**Alert Thresholds:**
- L0 files > 10: Compaction struggling
- L0 files > 20: Write stalls imminent
- Pending compaction > 50 GB: Compaction far behind
- Space amplification > 3.0x: Too many obsolete versions

### Compaction Impact on SSTables

```
┌─────────────────────────────────────────────────────────┐
│         SIZE-TIERED COMPACTION (STCS)                   │
├─────────────────────────────────────────────────────────┤
│  Many small SSTables → Few large SSTables               │
│                                                          │
│  L0: [10MB][10MB][10MB][10MB]                          │
│       ↓ Compact 4 similar-sized files                   │
│  L1: [────────── 40MB ──────────]                      │
│       ↓ Compact 4 similar-sized files                   │
│  L2: [────────────── 160MB ──────────────]             │
│                                                          │
│  ✅ Fewer SSTables over time                            │
│  ✅ Better read performance                             │
│  ❌ Temporary space: 2x during compaction               │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│         LEVELED COMPACTION (LCS)                        │
├─────────────────────────────────────────────────────────┤
│  Each level: Fixed-size, non-overlapping SSTables       │
│                                                          │
│  L0: [10MB][10MB][10MB][10MB]  (overlapping OK)        │
│       ↓ Merge to L1                                     │
│  L1: [10MB][10MB][10MB][10MB]  (non-overlapping)       │
│       ↓ Merge to L2                                     │
│  L2: [10MB][10MB]...[100 files]  (10x more)            │
│                                                          │
│  ✅ Predictable read performance                        │
│  ✅ Each level fully sorted                             │
│  ❌ More compaction I/O                                 │
└─────────────────────────────────────────────────────────┘
```

### Configuration Cheat Sheet

```java
// Write-optimized: Fewer, larger SSTables
Options writeHeavy = new Options()
    .setWriteBufferSize(256 * 1024 * 1024)        // Large MemTable: 256MB
    .setTargetFileSizeBase(128 * 1024 * 1024)     // Large SSTables: 128MB
    .setLevel0FileNumCompactionTrigger(10)        // Delay compaction
    .setCompressionType(CompressionType.LZ4_COMPRESSION);

// Read-optimized: More compaction, fewer SSTables to check
Options readHeavy = new Options()
    .setWriteBufferSize(64 * 1024 * 1024)         // Smaller MemTable: 64MB
    .setTargetFileSizeBase(64 * 1024 * 1024)      // Smaller SSTables: 64MB
    .setLevel0FileNumCompactionTrigger(2)         // Aggressive compaction
    .setMaxBackgroundCompactions(8)               // Many compaction threads
    .setTableFormatConfig(
        new BlockBasedTableConfig()
            .setBlockSize(16 * 1024)               // 16KB blocks
            .setFilterPolicy(new BloomFilter(10))  // Bloom filters
            .setBlockCache(new LRUCache(4L * 1024 * 1024 * 1024))  // 4GB cache
    );

// Balanced: General purpose
Options balanced = new Options()
    .setWriteBufferSize(128 * 1024 * 1024)        // 128MB MemTable
    .setTargetFileSizeBase(64 * 1024 * 1024)      // 64MB SSTables
    .setLevel0FileNumCompactionTrigger(4)         // Moderate compaction
    .setCompressionType(CompressionType.LZ4_COMPRESSION);
```

## Related Topics

### Within LSM Tree Architecture
- **MemTable**: In-memory buffer that flushes to SSTables ([memtable.md](./memtable.md))
- **LSM Tree**: Overall architecture using SSTables ([lsm-tree.md](./lsm-tree.md))
- **Compaction**: Background process that merges SSTables
- **Write-Ahead Log**: Durability mechanism ([../db-wal.md](../db-wal.md))

### Database Implementations
- **Cassandra**: Wide-column store using SSTables ([cassandara.md](./cassandara.md))
- **RocksDB**: Embedded key-value store with SSTable format
- **HBase**: Hadoop-based database with HFile (SSTable variant)
- **ScyllaDB**: High-performance C++ implementation

### Performance Topics
- **Bloom Filters**: Probabilistic data structure for existence checks
- **Block Cache**: In-memory cache for SSTable data blocks
- **Compression Algorithms**: LZ4, Snappy, Zstd trade-offs
- **Write Amplification**: How many times data is rewritten
- **Read Amplification**: How many SSTables checked per read

### Storage Comparisons
- **B-Tree vs SSTable**: In-place updates vs immutable files ([../db-storage-engines.md](../db-storage-engines.md))
- **Database Indexes**: Different indexing strategies ([../db-indexes.md](../db-indexes.md))

### Further Reading
- [LSM Tree Architecture](./lsm-tree.md)
- [MemTable Design](./memtable.md)
- [Cassandra Deep Dive](./cassandara.md)
- [Database Storage Engines](../db-storage-engines.md)

---

## Summary

**SSTables are immutable, sorted, on-disk files that power write-optimized databases:**

✅ **Strengths:**
- Enable extremely fast writes (sequential I/O, no in-place updates)
- Self-contained with indexes and bloom filters
- Compression reduces storage costs by 50-80%
- Immutability simplifies concurrency (no locks needed)
- Perfect for write-heavy workloads (logging, events, metrics)

⚠️ **Challenges:**
- Multiple SSTables slow down reads (until compaction)
- Space amplification (old versions kept until compaction)
- Compaction overhead (background I/O cost)
- Deletes not instant (tombstones until compaction)

**Bottom Line:** SSTables are the foundational storage format for modern write-optimized databases. They trade read complexity for write simplicity, making them perfect for high-throughput data ingestion. The combination of immutability, sorting, bloom filters, and background compaction creates a system that can handle millions of writes per second while maintaining acceptable read performance.

**Key Insight:** The genius of SSTables is converting expensive random I/O into cheap sequential I/O. By making files immutable and sorted, databases can:
1. Write entire MemTables as single sequential disk writes
2. Use bloom filters to skip 99% of SSTables during reads
3. Leverage OS page cache effectively (immutable files cache well)
4. Run compaction in the background without blocking operations

This design is why Cassandra, HBase, and RocksDB can handle multi-million writes/sec on commodity hardware.

