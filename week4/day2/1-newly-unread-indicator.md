
<!-- 
Need to inform users about presence of new messages (not unread/un acked)


Brainstorm
- GET /unread-indicator - get the number
- DEL /unread-indecator - vanish the number

- On the fly computation vs pre-computed -->


# Newly Unread Indicator - Technical Notes
```
Genrated! Need to watch the video
```
## 1. Quick Summary

**What it does:** Displays a badge/counter showing the number of new messages a user hasn't seen yet (distinct from unread/unacknowledged messages).

**Primary use case:** Messaging applications, notification systems, social media feeds

**Key feature highlights:**
- Real-time new message counting
- User-specific tracking
- Dismissible indicator

## 2. Core Concepts

### API Endpoints
- **GET /unread-indicator** - Retrieves the count of new messages
- **DELETE /unread-indicator** - Clears/dismisses the indicator

### Key Architectural Decision: **On-the-Fly Computation vs Pre-Computed**

#### A. **On-the-Fly Computation**
Calculate the count when requested by querying the database in real-time.

```java
@GetMapping("/unread-indicator")
public UnreadIndicatorResponse getUnreadIndicator(@RequestHeader("userId") String userId) {
    // Query database for messages created after user's last seen timestamp
    Instant lastSeen = userActivityRepo.getLastSeenTimestamp(userId);
    int newMessageCount = messageRepo.countNewMessages(userId, lastSeen);
    
    return new UnreadIndicatorResponse(newMessageCount);
}
```

#### B. **Pre-Computed (Cached Counter)**
Maintain a counter that's updated asynchronously whenever new messages arrive.

```java
@GetMapping("/unread-indicator")
public UnreadIndicatorResponse getUnreadIndicator(@RequestHeader("userId") String userId) {
    // Fetch pre-computed count from Redis/cache
    Integer count = redisTemplate.opsForValue().get("unread:" + userId);
    return new UnreadIndicatorResponse(count != null ? count : 0);
}

// Increment counter when new message arrives (via event handler)
@EventListener
public void onNewMessage(MessageCreatedEvent event) {
    for (String recipientId : event.getRecipientIds()) {
        redisTemplate.opsForValue().increment("unread:" + recipientId);
    }
}
```

## 3. Implementation Approaches

### Approach 1: On-the-Fly with Last Seen Timestamp

```java
public class UnreadIndicatorService {
    
    @Transactional(readOnly = true)
    public int getNewMessageCount(String userId) {
        UserActivity activity = userActivityRepo.findByUserId(userId);
        Instant lastSeen = activity.getLastSeenTimestamp();
        
        // Count messages created after last seen
        return messageRepo.countByRecipientIdAndCreatedAtAfter(userId, lastSeen);
    }
    
    public void markAsSeen(String userId) {
        userActivityRepo.updateLastSeen(userId, Instant.now());
    }
}
```

### Approach 2: Pre-Computed with Redis Counter

```java
public class UnreadIndicatorService {
    
    private final RedisTemplate<String, Integer> redisTemplate;
    
    public int getNewMessageCount(String userId) {
        Integer count = redisTemplate.opsForValue().get(getKey(userId));
        return count != null ? count : 0;
    }
    
    public void incrementCounter(String userId, int delta) {
        redisTemplate.opsForValue().increment(getKey(userId), delta);
    }
    
    public void clearCounter(String userId) {
        redisTemplate.delete(getKey(userId));
    }
    
    private String getKey(String userId) {
        return "unread_indicator:" + userId;
    }
}
```

### Approach 3: Hybrid (Pre-Computed with Fallback)

```java
public class UnreadIndicatorService {
    
    public int getNewMessageCount(String userId) {
        // Try cache first
        Integer cached = redisTemplate.opsForValue().get(getKey(userId));
        if (cached != null) {
            return cached;
        }
        
        // Fallback to computation if cache miss
        Instant lastSeen = userActivityRepo.getLastSeenTimestamp(userId);
        int count = messageRepo.countNewMessages(userId, lastSeen);
        
        // Store in cache for next time
        redisTemplate.opsForValue().set(getKey(userId), count, Duration.ofHours(24));
        return count;
    }
}
```

## 4. Pros and Cons

### On-the-Fly Computation ✅❌

**Pros:**
- ✅ **Always accurate** - No synchronization issues
- ✅ **Simple to implement** - Just query the database
- ✅ **No additional storage** - No cache/counter to maintain
- ✅ **Self-healing** - No risk of counter drift

**Cons:**
- ❌ **Slower response time** - Database query on every request
- ❌ **Higher database load** - Especially with many concurrent users
- ❌ **Doesn't scale well** - Performance degrades with data growth
- ❌ **Complex queries** - May need indexes on timestamp columns

### Pre-Computed (Cached) ⚡❌

**Pros:**
- ⚡ **Extremely fast** - O(1) cache lookup
- ⚡ **Scalable** - Handles high read traffic easily
- ⚡ **Low database load** - Minimal DB queries
- ⚡ **Predictable performance** - Consistent response times

**Cons:**
- ❌ **Synchronization complexity** - Must update counter on every message
- ❌ **Potential inconsistency** - Counter can drift if updates fail
- ❌ **Additional infrastructure** - Requires Redis/cache layer
- ❌ **Race conditions** - Need careful handling of concurrent updates

### Hybrid Approach 🎯

**Pros:**
- 🎯 **Best of both worlds** - Fast reads with accuracy fallback
- 🎯 **Resilient** - Works even if cache fails
- 🎯 **Self-correcting** - Rebuilds cache on miss

**Cons:**
- 🎯 **More complex** - Additional logic to maintain
- 🎯 **Inconsistent latency** - Cache miss causes slow response

## 5. Key Points to Remember

### Important Gotchas
- **"New" vs "Unread"** - Distinguish between messages never seen (new) vs messages seen but not marked as read
- **Race conditions** - User might see indicator, then new message arrives, causing confusion
- **Timezone issues** - Use UTC timestamps consistently
- **Counter drift** - Pre-computed counters can get out of sync if updates fail

### Performance Considerations
- **Index requirements** - On-the-fly needs indexes on `(recipient_id, created_at)` or `(recipient_id, last_seen)`
- **Cache TTL** - Set appropriate expiration for pre-computed values
- **Batch updates** - For group messages, update counters in batches
- **Read/write ratio** - High read ratio favors pre-computed

### Scalability Factors
- **User base size** - Millions of users need pre-computed approach
- **Message volume** - High message rate requires efficient counter updates
- **Real-time requirements** - Stricter real-time needs favor event-driven pre-computation

### Common Mistakes
- ❌ Not handling clock skew between servers
- ❌ Forgetting to clear indicator when user dismisses it
- ❌ Using non-atomic increment operations (causing lost updates)
- ❌ Not considering offline users (counters can grow unbounded)

## 6. Quick Reference

### Recommended Approach by Scale

| User Scale | Message Volume | Recommendation |
|------------|----------------|----------------|
| < 10K users | Low | On-the-fly computation |
| 10K-100K | Medium | Hybrid approach |
| > 100K | High | Pre-computed with Redis |

### Redis Commands (for Pre-Computed)
```java
// Increment counter
redisTemplate.opsForValue().increment("unread:userId123", 1);

// Get count
Integer count = redisTemplate.opsForValue().get("unread:userId123");

// Clear counter
redisTemplate.delete("unread:userId123");

// Set with expiration
redisTemplate.opsForValue().set("unread:userId123", 5, Duration.ofHours(24));
```

### Database Query (for On-the-Fly)
```java
// SQL query with prepared statement
String sql = "SELECT COUNT(*) FROM messages " +
             "WHERE recipient_id = ? AND created_at > ?";
```

## 7. Related Topics

- **Push Notifications** - Often triggered by same events as indicator updates
- **WebSocket/SSE** - Real-time indicator updates without polling
- **Message Queue (Kafka/RabbitMQ)** - Async counter updates at scale
- **Read Receipts** - More granular tracking than new/unread indicators
- **Database Partitioning** - For scaling message storage
- **Event Sourcing** - Alternative approach to tracking user activity

---

## Recommendation

For most applications, start with **Pre-Computed (Redis)** if you have >1000 active users. It provides the best user experience with fast response times. Use message queues to ensure reliable counter updates and implement monitoring to detect counter drift.