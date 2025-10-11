# Caching

## What is Caching?

Cache is a high-speed storage layer that sits between your application and your primary database. It stores frequently accessed data in memory to reduce database load and improve response times.

**Why Cache?**
- **Speed**: In-memory access (< 1ms) vs disk-based DB (10-100ms+)
- **Reduced DB Load**: Offload read traffic from primary database
- **Cost Savings**: Fewer DB reads = lower infrastructure costs
- **Better User Experience**: Faster response times

**Typical Latencies:**
- L1/L2 Cache: < 10ns
- RAM: ~100ns
- SSD: ~100μs
- Network + DB: 10-100ms+

## Cache Hit vs Cache Miss

- **Cache Hit**: Data found in cache, return immediately
- **Cache Miss**: Data not in cache, fetch from DB, populate cache, return to client

**Cache Hit Ratio = Cache Hits / (Cache Hits + Cache Misses)**

Target hit ratio: 80-90% for most applications

## Caching Strategies

### 1. Cache-Aside (Lazy Loading)

Most common pattern. Application manages both cache and database.

**Read Flow:**
1. Check cache for data
2. If cache hit → return data
3. If cache miss → fetch from DB
4. Store in cache
5. Return data

**Write Flow:**
1. Write to database
2. Invalidate/delete cache entry (don't update!)

**Pros:**
- Only requested data is cached (no unnecessary data)
- Cache failure doesn't break the system (graceful degradation)
- Works well for read-heavy workloads

**Cons:**
- Cache miss penalty (3 trips: cache check, DB read, cache write)
- Stale data possible if cache isn't invalidated on updates
- Cold start problem (empty cache after restart)

```java
public User getUser(String userId) {
    // Try cache first
    User user = cache.get(userId);
    
    if (user != null) {
        return user; // Cache hit
    }
    
    // Cache miss - fetch from DB
    user = database.query("SELECT * FROM users WHERE id = ?", userId);
    
    // Populate cache
    cache.set(userId, user, TTL_SECONDS);
    
    return user;
}

public void updateUser(User user) {
    // Update DB first
    database.update(user);
    
    // Invalidate cache (don't update!)
    cache.delete(user.getId());
}
```

**Why invalidate instead of update?**
- Race conditions: Another thread might update DB right after you update cache
- Failed DB writes: Cache would have new data but DB has old data (inconsistency)
- Simpler: No need to keep cache and DB update logic in sync


### 2. Read-Through Cache

Cache sits between application and database. Cache library handles DB access.

**Read Flow:**
1. Application requests data from cache
2. If cache miss, cache library automatically fetches from DB
3. Cache library stores in cache and returns data

**Pros:**
- Application code is simpler (cache handles DB logic)
- Consistent caching logic across application

**Cons:**
- Cache becomes a single point of failure
- Less control over cache behavior
- Cold start problem still exists


### 3. Write-Through Cache

Every write goes through cache to database.

**Write Flow:**
1. Application writes to cache
2. Cache synchronously writes to DB
3. Returns success only after both succeed

**Pros:**
- Cache and DB always consistent
- No stale data
- Read-heavy workloads benefit

**Cons:**
- Higher write latency (synchronous DB write)
- Cache filled with data that may never be read
- Doesn't help with write-heavy workloads


### 4. Write-Behind (Write-Back) Cache

Writes go to cache, then asynchronously to database.

**Write Flow:**
1. Write to cache immediately
2. Return success
3. Cache writes to DB asynchronously (batched)

**Pros:**
- Very fast writes (no DB wait)
- Can batch multiple writes (better throughput)
- Good for write-heavy workloads

**Cons:**
- Risk of data loss if cache fails before DB write
- Complex to implement (need reliable queue)
- Eventual consistency


## Cache Eviction Policies

When cache is full, which items should be removed?

### LRU (Least Recently Used)
- Evict items that haven't been accessed in longest time
- Most commonly used
- Good for general-purpose applications
- **Assumption**: Recently used data will be used again soon

### LFU (Least Frequently Used)
- Evict items with lowest access count
- Good when some items are consistently popular
- **Assumption**: Frequently accessed data should stay in cache

### FIFO (First In First Out)
- Evict oldest items first
- Simple to implement
- Not as effective as LRU/LFU in most cases

### TTL (Time To Live)
- Items expire after fixed time
- Good for data with known freshness requirements
- Often combined with LRU/LFU

**Redis Default:** approximated LRU (doesn't track all keys, samples randomly for performance)


## Cache Problems & Solutions

### 1. Cache Stampede (Thundering Herd)

**Problem:**
- Cache expires for popular key
- 1000s of concurrent requests see cache miss
- All 1000s hit database simultaneously
- Database overload → slow queries → timeouts → cascading failures

**Common Scenario:**
- Homepage data expires at midnight
- 10,000 users refresh at 12:00:01 AM
- All see cache miss
- Database gets 10,000 simultaneous queries for same data

**Solution Options:**

#### Option 1: Local Semaphore (Single Instance)
Only one thread can fetch data, others wait.

#### Option 2: Distributed Lock (Multiple Instances)
Only one server instance can fetch data using Redis locks.

#### Option 3: Probabilistic Early Expiration
Refresh cache before expiration with some probability.

```java
// Refresh early based on remaining TTL
long timeLeft = cache.getTTL(key);
double refreshProbability = 1.0 - (timeLeft / originalTTL);

if (Math.random() < refreshProbability) {
    // Refresh cache asynchronously
    asyncRefreshCache(key);
}
```

#### Option 4: Never Truly Expire
Background job refreshes cache before expiration.


### 2. Cache Penetration

**Problem:**
- Requests for non-existent data
- Every request hits database (cache can't help)
- Example: Malicious user queries user_id = "fake123" repeatedly

**Solutions:**
- Cache null/empty results with short TTL
- Bloom filter (probabilistic data structure to check existence)
- Request validation/rate limiting


### 3. Cache Avalanche

**Problem:**
- Many cache keys expire at same time
- Massive DB load spike

**Solutions:**
- Add random jitter to TTL: `TTL = base_ttl + random(0, 60)`
- Never truly expire popular keys (background refresh)
- Circuit breaker pattern on DB


## Cache Stampede Prevention - Java Implementation

### Solution 1: Local Semaphore (Single Instance)

Works when you have one application instance. Only one thread fetches data, others wait.

```java
import java.util.concurrent.*;
import java.util.Map;
import redis.clients.jedis.Jedis;

public class LocalSemaphoreCache {
    private final Jedis cache;
    private final Database database;
    private final ConcurrentHashMap<String, Semaphore> locks = new ConcurrentHashMap<>();
    private static final int TTL_SECONDS = 300; // 5 minutes
    
    public LocalSemaphoreCache(Jedis cache, Database database) {
        this.cache = cache;
        this.database = database;
    }
    
    public String getData(String key) throws Exception {
        // Try cache first
        String cachedValue = cache.get(key);
        if (cachedValue != null) {
            return cachedValue; // Cache hit
        }
        
        // Cache miss - acquire semaphore for this key
        // Only one thread per key can proceed
        Semaphore lock = locks.computeIfAbsent(key, k -> new Semaphore(1));
        
        try {
            // Try to acquire permit
            // First thread gets it immediately, others wait
            lock.acquire();
            
            // Double-check cache (another thread might have populated it)
            cachedValue = cache.get(key);
            if (cachedValue != null) {
                return cachedValue; // Another thread already loaded it
            }
            
            // Still not in cache - fetch from database
            System.out.println(Thread.currentThread().getName() + " - Fetching from DB for key: " + key);
            String dbValue = database.query(key); // Only ONE thread executes this
            
            // Populate cache
            cache.setex(key, TTL_SECONDS, dbValue);
            
            return dbValue;
            
        } finally {
            // Always release the lock
            lock.release();
            
            // Cleanup: Remove semaphore if no one is waiting
            // (prevents memory leak from accumulating semaphores)
            if (lock.availablePermits() == 1 && !lock.hasQueuedThreads()) {
                locks.remove(key);
            }
        }
    }
    
    // Simulate multiple concurrent requests
    public static void main(String[] args) throws Exception {
        Jedis jedis = new Jedis("localhost", 6379);
        Database db = new Database();
        LocalSemaphoreCache cacheService = new LocalSemaphoreCache(jedis, db);
        
        // Simulate 100 concurrent requests for same key
        ExecutorService executor = Executors.newFixedThreadPool(100);
        String testKey = "user:12345";
        
        // Clear cache to simulate cache miss
        jedis.del(testKey);
        
        CountDownLatch latch = new CountDownLatch(100);
        
        for (int i = 0; i < 100; i++) {
            executor.submit(() -> {
                try {
                    String result = cacheService.getData(testKey);
                    System.out.println(Thread.currentThread().getName() + " - Got result: " + result);
                } catch (Exception e) {
                    e.printStackTrace();
                } finally {
                    latch.countDown();
                }
            });
        }
        
        latch.await();
        executor.shutdown();
        
        System.out.println("\nTotal DB queries made: " + db.getQueryCount());
        System.out.println("Expected: 1 (all other threads waited)");
    }
}

// Simulated database class
class Database {
    private int queryCount = 0;
    
    public synchronized String query(String key) {
        queryCount++;
        // Simulate slow DB query
        try {
            Thread.sleep(1000);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        return "Data for " + key;
    }
    
    public synchronized int getQueryCount() {
        return queryCount;
    }
}
```

**How it works:**
1. 100 threads request same key
2. First thread acquires semaphore, others block
3. First thread fetches from DB (1 second)
4. First thread populates cache and releases semaphore
5. Other 99 threads wake up, check cache, find data
6. **Result: Only 1 DB query instead of 100**

**Limitation:** Only works for single application instance. If you have multiple servers, each server's threads compete separately.


### Solution 2: Redis Distributed Lock (Multi-Instance)

For distributed systems with multiple application instances. Uses Redis SET NX EX for distributed locking.

```java
import redis.clients.jedis.Jedis;
import redis.clients.jedis.params.SetParams;
import java.util.UUID;

public class DistributedLockCache {
    private final Jedis cache;
    private final Database database;
    private static final int TTL_SECONDS = 300; // Cache TTL
    private static final int LOCK_TTL_SECONDS = 10; // Lock expires after 10 seconds
    private static final int MAX_RETRY_ATTEMPTS = 50;
    private static final int RETRY_DELAY_MS = 100;
    
    public DistributedLockCache(Jedis cache, Database database) {
        this.cache = cache;
        this.database = database;
    }
    
    public String getData(String key) throws Exception {
        // Try cache first
        String cachedValue = cache.get(key);
        if (cachedValue != null) {
            return cachedValue; // Cache hit
        }
        
        // Cache miss - need to fetch from DB
        // Use distributed lock to ensure only ONE instance across ALL servers fetches
        String lockKey = "lock:" + key;
        String lockValue = UUID.randomUUID().toString(); // Unique identifier for this lock
        
        // Try to acquire distributed lock
        boolean lockAcquired = acquireLock(lockKey, lockValue, LOCK_TTL_SECONDS);
        
        if (lockAcquired) {
            try {
                // We got the lock! Double-check cache
                cachedValue = cache.get(key);
                if (cachedValue != null) {
                    return cachedValue;
                }
                
                // Fetch from database
                System.out.println(Thread.currentThread().getName() + " - Acquired lock, fetching from DB");
                String dbValue = database.query(key);
                
                // Populate cache
                cache.setex(key, TTL_SECONDS, dbValue);
                
                return dbValue;
                
            } finally {
                // Release lock safely (only if we still own it)
                releaseLock(lockKey, lockValue);
            }
        } else {
            // We didn't get the lock - another instance is fetching data
            // Wait and retry reading from cache
            System.out.println(Thread.currentThread().getName() + " - Waiting for cache hydration");
            return waitForCacheHydration(key);
        }
    }
    
    /**
     * Acquire distributed lock using Redis SET NX EX
     * SET key value NX EX seconds
     * NX = only set if key doesn't exist
     * EX = set expiration time
     */
    private boolean acquireLock(String lockKey, String lockValue, int ttlSeconds) {
        SetParams params = SetParams.setParams()
            .nx() // Only set if not exists
            .ex(ttlSeconds); // Set expiration
            
        String result = cache.set(lockKey, lockValue, params);
        return "OK".equals(result);
    }
    
    /**
     * Release lock only if we still own it
     * Use Lua script for atomic check-and-delete
     */
    private void releaseLock(String lockKey, String lockValue) {
        String luaScript = 
            "if redis.call('get', KEYS[1]) == ARGV[1] then " +
            "    return redis.call('del', KEYS[1]) " +
            "else " +
            "    return 0 " +
            "end";
        
        cache.eval(luaScript, 1, lockKey, lockValue);
    }
    
    /**
     * Wait for another instance to populate cache
     */
    private String waitForCacheHydration(String key) throws Exception {
        for (int i = 0; i < MAX_RETRY_ATTEMPTS; i++) {
            // Check cache
            String cachedValue = cache.get(key);
            if (cachedValue != null) {
                System.out.println(Thread.currentThread().getName() + " - Cache hydrated by another instance");
                return cachedValue;
            }
            
            // Not ready yet, wait a bit
            Thread.sleep(RETRY_DELAY_MS);
        }
        
        // Timeout - fallback to fetching ourselves
        System.out.println(Thread.currentThread().getName() + " - Timeout waiting, fetching directly");
        return fetchWithFallback(key);
    }
    
    /**
     * Fallback: fetch directly without lock
     * Used when waiting times out
     */
    private String fetchWithFallback(String key) throws Exception {
        String dbValue = database.query(key);
        cache.setex(key, TTL_SECONDS, dbValue);
        return dbValue;
    }
    
    // Test with multiple threads simulating multiple servers
    public static void main(String[] args) throws Exception {
        Jedis jedis = new Jedis("localhost", 6379);
        Database db = new Database();
        
        String testKey = "product:54321";
        jedis.del(testKey); // Clear cache
        
        // Simulate 200 concurrent requests from multiple instances
        DistributedLockCache cacheService = new DistributedLockCache(jedis, db);
        
        ExecutorService executor = Executors.newFixedThreadPool(200);
        CountDownLatch latch = new CountDownLatch(200);
        
        for (int i = 0; i < 200; i++) {
            executor.submit(() -> {
                try {
                    String result = cacheService.getData(testKey);
                    System.out.println(Thread.currentThread().getName() + " - Result: " + result);
                } catch (Exception e) {
                    e.printStackTrace();
                } finally {
                    latch.countDown();
                }
            });
        }
        
        latch.await();
        executor.shutdown();
        
        System.out.println("\nTotal DB queries: " + db.getQueryCount());
        System.out.println("Expected: 1 (distributed lock prevented stampede)");
    }
}
```

**How Distributed Lock Works:**

1. **Lock Acquisition:**
   ```
   SET lock:user:123 "unique-uuid" NX EX 10
   ```
   - `NX`: Only set if key doesn't exist (atomic operation)
   - `EX 10`: Lock expires after 10 seconds (prevents deadlock if instance crashes)
   - Returns "OK" if lock acquired, null otherwise

2. **Lock Release:**
   ```lua
   -- Lua script ensures atomic check-and-delete
   if redis.call('get', KEYS[1]) == ARGV[1] then
       return redis.call('del', KEYS[1])
   else
       return 0
   end
   ```
   - Only delete if we still own the lock (lockValue matches)
   - Prevents accidentally releasing another instance's lock

3. **Flow:**
   - Instance A acquires lock → fetches from DB → populates cache → releases lock
   - Instances B, C, D... wait and retry reading cache
   - Once Instance A finishes, B, C, D get cache hits


### Solution 3: Advanced - Lock with Polling (Production-Ready)

More robust implementation with better error handling and monitoring.

```java
import redis.clients.jedis.Jedis;
import redis.clients.jedis.params.SetParams;
import java.util.UUID;
import java.util.concurrent.*;

public class ProductionCacheStampedeProtection {
    private final Jedis cache;
    private final Database database;
    private final CacheMetrics metrics;
    
    // Configuration
    private static final int CACHE_TTL_SECONDS = 300;
    private static final int LOCK_TTL_SECONDS = 15;
    private static final int MAX_WAIT_MS = 5000;
    private static final int POLL_INTERVAL_MS = 50;
    
    public ProductionCacheStampedeProtection(Jedis cache, Database database, CacheMetrics metrics) {
        this.cache = cache;
        this.database = database;
        this.metrics = metrics;
    }
    
    public String getData(String key) throws Exception {
        // Try cache first
        String value = cache.get(key);
        if (value != null) {
            metrics.recordCacheHit(key);
            return value;
        }
        
        metrics.recordCacheMiss(key);
        
        // Attempt to acquire lock
        String lockKey = "lock:" + key;
        String lockValue = UUID.randomUUID().toString();
        
        if (acquireLock(lockKey, lockValue)) {
            // We got the lock
            try {
                return fetchAndCache(key);
            } finally {
                releaseLock(lockKey, lockValue);
            }
        } else {
            // Someone else has the lock - wait for cache hydration
            return waitForCacheOrFallback(key);
        }
    }
    
    private String fetchAndCache(String key) throws Exception {
        // Double-check cache
        String value = cache.get(key);
        if (value != null) {
            return value;
        }
        
        // Fetch from database
        long startTime = System.currentTimeMillis();
        value = database.query(key);
        long duration = System.currentTimeMillis() - startTime;
        
        metrics.recordDatabaseQuery(key, duration);
        
        // Cache the result
        cache.setex(key, CACHE_TTL_SECONDS, value);
        
        return value;
    }
    
    private String waitForCacheOrFallback(String key) throws Exception {
        long deadline = System.currentTimeMillis() + MAX_WAIT_MS;
        
        while (System.currentTimeMillis() < deadline) {
            // Check if cache is populated
            String value = cache.get(key);
            if (value != null) {
                metrics.recordCacheHydrationSuccess(key);
                return value;
            }
            
            // Wait a bit before retrying
            Thread.sleep(POLL_INTERVAL_MS);
        }
        
        // Timeout - fetch directly as fallback
        metrics.recordCacheHydrationTimeout(key);
        return fetchWithoutLock(key);
    }
    
    private String fetchWithoutLock(String key) throws Exception {
        String value = database.query(key);
        
        // Try to cache (best effort, might race with lock holder)
        try {
            cache.setex(key, CACHE_TTL_SECONDS, value);
        } catch (Exception e) {
            // Ignore cache write failures in fallback path
            metrics.recordCacheWriteFailure(key);
        }
        
        return value;
    }
    
    private boolean acquireLock(String lockKey, String lockValue) {
        SetParams params = SetParams.setParams().nx().ex(LOCK_TTL_SECONDS);
        String result = cache.set(lockKey, lockValue, params);
        
        boolean acquired = "OK".equals(result);
        if (acquired) {
            metrics.recordLockAcquired(lockKey);
        } else {
            metrics.recordLockContention(lockKey);
        }
        
        return acquired;
    }
    
    private void releaseLock(String lockKey, String lockValue) {
        String script = 
            "if redis.call('get', KEYS[1]) == ARGV[1] then " +
            "    return redis.call('del', KEYS[1]) " +
            "else " +
            "    return 0 " +
            "end";
        
        try {
            cache.eval(script, 1, lockKey, lockValue);
            metrics.recordLockReleased(lockKey);
        } catch (Exception e) {
            metrics.recordLockReleaseFailure(lockKey);
        }
    }
}

// Metrics class for monitoring
class CacheMetrics {
    private final ConcurrentHashMap<String, Long> cacheHits = new ConcurrentHashMap<>();
    private final ConcurrentHashMap<String, Long> cacheMisses = new ConcurrentHashMap<>();
    private final ConcurrentHashMap<String, Long> dbQueries = new ConcurrentHashMap<>();
    
    public void recordCacheHit(String key) {
        cacheHits.merge(key, 1L, Long::sum);
        System.out.println("METRIC: Cache hit for " + key);
    }
    
    public void recordCacheMiss(String key) {
        cacheMisses.merge(key, 1L, Long::sum);
        System.out.println("METRIC: Cache miss for " + key);
    }
    
    public void recordDatabaseQuery(String key, long durationMs) {
        dbQueries.merge(key, 1L, Long::sum);
        System.out.println("METRIC: DB query for " + key + " took " + durationMs + "ms");
    }
    
    public void recordLockAcquired(String lockKey) {
        System.out.println("METRIC: Lock acquired: " + lockKey);
    }
    
    public void recordLockContention(String lockKey) {
        System.out.println("METRIC: Lock contention: " + lockKey);
    }
    
    public void recordLockReleased(String lockKey) {
        System.out.println("METRIC: Lock released: " + lockKey);
    }
    
    public void recordLockReleaseFailure(String lockKey) {
        System.out.println("METRIC: Lock release failed: " + lockKey);
    }
    
    public void recordCacheHydrationSuccess(String key) {
        System.out.println("METRIC: Cache hydration success: " + key);
    }
    
    public void recordCacheHydrationTimeout(String key) {
        System.out.println("METRIC: Cache hydration timeout: " + key);
    }
    
    public void recordCacheWriteFailure(String key) {
        System.out.println("METRIC: Cache write failure: " + key);
    }
    
    public void printSummary() {
        System.out.println("\n=== CACHE METRICS SUMMARY ===");
        System.out.println("Cache Hits: " + cacheHits.values().stream().mapToLong(Long::longValue).sum());
        System.out.println("Cache Misses: " + cacheMisses.values().stream().mapToLong(Long::longValue).sum());
        System.out.println("DB Queries: " + dbQueries.values().stream().mapToLong(Long::longValue).sum());
    }
}
```

**Production Considerations:**

1. **Lock Expiration (TTL):**
   - Set reasonable lock TTL (10-15 seconds)
   - Prevents deadlock if lock holder crashes
   - Should be longer than typical DB query time

2. **Monitoring:**
   - Track cache hit ratio
   - Monitor lock contention
   - Alert on excessive DB queries
   - Track cache hydration timeouts

3. **Fallback Strategy:**
   - Don't wait indefinitely for cache hydration
   - Implement timeout and fetch directly
   - Prevents request pileup

4. **Error Handling:**
   - Cache failures shouldn't break application
   - Fallback to DB if Redis is down
   - Log failures for investigation

5. **Testing:**
   - Load test with concurrent requests
   - Test Redis failure scenarios
   - Test lock expiration edge cases


## Cache Sizing

How much memory do you need?

**Example: User Profile Cache**
```
User object: 2KB (JSON)
1 million active users
Total memory: 1M * 2KB = 2GB

Add 20% overhead for Redis metadata
Total: 2.4GB
```

**General Rules:**
- Redis can handle 250-500K requests/second per instance
- Single Redis instance: Up to 256GB RAM (AWS ElastiCache)
- For larger datasets: Use Redis Cluster (sharding)


## When NOT to Cache

- Data changes very frequently (hit ratio < 50%)
- Data is already very fast to fetch (< 5ms)
- Data is unique per user (can't be shared across requests)
- Security-sensitive data (credentials, tokens)
- Compliance requirements (audit logs, financial data)


## Popular Caching Solutions

**Redis:**
- Most popular
- Rich data structures (strings, lists, sets, hashes)
- Pub/sub support
- Cluster mode for scaling

**Memcached:**
- Simple key-value only
- Multi-threaded (better CPU utilization)
- Slightly faster for simple use cases

**In-Process Cache (Caffeine, Guava):**
- No network latency
- Limited to single instance memory
- Good for immutable reference data


## Summary: Best Practices

1. **Use Cache-Aside** for most applications
2. **Set appropriate TTL** based on data freshness requirements
3. **Add random jitter to TTL** to prevent cache avalanche
4. **Use distributed locks** to prevent cache stampedes
5. **Monitor cache hit ratio** (target 80-90%)
6. **Invalidate on writes**, don't update
7. **Handle cache failures gracefully** (fallback to DB)
8. **Don't cache everything** - only hot data
9. **Use connection pooling** for Redis clients
10. **Set memory limits** and eviction policies (LRU)

Happy caching! 🚀

-------
Things I learnt
- Caching
- Redis kv
    - Persistence
        - AOF (Append only file)
        -  RDB
- Issues
    - Stampade problem
