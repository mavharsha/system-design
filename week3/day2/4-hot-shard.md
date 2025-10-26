## Hot Shards

## Quick Summary

**What it does:** Hot sharding is a technique to handle uneven data distribution and traffic patterns by logically partitioning data into many small shards, allowing hot (high-traffic) shards to be moved to separate physical servers.

**Primary use case:** Handling celebrity/influencer accounts, viral content, or any scenario where a small subset of data receives disproportionately high traffic.

**Key features:**
- Logical sharding at scale (thousands of logical shards)
- Dynamic shard migration based on telemetry
- Prevents single-point bottlenecks in distributed systems

## Core Concepts

### The Hot Shard Problem

**Definition:** When data is unevenly distributed across shards, certain shards receive significantly more traffic than others, creating performance bottlenecks.

**Common causes:**
- **Celebrity users** (Instagram/Twitter influencers with millions of followers)
- **Viral content** (trending posts, videos, hashtags)
- **Geographic concentration** (major events in specific locations)
- **Temporal spikes** (Black Friday sales, live events)

### Logical vs Physical Sharding

**Logical Shards:**
- Many small logical partitions (e.g., 8000 logical shards)
- Multiple logical shards can exist on same physical server
- Easy to move individual shards

**Physical Shards:**
- Actual database servers/instances
- Limited by hardware availability
- Expensive to add/remove

### Instagram's Approach

Instagram created **8000 logical shards** for their POST tables:

```sql
CREATE DATABASE insta.post.0001;
CREATE DATABASE insta.post.0002;
CREATE DATABASE insta.post.0003;
...
CREATE DATABASE insta.post.8000;
```

**Shard assignment formula:**
```
shard_id = hash(user_id) % 8000
```

**Benefits:**
- Easy to monitor individual shard performance
- Simple to migrate hot shards to dedicated servers
- Minimal data movement during rebalancing

## Usage Examples

### Example 1: Shard Selection Logic (Java)

```java
public class ShardManager {
    private static final int TOTAL_LOGICAL_SHARDS = 8000;
    private Map<Integer, DataSource> shardToDataSource;
    
    public DataSource getShardForUser(long userId) {
        int logicalShardId = calculateShardId(userId);
        return shardToDataSource.get(logicalShardId);
    }
    
    private int calculateShardId(long userId) {
        // Consistent hashing to determine logical shard
        return Math.abs((int) (userId % TOTAL_LOGICAL_SHARDS));
    }
    
    public void insertPost(long userId, Post post) {
        DataSource ds = getShardForUser(userId);
        try (Connection conn = ds.getConnection()) {
            String sql = "INSERT INTO posts (id, user_id, content, created_at) VALUES (?, ?, ?, ?)";
            PreparedStatement stmt = conn.prepareStatement(sql);
            stmt.setLong(1, post.getId());
            stmt.setLong(2, userId);
            stmt.setString(3, post.getContent());
            stmt.setTimestamp(4, post.getCreatedAt());
            stmt.executeUpdate();
        }
    }
}
```

### Example 2: Hot Shard Detection

```java
public class HotShardDetector {
    private MetricsCollector metricsCollector;
    private static final double HOT_THRESHOLD = 0.05; // 5% of total traffic
    
    public List<Integer> detectHotShards() {
        Map<Integer, Long> shardRequestCounts = metricsCollector.getShardRequestCounts();
        long totalRequests = shardRequestCounts.values().stream()
            .mapToLong(Long::longValue)
            .sum();
        
        List<Integer> hotShards = new ArrayList<>();
        
        for (Map.Entry<Integer, Long> entry : shardRequestCounts.entrySet()) {
            double shardPercentage = (double) entry.getValue() / totalRequests;
            
            if (shardPercentage > HOT_THRESHOLD) {
                hotShards.add(entry.getKey());
                System.out.printf("Hot shard detected: %d (%.2f%% of traffic)%n", 
                    entry.getKey(), shardPercentage * 100);
            }
        }
        
        return hotShards;
    }
}
```

### Example 3: Shard Migration Strategy

```java
public class ShardMigration {
    private ShardManager shardManager;
    
    public void migrateHotShard(int logicalShardId, DataSource newPhysicalServer) {
        System.out.println("Starting migration for shard: " + logicalShardId);
        
        // Step 1: Enable read-only mode on source
        setShardReadOnly(logicalShardId, true);
        
        // Step 2: Export data from hot shard
        exportShardData(logicalShardId);
        
        // Step 3: Import to new physical server
        importShardData(logicalShardId, newPhysicalServer);
        
        // Step 4: Update routing table
        shardManager.updateShardMapping(logicalShardId, newPhysicalServer);
        
        // Step 5: Re-enable writes on new location
        setShardReadOnly(logicalShardId, false);
        
        System.out.println("Migration completed for shard: " + logicalShardId);
    }
    
    private void exportShardData(int shardId) {
        // Export only one logical shard instead of entire database
        String command = String.format(
            "pg_dump -t posts -t comments insta.post.%04d > shard_%04d.sql",
            shardId, shardId
        );
        // Execute export
    }
}
```

### Example 4: Monotonic ID Generation per Shard

```java
public class ShardSequenceGenerator {
    
    // Each shard has its own sequence generator
    public long generateId(Connection conn, int shardId) throws SQLException {
        // Use database stored procedure for monotonic IDs
        String sql = "{CALL generate_post_id()}";
        
        try (CallableStatement stmt = conn.prepareCall(sql)) {
            ResultSet rs = stmt.executeQuery();
            if (rs.next()) {
                return rs.getLong(1);
            }
            throw new SQLException("Failed to generate ID");
        }
    }
}
```

**Stored Procedure Example (PostgreSQL):**
```sql
CREATE SEQUENCE post_id_seq;

CREATE OR REPLACE FUNCTION generate_post_id()
RETURNS BIGINT AS $$
DECLARE
    new_id BIGINT;
BEGIN
    new_id := nextval('post_id_seq');
    RETURN new_id;
END;
$$ LANGUAGE plpgsql;
```

## Key Points to Remember

### Detection Strategies

- **Monitor continuously**: Track requests per shard using observability tools (Prometheus, Datadog)
- **Set thresholds**: Alert when shard receives >5% of total traffic
- **Track metrics**: CPU usage, query latency, connection pool saturation, disk I/O

### Migration Considerations

- **Minimize downtime**: Use read-only mode during migration, not full shutdown
- **Small data sets**: Each logical shard is small, making migration fast (minutes vs hours)
- **Rollback plan**: Keep source data until migration is verified

### Common Mistakes

❌ **Creating too few logical shards**: Limits flexibility for future splits
❌ **Ignoring celebrity accounts**: Plan for power-law distribution from day one
❌ **Equal shards = equal load**: Traffic distribution is rarely uniform
❌ **Manual sharding**: Always automate shard selection and routing
❌ **Ignoring read replicas**: Hot reads can be distributed even without migration

### Performance Considerations

- **Shard routing overhead**: Hash calculation is negligible (nanoseconds)
- **Connection pooling**: Maintain separate pools per physical server
- **Cross-shard queries**: Avoid joins across shards; denormalize when necessary
- **Rebalancing frequency**: Don't migrate too often (wait for sustained hotness)

### Prevention Strategies

1. **Read replicas for hot shards**: Scale reads horizontally
2. **Caching layer**: Redis/Memcached for celebrity profile data
3. **CDN for static content**: Images, videos of viral posts
4. **Rate limiting**: Protect against artificial hotness (DDoS)
5. **Dedicated infrastructure**: Reserve capacity for known celebrities

## Quick Reference

### Shard Detection Commands

```bash
# Monitor shard request distribution
SELECT shard_id, COUNT(*) as requests 
FROM access_logs 
WHERE timestamp > NOW() - INTERVAL '5 minutes'
GROUP BY shard_id 
ORDER BY requests DESC 
LIMIT 10;

# Check shard size
SELECT 
    schemaname, 
    tablename, 
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) 
FROM pg_tables 
WHERE schemaname LIKE 'insta.post.%';
```

### Migration Steps

1. **Identify hot shard**: Monitoring + alerting
2. **Provision new server**: Ensure adequate resources
3. **Enable read-only**: Pause writes to source
4. **Export data**: `pg_dump` for single shard
5. **Import to target**: `pg_restore` on new server
6. **Update routing**: Modify shard mapping table
7. **Verify**: Test queries on new location
8. **Enable writes**: Resume normal operations
9. **Monitor**: Ensure performance improvement

### Key Metrics to Monitor

| Metric | Threshold | Action |
|--------|-----------|--------|
| Request percentage | >5% per shard | Consider migration |
| Query latency | >100ms p99 | Investigate bottleneck |
| CPU usage | >80% sustained | Add read replicas or migrate |
| Connection pool | >90% utilized | Scale connections or migrate |

## Related Topics

- **Consistent Hashing**: Alternative to modulo-based sharding for easier rebalancing
- **Database Partitioning**: Native PostgreSQL/MySQL partitioning vs logical sharding
- **Read Replicas**: Scaling reads without moving data
- **Caching Strategies**: Reducing load on hot shards
- **CDN Architecture**: Offloading static content
- **Twitter Snowflake**: Distributed ID generation for sharded systems
- **Vitess (YouTube)**: MySQL sharding middleware
- **Cassandra**: Built-in partitioning and hotspot handling