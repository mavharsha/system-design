# Apache Cassandra (Generated need to review)

## Overview
Apache Cassandra is a distributed NoSQL database designed for handling large amounts of data across many commodity servers, providing high availability with no single point of failure.

## Cluster Architecture

### Node Structure
- **Nodes**: Individual servers in the Cassandra cluster
- **Datacenter**: Collection of related nodes
- **Cluster**: Complete set of one or more datacenters

### Ring Architecture
- Cassandra uses a **peer-to-peer** distributed system
- All nodes are equal - **no master/slave**
- Nodes arranged in a **ring topology**
- Data distributed across nodes using **consistent hashing**

### Data Distribution
- **Partition Key**: Determines which node stores the data
- **Token Range**: Each node is responsible for a range of hash values
- **Virtual Nodes (vnodes)**: Each physical node handles multiple token ranges for better distribution

### Replication
- **Replication Factor (RF)**: Number of copies of data across the cluster
  - RF=3 means 3 copies of each data piece
- **Replication Strategy**:
  - **SimpleStrategy**: For single datacenter (not for production)
  - **NetworkTopologyStrategy**: For multiple datacenters (production use)

## Coordinator Node

### Role
- **Any node can be a coordinator** for a given request
- The coordinator is the node that receives the client request
- Doesn't store data itself (for that request) - just orchestrates

### Responsibilities
1. **Route requests** to appropriate nodes based on partition key
2. **Collect responses** from replica nodes
3. **Return result** to client once consistency level is met
4. Handle **read repair** if inconsistencies detected
5. Manage **hinted handoff** for temporarily unavailable nodes

### Request Flow
```
Client → Coordinator → Replica Nodes → Coordinator → Client
```

## Tunable Consistency

### Concept
Cassandra allows you to **tune consistency per-operation**, balancing between:
- **Consistency**: How many nodes must acknowledge
- **Availability**: System remains operational
- **Partition Tolerance**: Works despite network failures

### Write Consistency Levels
- **ANY**: At least one node (can be hinted handoff)
- **ONE**: At least one replica node must acknowledge
- **TWO**: At least two replica nodes
- **THREE**: At least three replica nodes
- **QUORUM**: Majority of replicas (RF/2 + 1)
  - RF=3 → requires 2 nodes
  - RF=5 → requires 3 nodes
- **LOCAL_QUORUM**: Quorum in local datacenter only
- **EACH_QUORUM**: Quorum in each datacenter
- **ALL**: All replica nodes must acknowledge

### Read Consistency Levels
- **ONE**: Return from first responding replica
- **TWO**: Return after two replicas respond with matching data
- **THREE**: Return after three replicas respond
- **QUORUM**: Wait for majority of replicas
- **LOCAL_QUORUM**: Quorum in local datacenter
- **ALL**: All replicas must respond

### Strong Consistency Formula
To achieve strong consistency (like CP in CAP theorem):
```
R + W > RF
```
Where:
- R = Read consistency level
- W = Write consistency level
- RF = Replication factor

Example: RF=3, W=QUORUM(2), R=QUORUM(2) → 2+2 > 3 ✓ Strong consistency

## Writes in Cassandra

### Write Path
1. **Client sends write** to coordinator
2. **Coordinator determines replicas** using partition key
3. **Write to CommitLog** (append-only log on disk) - for durability
4. **Write to MemTable** (in-memory structure)
5. **Acknowledge to client** once consistency level met
6. **MemTable flush to SSTable** when threshold reached

### Write Components
```
Write → CommitLog (disk) → MemTable (memory) → SSTable (disk)
```

### Key Features
- **Always append, never update in place**
- **No read before write** - extremely fast writes
- **Timestamps** resolve conflicts (last-write-wins)
- **TTL support** for automatic expiration

### Hinted Handoff
- If replica node is down, coordinator stores a "hint"
- When node comes back up, hint is replayed
- Ensures eventual consistency

## Reads in Cassandra

### Read Path
1. **Client sends read** to coordinator
2. **Coordinator contacts replicas** based on consistency level
3. **Check MemTable** first (most recent data)
4. **Check Bloom Filters** (quickly determine if data might be in SSTable)
5. **Check SSTables** if needed (may need to check multiple)
6. **Merge results** from multiple SSTables (reconcile by timestamp)
7. **Return result** to client

### Read Optimization Techniques
- **Bloom Filters**: Probabilistic data structure to avoid unnecessary disk reads
- **Key Cache**: Caches partition key locations
- **Row Cache**: Caches entire rows
- **Partition Summary**: In-memory index of partition keys

### Read Repair
- When inconsistencies detected during read
- Coordinator pushes latest version to out-of-date replicas
- Happens in background (doesn't delay read response)

### Types of Reads
- **Direct Read**: Fetch full data from closest replica
- **Digest Read**: Fetch hash/checksum from other replicas for comparison
- **Background Read Repair**: Fix inconsistencies asynchronously

## LSM Tree (Log-Structured Merge Tree)

### Concept
Cassandra uses **LSM tree** storage architecture for high write throughput

### Structure
```
Writes → MemTable (RAM) → SSTable (Disk) → Compaction
```

### Components

#### MemTable
- **In-memory** sorted data structure
- Holds recent writes
- Flushed to disk as SSTable when:
  - Size threshold reached
  - CommitLog size exceeds threshold
  - Manual flush triggered

#### SSTable (Sorted String Table)
- **Immutable** files on disk
- Once written, never modified
- Contain sorted key-value pairs
- Multiple SSTables can exist for same data

#### CommitLog
- **Append-only** log for durability
- Replay on node restart to recover MemTable
- Can be safely deleted after MemTable flushed to SSTable

### Compaction
**Problem**: Multiple SSTables for same partition key over time

**Solution**: Compaction merges SSTables

#### Compaction Strategies
1. **SizeTieredCompactionStrategy (STCS)**
   - Merges SSTables of similar size
   - Good for write-heavy workloads
   - Can cause space amplification

2. **LeveledCompactionStrategy (LCS)**
   - Organizes SSTables into levels (like LevelDB)
   - Better read performance
   - More predictable space usage
   - Higher I/O overhead

3. **TimeWindowCompactionStrategy (TWCS)**
   - For time-series data
   - Groups data by time window
   - Easy to drop old time windows (TTL)

### Why LSM for Cassandra?
- **Fast writes**: Append-only, no random writes
- **Sequential I/O**: Much faster than random I/O
- **High write throughput**: Critical for distributed systems
- **Trade-off**: Reads can be slower (check multiple SSTables)

## Secondary Indexes

### Overview
Secondary indexes allow querying on columns **other than** the partition key

### Important Notes
⚠️ **Use with caution** - they have limitations in distributed systems

### How They Work
- Cassandra creates a **hidden table** for the index
- Maps indexed column values → partition keys
- Requires **scatter-gather** across all nodes

### Limitations
1. **Performance Issues**
   - Query touches **all nodes** (or many nodes) in cluster
   - Not recommended for:
     - High cardinality columns (e.g., timestamps, unique IDs)
     - Low cardinality columns in large datasets (e.g., boolean with billions of rows)

2. **Not Recommended For**
   - Columns frequently updated
   - Columns with high cardinality (many unique values)
   - Columns with very low cardinality (few unique values)
   - Large partitions

3. **Best Use Cases**
   - Medium cardinality columns
   - Known partition key + filtering on another column
   - Small datasets

### Creating Secondary Index
```cql
CREATE INDEX ON users (email);
CREATE INDEX user_status_idx ON users (status);
```

### Alternatives to Secondary Indexes
1. **Materialized Views**: Automatically maintained denormalized views
2. **Manual Denormalization**: Create separate tables with different partition keys
3. **Application-level indexing**: Use external search engine (Elasticsearch, Solr)

### SSTable Attached Secondary Index (SASI)
- Improved secondary index implementation
- Better performance characteristics
- Supports prefix and contains queries
- Still has limitations of secondary indexes

## Best Practices Summary

### Do's
- ✅ Use appropriate replication factor (typically 3)
- ✅ Choose partition keys carefully for even distribution
- ✅ Understand your consistency requirements
- ✅ Denormalize data based on query patterns
- ✅ Use time-series data with TWCS compaction
- ✅ Monitor compaction and tune as needed

### Don'ts
- ❌ Don't use secondary indexes for high cardinality columns
- ❌ Don't use ALL consistency for high write throughput needs
- ❌ Don't rely on read-before-write patterns
- ❌ Don't create large partitions (>100MB)
- ❌ Don't forget to plan for tombstones and TTL

## Key Takeaways
1. **Masterless architecture** - all nodes equal, no SPOF
2. **Tunable consistency** - balance CAP per operation
3. **Write-optimized** - LSM tree enables extremely fast writes
4. **Eventual consistency by default** - but can achieve strong consistency
5. **Denormalization is normal** - design tables per query pattern
6. **Secondary indexes are limited** - consider alternatives

