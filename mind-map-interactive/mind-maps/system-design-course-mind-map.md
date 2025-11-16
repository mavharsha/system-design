# System Design Course: Week by Week

## Week 1: Fundamentals & Core Patterns

### Online/Offline Indicator System
- Requirements: Track user online/offline status
- Access patterns: Key-value lookups
- GET operations: Batch requests for multiple users
- POST operations: Client heartbeat (push-based, every 20 seconds)
- Database schema: userId, epoch timestamp
- Storage strategies: Dense vs Sparse storage
- Dense: Store all users (including offline) - simpler queries
- Sparse: Only store online users - saves space
- Connection pooling: Blocking queue implementation
- Technologies: Redis, DynamoDB for key-value storage

### Database Design & Schema
- Schema design for blogging platforms (Medium.com)
- Users table: id, name, bio
- Blogs table: id, author_id, title, is_deleted, published_at, body
- Soft deletes: is_deleted flag
- Importance: Compliance, auditability, data recovery, analytics
- Index issues with soft deletes: Index bloat, query performance overhead
- Optimization: Partial/filtered indexes, composite indexes
- Hard deletes: DB rebalancing issues, index fragmentation
- Best practices: Batch deletions, VACUUM operations, archiving

### Caching Strategies
- What is caching: High-speed storage layer between app and DB
- Why cache: Speed (memory < 1ms vs disk 10-100ms+), reduced DB load, cost savings
- Cache hit vs miss: Hit ratio target 80-90%
- Cache-Aside (Lazy Loading): Most common pattern
- Read flow: Check cache → DB on miss → populate cache
- Write flow: Write to DB → invalidate cache (don't update!)
- Read-Through: Cache library handles DB access automatically
- Write-Through: Write to cache and DB simultaneously
- Write-Behind (Write-Back): Write to cache, async write to DB
- Refresh-Ahead: Proactive cache warming
- Cache invalidation strategies: TTL, event-based, manual
- Cache stampede prevention: Locks, jitter, probabilistic early expiration

### Scaling Fundamentals
- Types: Vertical vs Horizontal scaling
- Vertical: Add RAM, CPU, disk to existing server
- Horizontal: Add more servers (linear units)
- Philosophy: Prefer vertical, only horizontal when needed
- Load testing: Simulate real-world usage
- Unit of economics: Users per server/instance
- Application layer scaling: Stateful vs stateless components
- Database scaling: Read replicas, sharding, connection pooling
- Hash-based sharding: Prototype implementation

### B+ Trees in SQL Databases
- Fundamental data structure for SQL databases
- Optimized for disk access: Minimize I/O operations
- Balanced structure: All leaf nodes at same level
- Sequential access: Linked leaf nodes for range queries
- High fanout: Many children per node, low tree height
- Internal nodes: Store keys and pointers only
- Leaf nodes: Store keys and actual data
- Operations: Search O(log n), Insert O(log n), Delete O(log n)
- Database indexes: Primary, secondary, composite indexes
- Real-world: MySQL, PostgreSQL, SQL Server use B+ trees

### Day 2 Topics
- Scaling: Horizontal vs vertical, load testing, unit economics
- ProxySQL: Database proxy for connection pooling and query routing
- Delegation (Async Worker): Offload heavy tasks to background workers
- Communications: Message queues, event-driven architecture
- Kafka prototype: Event streaming platform

## Week 2: Database Deep Dive & Real-time Systems

### Database ACID Properties
- Atomicity: All or nothing - transactions as single logical unit
- Consistency: Valid state to valid state - maintains constraints
- Durability: Permanent changes - survives system failures
- Isolation: Concurrent transactions execute independently
- Write-Ahead Logging (WAL): Primary mechanism for durability
- Transaction management: BEGIN, COMMIT, ROLLBACK

### Database Locking
- Lock types: Shared locks, exclusive locks
- Lock granularity: Row-level, table-level
- Deadlock detection and prevention
- Lock escalation: Row to table locks
- Isolation levels: Read uncommitted, read committed, repeatable read, serializable

### Storage-Compute Separation
- Decoupling storage and compute layers
- Benefits: Independent scaling, cost optimization
- Use cases: Data warehouses, analytics platforms
- Technologies: Snowflake, BigQuery architecture

### Read Replica Lag
- Problem: Leader accepts writes → async propagation to replicas
- Replication lag: Replicas seconds/minutes behind leader
- Stale reads: User writes to leader, reads from replica
- GTID (Global Transaction ID): Unique identifier for transactions
- Measuring lag: GTID drift, Seconds_Behind_Master
- Solutions: Read from leader for critical reads, eventual consistency acceptance
- Monitoring: GTID position, replication status checks

### Non-Relational Databases
- Row-oriented DB (OLTP): Stored row by row
- Column-oriented DB (OLAP): Stored column by column
- Graph databases: Nodes, edges, properties
- Use cases: Social networks, recommendation engines
- Index-free adjacency: O(1) relationship traversal
- Wide column stores: Cassandra, HBase
- Document stores: MongoDB, CouchDB

### Real-time Messaging (Slack)
- WebSocket connections: Persistent connections for real-time
- Scaling WebSockets: Pub-sub backends, sticky sessions
- Message delivery: Guaranteed delivery, ordering
- Presence indicators: Online/offline status
- Architecture: Load balancer → WebSocket servers → Message queue
- Technologies: Socket.io, SignalR, custom WebSocket implementations

## Week 3: Distributed Systems & Coordination

### Load Balancers
- L4 (Network Layer): TCP/UDP, balances by IP and port
- L7 (Application Layer): HTTP, balances by URL path, headers, cookies
- Features: SSL termination, health checks, session persistence
- Algorithms: Round robin, weighted round robin, least connections
- IP hash: Same client to same server (caching benefits)
- Least response time: Route to fastest server
- Health checks: Active probes (TCP/HTTP), passive monitoring
- Scaling LBs: Network throughput, connection count, CPU usage
- Auto-scaling: 3-7 minutes response time, pre-warming for spikes

### Distributed Locks
- Mutual exclusion across multiple processes
- Use cases: Scheduled jobs, resource access control
- Implementations: Redis, Zookeeper, database-based
- Lock expiration: TTL to prevent deadlocks
- Lock acquisition: Retry with exponential backoff
- Distributed lock challenges: Network partitions, clock skew

### Zookeeper (Distributed Coordination)
- Centralized coordination service
- Problems solved: Leader election, distributed locking, service discovery
- Configuration management: Centralized, consistent config
- Group membership: Track alive nodes
- Consensus: Agreement on shared state
- Quorum: Minimum nodes needed for decisions
- ZNodes: Hierarchical namespace (like file system)
- Watches: Event notifications on data changes
- Use cases: Kafka, HBase, Hadoop coordination

### ID Generation
- Foundation of ID generation: Requirements and trade-offs
- Twitter Snowflake: 64-bit unique, monotonic, time-sortable IDs
- Architecture: Runs in application servers (not central service)
- ID structure: 1 bit sign, 41 bits timestamp, 10 bits machine ID, 12 bits sequence
- Benefits: No coordination needed, distributed generation
- Alternatives: UUID, database sequences, centralized ID service
- Amazon centralized ID generation: Single service approach

### Pagination
- Offset-based: LIMIT/OFFSET (performance issues at scale)
- Cursor-based: Use last seen ID/timestamp (better performance)
- Keyset pagination: More efficient for large datasets
- Trade-offs: Consistency vs performance

### Hot Shard Problem
- Uneven data distribution across shards
- Causes: Popular users, geographic concentration
- Solutions: Shard splitting, rebalancing, consistent hashing improvements
- Monitoring: Track shard sizes and access patterns

## Week 4: Content Delivery & Media Services

### CDN (Content Delivery Network)
- Purpose: Distribute content closer to users
- Edge servers: Cache content at geographic locations
- Benefits: Reduced latency, bandwidth savings
- Cache strategies: Cache-Control headers, ETags
- Providers: Cloudflare, Akamai, AWS CloudFront
- Use cases: Static assets, images, videos

### Image Upload Service
- Architecture: Pre-signed URLs for direct S3 upload
- Flow: Client → API (get pre-signed URL) → S3 (direct upload)
- Benefits: Offloads API server, reduces bandwidth
- Security: Time-limited URLs, signature validation
- Database schema: users, posts with img_path
- File validation: Type, size constraints
- Technologies: AWS S3, pre-signed URL generation

### Gravatar Service
- Global avatar service
- Architecture: Email hash → avatar URL
- Caching: CDN for avatar images
- Fallbacks: Default avatars for missing emails
- Integration: Simple URL-based API

### Hashtag Service (Instagram)
- Design trade-offs: On-the-fly vs precomputation
- Precomputation: Better UX, slight staleness acceptable
- API design: Single call returns tag info and top posts
- Storage: SQL/NoSQL for frequent updates
- Read path: CDN caching, lazy loading in UI
- Write path: Event-driven (Kafka), consumer groups
- Batching: Process 100 posts, update counters
- Pagination: Cursor-based for top posts

### Unread Indicator
- Real-time unread count updates
- Architecture: WebSocket for live updates
- Scaling: Pub-sub pattern, message queues
- Caching: User-specific unread counts
- Optimization: Batch updates, debouncing

## Week 5: Caching & Storage Systems

### Designing Single Node Cache
- Requirements: Fast key-value storage
- Communication: HTTP (verbose) vs custom protocol (Redis protocol)
- Storage: Hashmaps - string to abstract datatype
- Redis object (robj): Pointer to value, data type
- Single-threaded vs multithreaded: Event loop model
- Performance: 300K ops/second with single core
- Memory management: Global variable for memory usage
- Wrapper methods: malloc()/free() wrappers track memory
- Eviction: When TMU + n > available memory
- TTL: Time-to-live for automatic expiration
- Eviction policies: LRU, LFU, random, TTL-based

### Designing Distributed Cache
- Multi-node cache architecture
- Consistency: Strong vs eventual consistency
- Replication: Master-slave, master-master
- Sharding: Hash-based key distribution
- Cache coherence: Invalidation across nodes
- Network protocols: Custom binary protocol
- Load balancing: Consistent hashing for shard distribution
- Failure handling: Replication, failover mechanisms

### Word Dictionary (No Database)
- In-memory data structures: Trie, hashmap
- Trie: Prefix-based search, space-efficient
- Use cases: Autocomplete, spell check
- Compression: Trie compression techniques
- Memory optimization: Shared prefixes

### Superfast KV Store
- High-performance key-value storage
- Design goals: Low latency, high throughput
- Data structures: Optimized hash tables
- Memory layout: Cache-friendly data structures
- Persistence: Append-only log, snapshots

## Week 6: Advanced Storage & Data Systems

### LSM Tree (Log-Structured Merge Tree)
- Write-optimized data structure
- Components: Memtable, SSTables (Sorted String Tables)
- Write path: Write to memtable → flush to SSTable
- Read path: Check memtable → search SSTables
- Compaction: Merge SSTables to reduce read amplification
- Use cases: Cassandra, RocksDB, LevelDB
- Benefits: High write throughput, sequential writes
- Trade-offs: Read amplification, compaction overhead

### Multi-Tier Datastore
- Hot vs cold storage separation
- Access patterns: Recent data (hot), old data (cold)
- Orders example: 0-6 months (hot), 6+ months (cold)
- Data movement: ETL pipeline from hot to cold
- Dumper: Scheduled batch jobs, extract from hot storage
- Staging storage: Intermediate buffer for transformation
- Loader: Transform, compress, partition data
- Benefits: Cost optimization, performance improvement
- Technologies: S3 for cold storage, Hive for analytics

### Data Ingestion
- ETL/ELT pipelines: Extract, Transform, Load
- Batch processing: Scheduled jobs for large datasets
- Streaming: Real-time data ingestion (Kafka, Kinesis)
- Data validation: Schema validation, data quality checks
- Error handling: Dead letter queues, retry mechanisms

### Live Streaming
- Video streaming architecture
- Chunking: FFmpeg for video segmentation
- HLS (HTTP Live Streaming): m3u8 playlists, TS segments
- CDN distribution: Edge servers for video delivery
- Adaptive bitrate: Multiple quality levels
- Technologies: FFmpeg, HLS, DASH

### Designing S3 (Object Storage)
- Prerequisites: Google File System (GFS) concepts
- Components: Frontend service, GFS coordinator, storage chunks
- Architecture: Distributed object storage
- Consistency: Eventual consistency model
- Durability: Replication across multiple servers
- Scalability: Horizontal scaling of storage nodes
- API: RESTful API for object operations (PUT, GET, DELETE)

