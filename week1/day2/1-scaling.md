# Scaling

```
Assignment: 
Build a prototype to build a program that simulates hash based sharding. Like having multiple connections (shards)
```

---

## 1. What is Scaling?

Ability to handle **large number of concurrent** requests

---

## 2. Types of Scaling

### Vertical Scaling
- Add more RAM, CPU, disk to existing server
- Simpler to implement
- Limited by hardware capacity

### Horizontal Scaling
- Add more servers (linear units)
- More complex but unlimited potential
- Requires distributed system design

**Philosophy:**
- Don't over engineer
- If scaling can be done vertically, always prefer vertical
- Only think about horizontal when you need to
- Engineering will always be an enabler to the product (prefer minimal, don't make enginnering primary thing you think about)

---

## 3. Planning for Scaling

**Load testing:**
- Load tests simulating real world usage will help figure out your scaling needs

**Unit of economics:**
- Understanding how many users can be supported per server/instance
- Helps with capacity planning and cost estimation

---

## 4. Horizontal Scaling - Application Layer

When adding scale by adding servers, consider:
- Figure out if the stateful components are capable of handling load too
- Figure out if the upstream APIs are scaled to handle increased traffic
- Figure out if queuing/external systems can handle potential increased load

---

## 5. Database Scaling - Read Replicas

### When to Use
Most workloads are read-heavy (90%+ reads)

### How Replication Works
- Different DBs do replicas differently
    - Push vs Pull
    - Master pushes vs Replicas pull

### Trade-offs

**Replication lag:**
- Adding read replicas can potentially return stale data
- Time delay between write on master and replication to replicas
- Causes eventual consistency

**When stale data is acceptable:**
- Non-critical reads (feeds, dashboards, analytics)

**When stale data is NOT acceptable:**
- Critical reads (financial data, inventory counts)

**Mitigation strategies:**
- Reading from master after write
- One read from master to replica 1, and then update other replicas from replica 1
- Using read-your-writes consistency
- Monitoring replication lag

### Implementation (Application Layer)

**Two connection pools:**
- One for master DB connections
- One for replica DB connections

**Use Master connection for:**
- All writes (INSERT, UPDATE, DELETE)
- Critical reads that require latest data
- Reads immediately after writes (read-your-writes)
- Transactional operations requiring consistency

**Use Replica connection for:**
- Non-critical reads (user feeds, search results)
- Analytics and reporting queries
- Heavy read operations that can tolerate staleness
- Background jobs and batch processing

### Important: Master Can Also Serve Reads

**Common misconception:** reads MUST go to replicas only

**Reality:**
- If writes are minimal and master is underutilized, send reads to master too
- For read-heavy apps (99% reads):
    - Round-robin across ALL nodes (master + replicas)
    - Master participates in read load balancing
    - Only send writes exclusively to master

**Benefits:**
- Better resource utilization
- No replication lag for those reads
- More capacity without adding replicas

---

## 6. Database Scaling - Sharding

### When to Use
When most workloads are writes (read replicas don't help with write load)

### What is Sharding?
Shard aware - different nodes storing different data

### Sharding Strategies

**1. Hash-based ownership**
- Hash key to figure out shard
- Minimal metadata needed
- Even distribution
- Example: `user_id % num_shards`

**2. Range-based ownership**
- Ranges assigned to shards: (a-j) (k-t) (u-z)
- Some metadata needed to store ranges
- Can have uneven distribution (hotspots)

**3. Static-based ownership**
- Static metadata mapping: user1 → DB1, user2 → DB2, user3 → DB1
- Full control over placement
- Most metadata overhead

### Issues with Sharding

**Cross shard queries:**
- Need to query multiple shards and aggregate results at application layer
- Performance is as slow as the slowest shard
- Try to design schema to avoid cross-shard queries

**Shard key selection is critical:**
- Most important to identify proper partition key
- Choose key with even data distribution
- Should match your query patterns
- Hard to change later

**Uneven distribution:**
- Some shards may get more traffic than others (hotspots)
- Need to monitor and rebalance if needed

**Operational complexity:**
- Resharding requires data migration
- Need to backup each shard separately
- Schema changes must coordinate across all shards

**Joins and foreign keys don't work across shards:**
- Must handle at application layer
- May need to denormalize data

---
