# Pagination

## Quick Summary
- **What it does:** Divides large datasets into smaller chunks (pages) for efficient retrieval and better UX
- **Primary use case:** Fetching records from databases, API responses, infinite scroll, feed systems
- **Key feature highlights:** 
  - Reduces memory usage and network bandwidth
  - Monotonic IDs enable faster, more reliable pagination (cursor-based)
  - Prevents loading millions of rows at once

---

## Core Concepts

### 1. **Offset-Based Pagination** (Traditional)
- Uses `OFFSET` and `LIMIT` clauses
- Pages defined by page number: `page=3`, `limit=20`
- **Formula:** `OFFSET = (page - 1) × limit`

```sql
SELECT * FROM tweets 
ORDER BY created_at DESC 
LIMIT 20 OFFSET 40;  -- Page 3 (skip first 40 records)
```

**Problems:**
- ❌ Slow for large offsets (DB must scan and skip rows)
- ❌ Inconsistent results if data changes between pages (missed/duplicate records)
- ❌ `OFFSET 1000000` scans 1M rows just to skip them

---

### 2. **Cursor-Based Pagination** (Recommended)
- Uses a **cursor** (typically last seen ID) to fetch next page
- Only retrieves records **after** the cursor
- **Requires monotonic IDs** for reliable ordering

```sql
-- First page (no cursor)
SELECT * FROM tweets 
WHERE id > 0 
ORDER BY id DESC 
LIMIT 20;

-- Next page (use last ID as cursor)
SELECT * FROM tweets 
WHERE id < 12345678  -- cursor from previous page
ORDER BY id DESC 
LIMIT 20;
```

**Advantages:**
- ✅ Consistent results (immune to concurrent inserts)
- ✅ Fast (uses index efficiently, no scanning)
- ✅ Works with infinite scroll

---

### 3. **Keyset Pagination** (Seek Method)
- Similar to cursor-based but uses composite keys
- Uses timestamp + ID for orderingD

```sql
SELECT * FROM tweets 
WHERE (created_at, id) < ('2024-01-15 10:30:00', 999888)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

---

### 4. **How Monotonic IDs Help**
Monotonic IDs (always increasing) make pagination efficient because:

| Benefit | Why? |
|---------|------|
| **Natural ordering** | No separate timestamp column needed |
| **Index efficiency** | `WHERE id > cursor` uses B+ tree index directly |
| **Consistency** | IDs never change, cursors remain valid |
| **Time-based queries** | Can estimate time range from ID range |
| **Database performance** | Sequential inserts → better cache locality |

**Example with Snowflake IDs:**
```
Tweet IDs: 1001 → 1002 → 1003 → ... (always increasing)
```
- Fetch page 1: `WHERE id > 0 LIMIT 20` → gets IDs 1001-1020
- Fetch page 2: `WHERE id > 1020 LIMIT 20` → gets IDs 1021-1040
- Fast & consistent!

---

## Usage Examples

### Example 1: Offset-Based Pagination (Java)
```java
public class OffsetPagination {
    private final Connection conn;
    
    public List<Tweet> getTweets(int pageNumber, int pageSize) throws SQLException {
        int offset = (pageNumber - 1) * pageSize;
        
        String sql = "SELECT * FROM tweets " +
                     "ORDER BY created_at DESC " +
                     "LIMIT ? OFFSET ?";
        
        try (PreparedStatement stmt = conn.prepareStatement(sql)) {
            stmt.setInt(1, pageSize);
            stmt.setInt(2, offset);
            
            ResultSet rs = stmt.executeQuery();
            List<Tweet> tweets = new ArrayList<>();
            
            while (rs.next()) {
                tweets.add(mapToTweet(rs));
            }
            
            return tweets;
        }
    }
}

// Usage
List<Tweet> page1 = getTweets(1, 20);  // First 20 tweets
List<Tweet> page2 = getTweets(2, 20);  // Next 20 tweets (OFFSET 20)
```

**Output:**
```
Page 1: 20 tweets (fast)
Page 2: 20 tweets (still fast)
Page 100: 20 tweets (slow - must skip 1980 rows!)
```

---

### Example 2: Cursor-Based Pagination with Monotonic IDs (Java)
```java
public class CursorPagination {
    private final Connection conn;
    
    public PaginatedResult<Tweet> getTweets(Long cursor, int limit) throws SQLException {
        String sql = "SELECT * FROM tweets " +
                     "WHERE id < ? " +        // Use monotonic ID as cursor
                     "ORDER BY id DESC " +
                     "LIMIT ?";
        
        try (PreparedStatement stmt = conn.prepareStatement(sql)) {
            stmt.setLong(1, cursor != null ? cursor : Long.MAX_VALUE);
            stmt.setInt(2, limit + 1);  // Fetch one extra to check if more pages
            
            ResultSet rs = stmt.executeQuery();
            List<Tweet> tweets = new ArrayList<>();
            
            while (rs.next() && tweets.size() < limit) {
                tweets.add(mapToTweet(rs));
            }
            
            Long nextCursor = null;
            boolean hasMore = rs.next();  // Check if there's a next page
            
            if (hasMore && !tweets.isEmpty()) {
                nextCursor = tweets.get(tweets.size() - 1).getId();
            }
            
            return new PaginatedResult<>(tweets, nextCursor, hasMore);
        }
    }
}

// Usage
PaginatedResult<Tweet> page1 = getTweets(null, 20);  // First page
System.out.println("Next cursor: " + page1.getNextCursor());

PaginatedResult<Tweet> page2 = getTweets(page1.getNextCursor(), 20);  // Next page
// Always fast, no matter how many pages!
```

**Output:**
```
Page 1: 20 tweets, nextCursor=1020, hasMore=true (0.5ms)
Page 2: 20 tweets, nextCursor=1000, hasMore=true (0.5ms)
Page 1000: 20 tweets, nextCursor=100, hasMore=true (0.5ms) ← Still fast!
```

---

### Example 3: API Response Format
```java
public class TweetController {
    
    @GetMapping("/api/tweets")
    public ResponseEntity<PaginationResponse> getTweets(
            @RequestParam(required = false) Long cursor,
            @RequestParam(defaultValue = "20") int limit) {
        
        PaginatedResult<Tweet> result = cursorPagination.getTweets(cursor, limit);
        
        PaginationResponse response = PaginationResponse.builder()
            .data(result.getData())
            .nextCursor(result.getNextCursor())
            .hasMore(result.isHasMore())
            .build();
        
        return ResponseEntity.ok(response);
    }
}

// Response JSON
{
    "data": [
        {"id": 1020, "text": "Hello World", "created_at": "2024-01-15T10:30:00Z"},
        {"id": 1019, "text": "Another tweet", "created_at": "2024-01-15T10:29:00Z"},
        // ... 18 more
    ],
    "nextCursor": 1001,
    "hasMore": true
}

// Client fetches next page
GET /api/tweets?cursor=1001&limit=20
```

---

## Key Points to Remember

### Performance Gotchas
- ⚠️ **Offset pagination degrades with depth:** `OFFSET 1000000` is extremely slow
- ⚠️ **Always index the ordering column:** `CREATE INDEX idx_tweets_id ON tweets(id DESC)`
- ⚠️ **Concurrent inserts break offset pagination:** User sees duplicates or skips records
- ⚠️ **Never use offset for infinite scroll:** Use cursor-based instead

### Database Considerations
- **B+ Tree optimization:** Monotonic IDs keep index balanced, sequential writes are 2-10x faster
- **Range queries:** `id > cursor` leverages index efficiently (log N lookup)
- **Covering index:** Include frequently fetched columns in index to avoid table lookups
  ```sql
  CREATE INDEX idx_tweets_id_covering ON tweets(id DESC) 
  INCLUDE (user_id, text, created_at);
  ```

### Edge Cases
- ❌ **Empty pages:** Always check `hasMore` before fetching next page
- ❌ **Deleted records:** Cursor might point to deleted ID (handle gracefully)
- ❌ **Clock skew with timestamps:** Use monotonic IDs instead of `created_at` for reliability

### Common Mistakes
- ❌ Using offset pagination for deep pages (page 1000+)
- ❌ Not indexing the sort column (ORDER BY unindexed column is slow)
- ❌ Returning page numbers instead of cursors in APIs
- ❌ Assuming offset pagination is "good enough" at scale
- ✅ **Always use cursor-based pagination with monotonic IDs for production systems**

---

## Quick Reference

### Pagination Method Comparison

| Method | Performance | Consistency | Use Case |
|--------|-------------|-------------|----------|
| **Offset-based** | ❌ Degrades with depth | ❌ Inconsistent | Small datasets, page numbers required |
| **Cursor-based** | ✅ Always fast | ✅ Consistent | Infinite scroll, feeds, high-scale |
| **Keyset** | ✅ Fast | ✅ Consistent | Complex sorting (timestamp + ID) |

### SQL Patterns

**Offset (simple but slow for deep pages):**
```sql
SELECT * FROM tweets ORDER BY id DESC LIMIT 20 OFFSET 40;
```

**Cursor with monotonic ID (fast, recommended):**
```sql
-- Descending (newest first)
SELECT * FROM tweets WHERE id < ? ORDER BY id DESC LIMIT 20;

-- Ascending (oldest first)
SELECT * FROM tweets WHERE id > ? ORDER BY id ASC LIMIT 20;
```

**Composite cursor (timestamp + ID):**
```sql
SELECT * FROM tweets 
WHERE (created_at, id) < (?, ?)
ORDER BY created_at DESC, id DESC 
LIMIT 20;
```

### Index Requirements
```sql
-- For cursor-based pagination on ID
CREATE INDEX idx_tweets_id_desc ON tweets(id DESC);

-- For composite cursor (timestamp + ID)
CREATE INDEX idx_tweets_created_id ON tweets(created_at DESC, id DESC);

-- For filtered pagination (e.g., user timeline)
CREATE INDEX idx_tweets_user_id ON tweets(user_id, id DESC);
```

### Java Helper Classes
```java
@Data
@Builder
public class PaginatedResult<T> {
    private List<T> data;
    private Long nextCursor;
    private boolean hasMore;
}

@Data
@AllArgsConstructor
public class PaginationRequest {
    private Long cursor;
    private int limit = 20;
    
    public Long getCursorOrDefault() {
        return cursor != null ? cursor : Long.MAX_VALUE;
    }
}
```

---

## Why Monotonic IDs are Essential for Pagination

### Problem with Random IDs (UUIDs)
```java
// UUID v4: random, non-sequential
UUID id1 = UUID.randomUUID(); // 3f8b9c2a-...
UUID id2 = UUID.randomUUID(); // 1a2b3c4d-...
UUID id3 = UUID.randomUUID(); // 9e8f7g6h-...

// No natural ordering!
// Cannot do: SELECT * FROM tweets WHERE id > ? (unpredictable)
```

**Issues:**
- ❌ No temporal ordering
- ❌ Cannot use as cursor (no guarantee of sequence)
- ❌ Database index fragmentation (random inserts split B+ tree pages)
- ❌ Slower writes (random locations in index)

---

### Solution: Monotonic IDs (Twitter Snowflake)

```java
// Snowflake IDs: always increasing
long id1 = 1152945683951325184L;  // Tweet at 10:30:00
long id2 = 1152945683951325185L;  // Tweet at 10:30:01
long id3 = 1152945683951325186L;  // Tweet at 10:30:02

// Natural ordering: id1 < id2 < id3
```

**Format (64 bits):**
```
┌──────────────────────────────────────────────┐
│ 41 bits    │ 10 bits      │ 12 bits         │
│ Timestamp  │ Machine ID   │ Sequence        │
└──────────────────────────────────────────────┘
```

**Benefits for Pagination:**

1. **Efficient range queries:**
```java
// Get tweets after cursor (uses index)
String sql = "SELECT * FROM tweets WHERE id > ? ORDER BY id DESC LIMIT 20";
// Database: Index seek O(log N) + scan 20 rows ← FAST
```

2. **Time-based filtering without timestamp column:**
```java
// Get tweets from last hour using ID alone
long oneHourAgo = System.currentTimeMillis() - 3600000;
long minId = generateSnowflakeForTime(oneHourAgo);

String sql = "SELECT * FROM tweets WHERE id > ? LIMIT 100";
// No need for separate created_at column!
```

3. **Consistent pagination:**
```java
// Page 1
List<Tweet> page1 = getTweets(null, 20);  // IDs: 1000-1020

// New tweet inserted (ID: 1021)

// Page 2 still consistent
List<Tweet> page2 = getTweets(1020, 20);  // IDs: 1019-1000
// Doesn't see the new tweet (ID 1021) ← Good! No duplicates
```

4. **Database write performance:**
```
Sequential inserts (monotonic IDs):
┌───────────────────────────────────┐
│ 1000 │ 1001 │ 1002 │ 1003 │ 1004 │  ← Append at end
└───────────────────────────────────┘
Fast: No page splits, better cache locality

Random inserts (UUIDs):
┌───────────────────────────────────┐
│ a3f2 │ ???? │ 7e9d │ ???? │ 2b1c │  ← Insert anywhere
└───────────────────────────────────┘
Slow: Frequent page splits, poor cache performance
```

---

### Real-World Example: Twitter Timeline Pagination

```java
public class TwitterTimelineService {
    private final SnowflakeIdGenerator idGenerator;
    
    // POST /api/tweets - Create new tweet
    public Tweet createTweet(String text, Long userId) {
        long tweetId = idGenerator.generateId();  // Monotonic ID
        
        Tweet tweet = Tweet.builder()
            .id(tweetId)
            .userId(userId)
            .text(text)
            .build();
        
        tweetRepository.save(tweet);
        return tweet;
    }
    
    // GET /api/timeline?cursor=123&limit=20
    public PaginatedResult<Tweet> getTimeline(Long cursor, int limit) {
        String sql = """
            SELECT t.* FROM tweets t
            JOIN followers f ON f.followee_id = t.user_id
            WHERE f.follower_id = ?
              AND t.id < ?
            ORDER BY t.id DESC
            LIMIT ?
        """;
        
        // Uses index on (follower_id, tweet_id)
        // Fast for any page depth!
    }
}

// Client code
PaginatedResult<Tweet> page1 = getTimeline(null, 20);
// Response: { data: [...], nextCursor: 999888, hasMore: true }

PaginatedResult<Tweet> page2 = getTimeline(999888L, 20);
// Response: { data: [...], nextCursor: 999868, hasMore: true }

// Even at page 1000, still fast (<1ms)!
```

---

## Related Topics
- **Twitter Snowflake ID Generation** - Distributed monotonic ID system
- **Database B+ Tree Indexing** - Why sequential IDs perform better
- **Hot Shard Problem** - Mitigating uneven load with monotonic IDs
- **Clock Synchronization (NTP)** - Critical for monotonic ID generation
- **Zookeeper** - Coordinating machine IDs in distributed systems
- **Read Replica Lag** - Handling consistency in paginated queries across replicas
- **Offset vs Cursor Pagination** - Performance comparison at scale

---

## Summary

✅ **Use cursor-based pagination with monotonic IDs for production systems**

**Why?**
- 10-100x faster than offset pagination at scale
- Consistent results (no duplicates/skips)
- Database-friendly (sequential inserts, efficient indexes)
- Works with infinite scroll, feeds, and real-time systems

**Implementation Checklist:**
1. Use monotonic IDs (Snowflake, ULID, or auto-increment)
2. Index the ID column with `DESC` order
3. Return `nextCursor` and `hasMore` in API responses
4. Use `WHERE id < cursor ORDER BY id DESC LIMIT n+1` pattern
5. Never use offset for deep pagination

**Key Insight:** Monotonic IDs turn pagination from O(N) to O(log N) by leveraging database index structures efficiently.
