# B+ Trees vs LSM Trees in Databases

## Overview

Database storage engines fundamentally differ in how they organize data on disk. The two dominant approaches are **B+ Trees** (optimized for reads) and **LSM Trees** (optimized for writes). Understanding their trade-offs is crucial for selecting the right database for your use case.

## Database Implementations

### B+ Tree Based Databases

| Database | Type | Use Case | Key Features |
|----------|------|----------|--------------|
| **PostgreSQL** | RDBMS | General-purpose | MVCC, full ACID, complex queries |
| **MySQL (InnoDB)** | RDBMS | Web applications | Row-level locking, foreign keys |
| **Oracle** | RDBMS | Enterprise | Advanced features, partitioning |
| **SQL Server** | RDBMS | Windows enterprise | Integrated with .NET, BI tools |
| **SQLite** | Embedded | Local storage | Single file, serverless |
| **BerkeleyDB** | Key-Value | Embedded systems | Multiple access methods |
| **MongoDB (WiredTiger)** | Document | Flexible schema | Document-oriented, horizontal scaling |
| **CockroachDB** | NewSQL | Distributed SQL | Geo-distributed, strong consistency |

### LSM Tree Based Databases

| Database | Type | Use Case | Key Features |
|----------|------|----------|--------------|
| **Cassandra** | Wide-column | High write throughput | Distributed, eventually consistent |
| **ScyllaDB** | Wide-column | Ultra-low latency | C++ rewrite of Cassandra |
| **RocksDB** | Key-Value | Embedded | Facebook's fork of LevelDB |
| **LevelDB** | Key-Value | Embedded | Google's lightweight KV store |
| **HBase** | Wide-column | Big data | Hadoop integration, strong consistency |
| **BigTable** | Wide-column | Google-scale | Original LSM implementation |
| **TiKV** | Key-Value | Distributed KV | ACID transactions, Raft consensus |
| **DynamoDB** | Key-Value | AWS managed | Serverless, auto-scaling |
| **InfluxDB** | Time-series | Metrics/monitoring | Optimized for time-series data |

## How B+ Trees Improve Reads and Writes

### Read Performance Advantages

1. **Direct Access Path**
   - O(log N) lookups with predictable I/O patterns
   - Each level narrows the search space efficiently
   - Typically 3-4 disk seeks for any record in a billion-row table

2. **Cache-Friendly Structure**
   - Internal nodes contain only keys, maximizing cache utilization
   - Hot paths (root and upper levels) stay in memory
   - Sequential leaf nodes enable efficient range scans

3. **In-Place Updates**
   - No need to search through multiple files
   - Single source of truth for each record
   - No read amplification from checking multiple versions

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

2. **Write Amplification Reduction**
   - Batching writes before flushing to disk
   - No in-place updates means no page splits
   - Append-only nature eliminates seek time

3. **High Throughput**
   - Can sustain millions of writes per second
   - Writes never block on disk I/O (until memory pressure)
   - Perfect for write-heavy workloads

### Read Performance Characteristics

1. **Read Amplification**
   - Must check MemTable + multiple SSTable levels
   - Bloom filters reduce unnecessary disk reads
   - More files to check = slower reads

2. **Optimization Techniques**
   - Bloom filters (probabilistic data structure)
   - Block cache for frequently accessed data
   - Compaction to merge files and reduce levels

3. **Range Scans**
   - Requires merging results from multiple files
   - Sorted nature of SSTables helps
   - Less efficient than B+ tree sequential scans

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
- [ ] Need consistent read latency
- [ ] Complex queries with multiple indexes
- [ ] Frequent updates to existing records
- [ ] Small to medium dataset that fits mostly in memory

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
- [ ] Can tolerate some read latency variance
- [ ] Append-only or time-series data
- [ ] Massive scale (TBs to PBs)
- [ ] High ingestion rate (100K+ writes/sec)

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
1. **What's your read/write ratio?**
   - 90% reads → B+ Tree
   - 90% writes → LSM Tree
   - Mixed → Depends on other factors

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
- **Banking System**: PostgreSQL for ACID transactions
- **E-commerce Platform**: MySQL for inventory management
- **Analytics Dashboard**: SQL Server for complex reporting

### LSM Tree Success Stories
- **Time-Series Monitoring**: InfluxDB for metrics collection
- **Message Queue**: Cassandra for high-throughput messaging
- **CDN Logs**: ScyllaDB for billions of daily events
- **Social Feed**: RocksDB for user timeline storage

## Conclusion

The choice between B+ Tree and LSM Tree databases isn't just about performance—it's about aligning the storage engine with your specific use case. B+ Trees excel at balanced workloads with complex queries, while LSM Trees dominate write-intensive scenarios at scale. Understanding these trade-offs helps you make an informed decision that will scale with your application's growth.