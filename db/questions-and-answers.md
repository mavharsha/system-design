# Questions and Answers (generated needs review)

## 1. How do indexes make reads faster?

**Short answer**: Indexes are data structures that create shortcuts to find data without scanning the entire table.

### Without Index (Full Table Scan)
```
SELECT * FROM users WHERE email = 'john@example.com';

Database has to:
- Read row 1: Check email? No
- Read row 2: Check email? No
- Read row 3: Check email? No
- ... scan ALL rows until found
- Time complexity: O(n)
```

### With Index (B+Tree)
```
Index on email column creates a B+Tree:

                    [m]
                   /   \
            [a-l]         [n-z]
           /  |  \       /  |  \
     [a-d][e-h][i-l] [n-q][r-u][v-z]
       |
       └─> john@example.com → points to Row ID 42

Steps:
1. Navigate tree (3-4 comparisons)
2. Get row pointer
3. Read exact row
Time complexity: O(log n)
```

### Key Points

**Why B+Trees?**
- Logarithmic search time
- All leaf nodes linked (great for range queries)
- Fits disk block size well (minimize I/O)
- Self-balancing

**Example Numbers:**
- Table with 1 million rows
- Full scan: 1,000,000 comparisons
- B+Tree index (height 4): ~4 comparisons
- 250,000x faster!

**Trade-offs:**
- ✅ Faster reads
- ❌ Slower writes (index must be updated)
- ❌ More storage space
- ❌ Index maintenance overhead

**Types of indexes:**
- **Clustered**: Data stored in index order (one per table)
- **Non-clustered**: Index separate from data (multiple per table)
- **Composite**: Index on multiple columns
- **Covering**: Index contains all needed columns (no table lookup needed!)

**When indexes don't help:**
- Small tables (scan is fast anyway)
- High cardinality columns (many duplicates)
- Wildcard prefix searches: `LIKE '%john%'`
- Functions on columns: `WHERE UPPER(name) = 'JOHN'`

---

## 2. Sharding vs Partitioning

Both split data into smaller chunks, but different scope and use cases.

### Partitioning

**What**: Splitting a table into smaller pieces within the **same database instance**.

**Types:**

1. **Horizontal Partitioning** (most common)
   ```
   users table split by range:
   
   Partition 1: users with id 1-1000000
   Partition 2: users with id 1000001-2000000
   Partition 3: users with id 2000001-3000000
   
   All in same database server
   ```

2. **Vertical Partitioning**
   ```
   users table split by columns:
   
   Partition 1: id, name, email (frequently accessed)
   Partition 2: id, bio, preferences (rarely accessed)
   ```

**Strategies:**
- Range: `id < 1000`, `id >= 1000 AND id < 2000`
- List: `country IN ('USA', 'Canada')`, `country IN ('UK', 'Germany')`
- Hash: `HASH(user_id) % 4`

**Benefits:**
- Faster queries (scan only relevant partition)
- Better index management (smaller indexes per partition)
- Easier archival (drop old partitions)
- Query optimizer can prune partitions

**Limitations:**
- Still single server (single point of failure)
- Doesn't scale beyond one machine's capacity
- Same CPU, RAM, disk limits

### Sharding

**What**: Splitting data across **multiple database servers** (distributed system).

```
Application
    ↓
Shard Router/Proxy
    ↓
├─ Shard 1 (server1): users 0-999,999
├─ Shard 2 (server2): users 1M-1.9M
└─ Shard 3 (server3): users 2M-2.9M

Each shard = separate physical/logical database
```

**Sharding Keys:**
```
user_id = 12345
shard = hash(user_id) % num_shards
or
shard = user_id / 1000000  (range-based)
or
shard = country_code  (geo-based)
```

**Benefits:**
- **Horizontal scalability**: Add more servers for more capacity
- **Distributed load**: CPU, RAM, I/O spread across machines
- **Fault isolation**: One shard fails, others still work
- **Geo-distribution**: Users in US → US shard, EU → EU shard

**Challenges:**
- **Cross-shard queries**: Need to query multiple shards and merge results
- **Distributed transactions**: ACID across shards is hard
- **Resharding**: Changing shard count is painful
- **Hotspots**: Uneven data distribution (celebrity problem)
- **Joins**: Cross-shard joins are expensive or impossible
- **Complexity**: Application needs sharding logic

### Key Differences

| Aspect | Partitioning | Sharding |
|--------|-------------|----------|
| **Scope** | Single database | Multiple databases |
| **Scaling** | Vertical (bigger machine) | Horizontal (more machines) |
| **Transparency** | DB handles it automatically | App usually aware |
| **Complexity** | Low | High |
| **Failover** | Single point of failure | Can survive shard failures |
| **Joins** | Easy (same DB) | Hard/impossible |
| **Use case** | Performance optimization | Scale beyond single machine |

### When to use what?

**Partitioning**: 
- Data fits on one server but queries are slow
- Want better query performance with minimal complexity
- Need easy archival strategy

**Sharding**:
- Data doesn't fit on one server
- Need to scale beyond single machine limits
- Can tolerate application complexity
- Traffic exceeds single server capacity

**Analogy:**
- **Partitioning** = Organizing a big filing cabinet into multiple drawers (still one cabinet)
- **Sharding** = Splitting files across multiple filing cabinets in different offices

---

## 3. Local vs Global Indexes in Partitioned/Sharded DB

When you partition/shard data, you need to decide how to handle indexes.

### Local Indexes (Partitioned Indexes)

**What**: Each partition/shard has its own index covering only its data.

```
Users Table (sharded by user_id)

Shard 1 (users 0-999K):
  - Local index on user_id (only users 0-999K)
  - Local index on email (only emails in this shard)

Shard 2 (users 1M-1.9M):
  - Local index on user_id (only users 1M-1.9M)
  - Local index on email (only emails in this shard)

Shard 3 (users 2M-2.9M):
  - Local index on user_id (only users 2M-2.9M)
  - Local index on email (only emails in this shard)
```

**Characteristics:**
- Index and data are co-located
- Each shard maintains its own indexes independently
- Index partitioned same way as the data

**Pros:**
- ✅ **Fast writes**: Only update one shard's index
- ✅ **Maintenance easy**: Index maintenance is local
- ✅ **Parallel operations**: Each shard independent
- ✅ **Fault isolation**: One shard's index issues don't affect others

**Cons:**
- ❌ **Scatter queries**: Query by non-partition key requires checking ALL shards
  ```sql
  SELECT * FROM users WHERE email = 'john@example.com';
  
  Have to check:
  - Shard 1's email index
  - Shard 2's email index
  - Shard 3's email index
  Then merge results (expensive!)
  ```

**Example:**
```
Query: WHERE user_id = 1500000
→ Know it's in Shard 2 (partition key)
→ Use Shard 2's local index
→ Fast! (single shard access)

Query: WHERE email = 'john@example.com'
→ Don't know which shard
→ Fan out to all 3 shards
→ Wait for all responses
→ Merge results
→ Slow! (3x network calls)
```

### Global Indexes (Non-Partitioned Indexes)

**What**: One index spans all partitions/shards - covers entire dataset.

```
Users Table (sharded by user_id)

Shard 1: users 0-999K
Shard 2: users 1M-1.9M
Shard 3: users 2M-2.9M

Global Email Index (separate service/shard):
  john@example.com → Shard 2, Row 1234567
  jane@example.com → Shard 1, Row 456789
  bob@example.com → Shard 3, Row 2345678
  
(Index covers all emails across all shards)
```

**Characteristics:**
- Single index structure covering all data
- Can be partitioned differently than the data itself
- Often requires separate index servers/storage

**Pros:**
- ✅ **Fast lookups**: Single index lookup tells you exactly where data is
  ```sql
  SELECT * FROM users WHERE email = 'john@example.com';
  
  1. Check global email index → points to Shard 2
  2. Go directly to Shard 2
  3. Read the row
  Done! (no scatter)
  ```
- ✅ **Efficient for secondary lookups**: Any column can have global index
- ✅ **Better for reporting/analytics**: Can query by any indexed field efficiently

**Cons:**
- ❌ **Slow writes**: Every write must update global index
  ```
  INSERT into Shard 2
  → Also update global email index
  → Also update global username index
  → More network hops, more latency
  ```
- ❌ **Consistency challenges**: Distributed transaction needed (data + global index)
- ❌ **Single point of contention**: Global index can become bottleneck
- ❌ **Maintenance complexity**: Index failures affect entire system
- ❌ **Storage overhead**: Separate infrastructure for global indexes

### Comparison

| Aspect | Local Indexes | Global Indexes |
|--------|---------------|----------------|
| **Write performance** | Fast (single shard) | Slow (cross-shard) |
| **Read by partition key** | Fast | Fast |
| **Read by non-partition key** | Slow (scatter-gather) | Fast (direct lookup) |
| **Consistency** | Easy (local) | Hard (distributed) |
| **Maintenance** | Independent per shard | Centralized/complex |
| **Failure impact** | Isolated | System-wide |

### Real-World Strategy

Most systems use **hybrid approach**:

```
Users sharded by user_id:
  - Local indexes on: user_id (partition key), created_at
  - Global index on: email (need unique constraint)
  - No index on: bio, preferences (rarely queried alone)

Queries:
  - By user_id → Fast (local index)
  - By email → Fast (global index, but writes pay the cost)
  - By created_at + user_id range → Fast (local index)
  - By bio → Don't support (or use search service like Elasticsearch)
```

### Decision Framework

**Use Local Indexes when:**
- Most queries include the partition/sharding key
- Write performance is critical
- Can tolerate scatter-gather for occasional queries
- Each shard is independent

**Use Global Indexes when:**
- Need to enforce global uniqueness (email, username)
- Frequently query by non-partition columns
- Read performance on secondary keys is critical
- Can tolerate write overhead

**Example Services:**
- **DynamoDB**: Local indexes only (queries must specify partition key)
- **MongoDB**: Both local and global (configurable)
- **Vitess (MySQL sharding)**: Supports both via lookup vindexes
- **Spanner**: Global indexes with strong consistency

### Practical Example

```
E-commerce orders sharded by order_id:

Shard 1: orders 1-1M
Shard 2: orders 1M-2M
Shard 3: orders 2M-3M

Indexes:
  Local index: order_id, created_at (queries always have order_id or scan recent)
  Global index: user_id → order_ids (need to lookup "all orders by user")
  
Query patterns:
  - Get order by ID → Local index (fast)
  - Get recent orders → Local index + merge (acceptable)
  - Get user's orders → Global user_id index (critical UX, worth the cost)
  - Search by product → Elasticsearch (not SQL index at all)
```

---

## Summary

1. **Indexes**: B+Tree shortcuts that turn O(n) scans into O(log n) lookups
2. **Partitioning vs Sharding**: Partitioning = one DB split up; Sharding = multiple DBs split across servers
3. **Local vs Global Indexes**: Local = fast writes, scatter reads; Global = fast all reads, slow writes

