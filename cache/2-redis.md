# Redis

- Today, single node of redis can support 300K ops/second
- Uses single threaded event loop. Only utilizes one core cpu.

Supports GET/SET/DEL of the the following datastructers
- SETS
- LIST 
- JSON
- Bloomfilter
- Hyperlogs
- Full Text
..
..
..


## Notes: Single Threaded Event Loop in Redis

- Redis processes all commands from clients using a single thread and event loop. This keeps the design simple and eliminates the need for locks or complex concurrency logic.
- All commands (reads/writes) are executed one at a time, in the order they are received.
- Redis leverages “I/O multiplexing” (using `epoll`, `kqueue`, etc.) so it can handle many simultaneous connections efficiently, but still only one command is processed at any given moment.
- This design prevents race conditions and data corruption since only one thread ever modifies the data structures.
- Although the single-threaded model may sound like a bottleneck, Redis is extremely fast because commands are simple and memory-based, and modern CPUs can process hundreds of thousands of operations per second.
- Background tasks such as RDB/AOF persistence and key expiration can use additional threads, but they do not affect the single-threaded processing of client requests.
- Predictable performance: there’s no lock contention and all requests are executed sequentially, which ensures consistent latency and response order.
- If CPU-bound or multi-core performance becomes a concern, multiple Redis instances can be run on the same server—one per core.

> Core idea: "Single-threaded event loop = simple, safe, very fast for in-memory workloads."


## How Redis tracks connections and Event loop handles these connections
### Client Connections and File Descriptors in Redis

Complete flow:

### When a Client Connects to Redis

**Step-by-step:**

```
1. Client opens TCP connection to Redis (port 6379)
   ↓
2. OS creates a socket and assigns a file descriptor (e.g., FD 7)
   ↓
3. Redis accepts the connection and gets the FD
   ↓
4. Redis registers FD 7 with the event loop
   ↓
5. Event loop now monitors FD 7 for incoming data
```

### The Relationship

```
Client Connection = Network Socket = File Descriptor

One client connection = One socket = One FD
```

### Visual Representation

```
[Client A] ---TCP connection---> Redis gets FD 5
[Client B] ---TCP connection---> Redis gets FD 6  
[Client C] ---TCP connection---> Redis gets FD 8
[Client D] ---TCP connection---> Redis gets FD 9

Redis Event Loop watches: FD 5, 6, 8, 9
```

### Redis Internal Structure

```c
// Simplified Redis client structure
typedef struct client {
    int fd;                    // The file descriptor for this connection!
    sds querybuf;              // Buffer for incoming commands
    list *reply;               // Buffer for outgoing responses
    redisDb *db;              // Which database they're using
    // ... more fields
} client;
```

### Complete Flow Example

```
1. Client sends: SET mykey "hello"
   ↓
2. Data arrives on network → Activity on FD 7
   ↓
3. Event loop detects: "FD 7 has data!"
   ↓
4. Redis calls: readQueryFromClient(FD 7)
   ↓
5. Redis reads from FD 7: "SET mykey 'hello'"
   ↓
6. Redis processes command
   ↓
7. Redis writes response to FD 7: "+OK\r\n"
   ↓
8. Client receives response through same FD
   ↓
9. Event loop goes back to monitoring all FDs
```

### Code Flow in Redis

```c
// 1. Accept new connection
int clientfd = accept(serverfd, ...);  // OS gives us FD

// 2. Create client object
client *c = createClient(clientfd);    // Store FD in client struct

// 3. Register with event loop
aeCreateFileEvent(eventLoop, 
                  clientfd,            // Watch this FD
                  AE_READABLE,         // For read events
                  readQueryFromClient, // Call this function
                  c);                  // Pass client object

// 4. When data arrives, event loop calls:
void readQueryFromClient(aeEventLoop *el, int fd, void *privdata, int mask) {
    client *c = (client*) privdata;
    
    // Read from the file descriptor
    nread = read(fd, c->querybuf, PROTO_IOBUF_LEN);
    
    // Process the command
    processInputBuffer(c);
}

// 5. Send response back through same FD
void sendReplyToClient(int fd, client *c) {
    write(fd, c->reply, replylen);
}
```

### Why Multiple FDs?

```
Redis Server Socket (FD 4) - Listens for new connections
    ↓
Accepts connections and creates:
    Client 1 Socket (FD 5)
    Client 2 Socket (FD 6)
    Client 3 Socket (FD 7)
    ...

Event loop monitors ALL of them simultaneously
```

### Practical Example

```bash
# Terminal 1: Start Redis
redis-server

# Terminal 2: Connect client A
redis-cli
# Redis assigns FD (e.g., 8) to this connection

# Terminal 3: Connect client B  
redis-cli
# Redis assigns FD (e.g., 9) to this connection

# Terminal 4: Check Redis connections
redis-cli CLIENT LIST
# Output shows:
# id=1 addr=127.0.0.1:52341 fd=8 ...
# id=2 addr=127.0.0.1:52342 fd=9 ...
                          ↑
                    File Descriptor!
```

### The Magic of Event Loop

```
WITHOUT event loop (traditional approach):
Thread 1 → blocks on FD 5 waiting for Client A
Thread 2 → blocks on FD 6 waiting for Client B
Thread 3 → blocks on FD 7 waiting for Client C
// Need 10,000 threads for 10,000 clients!

WITH event loop (Redis approach):
Single thread → watches FD 5, 6, 7, ..., 10004
Only processes FDs that have activity
// 1 thread handles 10,000 clients!
```

### Connection Lifecycle

```
1. connect()     → OS assigns FD
2. accept()      → Redis gets FD
3. aeCreateFileEvent() → Event loop monitors FD
4. [Commands flow through FD]
5. close()       → FD is released and can be reused
```

### Checking FDs in Redis

```bash
# See client connections with FDs
redis-cli CLIENT LIST

# Sample output:
id=3 addr=127.0.0.1:50788 fd=8 name= age=5 idle=0 ...
                           ↑
                    This is the file descriptor
                    for this client's TCP connection
```

**Bottom line:** Each client connection is a TCP socket, and each socket has a file descriptor. Redis stores that FD and uses it to read commands from and write responses to that specific client. The event loop watches all these FDs simultaneously, waking up only when any of them has activity!





Java sample redis implementation

```java
import redis.clients.jedis.Jedis;
import redis.clients.jedis.JedisPool;
import redis.clients.jedis.JedisPoolConfig;
import java.util.Map;
import java.util.Set;

/**
 * Redis Key-Value Store Implementation
 * Add Jedis dependency to your pom.xml:
 * <dependency>
 *     <groupId>redis.clients</groupId>
 *     <artifactId>jedis</artifactId>
 *     <version>5.0.0</version>
 * </dependency>
 */
public class RedisKVStore {
    private JedisPool pool;
    
    // Constructor with default configuration
    public RedisKVStore() {
        this("localhost", 6379);
    }
    
    // Constructor with custom host and port
    public RedisKVStore(String host, int port) {
        JedisPoolConfig poolConfig = new JedisPoolConfig();
        poolConfig.setMaxTotal(128);
        poolConfig.setMaxIdle(128);
        poolConfig.setMinIdle(16);
        poolConfig.setTestOnBorrow(true);
        poolConfig.setTestOnReturn(true);
        poolConfig.setTestWhileIdle(true);
        
        this.pool = new JedisPool(poolConfig, host, port);
    }
    
    // Set a key-value pair
    public void set(String key, String value) {
        try (Jedis jedis = pool.getResource()) {
            jedis.set(key, value);
        }
    }
    
    // Set with expiration time (in seconds)
    public void setWithExpiry(String key, String value, int seconds) {
        try (Jedis jedis = pool.getResource()) {
            jedis.setex(key, seconds, value);
        }
    }
    
    // Get value by key
    public String get(String key) {
        try (Jedis jedis = pool.getResource()) {
            return jedis.get(key);
        }
    }
    
    // Check if key exists
    public boolean exists(String key) {
        try (Jedis jedis = pool.getResource()) {
            return jedis.exists(key);
        }
    }
    
    // Delete a key
    public long delete(String key) {
        try (Jedis jedis = pool.getResource()) {
            return jedis.del(key);
        }
    }
    
    // Delete multiple keys
    public long delete(String... keys) {
        try (Jedis jedis = pool.getResource()) {
            return jedis.del(keys);
        }
    }
    
    // Set multiple key-value pairs at once
    public void setMultiple(Map<String, String> keyValues) {
        try (Jedis jedis = pool.getResource()) {
            String[] keysAndValues = new String[keyValues.size() * 2];
            int i = 0;
            for (Map.Entry<String, String> entry : keyValues.entrySet()) {
                keysAndValues[i++] = entry.getKey();
                keysAndValues[i++] = entry.getValue();
            }
            jedis.mset(keysAndValues);
        }
    }
    
    // Get multiple values by keys
    public Map<String, String> getMultiple(String... keys) {
        try (Jedis jedis = pool.getResource()) {
            var values = jedis.mget(keys);
            return Map.ofEntries(
                java.util.stream.IntStream.range(0, keys.length)
                    .filter(i -> values.get(i) != null)
                    .mapToObj(i -> Map.entry(keys[i], values.get(i)))
                    .toArray(Map.Entry[]::new)
            );
        }
    }
    
    // Increment value (must be integer)
    public long increment(String key) {
        try (Jedis jedis = pool.getResource()) {
            return jedis.incr(key);
        }
    }
    
    // Increment by specific amount
    public long incrementBy(String key, long amount) {
        try (Jedis jedis = pool.getResource()) {
            return jedis.incrBy(key, amount);
        }
    }
    
    // Get all keys matching pattern
    public Set<String> getKeys(String pattern) {
        try (Jedis jedis = pool.getResource()) {
            return jedis.keys(pattern);
        }
    }
    
    // Set expiration on existing key
    public long expire(String key, int seconds) {
        try (Jedis jedis = pool.getResource()) {
            return jedis.expire(key, seconds);
        }
    }
    
    // Get time to live for key
    public long getTTL(String key) {
        try (Jedis jedis = pool.getResource()) {
            return jedis.ttl(key);
        }
    }
    
    // Close the connection pool
    public void close() {
        if (pool != null && !pool.isClosed()) {
            pool.close();
        }
    }
    
    // Example usage
    public static void main(String[] args) {
        RedisKVStore store = new RedisKVStore();
        
        try {
            // Basic set and get
            store.set("user:1:name", "John Doe");
            System.out.println("Name: " + store.get("user:1:name"));
            
            // Set with expiration (10 seconds)
            store.setWithExpiry("session:abc123", "active", 10);
            System.out.println("Session: " + store.get("session:abc123"));
            System.out.println("TTL: " + store.getTTL("session:abc123") + " seconds");
            
            // Multiple operations
            Map<String, String> users = Map.of(
                "user:2:name", "Jane Smith",
                "user:3:name", "Bob Johnson"
            );
            store.setMultiple(users);
            
            // Get multiple
            var retrieved = store.getMultiple("user:1:name", "user:2:name", "user:3:name");
            System.out.println("All users: " + retrieved);
            
            // Counter example
            store.set("page:views", "0");
            store.increment("page:views");
            store.incrementBy("page:views", 5);
            System.out.println("Page views: " + store.get("page:views"));
            
            // Pattern matching
            Set<String> userKeys = store.getKeys("user:*");
            System.out.println("User keys: " + userKeys);
            
            // Cleanup
            store.delete("user:1:name", "user:2:name", "user:3:name");
            
        } finally {
            store.close();
        }
    }
}
```