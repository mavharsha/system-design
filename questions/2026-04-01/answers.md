# System Design Answers — 2026-04-01

## A1: Online/Offline Indicator Design (Week 1)

**Redis** stores `user_id -> last_heartbeat_timestamp` as a key-value pair with a TTL (e.g., 30s). If the key expires, the user is offline. Redis is fast (sub-ms reads), but volatile — data is lost on restart unless persistence is configured. Best for ephemeral status where losing data on failure is acceptable.

**DynamoDB** stores the same data but with durable writes. It handles scale automatically (no need to manage cluster size), but has higher latency (~5-10ms) and costs more per operation. Better when you need to query historical presence data or have compliance requirements.

**Handling sudden disconnects**: since the client can't send a "going offline" event when it loses connectivity, you rely on **heartbeat expiry**. The server marks a user offline if no heartbeat arrives within 2x the heartbeat interval (e.g., 40 seconds for a 20-second heartbeat). This creates a "staleness window" — the user appears online for up to 40 seconds after disconnecting. For most use cases (chat apps), this is acceptable.

**Optimization**: use a **sparse data** approach — only store entries for online users. If there's no key in Redis, the user is offline. This is more efficient than storing status for every registered user.

**WebSocket alternative**: for real-time status, maintain a WebSocket connection. The server detects disconnect immediately (TCP FIN/RST). But WebSockets are harder to scale (sticky sessions, connection state management).

---

## A2: Cache Stampede Prevention (Week 1)

A **cache stampede** (thundering herd) happens when a cache entry expires and many concurrent requests all see a cache miss simultaneously, flooding the database with identical queries.

**Prevention with distributed lock (Redis `SET NX EX`)**: when a cache miss occurs, the first request acquires a lock (`SET product:42:lock NX EX 5`). Only this request queries the database and repopulates the cache. All other requests either wait briefly and retry the cache, or return a stale value if available.

```java
String lockKey = "product:42:lock";
Boolean acquired = redisTemplate.opsForValue().setIfAbsent(lockKey, "1", 5, TimeUnit.SECONDS);
if (acquired) {
    // Only this thread fetches from DB and repopulates cache
    Product product = db.findById(42);
    redisTemplate.opsForValue().set("product:42", product, 10, TimeUnit.MINUTES);
    redisTemplate.delete(lockKey);
} else {
    // Wait and retry cache, or return stale data
}
```

**Why invalidate, not update**: if you update the cache on every write, concurrent writes can race — write A arrives, write B arrives, B updates cache, A overwrites cache with older data. Invalidation is safer: delete the cache key on write, let the next read repopulate it from the source of truth (the database). This avoids stale cache data from race conditions.

Other mitigations from the notes: **cache penetration** (requests for non-existent data bypass cache) is solved with bloom filters. **Cache avalanche** (many keys expire at once) is solved by adding jitter to TTLs.

---

## A3: Database Sharding Strategy (Week 1)

**Hash-based sharding**: `shard = hash(key) % num_shards`. Distributes data evenly with minimal metadata. But adding/removing shards requires rehashing and data migration (unless using consistent hashing). Best for even write distribution.

**Range-based sharding**: partition by ranges (e.g., users A-M on shard 1, N-Z on shard 2). Enables efficient range queries within a shard. But creates **hotspots** — if most new users have names starting with certain letters, one shard gets disproportionate load.

**Static-based sharding**: a lookup table maps each entity to a specific shard. Full control over placement, but the lookup table itself becomes a dependency and potential bottleneck.

**For user-generated content**: shard by `user_id` with hash-based sharding. All content for a user lives on one shard, so "show me my posts" is a single-shard query (fast). The hash distributes users evenly across shards.

**Cross-shard queries** (global trending feed): this is the hard part. You must scatter the query to all shards, each returns its top results, and an aggregation layer merges them. This is expensive — `O(num_shards)` fan-out. Solutions: maintain a separate denormalized table/cache for global queries, or use a search index (Elasticsearch) that indexes across all shards.

**Key takeaway from the notes**: read replicas scale reads, sharding scales writes. Don't shard until you've exhausted vertical scaling and read replicas. Shard key selection is the most important decision — a bad key creates hot shards and makes queries painful.

---

## A4: Pessimistic vs Optimistic Locking (Week 2)

**`SELECT ... FOR UPDATE`**: acquires an exclusive lock on selected rows. Other transactions trying to lock the same rows **block and wait** until the lock is released. Safe but can cause long wait times and deadlocks under high contention.

**`FOR UPDATE NOWAIT`**: same as above, but if the row is already locked, the transaction **fails immediately** with an error instead of waiting. Good for user-facing flows where you'd rather show "try again" than make the user wait.

**`FOR UPDATE SKIP LOCKED`**: skips over any rows that are currently locked and returns only unlocked rows. This is **perfect for queue-like processing** — 10,000 users all try to grab seats, and each query grabs the next available unlocked seat without blocking.

**For ticket booking**: use `FOR UPDATE SKIP LOCKED`. Each booking request runs:

```sql
BEGIN;
SELECT id FROM seats WHERE event_id = 100 AND status = 'available'
FOR UPDATE SKIP LOCKED LIMIT 1;
-- If a row is returned, book it
UPDATE seats SET status = 'booked', user_id = ? WHERE id = ?;
COMMIT;
```

No waiting, no deadlocks. If all available seats are locked by concurrent transactions, the query returns empty — the user sees "sold out" or "try again in a moment." This pattern handles 10,000 concurrent users gracefully because each transaction grabs a different row.

**Optimistic locking** (version-based) would be a poor choice here: with 10,000 users competing for 50 seats, almost every transaction would fail the version check and need to retry, causing a retry storm.

---

## A5: Cassandra Data Modeling (Week 2)

**Why secondary indexes are problematic**: in Cassandra, data is partitioned across nodes by partition key. A secondary index on `sender_id` means the index is **local to each node** — it only knows about data on that node. A query on `sender_id` triggers a **scatter-gather**: the coordinator must ask **every node** in the cluster, each checks its local index, and results are merged. With 100 nodes, that's 100 network calls for one query. This doesn't scale.

**The correct approach**: design a **separate table per access pattern**. If you query messages by sender, create a table with `sender_id` as the partition key:

```sql
CREATE TABLE messages_by_sender (
    sender_id UUID,
    message_id TIMEUUID,
    body TEXT,
    PRIMARY KEY (sender_id, message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
```

If you also query by conversation, create another table:

```sql
CREATE TABLE messages_by_conversation (
    conversation_id UUID,
    message_id TIMEUUID,
    sender_id UUID,
    body TEXT,
    PRIMARY KEY (conversation_id, message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
```

Data duplication is expected and encouraged in Cassandra. Write to both tables on every message send. Writes are cheap (append-only LSM tree).

**Consistency levels**: `ONE` = fast, eventual consistency (one replica acknowledges). `QUORUM` = majority of replicas acknowledge (strong consistency if `R + W > RF`). `ALL` = every replica acknowledges (slowest, fails if any replica is down). For messaging, write at `LOCAL_QUORUM` and read at `LOCAL_QUORUM` for a good balance of consistency and availability within a datacenter.

---

## A6: Read Replica Lag and Consistency (Week 2)

**Why it happens**: the user writes to the master, which acknowledges the write. The master then asynchronously replicates to read replicas. If the next read is routed to a replica that hasn't received the replication yet, the user sees stale data. This is **replication lag** — typically milliseconds, but can spike to seconds under heavy write load.

**Pattern 1: Read-your-writes with session stickiness**: after a write, route that user's reads to the **master** for a short window (e.g., 5 seconds). Use a cookie or session flag: `last_write_timestamp`. If `now - last_write_timestamp < 5s`, route reads to master. After the window, resume reading from replicas. This doesn't waste master capacity for all reads — only for the brief post-write window.

**Pattern 2: GTID/LSN-based routing**: after a write, the master returns a **Global Transaction ID** (MySQL GTID) or **Log Sequence Number** (PostgreSQL LSN). The client sends this with subsequent reads. The read router checks if the replica has applied at least that transaction. If yes, route to replica. If not, either wait briefly or route to master.

```java
// After write
String gtid = masterConnection.getLastGtid();
session.setAttribute("min_gtid", gtid);

// On read
String minGtid = session.getAttribute("min_gtid");
if (replica.hasApplied(minGtid)) {
    return replica.query(sql);
} else {
    return master.query(sql);  // fallback to master
}
```

**Key insight from the notes**: the master can also serve reads — don't waste its capacity by routing everything to replicas. Use two connection pools (master and replica) and route intelligently.

---

## A7: Twitter Snowflake vs Amazon Centralized ID Generation (Week 3)

**Snowflake (decentralized)**: each machine generates IDs independently using `41-bit timestamp | 10-bit machine_id | 12-bit sequence`. No network call needed — generation is local and sub-microsecond. Supports 4,096 IDs/ms per machine, so 50 machines can generate ~200K IDs/ms. IDs are naturally time-sorted (timestamp is the most significant component), which makes B+ tree inserts sequential and pagination efficient.

**Amazon centralized**: a central ID authority microservice allocates blocks of 500-1000 IDs at a time. A machine requests a block, uses IDs from that block locally, and requests a new block when exhausted. Fewer network calls than requesting one ID at a time, but still requires a network round-trip per block. The central service is a potential SPOF (mitigated with replication).

**Clock drift in Snowflake**: if a machine's clock jumps backward (NTP correction), the generator **refuses to produce IDs** until the clock catches up. Otherwise, it would generate IDs with a past timestamp, colliding with previously generated IDs. This is a hard failure — the machine is blocked. Mitigation: use `chrony` for smooth clock adjustments rather than sudden jumps.

**Clock drift in Amazon's approach**: doesn't matter. The central service controls ID assignment; machine clocks are irrelevant to uniqueness. The trade-off is the network dependency and potential SPOF.

**For 100K IDs/sec across 50 machines**: Snowflake handles this easily (2K/sec per machine, well under the 4M/sec limit). Amazon's approach also works but adds latency from batch allocation requests. Snowflake wins on latency; Amazon wins on simplicity (no clock dependency).

---

## A8: Cursor-Based Pagination with Monotonic IDs (Week 3)

**Why offset breaks**: `SELECT * FROM posts ORDER BY created_at DESC LIMIT 20 OFFSET 40` means the database scans and discards 40 rows before returning 20. If a new post is inserted while the user is on page 2, all offsets shift — the user sees a duplicate (the item that was at position 20 is now at position 21, appearing on both page 1 and page 2). At page 1000, the DB scans 20,000 rows to return 20. Performance degrades linearly.

**Cursor-based solution**: use the last seen ID as a cursor.

```sql
-- First page (no cursor)
SELECT * FROM posts ORDER BY id DESC LIMIT 21;  -- fetch n+1 to check hasMore

-- Next page (use last ID from previous page as cursor)
SELECT * FROM posts WHERE id < 12345678 ORDER BY id DESC LIMIT 21;
```

**Why this works**: `WHERE id < cursor` uses the B+ tree index directly — it's a simple index seek, O(log N) regardless of which "page" you're on. Page 1 and page 1000 have identical performance. New inserts don't affect the cursor because IDs are monotonically increasing — new posts get higher IDs, and the cursor points at a fixed position in the log.

**Required index**: a B+ tree index on `id` (the primary key, which already has one). For composite cursors (e.g., `ORDER BY created_at DESC, id DESC`), you need a composite index on `(created_at, id)`.

**API design**: return `nextCursor` and `hasMore` in responses. The client sends `cursor` on the next request. Never expose page numbers in a cursor-based API.

---

## A9: Image Upload Service with Pre-Signed URLs (Week 4)

**Pre-signed URLs**: your server generates a temporary, signed URL that grants the client permission to upload directly to S3 for a limited time (e.g., 15 minutes). The signature includes the bucket, key, expiry, and allowed content type. The client uploads the file directly to S3 using an HTTP PUT — the file never touches your server.

**Flow:**

```
1. Client → Server: POST /api/uploads { filename: "photo.jpg", content_type: "image/jpeg" }
2. Server: generates pre-signed PUT URL for S3, saves metadata to DB (status: PENDING)
3. Server → Client: { upload_url: "https://s3.../photo.jpg?X-Amz-Signature=...", upload_id: "abc" }
4. Client → S3: PUT upload_url with file body (direct upload, bypasses server)
5. S3 → Server: S3 event notification (or client calls server to confirm)
6. Server: updates status to UPLOADED, triggers thumbnail generation
```

**Why better than proxying**: your server doesn't handle large file transfers (saves CPU, memory, bandwidth). Scales to thousands of concurrent uploads without server bottleneck. S3 handles retries, multipart upload, and storage durability (11 nines).

**Thumbnail generation**: use an S3 event notification to trigger a Lambda function or push a message to an SQS/Kafka queue. A worker picks up the message, downloads the original from S3, generates thumbnails at multiple sizes, uploads them back to S3, and updates the database with thumbnail URLs. This is an async worker pattern (delegation) — the user doesn't wait for thumbnails.

**Security considerations**: restrict the pre-signed URL to a specific content type and max file size. Set a short expiry (5-15 minutes). Validate the upload after it completes (check file headers, scan for malware).

---

## A10: LSM Trees vs B+ Trees for Storage Engines (Week 6)

**B+ tree (PostgreSQL, MySQL InnoDB)**: data is stored in a balanced tree structure on disk. Writes are **random I/O** — to update a row, the engine must find the correct leaf page, potentially split pages, and write in-place. Reads are efficient (O(log N) tree traversal). Indexes are always up to date.

**LSM tree (Cassandra, RocksDB, LevelDB)**: writes go to an in-memory **memtable** (sorted structure). When the memtable is full, it's flushed to disk as an immutable **SSTable** (Sorted String Table). Writes are **sequential I/O** (append-only), which is 10-100x faster than random I/O on spinning disks and still significantly faster on SSDs. Reads must check the memtable and potentially multiple SSTables, then merge results. Background **compaction** merges SSTables to reduce read amplification.

**Why LSM is faster for writes**: every write is an append to the memtable (in-memory) and eventually a sequential disk flush. No random seeks. No page splits. No in-place updates. This makes LSM trees ideal for write-heavy workloads like logging, time-series data, and event ingestion.

**When B+ trees win**:
- **Read-heavy workloads**: B+ tree reads are a single tree traversal. LSM reads may check multiple SSTables (read amplification).
- **Point lookups on indexed columns**: B+ tree index is always current. LSM may need to check memtable + multiple levels of SSTables.
- **Transactions and ACID**: B+ tree databases (PostgreSQL) have mature transaction support. Cassandra trades this for availability.
- **Space efficiency**: LSM trees can temporarily store multiple versions of the same key across SSTables until compaction runs (write amplification and space amplification).

**For a high-write logging service**: LSM tree (Cassandra) is the clear winner. Append-only writes, horizontal scaling, tunable consistency. B+ tree databases would require sharding to handle the write throughput, adding operational complexity that Cassandra handles natively.
