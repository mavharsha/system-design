## Schema for medium.com

Schema

users table
    id
    name
    bio


blogs table
    id
    author_id
    title
    is_deleted
    published_at
    body

Importance of is_deleted [soft delete]
- **Compliance**: Legal requirements (GDPR, CCPA) may require keeping records for audit trails
- **Auditability**: Track what was deleted, when, and by whom for security investigations
- **Data Recovery**: Easy to restore accidentally deleted content
- **Analytics/Reporting**: Historical data remains intact for business intelligence
- **Referential Integrity**: Prevents breaking foreign key relationships in other tables
- **User Experience**: Allows "undo" functionality or showing "This post was deleted" messages
- **Gradual Deletion**: Can implement multi-stage deletion (deleted → archived → purged)

**Index Issues with Soft Deletes:**
- **Index Bloat**: Soft-deleted rows remain in all indexes, causing them to grow indefinitely
- **Query Performance**: Every query must filter `WHERE is_deleted = false`, adding overhead
- **Index Maintenance**: Updates to `is_deleted` require updating all indexes that include that column
- **Write Amplification**: Changing `is_deleted` from false → true updates the row + all associated indexes
- **Cache Pollution**: Deleted rows still occupy space in index caches, reducing hit rates for active data

**Optimization Strategies:**
- **Partial/Filtered Indexes**: Create indexes only on non-deleted rows
  ```sql
  CREATE INDEX idx_active_blogs ON blogs(author_id) WHERE is_deleted = false;
  ```
- **Composite Indexes**: Include `is_deleted` as first column for better filtering
  ```sql
  CREATE INDEX idx_blogs_lookup ON blogs(is_deleted, author_id, published_at);
  ```
- **Periodic Archiving**: Move old soft-deleted data to archive tables to reduce index size
- **Separate Tables**: Consider separate `active_blogs` and `deleted_blogs` tables for large datasets


And incase of actual data deletion (hard delete), there are potential issues:

**DB Rebalancing Issues:**
- **Index Fragmentation**: When rows are deleted, B-tree indexes can become fragmented, leading to inefficient queries
- **Storage Reclamation**: Deleted data may not immediately free up disk space; DB needs to run VACUUM/OPTIMIZE operations
- **Partition Rebalancing**: In sharded/partitioned tables, deletions can cause uneven data distribution
- **Performance Impact**: Large batch deletions can lock tables and impact read/write operations
- **Replication Lag**: Hard deletes need to propagate to replicas, causing temporary inconsistency

**Best Practices:**
- Schedule hard deletes during low-traffic periods (where even possible batching)
- **Use batch deletions with limits (e.g., DELETE LIMIT 1000) to avoid long-running transactions**
- Run VACUUM/ANALYZE after bulk deletions to reclaim space and update statistics
- Consider archiving old soft-deleted data to separate "cold storage" tables/databases

------
Things I learnt
- Need for soft deletes (legal, auditablitiy)
- How hard deletes effect DB rebalances
- How hard deletes fragments indexes

