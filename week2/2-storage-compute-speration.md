## Storage Compute Separation

### Requirements

Build: 
1. Distributed HTTP KV store on top of relational dbs

Requirements:
- Heavily scalable
- GET, PUT, DELETE, TTL - all ops are sync (per key)

Flow: `User --- HTTP ---> API ----> SQL`

### Brainstorm
- Storage
- Optimize storage
- Insert
- Update
- TTL
- Partitioning schemes
- Tradeoffs

2. Build dynamodb on relation DB
 - local and global secondary index implementation
---

## Schema

| Key | Value | Expired_at |
|-----|-------|------------|
| string | string | epoch + duration |


## PUT(K, V)

- INSERT and UPDATE

### Two strategies

#### Transaction (GET, if value exists, update else insert)
```sql
tx
    v = get(k)
    if v
        update(k,v)
    else
        insert(k, v)
```
- 2 DB round trips for TX
- at max 2 DB to get and update/insert

#### Upsert
```sql
upsert(k,v)
```
- 1 DB round trip

---

## DEL(K)

```sql
update store set expired_at = 0 
    and expired_at > now()
```

`expired_at > now()` micro optimization to avoid disk write. If the row is already expired, why is there a need do a disk I/O to update to 0. (Writes are buffered and batched to disk io)

### Clean up cron job
```sql
delete from store
    where expired_at < now() limit 1000
```

**Clean up or cron queries should always be limited.** We don't want cron jobs affect the DB performance.

---

## GET(k)

```sql
select * from store
    where key=k
    and expired_at > now();
```

---

## Scaling of KV Store

- When one node is not enough
    - for reads - add read replicas
    - for writes - add sharding

When sharding, it's really important to figure out access patterns.
When sharding is not done well, will lead to cross shard joins.
It's important to know access patterns to colocate tables.

### Sharding Strategy

- hash based (ownership no control, no metadata)
- range based (ownership is partial, little metadata)
- static based (ownership is complete, lot of metadata)


Incase of Read replicas, when there are lag in writes from master to replicas
How can our solution provide consistent reads.

```
    select value from store where k=k1 with consistency=reads
```

(fill after watching vod)