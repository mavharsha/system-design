## Indexes

**What are Indexes?**
- Indexes are special data structures that improve the speed of data retrieval operations on a database table.
- They work like a book’s index: helping you find specific rows quickly without scanning the whole table.

**Why use Indexes?**
- **Faster queries:** Indexes make SELECT queries much faster, especially on large tables.
- **Efficient lookups:** Useful for searching, filtering (`WHERE`), joining, and sorting (`ORDER BY`) data.
- **Enforce constraints:** Unique indexes ensure data integrity by preventing duplicates.

**Trade-offs:**
- **Write overhead:** Insert, update, and delete operations are slower, as indexes must be updated.
- **Storage:** Indexes require extra disk space.

**Typical Use Cases:**
- Speeding up frequent queries on large tables
- Enforcing uniqueness of key columns (like emails or usernames)
- Optimizing performance for reporting or analytics workloads



#### Types of indexes (for SQL)

1. **Primary/Clustered indexes** - determines physical order of data
   ```sql
   CREATE TABLE users (
     id INT PRIMARY KEY,  -- Automatically creates clustered index
     name VARCHAR(100)
   );
   ```
   **When to use:**
   - Every table should have one (usually on primary key)
   - Choose column(s) frequently used in range queries or ORDER BY
   - Ideal for sequential data like timestamps, auto-increment IDs
   - Only ONE per table (data can only be physically sorted one way)

2. **Secondary/Non-clustered indexes** - separate structure with pointers to data
   ```sql
   CREATE INDEX idx_users_email ON users(email);
   ```
   **When to use:**
   - Columns frequently used in WHERE, JOIN, or ORDER BY clauses
   - When you need fast lookups on non-primary key columns
   - Example: Looking up users by email, orders by status, posts by author
   - Can have multiple per table

3. **Unique indexes** - enforce uniqueness constraint
   ```sql
   CREATE UNIQUE INDEX idx_users_username ON users(username);
   -- Ensures no duplicate usernames
   ```
   **When to use:**
   - Enforce business rules (no duplicate emails, usernames, SSNs)
   - Natural keys that must be unique (phone numbers, license plates)
   - Provides both data integrity AND fast lookups
   - Alternative to UNIQUE constraint with more control

4. **Composite indexes** - index on multiple columns
   ```sql
   CREATE INDEX idx_orders_user_date ON orders(user_id, order_date);
   -- Useful for queries filtering by both user_id and order_date
   ```
   **When to use:**
   - Queries that filter on multiple columns together
   - Example: `WHERE user_id = 123 AND order_date > '2024-01-01'`
   - Column order matters! Index (A, B) helps queries on A or (A, B), but NOT just B
   - Common pattern: (tenant_id, created_at) for multi-tenant apps

5. **Partial/Filtered indexes** - index with WHERE clause condition
   ```sql
   CREATE INDEX idx_active_users ON users(email) WHERE status = 'active';
   -- Only indexes active users, smaller and faster
   ```
   **When to use:**
   - Queries that always filter on a specific condition
   - Large tables where only small subset is queried (active vs archived)
   - Examples: WHERE is_deleted = false, WHERE payment_status = 'pending'
   - Saves disk space and improves performance (smaller index)

6. **Covering indexes** - includes all columns needed for a query
   ```sql
   CREATE INDEX idx_orders_covering ON orders(user_id) INCLUDE (order_date, total_amount);
   -- Query can be satisfied entirely from the index without accessing table
   ```
   **When to use:**
   - Frequently run queries that need specific columns
   - Avoid expensive table lookups (index-only scans)
   - Example: Dashboard queries, reporting, API endpoints with fixed responses
   - Trade-off: Larger index size but much faster queries

7. **Full-text indexes** - specialized for text search
   ```sql
   CREATE FULLTEXT INDEX idx_articles_content ON articles(title, body);
   -- Enables efficient text searching: MATCH(title, body) AGAINST ('search term')
   ```
   **When to use:**
   - Search functionality in articles, posts, comments, documents
   - When users need to search by keywords or phrases
   - Better than LIKE '%keyword%' which doesn't use regular indexes
   - Supports stemming, relevance ranking, natural language search
   - Alternative: Use dedicated search engines (Elasticsearch) for complex needs


### Dense vs Sparse Indexes

**Index structures** fall into two main types: **dense** and **sparse**. Understanding when to use each is essential for tuning performance and storage.

#### 1. **Dense Index**
- Has an index entry for **every row** in the table (duplicates included).
- **Use cases:** Needed for uniqueness enforcement and very fast lookups.
- **Pros:**
  - Fast point queries (`SELECT ... WHERE id = ...`)
  - Enforces unique constraints efficiently
- **Cons:**
  - Larger index size (can affect writes and disk usage)
  - More updates on data changes

#### 2. **Sparse Index**
- Only adds index entries for **some rows**—usually the first row of each disk block or a group (common with physically sorted/clustered tables).
- **Use cases:** Range scans, big tables where only some data is frequently queried.
- **Pros:**
  - Smaller and faster index scans
  - Fewer updates when table changes
- **Cons:**
  - Slower for point lookups unless data is well-clustered
  - Cannot enforce row-level uniqueness

---

#### PostgreSQL Example

- **Default**: All standard PostgreSQL indexes (B-tree, hash, GIN, GiST) are **dense**—every matching row is indexed.
- **Clustered table:** Using `CLUSTER` physically arranges data on disk but still uses a dense B-tree index.
- **Partial indexes** (with `WHERE` conditions): Behave somewhat like sparse; only rows matching the condition are indexed, helping reduce index size for large tables.

**Example:**
```sql
-- Dense index (default in Postgres)
CREATE INDEX idx_users_email ON users(email);

-- Partial index = "sparse" on filtered rows
CREATE INDEX idx_active_users ON users(email) WHERE status = 'active';
```

---

#### When to Use Dense vs Sparse/Partial

|                        | **Dense Index**                                      | **Sparse/Partial Index**                      |
|------------------------|------------------------------------------------------|-----------------------------------------------|
| **Best for**           | Fast lookups, uniqueness, random access              | Range scans, queries on subsets               |
| **In PostgreSQL**      | All regular indexes are dense                        | Partial indexes approximate sparse indexes    |
| **Choose if…**         | You frequently query or validate all rows            | You only care about a subset (e.g., status)   |

**Key Points:**
- **PostgreSQL:** Dense by default; use partial indexes for subset queries to save space.
- **Sparse indexes:** Common in NoSQL/clustered storage, not typical in Postgres.
- For big tables with lots of unneeded rows, partial indexes can deliver many of the benefits of sparsity.

