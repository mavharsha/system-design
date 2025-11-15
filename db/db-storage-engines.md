# Database Storage Engines: Complete Guide

## Table of Contents
1. [Responsibilities: Database vs Storage Engine](#responsibilities)
2. [Storage Engine Types & Architecture](#architecture)
3. [Complete List: Databases & Storage Engines](#list)
4. [Performance Characteristics](#performance)
5. [Decision Guide](#decision-guide)
6. [Trade-offs & Use Cases](#tradeoffs)
7. [Real-World Examples](#examples)

---

## Responsibilities: Database vs Storage Engine

### **Storage Engine Responsibilities** (Low-level, embedded library)
- **Physical data storage**: How data is written to disk/memory
- **Data structures**: B-trees, LSM-trees, hash tables, etc.
- **Page/block management**: Managing fixed-size storage units
- **Indexing mechanisms**: Primary and secondary index structures
- **Caching**: Buffer pools, memory management
- **Crash recovery**: Write-ahead logging (WAL), checkpointing
- **Single-node ACID**: Transactions on a single machine
- **Compression**: Data encoding and compression algorithms
- **Read/write operations**: Get, put, delete, scan primitives

**What storage engines DON'T do:**
- Network communication
- SQL/query parsing
- Distributed coordination
- User authentication
- Query optimization
- Multi-node transactions

### **Database System Responsibilities** (Complete application)
- **Query language**: SQL, CQL, MongoDB query language, etc.
- **Query parser & optimizer**: Converting queries to execution plans
- **Network protocol**: Client-server communication (TCP, HTTP, etc.)
- **Authentication & authorization**: User management, permissions
- **Catalog management**: Schema, metadata storage
- **Distributed coordination**: Sharding, replication, consensus (for distributed DBs)
- **Distributed transactions**: ACID across multiple nodes
- **Connection pooling**: Managing client connections
- **Backup/restore**: High-level data management tools
- **Monitoring & logging**: Operational tools
- **Uses storage engine(s)** as the underlying persistence layer

---

## Storage Engine Types & Architecture

### **Architecture Layers: Full Stack View**

```
┌─────────────────────────────────────────────────────────────┐
│                    APPLICATION LAYER                         │
│              (Your Java/Python/Node.js App)                  │
└─────────────────────────────────────────────────────────────┘
                           ↓ (Database Driver/Client)
┌─────────────────────────────────────────────────────────────┐
│                   DATABASE SYSTEM                            │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Query Parser & Optimizer                             │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Network Layer (TCP/HTTP)                             │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Transaction Manager                                   │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Catalog & Schema Management                          │  │
│  └───────────────────────────────────────────────────────┘  │
│                           ↓                                  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │           STORAGE ENGINE INTERFACE                    │  │
│  │        (get, put, delete, scan, batch)                │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                    STORAGE ENGINE                            │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Buffer Pool / Cache Manager                          │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Index Structures (B-tree/LSM-tree/Hash)              │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Write-Ahead Log (WAL)                                │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Compaction / Merging (for LSM)                       │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                    PHYSICAL STORAGE                          │
│              (Disk: SSD/HDD, Memory: RAM)                    │
└─────────────────────────────────────────────────────────────┘
```

### **Three Major Storage Engine Types**

#### **1. B-Tree Based Storage Engines**

```
Structure: Balanced Tree (in-place updates)

                    [Root Node]
                   /     |     \
            [Internal] [Internal] [Internal]
            /    \      /    \      /    \
        [Leaf] [Leaf] [Leaf] [Leaf] [Leaf] [Leaf]
         ↓       ↓      ↓      ↓      ↓      ↓
       Disk    Disk   Disk   Disk   Disk   Disk
       Page    Page   Page   Page   Page   Page

Write Operation:
1. Find leaf node (multiple disk seeks)
2. Update page in-place
3. Write updated page back to disk
4. WAL for crash recovery
```

**Characteristics:**
- **Read:** Fast (O(log n) seeks), predictable
- **Write:** Slower (random I/O, in-place updates)
- **Space:** No write amplification from compaction
- **Best for:** Read-heavy workloads, transactional systems

**Examples:** InnoDB (MySQL), PostgreSQL heap storage

#### **2. LSM-Tree Based Storage Engines**

```
Structure: Log-Structured Merge Tree (append-only, periodic compaction)

┌─────────────────────────────────────────────────────────┐
│                      MEMORY                              │
│  ┌─────────────┐         ┌──────────────┐              │
│  │  MemTable   │  ───→   │   Immutable  │              │
│  │  (Active)   │ (Full)  │   MemTable   │              │
│  │  (Sorted)   │         │   (Being     │              │
│  │             │         │   Flushed)   │              │
│  └─────────────┘         └──────────────┘              │
└─────────────────────────────────────────────────────────┘
                                ↓ Flush to disk
┌─────────────────────────────────────────────────────────┐
│                       DISK                               │
│  Level 0:  [SST-1] [SST-2] [SST-3] [SST-4]              │
│            (May have overlapping keys)                   │
│                         ↓ Compaction                     │
│  Level 1:  [SST-5────────] [SST-6────────]              │
│            (No overlapping keys)                         │
│                         ↓ Compaction                     │
│  Level 2:  [SST-7──────────────────] [SST-8──────]      │
│            (No overlapping keys, larger files)           │
│                         ↓ Compaction                     │
│  Level 3:  [SST-9────────────────────────────────────]  │
│            (Largest files)                               │
└─────────────────────────────────────────────────────────┘

Write Path:
1. Write to WAL (sequential write)
2. Write to MemTable (in memory)
3. When MemTable full → flush to SSTable (Level 0)
4. Background compaction merges SSTables

Read Path:
1. Check MemTable
2. Check Immutable MemTable
3. Check Bloom filters for each SSTable
4. Read from SSTables (may need multiple levels)
```

**Characteristics:**
- **Write:** Very fast (sequential writes, no random I/O)
- **Read:** Slower (may check multiple levels)
- **Space:** Write amplification due to compaction
- **Best for:** Write-heavy workloads, time-series data

**Examples:** RocksDB, LevelDB, Cassandra, HBase

#### **3. Log-Structured Hash Storage (Bitcask-style)**

```
Structure: Append-only log + in-memory hash index

┌─────────────────────────────────────────────────────────┐
│                      MEMORY                              │
│  ┌────────────────────────────────────────────────────┐ │
│  │         Hash Index (All keys in RAM)               │ │
│  │  ┌────────┬──────────────────┐                     │ │
│  │  │  Key   │  File Offset     │                     │ │
│  │  ├────────┼──────────────────┤                     │ │
│  │  │ "user1"│  offset: 1024    │  ───→ Fast lookup  │ │
│  │  │ "user2"│  offset: 2048    │                     │ │
│  │  │ "user3"│  offset: 4096    │                     │ │
│  │  └────────┴──────────────────┘                     │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
                           ↓ Points to
┌─────────────────────────────────────────────────────────┐
│                      DISK                                │
│  ┌────────────────────────────────────────────────────┐ │
│  │         Active Data File (Append-only)             │ │
│  │  [Record1][Record2][Record3][Record4]...           │ │
│  │   ↑        ↑        ↑                               │ │
│  │   1024     2048     4096                            │ │
│  └────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────┐ │
│  │         Older Data Files (Immutable)               │ │
│  │  [File-001.data] [File-002.data] [File-003.data]  │ │
│  └────────────────────────────────────────────────────┘ │
│                          ↓ Compaction                    │
│  ┌────────────────────────────────────────────────────┐ │
│  │  Merge old files, remove deleted/old values        │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘

Write: O(1) - append to log + update hash
Read:  O(1) - hash lookup + single disk seek
```

**Characteristics:**
- **Write:** Extremely fast (single sequential write)
- **Read:** Extremely fast (one disk seek)
- **Limitation:** All keys must fit in memory
- **Best for:** High-throughput key-value stores with limited key space

**Examples:** Bitcask (Riak), Redis persistence (RDB + AOF hybrid)

---

## Complete List: Most Used Databases & Storage Engines

### **Full Database Systems**

**PostgreSQL**
- **Type:** Full relational database (RDBMS)
- **Storage Engine:** Built-in heap storage (proprietary)
- **Architecture:** Not pluggable; single integrated storage layer
- **Notes:** Since PostgreSQL 12, supports Table Access Methods (experimental pluggable storage)

**MySQL**
- **Type:** Full relational database (RDBMS)
- **Default Storage Engine:** InnoDB (B-tree based)
- **Alternative Engines:** MyISAM, MyRocks (RocksDB), Memory, CSV, Archive
- **Architecture:** Pluggable storage engine architecture
- **Notes:** Can switch engines per table

**MongoDB**
- **Type:** Full document database (NoSQL)
- **Default Storage Engine:** WiredTiger (since MongoDB 3.2)
- **Architecture:** Pluggable storage engines
- **Notes:** WiredTiger uses document-level concurrency and MVCC

**YugabyteDB**
- **Type:** Full distributed SQL database
- **Storage Engine:** DocDB (built on RocksDB)
- **Architecture:** DocDB manages multiple RocksDB instances, one per tablet
- **Additional Layers:** Raft consensus, distributed transaction manager, query layer

**Riak** (note: project discontinued in 2021)
- **Type:** Full distributed key-value database
- **Storage Engine:** Bitcask (default), LevelDB, or Memory backend
- **Architecture:** Pluggable storage backends
- **Notes:** Bitcask was created specifically for Riak

### **Storage Engines Only** (Embedded Libraries)

**LevelDB**
- **Type:** Storage engine library ONLY
- **Architecture:** LSM-tree based
- **Usage:** Embedded in applications (Chrome, Bitcoin Core, etc.)
- **Created by:** Google
- **Notes:** Single-threaded compaction, no network layer

**RocksDB**
- **Type:** Storage engine library ONLY
- **Architecture:** LSM-tree based (fork of LevelDB)
- **Usage:** Embedded in databases (YugabyteDB, CockroachDB, TiDB, MySQL as MyRocks)
- **Created by:** Facebook
- **Notes:** Multi-threaded compaction, optimized for SSDs

**Bitcask**
- **Type:** Storage engine library ONLY
- **Architecture:** Log-structured hash table
- **Usage:** Embedded in Riak and other applications
- **Created by:** Basho (Riak creators)
- **Notes:** All keys must fit in memory, extremely simple design

## Visual Hierarchy

```
DATABASE SYSTEMS (Complete applications with network, query language, etc.)
├── PostgreSQL
│   └── Uses: Built-in heap storage
├── MySQL
│   └── Uses: InnoDB (or MyRocks/RocksDB)
├── MongoDB
│   └── Uses: WiredTiger
├── YugabyteDB
│   └── Uses: DocDB → RocksDB
└── Riak
    └── Uses: Bitcask (or LevelDB)

STORAGE ENGINES (Embedded libraries only)
├── LevelDB
├── RocksDB
└── Bitcask
```

## Key Distinctions

| Aspect | Storage Engine | Database System |
|--------|---------------|-----------------|
| **Interface** | Low-level API (get/put/delete) | SQL, CQL, or high-level query language |
| **Network** | No network layer | TCP/HTTP server |
| **Deployment** | Library linked into app | Standalone server process |
| **Queries** | Key-value operations only | Complex queries, joins, aggregations |
| **Users** | No user management | Authentication & authorization |
| **Distribution** | Single node only | Can be distributed (sharding, replication) |
| **Examples** | LevelDB, RocksDB, Bitcask | PostgreSQL, MySQL, MongoDB |

## Real-World Analogy

**Storage Engine** = Engine in a car
- Provides power and movement
- No steering wheel, no dashboard, no seats
- Just the core mechanical function

**Database System** = Complete car
- Has the engine PLUS steering, dashboard, seats, AC, radio
- Ready to drive
- User-friendly interface

---

## Performance Characteristics

### **Detailed Comparison: B-Tree vs LSM-Tree vs Hash-Log**

| Characteristic | B-Tree (InnoDB) | LSM-Tree (RocksDB) | Hash-Log (Bitcask) |
|---|---|---|---|
| **Read Latency** | O(log n) seeks<br>3-4 disk I/O typically | O(k) where k = levels<br>Can be 4-10 I/O | O(1) seek<br>Single disk I/O |
| **Write Latency** | Random I/O<br>Slower, unpredictable | Sequential I/O<br>Very fast, consistent | Sequential append<br>Fastest possible |
| **Read Throughput** | 5,000-15,000 IOPS<br>(with proper indexing) | 3,000-10,000 IOPS<br>(depends on levels) | 50,000+ IOPS<br>(memory-limited) |
| **Write Throughput** | 2,000-5,000 IOPS<br>(random writes) | 50,000-200,000 IOPS<br>(sequential writes) | 100,000+ IOPS<br>(append-only) |
| **Space Amplification** | 1.0-1.2x<br>Minimal overhead | 1.5-3.0x<br>Due to compaction | 1.2-2.0x<br>During compaction |
| **Write Amplification** | 2-3x<br>(WAL + page update) | 10-30x<br>(multiple compactions) | 2-4x<br>(during merge) |
| **Memory Usage** | Buffer pool<br>Configurable | MemTable + Block cache<br>Higher memory need | All keys in RAM<br>High memory requirement |
| **Bloom Filters** | Not typically used | Essential for reads<br>Reduces I/O | Not needed<br>(hash lookup) |
| **Compaction Impact** | No compaction<br>Stable performance | Background compaction<br>Can cause latency spikes | Infrequent merges<br>Minimal impact |
| **Range Scans** | Excellent<br>Sequential leaf reads | Good<br>May span levels | Poor/Not supported<br>Random access only |
| **Point Queries** | Good<br>Log(n) seeks | Fair<br>Check multiple levels | Excellent<br>Single seek |
| **Update Performance** | In-place update<br>Moderate cost | Append new version<br>Very fast | Append new version<br>Very fast |
| **Delete Performance** | In-place delete<br>Moderate cost | Tombstone marker<br>Very fast | Tombstone marker<br>Very fast |

### **Real-World Performance Numbers**

**B-Tree (InnoDB on SSD):**
```
Point Read:      0.1-1 ms
Range Scan:      100-500 μs per row
Write:           0.5-2 ms
Bulk Insert:     2,000-5,000 rows/sec
```

**LSM-Tree (RocksDB on SSD):**
```
Point Read:      0.2-2 ms (depends on levels)
Range Scan:      200-800 μs per row
Write:           50-200 μs (to MemTable)
Bulk Insert:     50,000-200,000 rows/sec
Compaction:      Can consume 50-80% I/O bandwidth
```

**Hash-Log (Bitcask on SSD):**
```
Point Read:      50-200 μs
Range Scan:      Not supported
Write:           20-100 μs
Bulk Insert:     100,000+ rows/sec
Memory:          ~32 bytes per key (minimum)
```

### **I/O Pattern Visualization**

**B-Tree Write Pattern:**
```
Time ──────────────────────────────────────────→

Disk:  [Random][Random][Random][Random][Random]
       ↓ Seek   ↓ Seek  ↓ Seek  ↓ Seek  ↓ Seek
       Write A  Write B Write C Write D Write E

Issue: Random I/O, high latency per write
```

**LSM-Tree Write Pattern:**
```
Time ──────────────────────────────────────────→

Memory: [A][B][C][D][E]... (MemTable fills up)
              ↓ Flush
Disk:   [Sequential write of sorted data]
              ↓ Later
        [Compaction merges multiple files]

Benefit: Sequential writes, very fast
Cost: Read amplification, write amplification
```

**Hash-Log Write Pattern:**
```
Time ──────────────────────────────────────────→

Disk:  [A][B][C][D][E][F][G][H]... (continuous append)

Benefit: Simplest possible writes
Cost: All keys must fit in memory
```

---

## Decision Guide

### **When to Use B-Tree Storage Engines**

✅ **Choose B-Tree if:**
- Read-heavy workload (80%+ reads)
- Need strong ACID guarantees with transactions
- Require efficient range scans
- Have mixed read/write patterns
- Need predictable, consistent performance
- Working with relational data and complex queries
- Space efficiency is important

❌ **Avoid B-Tree if:**
- Write-heavy workload (80%+ writes)
- High write throughput is critical (>10K writes/sec)
- Can tolerate slightly higher read latency
- Write latency spikes are unacceptable

**Best Use Cases:**
- E-commerce transactional systems
- Banking and financial applications
- Traditional OLTP workloads
- User profile management
- Inventory management systems

**Example Databases:** MySQL (InnoDB), PostgreSQL

---

### **When to Use LSM-Tree Storage Engines**

✅ **Choose LSM-Tree if:**
- Write-heavy workload (60%+ writes)
- Need very high write throughput (>50K writes/sec)
- Time-series data or logs
- Can trade read performance for write performance
- Have sufficient storage for write amplification
- Background compaction is acceptable

❌ **Avoid LSM-Tree if:**
- Read latency must be ultra-low and predictable
- Cannot tolerate compaction I/O spikes
- Storage space is extremely limited
- Range scans are the primary operation
- Need consistent, predictable performance

**Best Use Cases:**
- Time-series databases (metrics, logs)
- Event streaming platforms
- IoT sensor data ingestion
- Social media feeds and activity logs
- Write-heavy analytics pipelines
- Distributed databases (Cassandra, HBase)

**Example Databases:** Cassandra, HBase, RocksDB, ScyllaDB

---

### **When to Use Hash-Log Storage Engines**

✅ **Choose Hash-Log if:**
- Key-value workload only (no range scans)
- All keys fit in memory comfortably
- Need absolute maximum throughput
- Simplicity is valued
- Point queries are primary operation
- Write and read both need to be extremely fast

❌ **Avoid Hash-Log if:**
- Key space is very large (billions of keys)
- Limited memory availability
- Need range scans or complex queries
- Key size is unpredictable or very large

**Best Use Cases:**
- Session stores
- Cache implementations
- Configuration storage
- Small-scale key-value stores
- Real-time analytics with limited key space

**Example Databases:** Riak (Bitcask mode), Redis (hybrid)

---

### **Decision Flowchart**

```
Start: What is your primary workload?
    |
    ├─→ [Key-Value Only, Keys Fit in Memory]
    |       → Use Hash-Log (Bitcask)
    |
    ├─→ [Write-Heavy (>60% writes), High Throughput]
    |       |
    |       ├─→ Time-series/Logs → Use LSM-Tree (RocksDB/Cassandra)
    |       └─→ Can tolerate compaction → Use LSM-Tree
    |
    └─→ [Read-Heavy OR Balanced OR Need Transactions]
            |
            ├─→ Complex queries, ACID → Use B-Tree (PostgreSQL/InnoDB)
            ├─→ Range scans critical → Use B-Tree
            └─→ Predictable performance → Use B-Tree
```

---

## Trade-offs & Use Cases

### **The CAP of Storage Engines: Pick Your Priorities**

```
        Write Throughput
              ▲
              │
              │    LSM-Tree
              │       ●
              │      / \
              │     /   \
              │    /     \
              │   /       \
              │  /         \  Hash-Log
              │ /           \    ●
              │/             \  /
              ●───────────────●───────→ Read Performance
           B-Tree         /
                        /
                      /
                    ▼
              Consistency & ACID
```

### **Key Trade-offs Explained**

#### **1. Write Amplification**

**Definition:** How many times data is written to disk per logical write

```
Logical Write: User writes 1 KB of data

B-Tree:
├─ Write to WAL: 1 KB
├─ Update page: 16 KB (entire page must be written)
├─ Update parent nodes: 16 KB × 2
└─ Total: ~50 KB written (50x amplification)

LSM-Tree:
├─ Write to WAL: 1 KB
├─ Write to MemTable: 0 KB (in memory)
├─ Flush to L0: 1 KB
├─ Compact L0→L1: 10 KB
├─ Compact L1→L2: 50 KB
├─ Compact L2→L3: 200 KB
└─ Total: ~260 KB written (260x amplification worst case)

Hash-Log:
├─ Append to log: 1 KB
├─ Update hash index: 0 KB (in memory)
└─ Total: 1 KB written (1x amplification)

Mitigation Strategies:
- LSM: Tune compaction parameters, use leveled compaction
- B-Tree: Larger page sizes, batch writes
- Hash-Log: Infrequent merging, size-tiered compaction
```

#### **2. Read Amplification**

**Definition:** How many disk I/Os needed per logical read

```
Point Query: Get value for key "user:12345"

B-Tree:
├─ Seek to root: 1 I/O
├─ Seek to internal node: 1 I/O
├─ Seek to leaf: 1 I/O
└─ Total: 3 I/Os (best case with warm cache: 1 I/O)

LSM-Tree:
├─ Check MemTable: 0 I/O (in memory)
├─ Check bloom filters: 0 I/O (in memory)
├─ Read from L0: 0-4 I/Os (may have overlapping files)
├─ Read from L1: 0-1 I/O
├─ Read from L2: 0-1 I/O
├─ Read from L3: 0-1 I/O
└─ Total: 1-7 I/Os (worst case)

Hash-Log:
├─ Hash lookup: 0 I/O (in memory)
├─ Seek to offset: 1 I/O
└─ Total: 1 I/O (always)

Mitigation Strategies:
- LSM: Bloom filters, block cache, compaction to reduce levels
- B-Tree: Buffer pool cache, covering indexes
- Hash-Log: None needed (already optimal)
```

#### **3. Space Amplification**

**Definition:** Storage overhead beyond actual data size

```
Storing 100 GB of user data:

B-Tree:
├─ Data: 100 GB
├─ Index overhead: 10 GB (10%)
├─ Page fragmentation: 10 GB (10%)
└─ Total: 120 GB (1.2x)

LSM-Tree:
├─ Data: 100 GB
├─ Older versions during compaction: 50 GB
├─ Deleted tombstones: 20 GB
└─ Total: 170 GB (1.7x, can be 2-3x during heavy compaction)

Hash-Log:
├─ Data: 100 GB
├─ Old values before merge: 30 GB
└─ Total: 130 GB (1.3x)
```

---

## Real-World Examples

### **Example 1: E-commerce Product Catalog (B-Tree)**

**Scenario:** Amazon-like product database
- 10M products
- 100 reads/sec per product on average
- 1,000 updates/sec (inventory, pricing)
- Need ACID transactions
- Complex queries (search by category, price range)

**Why B-Tree (PostgreSQL/MySQL InnoDB)?**
```sql
-- Typical queries benefit from B-tree indexes
SELECT * FROM products 
WHERE category = 'Electronics' 
  AND price BETWEEN 100 AND 500
ORDER BY rating DESC
LIMIT 20;

-- Updates are moderate, not overwhelming
UPDATE products 
SET inventory = inventory - 1 
WHERE product_id = 12345 
  AND inventory > 0;
```

**Architecture:**
```
PostgreSQL with B-Tree Indexes
├─ Primary Key Index: B-Tree on product_id
├─ Secondary Index: B-Tree on (category, price)
├─ Secondary Index: B-Tree on rating
└─ Buffer Pool: 16 GB (cache hot data)

Performance:
- Reads: 1-2 ms (p99)
- Writes: 2-5 ms (p99)
- Range scans: Excellent
- ACID: Full compliance
```

---

### **Example 2: IoT Sensor Data (LSM-Tree)**

**Scenario:** Temperature sensors for data centers
- 100,000 sensors
- Each sensor reports every 10 seconds
- 10,000 writes/sec sustained
- Queries are mostly time-range (last hour, last day)
- Retention: 90 days

**Why LSM-Tree (Cassandra/ScyllaDB)?**
```java
// Write path: Sequential, high throughput
public void recordSensorData(String sensorId, long timestamp, double temp) {
    // LSM-tree excels at sequential writes
    INSERT INTO sensor_data (sensor_id, timestamp, temperature)
    VALUES (?, ?, ?);
}

// Read path: Time-range scans
public List<Reading> getLastHour(String sensorId) {
    long now = System.currentTimeMillis();
    SELECT * FROM sensor_data 
    WHERE sensor_id = ? 
      AND timestamp > ? 
      AND timestamp < ?;
}
```

**Architecture:**
```
Cassandra (LSM-Tree based)
├─ Partition Key: sensor_id
├─ Clustering Key: timestamp (sorted)
├─ MemTable: 256 MB per node
├─ SSTable Levels: 4 levels
└─ Compaction: Time-window compaction strategy

Performance:
- Writes: 50-100 μs (p99)
- Write throughput: 50,000/sec per node
- Reads: 5-20 ms (p99, scanning multiple SSTables)
- Storage: 2x amplification due to compaction
```

---

### **Example 3: Session Store (Hash-Log / Redis)**

**Scenario:** Web application session management
- 1M active sessions
- Average session size: 5 KB
- 100,000 reads/sec
- 10,000 writes/sec
- TTL: 1 hour

**Why Hash-Log (Redis with RDB+AOF)?**
```java
// Session operations: Simple get/set
public class SessionStore {
    private Redis redis;
    
    // O(1) write
    public void saveSession(String sessionId, SessionData data) {
        redis.set(sessionId, serialize(data), 3600); // 1 hour TTL
    }
    
    // O(1) read
    public SessionData getSession(String sessionId) {
        return deserialize(redis.get(sessionId));
    }
}
```

**Architecture:**
```
Redis (In-memory + AOF persistence)
├─ All keys in memory: 5 GB (1M × 5 KB)
├─ Hash index overhead: 64 MB
├─ AOF log: Append-only file on disk
└─ RDB snapshots: Every 5 minutes

Performance:
- Reads: 50-200 μs
- Writes: 20-100 μs
- Throughput: 100,000+ ops/sec (single instance)
- Memory: Must fit all data
```

---

### **Quick Reference: Real Systems**

| System | Storage Engine | Primary Use Case |
|--------|---------------|------------------|
| **MySQL** | InnoDB (B-Tree) | OLTP, e-commerce, banking |
| **PostgreSQL** | Heap + B-Tree | OLTP, complex queries, ACID |
| **Cassandra** | LSM-Tree | Time-series, IoT, high writes |
| **HBase** | LSM-Tree | BigTable-like, analytics |
| **RocksDB** | LSM-Tree | Embedded DB, blockchain |
| **MongoDB** | WiredTiger (B-Tree) | Document store, flexible schema |
| **Redis** | In-memory + AOF | Caching, sessions, real-time |
| **Riak** | Bitcask (Hash-Log) | KV store, high availability |
| **CockroachDB** | RocksDB (LSM) | Distributed SQL, geo-replication |
| **TiDB** | RocksDB (LSM) | Distributed SQL, HTAP |

---

## Summary: The Bottom Line

**Choose Your Storage Engine Based On:**

1. **Workload Pattern**
   - Read-heavy → B-Tree
   - Write-heavy → LSM-Tree
   - Key-value + memory fits → Hash-Log

2. **Query Types**
   - Complex queries + range scans → B-Tree
   - Time-series + point queries → LSM-Tree
   - Simple get/set → Hash-Log

3. **Performance Priorities**
   - Read latency critical → B-Tree or Hash-Log
   - Write throughput critical → LSM-Tree or Hash-Log
   - Balanced → B-Tree

4. **Operational Constraints**
   - Limited memory → B-Tree or LSM-Tree
   - Limited storage → B-Tree
   - Limited operational complexity → Hash-Log

**The Golden Rule:**
> "There is no universally best storage engine. The right choice depends on your specific workload, query patterns, and operational constraints. Measure, profile, and choose accordingly."