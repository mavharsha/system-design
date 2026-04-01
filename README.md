# System Design Preparation Notes

This repo is a snapshot of all the topics and resources explored regarding system design.
Initially planned to complete the System Design Master Class course.

## 📑 Quick Navigation

**[Week 1](#week-1-fundamentals--scaling-basics)** • **[Week 2](#week-2-database-deep-dive--real-time-systems)** • **[Week 3](#week-3-distributed-systems--id-generation)** • **[Week 4](#week-4-cdn-image-services--hashtags)** • **[Week 5](#week-5-cache-design--data-structures)** • **[Week 6](#week-6-storage-engines--streaming)** • **[Week 7](#week-7-search--scheduling--flash-sales)** • **[Week 8](#week-8-analytics--counting)**

---

## 📚 Repository Structure

### 🔧 `basics/`
Fundamental concurrency and synchronization concepts:
- **CPU vs I/O Bound** - Understanding operation types and optimization
- **Mutex** - Mutual exclusion for critical sections
- **Semaphore** - Resource count-based synchronization
- **Mind Map** - Technology stack overview

### 💾 `db/`
Deep dive into database concepts and design:
- **ACID Properties** - Atomicity, Consistency, Isolation, Durability
- **Write-Ahead Logging (WAL)** - Durability mechanism
- **Write Sequence** - Understanding database write operations
- **Storage Engines** - How databases store data
- **Indexes** - Dense vs sparse, types and use cases
- **Two-Phase Commit** - Distributed transaction protocol
- **B+ Trees vs LSM Trees** - Database engine comparison
- **NoSQL** - Cassandra, Graph databases, LSM trees, SSTables, Memtables
- **Q&A** - Common database interview questions

### 🗃️ `cache/`
- **Single Node Cache** - `TODO` - Stub, needs implementation details
- **Distributed Cache** - `TODO` - Stub, needs detailed content
- **Redis** - Single-threaded event loop, connection handling, Java implementation

### 🌳 `data-structures/`
- **B+ Tree** - Structure, operations, database usage
- **Bloom Filter** - `TODO` - Stub, only has minimal use case notes

### 🧠 `mind-map-interactive/`
Interactive web-based mind map viewer for system design, Java, and Angular topics. Deployed via GitHub Actions.

### 🤖 `prompts/`
Claude Code prompt templates: code review, debugging, architecture, performance, technical writing.

---

## 📅 Weekly Learning Path

### Week 1: Fundamentals & Scaling Basics

#### Day 1: Core Patterns & Caching
**Topics:**

1. **Online/Offline Indicator System**
   - Heartbeat mechanisms (client pushes status every 20 seconds)
   - Redis vs DynamoDB trade-offs for key-value storage
   - Connection pooling with blocking queues
   - Sparse vs dense data storage strategies
   - WebSocket implementations for real-time status
   - Assignment: Build blocking queue implementation (connection pool)

2. **Database Design**
   - Schema design for blogging platforms (Medium-like)
   - Soft deletes vs hard deletes
   - Index optimization strategies (partial/filtered indexes, composite indexes)
   - Managing database rebalancing and fragmentation
   - Impact of soft deletes on index performance

3. **Caching Strategies**
   - **Caching Patterns:**
     - Cache-aside (lazy loading) - most common
     - Read-through, write-through, write-behind
   - **Cache Eviction Policies:** LRU, LFU, FIFO, TTL
   - **Cache Problems & Solutions:**
     - Cache stampede (thundering herd) with distributed locks (Only this was done in class)
     - Cache penetration with bloom filters
     - Cache avalanche prevention
   - Production-ready cache implementations in Java
   - Cache sizing and capacity planning
   - When NOT to cache

**Key Learnings:**
- Batching requests for efficiency
- Connection pooling reduces 3-way handshake overhead
- Redis distributed locking with SET NX EX
- Invalidate cache on writes (don't update!)
- Trade-offs in consistency vs availability

---

#### Day 2: Scaling & Distributed Systems
**Topics:**

1. **Scaling Fundamentals**
   - Vertical vs horizontal scaling philosophy
   - Load testing and unit economics
   - Planning for scale (don't over-engineer early)
   - Assignment: Build hash-based sharding simulation

2. **Database Read Replicas**
   - Master-replica architecture
   - Replication lag and eventual consistency
   - When stale data is acceptable vs critical reads
   - Read-your-writes consistency patterns
   - Master can also serve reads (don't waste capacity!)
   - Two connection pools: master vs replica

3. **Database Sharding**
   - When to shard: write-heavy workloads
   - **Sharding Strategies:**
     - Hash-based ownership (minimal metadata)
     - Range-based ownership (can have hotspots)
     - Static-based ownership (full control)
   - Cross-shard queries and aggregation
   - Shard key selection is critical

4. **ProxySQL**
   - Database routing and query distribution
   - Connection multiplexing

5. **Delegation (Async Workers)**
   - Decoupling services with queues
   - Background job processing

6. **Communication Patterns**
   - Kafka for event streaming
   - Message brokers and pub-sub

**Key Learnings:**
- Don't over-engineer - vertical scaling first
- Read replicas scale reads, sharding scales writes
- Replication lag is real - design for eventual consistency
- Shard key selection is the most important decision
- Cross-shard queries are expensive

---

### Week 2: Database Deep Dive & Real-time Systems

#### Day 1: ACID & Locking Mechanisms
**Topics:**

1. **ACID Properties**
   - **Atomicity:** All or nothing transactions
   - **Consistency:** Valid state to valid state transitions
   - **Durability:** Changes are permanent after commit (via WAL)
   - **Isolation:** Concurrent transactions execute independently

2. **Database Locking** - `TODO` - Partial content, sections marked as "(fill)"
   - Need for locking in high-contention scenarios
   - Use cases: Fixed inventory (IRCTC, BookMyShow, flash sales)
   - **Pessimistic Locking:**
     - Shared locks (read locks) vs Exclusive locks (write locks)
     - `FOR UPDATE` - lock rows for update
     - `FOR UPDATE NOWAIT` - fail immediately if locked
     - `FOR UPDATE SKIP LOCKED` - skip locked rows (great for queue processing)
     - Lock manager and deadlock detection
   - **Optimistic Locking:**
     - Assumes conflicts are rare
     - Checks for conflicts only at commit time
     - Version-based concurrency control
   - Assignment: Explore different locking strategies with contention scenarios

3. **Storage-Compute Separation**
   - Modern database architectures
   - Cloud-native database design patterns
   - Decoupling storage and compute layers

4. **Read Replica Lag Management**
   - Understanding eventual consistency
   - Measuring and monitoring replication lag
   - Read-your-writes consistency patterns
   - Master vs replica routing strategies

**Key Learnings:**
- Lock manager internals and deadlock detection
- `SKIP LOCKED` is perfect for job queues
- Trade-offs between pessimistic and optimistic locking
- Replication lag monitoring is critical

---

#### Day 2: NoSQL & Real-time Systems
**Topics:**

1. **Non-Relational Databases**
   
   **Column-Oriented (OLAP):**
   - Data stored column-by-column (vs row-by-row in OLTP)
   - Optimized for analytics and aggregations
   - Examples: Redshift, BigQuery
   
   **Graph Databases:**
   - Nodes, Edges, and Properties model
   - Index-free adjacency (O(1) traversals)
   - Use cases: Social networks, recommendations, fraud detection
   - Examples: Neo4j, Amazon Neptune, Dgraph
   
   **Wide-Column (Cassandra):**
   - LSM tree architecture (Log-Structured Merge tree)
   - Fast writes (append-only, no random disk I/O)
   - Tunable consistency levels: ONE, QUORUM, ALL, LOCAL_QUORUM
   - Replication Factor (RF) determines redundancy
   - Partition keys and clustering (sort) keys
   - Secondary indexes (scatter-gather pattern)
   - Design tables per access pattern (data duplication is OK)
   - Trade-off: Availability + Partition tolerance over Consistency (AP in CAP)

2. **Cassandra Deep Dive**
   - Writing with different consistency levels
   - Primary vs secondary indexes
   - Query-specific table design patterns
   - When to use secondary indexes vs duplicate tables

3. **Slack Real-time Messaging System**
   - WebSocket architecture for persistent connections
   - Scaling WebSocket connections across servers
   - Message delivery guarantees
   - Final design considerations

**Key Learnings:**
- NoSQL databases are specialized for specific access patterns
- Graph databases excel at relationship queries
- Cassandra trades consistency for availability and scale
- Index-free adjacency makes graph traversals O(1)
- Design Cassandra tables for queries, not for normalized data
- Secondary indexes in distributed systems require scatter-gather

---

### Week 3: Distributed Systems & ID Generation

#### Day 1: Load Balancing & Coordination
**Topics:**

1. **Load Balancer**
   - Layer 4 (Transport layer) vs Layer 7 (Application layer) load balancing
   - Health checks and failover mechanisms
   - Load balancing algorithms (round-robin, least connections, etc.)
   - Sticky sessions and session affinity

2. **Distributed Locks** - `TODO` - Partial content, needs review
   - Lock-free data structures
   - Distributed consensus protocols
   - Handling network partitions

3. **Zookeeper**
   - Distributed coordination service
   - Leader election algorithms
   - Configuration management across clusters
   - Service discovery

**Key Learnings:**
- Layer 7 load balancers can route based on content
- Zookeeper provides strong consistency for coordination
- Distributed locks are complex - use carefully
- Leader election prevents split-brain scenarios

---

#### Day 2: ID Generation & Distribution Challenges
**Topics:**

1. **Foundation of Distributed IDs**
   - Problem: Assign globally unique ID to anything
   - Why auto-increment fails in distributed/sharded systems
   - Collision prevention strategies
   - Evolution: Timestamp → Machine ID + Timestamp → Machine ID + Counter
   - Persistence and recovery (buffer and flush optimization)

2. **Monotonic Increasing IDs**
   - **What:** IDs that always increase over time
   - **Why Needed:**
     - Database performance (B+ tree efficiency)
     - Sequential inserts are 2-10x faster
     - Reduced page splits and fragmentation
     - Natural ordering for time-range queries
     - Efficient sharding by range
     - Easy debugging and observability
   - **Issues:**
     - Hot shard problem (all new writes go to latest shard)
     - Reveals business metrics to competitors
     - Clock skew can break monotonicity

3. **Twitter Snowflake**
   - 64-bit ID structure: `timestamp (41 bits) | machine_id (10 bits) | sequence (12 bits)`
   - Decentralized generation (no single point of failure)
   - Clock skew handling strategies
   - Generates ~4000 IDs per millisecond per machine

4. **Amazon's Centralized ID Service**
   - Central "ID authority" microservice
   - Batch allocation (allocate blocks of 500-1000 IDs at a time)
   - Reduces network calls and improves performance
   - High availability through replication
   - Trade-offs: complexity vs guaranteed uniqueness
   - ID format encodes timestamp, datacenter, machine, sequence

5. **Pagination**
   - Cursor-based pagination (efficient with monotonic IDs)
   - Offset-based pagination (can be slow and inconsistent)
   - `WHERE id > last_seen_id LIMIT 20` pattern

6. **Hot Shard Problem**
   - Monotonic IDs cause uneven load (newest shard gets all writes)
   - Mitigation strategies:
     - Reverse ID ordering
     - Add randomness to shard key
     - Monitor and rebalance
     - Use hash-based distribution for writes

**Key Learnings:**
- Monotonic IDs optimize database inserts but create hot shards
- Twitter Snowflake: decentralized, Amazon: centralized with batching
- Both approaches solve collision problem differently
- Clock synchronization is critical for distributed ID generation
- Batch allocation reduces network overhead significantly
- Trade-off: sequential ordering vs even load distribution

---

### Week 4: CDN, Image Services & Hashtags

#### Day 1: Content Delivery & Image Services
**Topics:**

1. **CDN** - `TODO` - Stub, only has basic overview
   - What CDN is and how it works

2. **Image Upload Service**
   - Architecture and database schema
   - Pre-signed URLs for direct-to-S3 uploads
   - Java examples and client-side upload code
   - Security considerations

3. **Gravatar System Design**
   - Requirements, database schema, indexes
   - API design and CDN architecture
   - Workflow and scaling considerations

---

#### Day 2: Social Features
**Topics:**

1. **Hashtag Service (Instagram)** - `TODO` - Stub, only has high-level brainstorm notes
   - Trade-off questions and initial design thoughts

2. **Newly Unread Indicator**
   - On-the-fly vs pre-computed approaches
   - Implementation examples, pros/cons analysis
   - Scaling recommendations

---

### Week 5: Cache Design & Data Structures

#### Day 1: Cache Implementation
**Topics:**

1. **Designing a Single Node Cache** - `TODO` - Stub, needs implementation details
   - Communication protocol, storage, threading, memory measurement

2. **Designing a Distributed Cache** - `TODO` - Stub, needs detailed content
   - Overview of approaches: proxy-based, client-side, etc.

---

#### Day 2: Storage Without Traditional DBs
**Topics:**

1. **Word Dictionary Without Any DB** - `TODO` - Stub, only has requirements and basic CSV approach
2. **`0.md`** - `TODO` - Empty file
3. **Superfast Persistent KV Store** - `TODO` - Stub, only has requirements and beginning of solution

---

### Week 6: Storage Engines & Streaming

#### Day 1: Data Storage & Ingestion
**Topics:**

1. **LSM Tree**
   - Comprehensive explanation of Log-Structured Merge trees
   - Links to related topics (SSTables, Memtables)

2. **Multi-tier Datastore**
   - Scaling data across hot/cold storage
   - Access patterns and data movement process
   - Read path decision logic and architecture flow

3. **Data Ingestion** - `TODO` - Stub, only has a video reference note

---

#### Day 2: Streaming & Object Storage
**Topics:**

1. **Live Streaming**
   - HLS protocol and FFmpeg chunking
   - Adaptive bitrate streaming
   - Server-side ad insertion with Java pseudocode

2. **Designing S3** - `TODO` - Stub, only has high-level three-piece architecture overview

---

### Week 7: Search, Scheduling & Flash Sales

> **`TODO`** - All files in this week are empty

#### Day 1:
1. **Recent Searches** - `TODO` - Empty

#### Day 2:
1. **Distributed Task Scheduler** - `TODO` - Empty
2. **Flash Sale** - `TODO` - Empty
3. **Flash Sale with High Volume Items** - `TODO` - Empty

---

### Week 8: Analytics & Counting

> **`TODO`** - Content not yet written

1. **Impression Counting** - `TODO` - Stub, only has introductory question

---

## 🎯 Learning Methodology

This repository follows a structured approach:
1. **Understand the Problem** - Requirements and constraints
2. **Explore Solutions** - Multiple approaches and trade-offs
3. **Deep Dive** - Implementation details and edge cases
4. **Practical Examples** - Code samples in Java, Go, Python
5. **Production Considerations** - Monitoring, scaling, error handling

---

## 🚀 Key Takeaways

### System Design Principles:
- Don't over-engineer - prefer vertical scaling when possible
- Load test to understand your scaling needs
- Design for failures - everything will fail eventually
- Choose consistency vs availability based on business requirements
- Cache intelligently - not everything needs caching
- Monitor everything - you can't improve what you don't measure

### Performance Patterns:
- Connection pooling reduces overhead
- Batching improves throughput
- Async processing decouples services
- Read replicas scale reads, sharding scales writes
- Distributed locks prevent thundering herd

### Database Design:
- Index strategy matters more than you think
- Soft deletes for compliance, hard deletes for performance
- Understand your access patterns before choosing NoSQL
- Replication lag is real - design for eventual consistency
- Sharding is complex - only do it when necessary

---

## 📝 Note Structure

Each note typically includes:
- **Problem Statement** - What are we solving?
- **Requirements** - Functional and non-functional
- **Solutions** - Multiple approaches with pros/cons
- **Implementation** - Code examples where applicable
- **Trade-offs** - What you gain and lose
- **Key Learnings** - Summary of important concepts

---
