## Locking (paritially generated needs to be reviewed)
```
 Theme: Locking and concurrency control in software systems.

 Description: This document introduces locking, which is a fundamental mechanism used to control concurrent access to shared resources in computing. Locks ensure that only one process or thread accesses a resource (such as a file, database row, or critical section of memory) at a time, preventing data corruption, race conditions, and inconsistent state.

 Potential Use Cases:
 - **Database row or table locking**: Ensuring that two transactions do not modify the same data simultaneously.
 - **Critical section protection in multithreaded programs**: Preventing data races in shared memory.
 - **File locking**: Coordinating access by multiple processes to a file to prevent corruption.
 - **Distributed/Remote locking**: Coordinating work in distributed systems (e.g., microservices, distributed background job workers, leader election, cron job scheduling).

 
```

#### Remote locking

**Explanation:**
Remote locking (or distributed locking) coordinates access to shared resources across multiple processes or servers. Unlike local locks (mutex, semaphore) that work within a single process, remote locks use an external system (like Redis, ZooKeeper) to ensure mutual exclusion across distributed systems.

**Example use case:**
Distributed background workers processing jobs from a queue (e.g., email sending, order processing). To avoid two workers processing the same job at the same time, they acquire a remote lock on the job before working on it. Only the worker with the lock proceeds; others skip or wait.


**Example:**
```python
import redis
import uuid

client = redis.Redis(host='localhost', port=6379)

# Acquire lock
lock_id = str(uuid.uuid4())
acquired = client.set('lock:payment_123', lock_id, nx=True, ex=30)

if acquired:
    try:
        # Process payment (critical section)
        process_payment()
    finally:
        # Release lock safely (check ownership)
        script = """
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        end
        """
        client.eval(script, 1, 'lock:payment_123', lock_id)
```

**Key Properties:**
-  **Atomic execution** - Lock acquire/release are atomic operations (SET NX, Lua scripts)
-  **Timeouts (automatic expiration)** - Locks expire automatically to prevent deadlocks if holder crashes

---

##### Distributed locks (Redis-based)

**Explanation:**
Distributed locks use Redis as a coordination service. The lock is stored as a key-value pair with:
- **Key**: Resource identifier (e.g., `lock:payment_123`)
- **Value**: Unique lock ID (UUID) to verify ownership
- **TTL**: Expiration time for automatic cleanup

**Example use case:**
Suppose you have a fleet of web servers that need to periodically perform a maintenance task (like cleaning up expired sessions), but you want to make sure only one server performs the task at a time to avoid conflicts or redundant work. Each server tries to acquire the distributed lock before running the task. Only the server that successfully acquires the lock does the cleanup; others skip and try again later.


**Simple Pattern:**
```redis
SET lock:resource unique_id NX EX 30
```
- `NX` = Only set if not exists (atomic)
- `EX 30` = Expire in 30 seconds

**Redlock Algorithm (High Availability):**
For production systems, use multiple Redis instances:
1. Try to acquire lock on N/2 + 1 instances (e.g., 3 out of 5)
2. If majority acquired within timeout → lock successful
3. Otherwise, release all locks and retry

**Example (Redlock):**
```python
from redlock import Redlock

# 5 independent Redis masters
dlm = Redlock([
    {"host": "redis1", "port": 6379},
    {"host": "redis2", "port": 6379},
    {"host": "redis3", "port": 6379},
    {"host": "redis4", "port": 6379},
    {"host": "redis5", "port": 6379}
])

lock = dlm.lock("payment_123", 10000)  # 10 second TTL
if lock:
    try:
        process_payment()
    finally:
        dlm.unlock(lock)
```

---

#### Comparison: Lock Approaches

| Approach | Pros | Cons | Use Case |
|----------|------|------|----------|
| **Single Redis** | Simple, fast, low latency | Single point of failure | Dev environments, non-critical locks |
| **Redlock (Multiple Redis)** | Fault tolerant, no SPOF | More complex, slower | Production, critical operations |
| **ZooKeeper** | Strong consistency, proven | Higher latency, complex setup | Leader election, long-lived locks |
| **Database locks** | Transactional guarantees | Not scalable, DB bottleneck | Single DB systems |
| **etcd** | Strong consistency, Raft | Operational overhead | Kubernetes, cloud-native |

**When to use each:**
- **Redis (single)**: High-performance, short-lived locks, acceptable failure risk
- **Redlock**: Critical business operations, need fault tolerance
- **ZooKeeper/etcd**: Complex coordination, leader election, strong consistency requirements

----
Read about locking
Process
1. Mutex
2. Semaphore
Distrubed 
1. Remote 