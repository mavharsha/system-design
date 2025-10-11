# B+ Trees in SQL Databases

## Table of Contents
1. [Introduction](#introduction)
2. [B+ Tree Structure](#b-tree-structure)
3. [Visual Diagrams](#visual-diagrams)
4. [How SQL Uses B+ Trees](#how-sql-uses-b-trees)
5. [Database Indexes](#database-indexes)
6. [Operations](#operations)
7. [Advantages](#advantages)
8. [Real-World Examples](#real-world-examples)

---

## Introduction

B+ trees are the fundamental data structure used by most SQL databases (MySQL, PostgreSQL, SQL Server, Oracle, etc.) for storing and retrieving data efficiently. They provide **O(log n)** time complexity for search, insert, and delete operations, making them ideal for disk-based storage systems.

### Why B+ Trees?

- **Optimized for disk access**: Minimize the number of disk I/O operations
- **Balanced structure**: All leaf nodes are at the same level
- **Sequential access**: Leaf nodes are linked, enabling efficient range queries
- **High fanout**: Each node can have many children, keeping tree height low

---

## B+ Tree Structure

### Key Properties

1. **Order (m)**: Maximum number of children a node can have
2. **Internal nodes**: Store only keys and pointers to children (no actual data)
3. **Leaf nodes**: Store keys and actual data (or pointers to data)
4. **Linked leaves**: All leaf nodes are connected via a linked list
5. **Balanced**: All leaf nodes are at the same depth

### Node Structure

**Internal Node:**
```
┌──────────────────────────────────────┐
│ [Key1 | Key2 | Key3 | ... | KeyN]   │
│ [Ptr0 | Ptr1 | Ptr2 | ... | PtrN]   │
└──────────────────────────────────────┘
```

**Leaf Node:**
```
┌──────────────────────────────────────┐
│ [Key1:Data1 | Key2:Data2 | ... ]    │
│ [Next Leaf Pointer] ──────────────>  │
└──────────────────────────────────────┘
```

---

## Visual Diagrams

### Example 1: Simple B+ Tree (Order 3)

Let's say we have a table with IDs: 1, 4, 7, 10, 13, 16, 19, 22, 25

```
                    [10, 19]                    ← Root (Internal Node)
                   /    |    \
                  /     |     \
                 /      |      \
         [1,4,7]    [10,13,16]   [19,22,25]    ← Leaf Nodes (contain data)
            ↓           ↓            ↓
         [Data]      [Data]       [Data]
            
         ← - - - - - - - - - - - - - →         (Linked list for sequential access)
```

**Explanation:**
- Root node contains keys `10` and `19` (separators)
- Values < 10 go to left child
- Values >= 10 and < 19 go to middle child
- Values >= 19 go to right child
- Leaf nodes contain actual data and are linked

---

### Example 2: Detailed B+ Tree (Order 4)

Storing keys: 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55

```
                            [25]                              ← Root
                          /      \
                         /        \
                    [15]            [35, 45]                  ← Internal Nodes
                   /    \          /    |    \
                  /      \        /     |     \
            [5,10]  [15,20]  [25,30]  [35,40]  [45,50,55]   ← Leaf Nodes
              ↓       ↓        ↓        ↓         ↓
           [Data]  [Data]   [Data]   [Data]    [Data]
              
            ← - - - - - - - - - - - - - - - - - - - - →      (Linked list)
```

**Search Example: Finding key 35**
1. Start at root: 35 >= 25, go right
2. At [35, 45]: 35 < 45, go left
3. Reach leaf [35, 40]: Found 35!

**Total disk reads: 3** (tree height = 3)

---

### Example 3: B+ Tree with Actual Database Records

Consider a `Users` table:

```sql
CREATE TABLE Users (
    user_id INT PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100),
    age INT
)
```

**B+ Tree on user_id (Primary Key):**

```
                                [100]
                              /       \
                             /         \
                        [50]            [150]
                       /    \          /     \
                      /      \        /       \
              [10,30]  [50,75]  [100,125]  [150,175]
                ↓         ↓         ↓          ↓
        
        user_id: 10              user_id: 50
        name: Alice              name: Charlie
        email: alice@...         email: charlie@...
        age: 25                  age: 30
        
        user_id: 30              user_id: 75
        name: Bob                name: David
        ...                      ...
```

---

## How SQL Uses B+ Trees

### 1. **Primary Storage Structure**

In databases like InnoDB (MySQL), the entire table is stored as a B+ tree:
- **Clustered Index**: The table itself IS a B+ tree ordered by primary key
- Leaf nodes contain the complete row data
- This is why choosing a good primary key is crucial for performance

### 2. **Page-Based Storage**

```
┌─────────────────────────────────────────┐
│         Database File (on disk)         │
├─────────────────────────────────────────┤
│  Page 1 (16KB)  │  Page 2  │  Page 3   │  ← Fixed-size pages
├─────────────────────────────────────────┤
│  B+ Tree Node   │  B+ Node │  B+ Node  │
└─────────────────────────────────────────┘
```

- Each B+ tree node fits into one or more disk pages
- Typical page size: 4KB, 8KB, or 16KB
- Order of B+ tree is determined by page size and key size

### 3. **Calculating Fanout**

For a page size of 16KB and key+pointer size of 12 bytes:
```
Fanout = 16,384 bytes / 12 bytes ≈ 1,365 children per node
```

**Impact on tree height:**
- 1 million records: ~2 levels (1,365² = 1,863,225)
- 1 billion records: ~3 levels (1,365³ = 2.5 billion)

**This is why B+ trees are so efficient!**

---

## Database Indexes

### What is an Index?

An index is a separate data structure (usually a B+ tree) that improves the speed of data retrieval operations.

```
┌─────────────────────────────────────────────────────────┐
│                     Table (Heap)                        │
│  ┌──────┬───────┬────────────┬─────┬─────────────────┐ │
│  │ ID   │ Name  │ Email      │ Age │ City            │ │
│  ├──────┼───────┼────────────┼─────┼─────────────────┤ │
│  │ 101  │ Alice │ alice@...  │ 25  │ New York        │ │
│  │ 205  │ Bob   │ bob@...    │ 30  │ San Francisco   │ │
│  │ 150  │ Carol │ carol@...  │ 28  │ Boston          │ │
│  └──────┴───────┴────────────┴─────┴─────────────────┘ │
└─────────────────────────────────────────────────────────┘
                            ↑
                            │
            ┌───────────────┴────────────────┐
            │                                 │
    ┌───────┴────────┐              ┌────────┴────────┐
    │  Index on ID   │              │ Index on Email  │
    │   (B+ Tree)    │              │   (B+ Tree)     │
    └────────────────┘              └─────────────────┘
```

### Types of Indexes

#### 1. **Clustered Index (Primary Index)**

- **Definition**: The physical order of rows matches the index order
- **Storage**: The table itself is organized as a B+ tree
- **Limit**: Only ONE per table (because data can only be physically sorted one way)
- **Performance**: Fastest for range queries on the indexed column

**Diagram:**
```
Clustered Index on user_id:
[B+ Tree Structure]
      ↓
   [1] → [Complete Row: 1, Alice, alice@email.com, 25]
   [2] → [Complete Row: 2, Bob, bob@email.com, 30]
   [5] → [Complete Row: 5, Carol, carol@email.com, 28]
```

#### 2. **Non-Clustered Index (Secondary Index)**

- **Definition**: Separate structure from the table
- **Storage**: B+ tree where leaf nodes contain pointers to actual data
- **Limit**: Multiple per table
- **Performance**: Requires extra lookup to get full row

**Diagram:**
```
Non-Clustered Index on email:
[B+ Tree Structure]
      ↓
   [alice@...] → Pointer to Row with user_id=1
   [bob@...]   → Pointer to Row with user_id=2
   [carol@...] → Pointer to Row with user_id=5
                           ↓
                    [Actual Row Data in Table]
```

### Index Example in SQL

```sql
-- Create a non-clustered index on email
CREATE INDEX idx_email ON Users(email)

-- Create a composite index (multi-column)
CREATE INDEX idx_city_age ON Users(city, age)
```

**Composite Index B+ Tree Structure:**
```
                     [(Boston,25)]
                    /              \
                   /                \
     [(Atlanta,20),(Atlanta,30)]  [(Boston,28),(Chicago,35)]
              ↓                            ↓
        Pointers to rows             Pointers to rows
```

**Query optimization:**
```sql
-- This query uses idx_city_age efficiently
SELECT * FROM Users WHERE city = 'Boston' AND age = 28

-- This query can partially use idx_city_age (only city part)
SELECT * FROM Users WHERE city = 'Boston'

-- This query CANNOT use idx_city_age (age is not the leading column)
SELECT * FROM Users WHERE age = 28
```

### Covering Index

A **covering index** contains all columns needed by a query, avoiding table lookup:

```sql
CREATE INDEX idx_email_name ON Users(email, name)

-- This query is fully covered by the index (no table access needed!)
SELECT name FROM Users WHERE email = 'alice@email.com'
```

---

## Operations

### 1. Search Operation

**Example: Search for key 35**

```
Step 1: Read root page
                [25]
              /      \
             /        \
        [15]            [35,45]  ← 35 >= 25, go right
        
Step 2: Read internal node
        [35,45]  ← 35 < 45, go to first child
       /   |   \
       
Step 3: Read leaf node
     [35,40]  ← Found! Return data
```

**Pseudocode:**
```python
def search(key, node):
    if node.is_leaf:
        return node.find(key)
    else:
        child_index = node.find_child_index(key)
        return search(key, node.children[child_index])
```

**Time Complexity: O(log n)**  
**Disk I/O: O(log_m n)** where m is the tree order

---

### 2. Range Query

**Example: Find all records where 15 <= key <= 40**

```
                [25]
              /      \
        [15]            [35,45]
       /    \          /   |   \
  [5,10] [15,20] [25,30] [35,40] [45,50]
            ↑________________________↑
            Start here     →      End here
            
Sequential scan using leaf linked list:
[15,20] → [25,30] → [35,40]
   ↓         ↓         ↓
Results: 15,20,25,30,35,40
```

**This is why B+ trees are superior to B-trees for range queries!**

---

### 3. Insert Operation

**Example: Insert key 23**

**Before:**
```
            [25]
           /    \
      [15]      [35]
     /    \    /    \
  [10] [15,20] [25,30] [35,40]
```

**After:**
```
            [25]
           /    \
      [15]      [35]
     /    \    /    \
  [10] [15,20,23] [25,30] [35,40]  ← 23 inserted here
```

**Insert Algorithm:**
1. Find correct leaf node
2. If leaf has space, insert key
3. If leaf is full, **split** the node:
   - Create new node
   - Move half the keys to new node
   - Insert middle key into parent
4. If parent is full, split recursively up the tree
5. If root splits, create new root (tree height increases)

---

### 4. Delete Operation

**Example: Delete key 20**

**Before:**
```
      [15]
     /    \
  [10] [15,20,23]
```

**After:**
```
      [15]
     /    \
  [10] [15,23]  ← 20 removed
```

**Delete Algorithm:**
1. Find and remove key from leaf
2. If leaf has too few keys (< ⌈m/2⌉), **rebalance**:
   - Try to borrow from sibling
   - If sibling can't spare, **merge** with sibling
3. Update parent keys if needed
4. Propagate merging up if necessary

---

## Advantages

### 1. **Minimal Disk I/O**
- Logarithmic height means few disk reads
- Example: 1 billion records, 3 disk reads!

### 2. **Excellent for Range Queries**
```sql
SELECT * FROM Users WHERE user_id BETWEEN 100 AND 200
```
- Navigate to first key (100)
- Follow linked list to last key (200)
- Efficient sequential access

### 3. **Balanced Structure**
- Always O(log n) performance
- No worst-case scenarios (unlike binary search trees)

### 4. **Cache-Friendly**
- Internal nodes fit in memory
- Only leaf nodes need disk access
- Buffer pool caches hot pages

### 5. **Sequential Inserts are Efficient**
```sql
INSERT INTO Users VALUES (1, ...), (2, ...), (3, ...)
```
- New records go to rightmost leaf
- Minimal tree reorganization

---

## Real-World Examples

### MySQL InnoDB

```sql
-- Primary key creates clustered index (B+ tree)
CREATE TABLE Orders (
    order_id INT PRIMARY KEY,      -- Clustered B+ tree
    user_id INT,
    total DECIMAL(10,2),
    order_date DATE,
    INDEX idx_user (user_id),      -- Non-clustered B+ tree
    INDEX idx_date (order_date)    -- Non-clustered B+ tree
)
```

**Storage layout:**
```
orders.ibd (InnoDB file)
├── Clustered B+ Tree (order_id)
│   └── Leaf nodes contain full rows
├── Secondary B+ Tree (user_id)
│   └── Leaf nodes contain (user_id → order_id)
└── Secondary B+ Tree (order_date)
    └── Leaf nodes contain (order_date → order_id)
```

### PostgreSQL

PostgreSQL also uses B+ trees (called btree) as the default index type:

```sql
-- Create B-tree index (actually B+ tree)
CREATE INDEX idx_customer_email ON customers(email)

-- View index information
\d customers
```

### Performance Example

**Without Index:**
```sql
SELECT * FROM Orders WHERE user_id = 12345
-- Full table scan: Read ALL pages (could be millions!)
-- Time: O(n)
```

**With Index:**
```sql
-- Uses idx_user (B+ tree)
-- 1. Traverse B+ tree: 3-4 page reads to find user_id=12345
-- 2. Follow pointer to get full row: 1 page read
-- Time: O(log n)
-- Total disk I/O: 4-5 pages vs. millions!
```

---

## Best Practices

### 1. **Choose Primary Key Wisely**
```sql
-- Good: Sequential, compact
user_id INT AUTO_INCREMENT PRIMARY KEY

-- Bad: Large, random UUIDs cause page splits
user_id CHAR(36) PRIMARY KEY  -- UUID
```

### 2. **Use Composite Indexes for Multi-Column Queries**
```sql
-- Query pattern
SELECT * FROM Orders WHERE user_id = ? AND status = ?

-- Optimal index (most selective column first)
CREATE INDEX idx_user_status ON Orders(user_id, status)
```

### 3. **Monitor Index Usage**
```sql
-- MySQL
EXPLAIN SELECT * FROM Orders WHERE user_id = 12345

-- Check for unused indexes
SELECT * FROM sys.schema_unused_indexes
```

### 4. **Don't Over-Index**
- Indexes speed up reads but slow down writes
- Each insert/update/delete must update all indexes
- Balance between read and write performance

---

## Summary

**B+ Trees are the backbone of SQL databases because they:**

1. ✅ Provide O(log n) performance for search, insert, delete
2. ✅ Minimize disk I/O (critical for databases)
3. ✅ Support efficient range queries via linked leaves
4. ✅ Maintain balance automatically
5. ✅ Enable both clustered and non-clustered indexes
6. ✅ Scale to billions of records with minimal depth

**Key Takeaway:** Understanding B+ trees helps you:
- Design better database schemas
- Create optimal indexes
- Write efficient queries
- Debug performance issues

---

## Further Reading

- [Database Internals by Alex Petrov](https://www.databass.dev/)
- [MySQL InnoDB Architecture](https://dev.mysql.com/doc/refman/8.0/en/innodb-architecture.html)
- [PostgreSQL B-Tree Implementation](https://www.postgresql.org/docs/current/btree-implementation.html)
- [B-tree vs B+ tree differences](https://www.geeksforgeeks.org/difference-between-b-tree-and-b-tree/)

---

