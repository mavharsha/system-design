# System Design Questions — 2026-04-01

## Q1: Online/Offline Indicator Design (Week 1)
You're designing a user presence system (like WhatsApp "last seen"). The client pushes a heartbeat every 20 seconds. Compare using Redis vs DynamoDB for storing user status. How would you handle the case where a user's device loses connectivity without sending a "going offline" event?

## Q2: Cache Stampede Prevention (Week 1)
During a product launch, your cache TTL expires on a hot product page and 50,000 concurrent requests hit the database simultaneously. Explain what a cache stampede is, how a distributed lock (Redis `SET NX EX`) prevents it, and why you should invalidate cache on writes rather than update it.

## Q3: Database Sharding Strategy (Week 1)
Your social media platform's write throughput has outgrown a single database. Compare hash-based, range-based, and static-based sharding strategies. For a user-generated content platform, which shard key would you pick and why? What happens when you need cross-shard queries (e.g., a global trending feed)?

## Q4: Pessimistic vs Optimistic Locking (Week 2)
You're building a ticket booking system (like BookMyShow) where 10,000 users try to book the last 50 seats simultaneously. Compare `SELECT ... FOR UPDATE`, `FOR UPDATE NOWAIT`, and `FOR UPDATE SKIP LOCKED`. Which locking strategy would you use here and why?

## Q5: Cassandra Data Modeling (Week 2)
Your team is building a messaging system and chose Cassandra for message storage. A colleague suggests creating a `messages` table with a secondary index on `sender_id` for lookups. Explain why this is problematic in Cassandra and what the correct approach is. How do Cassandra's consistency levels (`ONE`, `QUORUM`, `ALL`) affect this design?

## Q6: Read Replica Lag and Consistency (Week 2)
A user updates their profile photo and immediately refreshes the page, but sees the old photo. Explain why this happens with read replicas and describe two patterns to ensure read-your-writes consistency without routing all reads to the master.

## Q7: Twitter Snowflake vs Amazon Centralized ID Generation (Week 3)
Your distributed system needs to generate 100,000 unique, time-sortable IDs per second across 50 machines. Compare Twitter Snowflake's decentralized approach with Amazon's centralized batch allocation. What happens in each system when a machine's clock drifts backward?

## Q8: Cursor-Based Pagination with Monotonic IDs (Week 3)
Your API currently uses `OFFSET/LIMIT` pagination for a feed, and users report that items are duplicated or skipped as they scroll. Explain why offset-based pagination breaks under concurrent writes, and how cursor-based pagination with monotonic IDs (`WHERE id < cursor ORDER BY id DESC LIMIT n`) solves this. What index do you need?

## Q9: Image Upload Service with Pre-Signed URLs (Week 4)
Design an image upload flow where clients upload directly to S3 without the image passing through your server. Explain what pre-signed URLs are, why they're better than proxying through your backend, and how you'd handle generating thumbnails after upload.

## Q10: LSM Trees vs B+ Trees for Storage Engines (Week 6)
Your team is debating whether to use PostgreSQL (B+ tree) or Cassandra (LSM tree) for a high-write-throughput logging service. Explain the fundamental difference in how each stores data, why LSM trees are faster for writes, and when B+ trees still win.
