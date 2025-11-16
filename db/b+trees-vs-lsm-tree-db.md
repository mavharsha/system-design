# B+ Trees vs LSM Trees in Databases

## Overview

Database storage engines fundamentally differ in how they organize data on disk. The two dominant approaches are **B+ Trees** (optimized for consistent reads) and **LSM Trees** (optimized for high-throughput writes). Beyond raw performance, a critical differentiator is how they handle **temporal access patterns**—B+ Trees provide uniform performance regardless of data age, while LSM Trees excel when recent data is accessed far more frequently than historical data.


(Notes: High throughput systems-1 - 1hour 25 mins)

## Database Implementations

### B+ Tree Based Databases

| Database | Type | Use Case | Key Features | Temporal Pattern |
|----------|------|----------|--------------|------------------|
| **PostgreSQL** | RDBMS | General-purpose | MVCC, full ACID, complex queries | Uniform access |
| **MySQL (InnoDB)** | RDBMS | Web applications | Row-level locking, foreign keys | Consistent latency |
| **Oracle** | RDBMS | Enterprise | Advanced features, partitioning | Age-independent |
| **SQL Server** | RDBMS | Windows enterprise | Integrated with .NET, BI tools | Historical queries |
| **SQLite** | Embedded | Local storage | Single file, serverless | Random access |
| **BerkeleyDB** | Key-Value | Embedded systems | Multiple access methods | Uniform performance |
| **MongoDB (WiredTiger)** | Document | Flexible schema | Document-oriented, horizontal scaling | Mixed workloads |
| **CockroachDB** | NewSQL | Distributed SQL | Geo-distributed, strong consistency | Global consistency |

### LSM Tree Based Databases

| Database | Type | Use Case | Key Features | Temporal Pattern |
|----------|------|----------|--------------|------------------|
| **Cassandra** | Wide-column | High write throughput | Distributed, TTL support | Recent data hot |
| **ScyllaDB** | Wide-column | Ultra-low latency | C++ performance, cache-aware | Time-decaying access |
| **RocksDB** | Key-Value | Embedded | Level compaction, compression | Recent in memory |
| **LevelDB** | Key-Value | Embedded | Google's lightweight KV store | Write-optimized |
| **HBase** | Wide-column | Big data | Hadoop integration, time-series | Recent data cached |
| **BigTable** | Wide-column | Google-scale | Original LSM implementation | Age-tiered storage |
| **TiKV** | Key-Value | Distributed KV | ACID transactions, Raft consensus | Hot/cold separation |
| **DynamoDB** | Key-Value | AWS managed | Serverless, auto-scaling | Recent item cache |
| **InfluxDB** | Time-series | Metrics/monitoring | Time-series optimized | Temporal sharding |

## How B+ Trees Improve Reads and Writes

### Read Performance Advantages

1. **Direct Access Path**
   - O(log N) lookups with predictable I/O patterns
   - Each level narrows the search space efficiently
   - Typically 3-4 disk seeks for any record in a billion-row table
   - **Uniform latency**: 10-year-old data accessed as fast as 10-minute-old data

2. **Cache-Friendly Structure**
   - Internal nodes contain only keys, maximizing cache utilization
   - Hot paths (root and upper levels) stay in memory
   - Sequential leaf nodes enable efficient range scans
   - **Age-agnostic caching**: LRU based on access frequency, not data age

3. **In-Place Updates**
   - No need to search through multiple files
   - Single source of truth for each record
   - No read amplification from checking multiple versions
   - **Consistent performance**: No degradation for historical data access

### Write Performance Characteristics

1. **Random I/O Pattern**
   - Updates require finding the exact page location
   - May trigger page splits and tree rebalancing
   - Write amplification from updating indexes

2. **Optimization Techniques**
   - Write-ahead logging (WAL) for durability
   - Buffer pool to batch writes
   - Group commit to amortize fsync costs

3. **Concurrency Control**
   - Fine-grained locking at page/row level
   - MVCC (Multi-Version Concurrency Control) in modern implementations
   - Read queries don't block writes

### Example Write Path in PostgreSQL (B+ Tree)
```
1. Write to WAL (sequential, fast)
2. Update buffer pool page in memory
3. Mark page as dirty
4. Background writer flushes dirty pages to disk
5. Checkpoint periodically for crash recovery
```

## How LSM Trees Improve Reads and Writes

### Write Performance Advantages

1. **Sequential Write Pattern**
   - All writes go to in-memory MemTable first (O(1))
   - Flush to disk as sorted, immutable SSTable files
   - No random I/O - all disk writes are sequential
   - **Temporal benefit**: Recent writes stay in RAM until flushed

2. **Write Amplification Reduction**
   - Batching writes before flushing to disk
   - No in-place updates means no page splits
   - Append-only nature eliminates seek time
   - **Natural time-ordering**: Newer data automatically in faster storage tiers

3. **High Throughput**
   - Can sustain millions of writes per second
   - Writes never block on disk I/O (until memory pressure)
   - Perfect for write-heavy workloads
   - **Time-series friendly**: Append-only model ideal for temporal data

### Read Performance Characteristics

1. **Read Amplification with Temporal Bias**
   - Must check MemTable + multiple SSTable levels
   - **Performance by age**: Recent data (MemTable) = microseconds, older data (deep SSTables) = milliseconds
   - Bloom filters reduce unnecessary disk reads
   - **Exponential latency growth**: Each compaction level adds 10x latency

2. **Optimization Techniques**
   - Bloom filters (probabilistic data structure)
   - Block cache for frequently accessed data
   - **Temporal locality optimization**: Recent SSTables cached more aggressively
   - Compaction to merge files and reduce levels

3. **Range Scans**
   - Requires merging results from multiple files
   - Sorted nature of SSTables helps
   - **Time-range queries**: Extremely fast for recent time windows
   - Less efficient than B+ tree for full historical scans

### Example Read Path in Cassandra (LSM Tree)
```
1. Check MemTable (in memory)
2. Check row cache (if enabled)
3. Use bloom filters to identify SSTables
4. Binary search in SSTable indexes
5. Read data blocks from SSTables
6. Merge results from all sources
7. Return most recent version
```

## Decision Guide: B+ Tree vs LSM Tree

### Choose B+ Tree Based Databases When:

**Workload Characteristics:**
- [ ] Read-heavy workload (>70% reads)
- [ ] Need consistent read latency regardless of data age
- [ ] Complex queries with multiple indexes
- [ ] Frequent updates to existing records
- [ ] Small to medium dataset that fits mostly in memory
- [ ] **Uniform temporal access**: Old data accessed as frequently as new data

**Query Patterns:**
- [ ] Complex JOIN operations
- [ ] Ad-hoc analytical queries
- [ ] Need for secondary indexes
- [ ] Range scans are common
- [ ] Point lookups require predictable latency

**Consistency Requirements:**
- [ ] Strong consistency is mandatory
- [ ] ACID transactions across multiple tables
- [ ] Foreign key constraints needed
- [ ] Complex transaction isolation levels

**Operational Considerations:**
- [ ] Limited operational complexity desired
- [ ] Mature tooling and expertise available
- [ ] Need for SQL compatibility
- [ ] Backup/restore must be straightforward

### Choose LSM Tree Based Databases When:

**Workload Characteristics:**
- [ ] Write-heavy workload (>70% writes)
- [ ] Can tolerate some read latency variance based on data age
- [ ] Append-only or time-series data
- [ ] Massive scale (TBs to PBs)
- [ ] High ingestion rate (100K+ writes/sec)
- [ ] **Temporal access pattern**: Recent data accessed 10-1000x more than old data

**Query Patterns:**
- [ ] Simple key-value lookups
- [ ] Limited need for complex queries
- [ ] Can design around eventual consistency
- [ ] Batch processing is acceptable
- [ ] Recent data accessed more than old data

**Consistency Requirements:**
- [ ] Eventual consistency is acceptable
- [ ] Can work with last-write-wins
- [ ] No complex transaction requirements
- [ ] Can handle duplicate data temporarily

**Operational Considerations:**
- [ ] Horizontal scaling is critical
- [ ] Can handle compaction overhead
- [ ] Storage cost is a major factor
- [ ] Write availability is critical

## Key Questions to Ask

### Data Access Patterns
1. **What's your read/write ratio and temporal pattern?**
   - 90% reads with uniform access → B+ Tree
   - 90% writes with recent-data bias → LSM Tree
   - Mixed with historical queries → B+ Tree
   - Mixed with time-decay access → LSM Tree

2. **What's your query complexity?**
   - Complex JOINs, aggregations → B+ Tree
   - Simple key lookups → LSM Tree

3. **What's your latency requirement?**
   - Consistent low latency → B+ Tree
   - Can tolerate variance → LSM Tree

### Scale and Growth
4. **What's your data size?**
   - < 1TB → Either works
   - > 10TB → LSM Tree advantages grow
   - > 100TB → LSM Tree strongly preferred

5. **What's your write throughput?**
   - < 10K writes/sec → B+ Tree is fine
   - > 100K writes/sec → LSM Tree excels
   - > 1M writes/sec → LSM Tree required

### Consistency and Durability
6. **What are your consistency requirements?**
   - Banking/financial → B+ Tree (ACID)
   - Social media → LSM Tree (eventual)
   - E-commerce → Either (design-dependent)

7. **How do you handle conflicts?**
   - Need transactions → B+ Tree
   - Last-write-wins OK → LSM Tree
   - Event sourcing → LSM Tree

### Operational Constraints
8. **What's your operational expertise?**
   - Traditional DBA skills → B+ Tree
   - NoSQL experience → LSM Tree
   - Limited ops resources → Managed services

9. **What's your infrastructure?**
   - SSDs available → Both work well
   - HDDs only → LSM Tree (sequential I/O)
   - Cloud with IOPS limits → LSM Tree

## Real-World Examples

### B+ Tree Success Stories
- **Banking System**: PostgreSQL for ACID transactions - customers check both recent and historical transactions equally
- **E-commerce Platform**: MySQL for inventory management - product lookups independent of when added
- **Analytics Dashboard**: SQL Server for complex reporting - queries span entire historical dataset
- **Medical Records**: Oracle for patient history - 20-year-old records as important as today's

### LSM Tree Success Stories
- **Time-Series Monitoring**: InfluxDB for metrics - 99% queries for last 24 hours, 1% for historical
- **Message Queue**: Cassandra for messaging - recent messages hot, old messages rarely accessed
- **CDN Logs**: ScyllaDB for billions of events - real-time analysis, historical data compressed/archived
- **Social Feed**: RocksDB for user timeline - exponential decay in access frequency by post age
- **IoT Sensors**: HBase for sensor data - recent readings critical, historical data for trends only

## Conclusion

The choice between B+ Tree and LSM Tree databases isn't just about performance—it's about aligning the storage engine with your specific use case. B+ Trees excel at balanced workloads with complex queries, while LSM Trees dominate write-intensive scenarios at scale. Understanding these trade-offs helps you make an informed decision that will scale with your application's growth.

## Key Insights: Time-Based Access Patterns

### LSM Trees and Temporal Locality

**Why LSM Trees Excel at Recent Data Access:**
1. **Memory-First Architecture**
   - Recent writes are in MemTable (RAM) - microsecond access
   - No disk I/O needed for very recent data
   - Perfect for "hot" data that's frequently accessed soon after writing

2. **Natural Time Ordering**
   - Newer SSTables contain more recent data
   - Compaction preserves temporal locality
   - Level 0 has the freshest data, higher levels have older data

3. **Exponential Read Performance Decay**
   ```
   Data Age → Performance Impact
   < 1 hour:    MemTable (RAM) - 10μs
   < 1 day:     Level 0 SSTable - 100μs  
   < 1 week:    Level 1-2 - 1ms
   < 1 month:   Level 3-4 - 10ms
   > 1 month:   Level 5+ - 100ms+
   ```
   
4. **Real-World Applications**
   - **Social Media Feeds**: Recent posts accessed 1000x more than old ones
   - **Log Analytics**: Last 24 hours queried most frequently
   - **IoT Sensors**: Real-time data processing with historical archival
   - **Financial Trading**: Recent trades hot, historical data cold

### B+ Trees and Uniform Access

**Why B+ Trees Provide Consistent Performance:**
1. **Predictable Access Path**
   - Every record requires the same number of tree levels to traverse
   - O(log N) complexity regardless of data age
   - No performance penalty for accessing old vs new data

2. **Write Cost Breakdown**
   ```
   B+ Tree Write Operations:
   1. Tree Traversal: O(log N) - Find leaf page
   2. Page Modification: O(1) - Update in place
   3. Potential Split: O(log N) - Propagate up tree
   4. Index Updates: O(k log N) - k secondary indexes
   ```

3. **Rebalancing Overhead**
   - **Page Splits**: Occur when pages reach capacity (typically 50-90% full)
   - **Node Splits**: Can cascade up the tree, worst case touching O(log N) nodes
   - **Merge Operations**: When deletes leave pages under-utilized
   - **Impact**: 1-5% of writes trigger splits, but each split is expensive

4. **Optimization Strategies**
   - **Fill Factor**: Leave space in pages to reduce splits (e.g., 70% full)
   - **Batch Loading**: Pre-sort data to build optimal tree structure
   - **Deferred Splits**: Some implementations delay splits until necessary
   
## Summary: Choose Based on Access Patterns

### Choose LSM Trees When:
- ✅ Recent data accessed 10-100x more than old data
- ✅ Can tolerate slower access to historical data
- ✅ Write throughput is critical (>100K writes/sec)
- ✅ Time-series or append-only workloads
- ✅ Storage efficiency matters (compression friendly)

### Choose B+ Trees When:
- ✅ Uniform access across all data regardless of age
- ✅ Need consistent, predictable query latency
- ✅ Complex queries with multiple access patterns
- ✅ Cannot tolerate performance degradation for old data
- ✅ Update-heavy workloads (not just inserts)