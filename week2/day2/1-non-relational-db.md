## NoSQL
- Columnar
- Graph
- Wide Column

---

## Row-Oriented DB (OLTP)

**How is data stored:**
- Stored row by row: all columns of a row together on disk

**How data is represented:**

Example: Users table with id, name, age
```
[1, "Alice", 25][2, "Bob", 30][3, "Carol", 28]
```

---

## Column-Oriented DB (OLAP)

**How is data stored:**
- Stored column by column: all values of a column together on disk

**How data is represented:**

Example: Same users table
```
id: [1, 2, 3]
name: ["Alice", "Bob", "Carol"]
age: [25, 30, 28]
```

---

## Graph DBs

**Model:**
- Nodes (entities) + Edges (relationships) + Properties
- Example: `(Person:Alice) -[WORKS_AT]-> (Company:Google)`

**Use Case:**
- Find friends of friends of friends who work at same company and share two hobbies
- Relational DB → joins grow exponentially, complex queries
- Graph DB → traversing a relationship is O(1)
  - `A -[is_friends_with]-> B`

**Storage (Index-Free Adjacency):**
- Each node physically contains direct references to its adjacent nodes and relationships
- No index lookups needed for traversals

**When to Use:**
- Social networks, recommendation engines (collaborative filtering), fraud detection, knowledge graphs

**Examples:** Neo4j, Amazon Neptune, Dgraph, TigerGraph

---

## Wide Column DB

**Storage:**
- LSM-based (Log-Structured Merge tree)
  - Fast writes (append-only, no random disk I/O)
  - Periodic compaction merges sorted files

**Consistency:**
- Avoids ACID and locking
- Tunable eventually consistency
- Trade-off: availability + partition tolerance over consistency (AP in CAP)

**Replication:**
- Configurable consistency levels per operation:
  - ONE: fastest, least consistent
  - QUORUM: majority of replicas (balance)
  - ALL: slowest, most consistent
  - LOCAL_QUORUM: quorum within datacenter

**When to Use:**
- High write throughput, massive scale, multi-datacenter
- Time-series data, IoT, messaging

**Example:** Cassandra

---

### Cassandra Write Example

**Setup:**
- 3 nodes: Node A, Node B, Node C
- Replication Factor (RF) = 3 (data copied to 3 nodes)
```sql
// Cassandra CQL command to create a namespace (keyspace) with replication factor 3:
CREATE KEYSPACE my_keyspace WITH REPLICATION = {
  'class': 'SimpleStrategy',
  'replication_factor': 3
};
```


**Write Operation:**
Client writes: `INSERT INTO users (id, name) VALUES (1, 'Alice')`

**Different Consistency Levels:**

1. **Consistency Level = ONE**
   - Client sends write to coordinator
   - Coordinator sends to all 3 replicas (A, B, C)
   - Waits for 1 acknowledgment
   - Returns success to client
   - **Result:** Fast, but if Node A acknowledges and crashes before replicating, data may be lost

2. **Consistency Level = QUORUM** (RF/2 + 1 = 2)
   - Coordinator sends to all 3 replicas
   - Waits for 2 acknowledgments (majority)
   - Returns success to client
   - **Result:** Balanced - ensures majority have data before confirming

3. **Consistency Level = ALL**
   - Coordinator sends to all 3 replicas
   - Waits for all 3 acknowledgments
   - Returns success to client
   - **Result:** Slowest, but highest consistency. If any node is down, write fails

**Replication Factor Impact:**
- RF=1: No redundancy, single point of failure
- RF=3: Can tolerate 2 node failures (for reads with CL=ONE)
- Higher RF = more durability, more storage cost


**Example: Orders Table with Partition Key and Sort Key**

Suppose you have an `orders` table. In Cassandra (and other non-relational databases), you design tables based on how you want to access the data.

#### 1. Table optimized for listing all orders by a customer (partition key: `customer_id`, sort key: `order_date`)
```sql
CREATE TABLE orders_by_customer_id (
    customer_id UUID,      -- Partition Key
    order_date  TIMESTAMP, -- Clustering (Sort) Key
    order_id    UUID,
    total       DECIMAL,
    status      TEXT,
    PRIMARY KEY (customer_id, order_date)
);
-- This table is optimized for queries like:
-- "Get all orders by customer X, sorted by date"
```

#### 2. Table optimized for quick lookup by order id (partition key: `order_id`)
```sql
CREATE TABLE orders_by_order_id (
    order_id    UUID PRIMARY KEY,  -- Partition Key (no clustering key)
    customer_id UUID,
    order_date  TIMESTAMP,
    total       DECIMAL,
    status      TEXT
);
-- This table is optimized for queries like:
-- "Get details for order id X"
```

> In Cassandra, it's common to duplicate data across multiple tables, with each table tailored to a specific access pattern.

### Secondary Indexes

**Definition:**
- Allow querying by columns that are *not* part of the primary key
- Creates internal structure mapping non-primary key column values to rows
- Used when only `orders_by_customer_id` table. Instead of creating other tables, create extra secondary indexes. Tradeoff is that when indexes are used, reads are slower as they have to scatter and gather
**Example:**
Query orders by `status` (e.g., "all PENDING orders"):
```sql
CREATE INDEX ON orders_by_customer_id (status);

-- Now you can query:
SELECT * FROM orders_by_customer_id WHERE status = 'PENDING';
```

**How They Work:**
- **Local indexes:** Each node only indexes its own data
- **Scatter-gather:** Query hits all nodes, each returns matching local results, coordinator gathers and returns
- More resource-intensive than partition key queries

**Why are they called secondary indexes? Are there things called primary indexes?**

- The term **secondary index** refers to an index created on columns *other than* those making up the table's primary key. 
- The **primary index** is usually the data structure determined by the *primary key* you specify when defining a table. This key dictates how data is distributed and stored (e.g., which node stores which rows in a distributed system, or how data is physically sorted on disk).

- **Secondary indexes** allow querying by other, non-primary-key columns, offering additional query flexibility at the cost of some efficiency.
- So:  
  - **Primary index** = Provided by the table's *primary key*; it's the main access method for retrieving rows.  
  - **Secondary index** = An *extra* data structure to access rows by non-primary-key columns.

**In summary:**  
They're called "secondary" because the primary key/index comes first and fundamentally organizes the data. Any other index is "secondary"—a supplement for added flexibility.


**Limitations:**
- Best when indexed value has high cardinality (many unique values)
- Performance degrades with:
  - Low selectivity (many matching results)
  - Common values across distributed data
- At high scale, prefer duplicating data into query-specific tables

**Trade-off:** Flexibility vs. performance. Design tables for access patterns first.



---
**Reading:**
- How Redshift stores data or how BigQuery works
