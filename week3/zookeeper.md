# Distributed Coordination Services (Zookeeper) (generated needs to be reviewed)

## 1. Quick Summary

**What it is:**
A distributed coordination service is a centralized system that helps multiple distributed nodes agree on shared state, configuration, and synchronization primitives.

**Primary use case:**
Solving coordination problems in distributed systems where multiple processes need to work together reliably.

**Key problems solved:**
- **Leader Election** - Ensuring only one node is the primary
- **Distributed Locking** - Mutual exclusion across multiple processes
- **Service Discovery** - Finding available services dynamically
- **Configuration Management** - Centralized, consistent configuration
- **Group Membership** - Tracking which nodes are alive
- **Consensus** - Agreement on shared state across nodes

**Common implementations:**
- **Zookeeper** - Apache coordination service (used by Kafka, HBase, Hadoop)
- **etcd** - Kubernetes-native (uses Raft consensus)
- **Consul** - Service mesh with service discovery
- **Chubby** - Google's internal lock service (not open source)

---

## 2. Why Do We Need Coordination Services?

### The Problem: Distributed Systems Chaos

When you have multiple servers/processes that need to work together:

**Without Coordination:**
```
Server 1: "I'm the leader!"
Server 2: "No, I'm the leader!"
Server 3: "Both of you are down, I'm the leader!"
Result: 💥 Split-brain, data corruption, chaos
```

**With Coordination Service:**
```
Server 1 → [Coordination Service] → "You're the leader"
Server 2 → [Coordination Service] → "You're standby"
Server 3 → [Coordination Service] → "You're standby"
Result: ✅ Clear consensus, one source of truth
```

### Real-World Scenarios

| Scenario | Problem | Coordination Solution |
|----------|---------|----------------------|
| **Database Failover** | Two masters write conflicting data | Leader election ensures one primary |
| **Scheduled Jobs** | Cron runs on multiple servers → duplicate work | Distributed lock ensures one executor |
| **Microservices** | How does Service A find Service B instances? | Service discovery registry |
| **Feature Flags** | Config changes need to propagate instantly | Centralized configuration with watches |
| **Cluster Management** | Need to know which nodes are alive | Heartbeat + membership tracking |

---

## 3. Core Concepts

### **Consensus & Consistency**

**Consensus:** Agreement among distributed nodes on a single value/state

**Key Properties (needed for coordination):**
- **Consistency** - All nodes see the same data at any time
- **Agreement** - All nodes agree on decisions
- **Ordering** - Operations happen in a defined order
- **Durability** - Decisions persist across failures

**Trade-offs (CAP Theorem):**
- Coordination services choose **CP** (Consistency + Partition Tolerance)
- They sacrifice availability during network partitions
- Why? Better to be unavailable than give wrong answers

### **Quorum & Replication**

**Quorum:** Minimum number of nodes needed to make decisions

```
Formula: Quorum = (N / 2) + 1

3 servers → Quorum = 2 (can tolerate 1 failure)
5 servers → Quorum = 3 (can tolerate 2 failures)
7 servers → Quorum = 4 (can tolerate 3 failures)
```

**Why odd numbers?**
- 4 servers → Quorum = 3 (tolerates 1 failure) ❌ Same as 3 servers
- 5 servers → Quorum = 3 (tolerates 2 failures) ✅ Better tolerance

### **Sessions & Ephemeral State**

**Session:** Connection between client and coordination service

```
Client connects → Session created (with timeout)
        ↓
Heartbeat every N seconds → Session alive
        ↓
No heartbeat for timeout period → Session expired
        ↓
Ephemeral data deleted → Triggers notifications
```

**Why useful?** Automatic cleanup when nodes crash (no manual cleanup needed)

### **Watches & Notifications**

**Watch:** Trigger that fires when data changes

```
Client sets watch on "/leader"
        ↓
Current leader crashes
        ↓
"/leader" deleted
        ↓
Watch fires → All clients notified
        ↓
Clients react (e.g., trigger new election)
```

**Characteristics:**
- Usually **one-time** (must re-register after firing)
- **Asynchronous** (don't block operations)
- **Ordered** (see changes in order they occurred)

### **Data Model**

Most coordination services use a **hierarchical namespace** (like a filesystem):

```
/
├── services/
│   ├── api-server-1    (ephemeral, contains IP:port)
│   ├── api-server-2    (ephemeral)
│   └── api-server-3    (ephemeral)
├── config/
│   ├── database        (persistent, contains connection string)
│   └── feature-flags   (persistent, contains JSON)
└── locks/
    ├── payment-lock    (ephemeral)
    └── job-lock_0001   (ephemeral + sequential)
```

**Key points:**
- Small data (< 1KB typically) - metadata, not storage
- Persistent or ephemeral nodes
- Can have children (like directories)
- Sequential numbering for ordering

---

## 4. Common Coordination Patterns

### Pattern 1: Leader Election

**Problem:** Multiple servers, only one should be active (primary/master)

**Algorithm:**
```
1. Each candidate creates ephemeral sequential node: /election/node_0001, node_0002, etc.
2. Get all children of /election
3. Node with smallest number is the leader
4. Non-leaders watch the node immediately before them
5. When watched node disappears → re-check leadership
```

**Flow Diagram:**
```
Server A creates /election/node_0001 → Smallest number → LEADER ✓
Server B creates /election/node_0002 → Watches node_0001
Server C creates /election/node_0003 → Watches node_0002

Server A crashes → node_0001 deleted (ephemeral)
                 ↓
Server B's watch fires → Checks again → Now smallest → NEW LEADER ✓
Server C still watches node_0002
```

**Why ephemeral + sequential?**
- **Ephemeral**: Auto-cleanup when server crashes
- **Sequential**: Clear ordering, no race conditions
- **Watch previous**: Avoids "herd effect" (everyone waking up at once)

**Real Example: Database Master Election**
```
PostgreSQL Cluster:
- 3 replicas want to be master
- Create nodes in coordination service
- Smallest node ID becomes master
- Others become read replicas
- If master dies → automatic failover
```

---

### Pattern 2: Distributed Locks (Mutual Exclusion)

**Problem:** Ensure only one process executes critical section at a time

**Algorithm:**
```
Acquire Lock:
1. Create ephemeral sequential node: /locks/lock_0001
2. Get all children of /locks
3. If yours is smallest → Lock acquired ✓
4. If not → Watch the node before yours, wait
5. When watch fires → Go to step 2

Release Lock:
1. Delete your node
2. Next waiting node's watch fires → They acquire lock
```

**Flow:**
```
Process A: Creates lock_0001 → Smallest → Lock acquired → Processing...
Process B: Creates lock_0002 → Not smallest → Watches lock_0001 → Waiting...
Process C: Creates lock_0003 → Not smallest → Watches lock_0002 → Waiting...

Process A completes → Deletes lock_0001
Process B's watch fires → lock_0002 now smallest → Lock acquired → Processing...
Process C still waiting...
```

**Why this design?**
- **Fair**: First-come, first-served (FIFO order)
- **No "herd effect"**: Only next in line wakes up
- **Crash-safe**: If holder crashes, ephemeral node deleted → lock released

**Real Example: Cron Job Deduplication**
```
Cron job runs on 10 servers at same time
Without lock: Job executes 10 times (bad!)
With lock: First to acquire runs, others skip
```

---

### Pattern 3: Service Discovery

**Problem:** Dynamic list of service instances, consumers need to find them

**Algorithm:**
```
Service Registration (Provider):
1. Service starts → Create ephemeral node /services/api/instance_0001
2. Store service address (IP:port) in node data
3. Maintain heartbeat → Node stays alive
4. Service crashes → Node deleted automatically

Service Discovery (Consumer):
1. List all children of /services/api → Get all instances
2. Set watch on /services/api
3. When watch fires (new service or removal) → Update local list
4. Load balance across available instances
```

**Example:**
```
Initial State:
/services/api-service/
  ├── instance_0001 → "192.168.1.10:8080"
  ├── instance_0002 → "192.168.1.11:8080"
  └── instance_0003 → "192.168.1.12:8080"

Consumer reads → Has list of 3 servers → Load balances

New instance starts:
  ├── instance_0004 → "192.168.1.13:8080"
Consumer's watch fires → Updates list → Now balances across 4 servers

Instance 0002 crashes:
  (instance_0002 deleted - ephemeral)
Consumer's watch fires → Updates list → Balances across remaining 3 servers
```

**Benefits:**
- **No manual config**: Services auto-register
- **Auto cleanup**: Dead services removed automatically
- **Real-time updates**: Clients notified immediately

---

### Pattern 4: Configuration Management

**Problem:** Need centralized config that updates dynamically across all servers

**Algorithm:**
```
Write Config:
1. Admin updates /config/database with new connection string
2. All watching clients notified

Read Config:
1. Client reads /config/database → Gets current value
2. Sets watch on /config/database
3. When config changes → Watch fires → Re-read config → Apply changes
```

**Example:**
```
/config/
  ├── feature-flags → {"dark_mode": true, "new_checkout": false}
  ├── rate-limits → {"api": 1000, "upload": 100}
  └── database → "postgres://prod-db:5432/myapp"

100 servers watching these configs
Admin changes feature flag → All 100 servers notified instantly
```

**Use Cases:**
- Feature flags (enable/disable features without deploy)
- Rate limits (adjust during traffic spikes)
- Database connection strings (for migrations)
- Circuit breaker thresholds

---

### Pattern 5: Group Membership / Health Monitoring

**Problem:** Track which nodes in cluster are alive

**Algorithm:**
```
Node Joins:
1. Create ephemeral node /cluster/members/node_X
2. Store node metadata (IP, role, capacity)

Monitoring:
1. List children of /cluster/members → Get all alive nodes
2. Set watch on /cluster/members
3. Watch fires when node joins or leaves → Update member list

Node Leaves:
1. Node crashes or disconnects → Ephemeral node deleted
2. Other nodes notified via watch
```

**Example:**
```
/cluster/members/
  ├── web-server-1 (ephemeral) → {"ip": "10.0.1.5", "capacity": 100}
  ├── web-server-2 (ephemeral) → {"ip": "10.0.1.6", "capacity": 100}
  └── web-server-3 (ephemeral) → {"ip": "10.0.1.7", "capacity": 50}

Load Balancer watches → Knows 3 healthy servers

web-server-2 crashes → Node deleted → Load balancer notified
                     → Routes traffic only to server-1 and server-3
```

**Use Cases:**
- Load balancer backend pools
- Kafka brokers tracking
- Elasticsearch cluster membership
- Cache server pools

---

## 5. Key Points to Remember

### Important Gotchas

**1. Watches are typically one-time triggers**
- After firing, must re-register to get future notifications
- Easy to miss events if not re-registered immediately
- Solution: Re-register in the watch callback itself

**2. Not a database - Store metadata only**
- Coordination services are for **small data** (< 1KB ideal)
- Store: service addresses, locks, leader info, config
- **Don't store**: user data, logs, large documents, images
- Why: All data kept in memory for speed; not optimized for storage

**3. Session expiration handling is critical**
- If client can't heartbeat → session times out → ephemeral data deleted
- Must handle: reconnection logic, session re-establishment, state recovery
- Network blip ≠ crash, but coordination service can't tell the difference

**4. Split-brain scenarios still possible**
- Client thinks it's the leader (session expired but client doesn't know yet)
- **Solution**: "Fencing tokens" - increment version number on each leader election
- Old leader's operations rejected if not using current token

**5. Write latency (quorum required)**
- Writes need majority agreement → higher latency than reads
- Not suitable for high-frequency writes
- Typical use: infrequent coordination operations (elections, lock acquisitions)

### Performance Characteristics

| Operation | Performance | Why |
|-----------|-------------|-----|
| **Reads** | Fast (local) ✅ | Any server can serve reads |
| **Writes** | Slower (quorum) ⚠️ | Must replicate to majority |
| **Watches** | Efficient ✅ | Push-based notifications |
| **Throughput** | 10K-100K ops/sec | Good for coordination, not storage |

### Consensus Algorithm Trade-offs

Coordination services typically use:
- **Paxos** (theory): Complex but proven correct
- **Raft** (practical): Easier to understand, widely used (etcd, Consul)
- **ZAB** (Zookeeper Atomic Broadcast): Zookeeper-specific

**Trade-offs:**
- ✅ Strong consistency (everyone sees same state)
- ✅ Fault tolerance (survives minority failures)
- ❌ Availability during partitions (CP, not AP)
- ❌ Write latency (need quorum)

### Common Mistakes

**1. Using coordination service as primary database**
```
❌ Bad: Storing user profiles, posts, analytics in coordination service
✅ Good: Storing which database is primary, service addresses, feature flags
```

**2. Too many clients/watches**
```
❌ Bad: 10,000 clients all watching same node
✅ Good: Use tiered architecture - few coordinators watch, broadcast to clients
```

**3. Large data payloads**
```
❌ Bad: Storing 1MB JSON document in node
✅ Good: Store reference/pointer to actual data in database
```

**4. Ignoring watch firing edge cases**
```
❌ Bad: Assume watch fired = exact change you expect
✅ Good: Read current state when watch fires, handle all possible states
```

**5. Not handling session timeouts**
```
❌ Bad: Assume session stays connected forever
✅ Good: Implement reconnection, re-registration, state validation
```

**6. Even number of servers**
```
❌ Bad: 4 servers (quorum=3, tolerates 1 failure) - wastes resources
✅ Good: 3 servers (quorum=2, tolerates 1 failure) or 5 (quorum=3, tolerates 2)
```

---

## 6. Quick Reference

### Core Operations (Common Across Implementations)

| Operation | Purpose | Example Path |
|-----------|---------|--------------|
| **Create** | Add new node with data | `/services/api-server-1` |
| **Read** | Get data from node | `/config/database` |
| **Update** | Modify existing node data | `/config/feature-flags` |
| **Delete** | Remove node | `/locks/job-lock-001` |
| **List Children** | Get child nodes | `/services/` → list all services |
| **Watch** | Get notified on changes | Watch `/leader` for changes |
| **Check Exists** | See if node exists (with optional watch) | Does `/leader` exist? |

### Node Types

| Type | Lifecycle | Use Case |
|------|-----------|----------|
| **Persistent** | Exists until explicitly deleted | Configuration, structure |
| **Ephemeral** | Deleted when session ends | Locks, membership, presence |
| **Sequential** | Auto-appends number | Leader election, queues |
| **Ephemeral + Sequential** | Both properties | Distributed locks, fair queues |

### Comparison of Popular Implementations

| Feature | Zookeeper | etcd | Consul |
|---------|-----------|------|--------|
| **Algorithm** | ZAB | Raft | Raft |
| **Data Model** | Hierarchical tree | Key-value | Key-value + catalog |
| **Language** | Java | Go | Go |
| **Use Case** | Big Data (Kafka, Hadoop) | Kubernetes | Service mesh |
| **Watch** | Yes (one-time) | Yes (streaming) | Yes (blocking queries) |
| **HTTP API** | Basic | Full REST | Full REST |
| **Service Discovery** | DIY | Basic | Built-in + DNS |
| **Health Checks** | Session-based | TTL-based | Active checks |
| **Performance** | ~100K ops/sec | ~10K writes/sec | Varies |

### When to Use What

```
Zookeeper:
✅ Hadoop/Big Data ecosystem
✅ Legacy systems already using it
✅ Need mature, battle-tested solution
❌ Starting greenfield project

etcd:
✅ Kubernetes environments
✅ Modern cloud-native apps
✅ Need good HTTP API
❌ Complex hierarchical data

Consul:
✅ Service mesh architecture
✅ Built-in health checks needed
✅ Multi-datacenter setups
✅ Service discovery primary concern
❌ Just need simple coordination
```

---

## 7. Real-World Examples & Use Cases

### Where Coordination Services Are Used

#### **1. Apache Kafka - Cluster Metadata**
**Problem:** 100+ brokers need to agree on partition leadership and replicas
**Coordination Pattern:** Leader election + configuration management
**How:**
- Controller election: One broker becomes cluster controller
- Topic/partition metadata stored centrally
- Broker membership tracking (which brokers are alive)
- **Note:** Moving away from Zookeeper to internal Raft (KIP-500)

#### **2. Kubernetes - Cluster State**
**Problem:** API server, schedulers, controllers need shared state
**Coordination Pattern:** etcd for all cluster state
**How:**
- Stores: pod definitions, services, config maps, secrets
- Leader election for controller-manager and scheduler
- Watch mechanism for real-time updates
- Multiple masters stay in sync via etcd consensus

#### **3. Database High Availability (PostgreSQL, MySQL)**
**Problem:** Need automatic failover if primary database fails
**Coordination Pattern:** Leader election + health monitoring
**How:**
- Primary database registers as leader
- Replicas monitor leader via coordination service
- Leader fails → Replicas detect → New leader elected
- Applications redirect to new primary automatically

#### **4. Distributed Cron Jobs**
**Problem:** Scheduled job runs on 50 servers, should execute only once
**Coordination Pattern:** Distributed lock
**How:**
```
12:00 AM - Job triggers on all 50 servers simultaneously
Server 1 acquires lock → Runs job → Releases lock
Servers 2-50 try lock → Fail → Skip execution
Result: Job runs exactly once ✅
```

#### **5. Netflix Microservices**
**Problem:** 1000+ microservices, instances constantly starting/stopping
**Coordination Pattern:** Service discovery + configuration
**How:**
- Service instances register on startup (ephemeral nodes)
- Load balancers watch service registry → Always have current list
- Feature flags stored centrally → Roll out features instantly
- Circuit breaker configs updated dynamically

#### **6. Uber Payment Processing**
**Problem:** Duplicate payment charges must be prevented
**Coordination Pattern:** Distributed lock with timeout
**How:**
```
User clicks "Pay" button
Payment service acquires lock on transaction_ID
Process payment
Release lock
(If service crashes, lock auto-released via session timeout)
```

#### **7. Elasticsearch Cluster Management**
**Problem:** Cluster of 20 nodes needs to coordinate shards and replicas
**Coordination Pattern:** Leader election + membership
**How:**
- Master node election (coordinates cluster operations)
- Node membership tracking (which nodes available)
- Shard allocation decisions centralized
- Cluster state changes replicated via consensus

### Pattern-to-Use-Case Matrix

| Use Case | Leader Election | Distributed Lock | Service Discovery | Config Mgmt | Membership |
|----------|----------------|------------------|-------------------|-------------|------------|
| **Database HA** | ✅ Primary | ❌ | ❌ | ⚠️ Connection info | ✅ Replicas |
| **Kafka Cluster** | ✅ Controller | ❌ | ❌ | ✅ Topics | ✅ Brokers |
| **Kubernetes** | ✅ Controllers | ⚠️ Rare | ❌ (DNS) | ✅ All state | ✅ Nodes |
| **Microservices** | ⚠️ Sometimes | ⚠️ Critical sections | ✅ Primary | ✅ Feature flags | ✅ Instances |
| **Batch Jobs** | ⚠️ Or lock | ✅ Prevent dupes | ❌ | ⚠️ Job params | ❌ |
| **Cache Cluster** | ❌ | ❌ | ✅ Find caches | ✅ Cache params | ✅ Cache nodes |
| **Message Queue** | ✅ Primary | ❌ | ✅ Brokers | ✅ Queues config | ✅ Consumers |

### Common Coordination Patterns Summary

| Pattern | When to Use | Example |
|---------|-------------|---------|
| **Leader Election** | Only one active instance needed | Database primary, Kafka controller, Job scheduler |
| **Distributed Lock** | Mutual exclusion across servers | Payment processing, File writes, Dedup jobs |
| **Service Discovery** | Dynamic service instances | Microservices, Load balancing, Proxy routing |
| **Configuration Management** | Centralized config updates | Feature flags, DB connections, Rate limits |
| **Group Membership** | Track live nodes | Cluster management, Health checks, Shard assignment |
| **Barriers** | Wait for all participants | MapReduce phases, Batch job coordination |
| **Queues** | Fair work distribution | Task queues, Job processing

---

## 8. Related Topics & Further Learning

### Coordination Service Implementations

| Service | Best For | Learn More |
|---------|----------|------------|
| **Zookeeper** | Hadoop/Kafka ecosystem, battle-tested | [Apache Zookeeper Docs](https://zookeeper.apache.org/) |
| **etcd** | Kubernetes, modern cloud-native | [etcd Documentation](https://etcd.io/docs/) |
| **Consul** | Service mesh, multi-DC | [HashiCorp Consul](https://www.consul.io/) |
| **Redis** | Simple locks (less reliable) | [Redlock Algorithm](https://redis.io/docs/manual/patterns/distributed-locks/) |

### Consensus Algorithms (How They Work Internally)

**Paxos** (Theory)
- Original consensus algorithm (1989)
- Notoriously hard to understand and implement
- Proven correct mathematically
- [Paxos Made Simple Paper](https://lamport.azurewebsites.net/pubs/paxos-simple.pdf)

**Raft** (Practical)
- Designed for understandability (2014)
- Leader election → Log replication → Safety
- Used by: etcd, Consul, CockroachDB
- [Raft Visualization](https://raft.github.io/) - Interactive demo!

**ZAB** (Zookeeper Atomic Broadcast)
- Zookeeper-specific protocol
- Similar to Paxos but optimized for Zookeeper use case
- [ZAB Paper](https://marcoserafini.github.io/papers/zab.pdf)

### Fundamental Concepts to Study

**1. CAP Theorem**
- **C**onsistency, **A**vailability, **P**artition Tolerance
- Can only guarantee 2 out of 3
- Coordination services choose **CP** (sacrifice availability during partitions)
- [Visual Guide to CAP](https://mwhittaker.github.io/blog/an_illustrated_proof_of_the_cap_theorem/)

**2. Distributed Transactions**
- Two-Phase Commit (2PC)
- Three-Phase Commit (3PC)
- Saga pattern (for microservices)
- Related but different from coordination

**3. Failure Detection**
- Heartbeat mechanisms
- Phi Accrual Failure Detector (Cassandra)
- Session timeouts
- How systems detect crashed nodes

**4. Split-Brain Problem**
- Network partition → Multiple leaders
- Fencing tokens solution
- Quorum-based prevention
- Why coordination services exist!

**5. Eventual Consistency vs Strong Consistency**
- Coordination services: Strong consistency
- NoSQL databases: Often eventual consistency
- Trade-offs and when to use each

### Architectural Patterns

**Service Mesh**
- Modern alternative for service discovery
- Examples: Istio, Linkerd, Consul Connect
- Built-in health checks, routing, security
- May still use coordination service internally

**Configuration as Code**
- GitOps approach (config in Git)
- Infrastructure as Code (Terraform, Ansible)
- vs. Dynamic config (coordination service)
- Often used together

**Event-Driven Architecture**
- Watches/notifications are event-driven
- Compare to: polling, webhooks, message queues
- Push vs Pull models

### Books & Papers to Read

**Distributed Systems**
- "Designing Data-Intensive Applications" by Martin Kleppmann (Chapter 9: Consistency & Consensus)
- "Database Internals" by Alex Petrov (Part 2: Distributed Systems)

**Papers**
- [Time, Clocks, and Ordering of Events](https://lamport.azurewebsites.net/pubs/time-clocks.pdf) - Lamport (foundational)
- [The Chubby Lock Service](https://research.google/pubs/pub27897/) - Google's coordination service
- [In Search of an Understandable Consensus Algorithm](https://raft.github.io/raft.pdf) - Raft paper

**Online Courses**
- MIT 6.824 Distributed Systems (free lectures online)
- [Distributed Systems lecture series](https://www.youtube.com/playlist?list=PLeKd45zvjcDFUEv_ohr_HdUFe97RItdiB) - Martin Kleppmann

### Debug & Monitoring Considerations

**What to Monitor:**
- Quorum health (are enough nodes alive?)
- Write/read latency
- Session timeout rate (high = network issues)
- Leadership flapping (frequent re-elections = problem)
- Watch notification delays

**Common Issues:**
- Network partitions → Loss of quorum
- Session timeouts → False failure detection
- High latency → Consensus delays
- Too many clients → Overwhelmed service

### When NOT to Use Coordination Services

❌ **Don't use for:**
- Primary data storage (use databases)
- High-frequency updates (too slow)
- Large data (not designed for it)
- Simple single-server apps (overkill)
- Eventually consistent data (use NoSQL)

✅ **Use for:**
- Leader election
- Distributed locking
- Service discovery
- Configuration management
- Group membership
- Rare but critical coordination

---

## Summary

**Coordination services solve a specific problem:** Getting distributed nodes to agree on shared state when failures happen.

**Key takeaways:**
1. Use **ephemeral nodes** for presence (auto-cleanup on crash)
2. Use **sequential nodes** for ordering (leader election, queues)
3. Use **watches** for real-time updates (reactive architecture)
4. Choose **odd number of servers** for quorum (fault tolerance)
5. Store **metadata only** (not application data)
6. Handle **session timeouts** gracefully (network isn't reliable)

**Remember:** Coordination services are infrastructure. They should be invisible to end users, but critical for system reliability.