# ProxySQL

---

## 1. Overview

### When to Use ProxySQL for DB Scaling

- **Read scaling with replicas:** Use ProxySQL to automatically route reads to replicas and writes to the master, enabling horizontal read scaling.
- **High connection overhead:** Connection pooling/multiplexing in ProxySQL reduces backend MySQL connections, supporting high concurrent app connections efficiently.
- **Minimal app changes:** ProxySQL enables you to add/remove replicas, reroute traffic, and handle failovers without changing application code.
- **Centralized routing and analytics:** Useful if you need to manage query routing, caching, and monitoring in one place as your database architecture grows.

*In summary: Use ProxySQL when scaling out MySQL/MariaDB with read replicas, high connection loads, or need seamless, centralized query management.*


**What is it?**
- Proxy layer between application and MySQL/MariaDB
- Handles connection pooling and query routing

**Key features:**
- Query routing (read/write split)
    - Route writes to master
    - Route reads to replicas
    - No application code changes needed
- Connection pooling / Multiplexing (in ProxySQL layer)
    - Many app connections → few DB connections
    - Example: 1000 app connections → 10 actual DB connections
    - **Two layers:**
        - App → ProxySQL: Apps can have many connections (app may or may not pool)
        - ProxySQL → DB: ProxySQL maintains connection pool to DB servers
    - How it works:
        - App opens connection to ProxySQL (stays open, idle most of time)
        - When query arrives, ProxySQL borrows DB connection from its pool
        - Executes query, returns result
        - Returns DB connection to pool (ready for next query from any app)
        - Multiple app connections share same DB connection (time-multiplexed)
    - Reduces: TCP handshakes to DB, auth overhead, memory on DB server
    - Config: 
        - `mysql-max_connections`: ProxySQL accepts from apps (default: 2048)
        - `mysql-default_max_connections`: ProxySQL opens to DB per hostgroup (default: 1000)
- Query caching
- Load balancing across replicas
- Query rewriting and filtering

**Benefits:**
- Offload routing logic from application
- Connection pooling improves performance
- Can add/remove DB nodes without app changes
- Monitoring and query analytics built-in

**When to use:**
- When you have read replicas
- High connection overhead
- Need centralized query routing
- Want to avoid managing connection pools in every service

---

## 2. Basic Configuration

### Hostgroups
Logical grouping of servers
- Hostgroup 0 = Writer (Master)
- Hostgroup 1 = Readers (Replicas)
- Can have multiple hostgroups for different clusters

```sql
-- Add servers to hostgroups
INSERT INTO mysql_servers (hostgroup_id, hostname, port)
VALUES 
  (0, 'master.db.com', 3306),      -- writer hostgroup
  (1, 'replica1.db.com', 3306),    -- reader hostgroup
  (1, 'replica2.db.com', 3306),    -- reader hostgroup
  (1, 'replica3.db.com', 3306);    -- reader hostgroup

LOAD MYSQL SERVERS TO RUNTIME;
SAVE MYSQL SERVERS TO DISK;
```

### Query Routing Rules

```sql
-- Route SELECT to replicas (hostgroup 1)
INSERT INTO mysql_query_rules (rule_id, active, match_pattern, destination_hostgroup)
VALUES (1, 1, '^SELECT.*FOR UPDATE', 0);  -- SELECT FOR UPDATE (+ NOWAIT/SKIP LOCKED) to master

INSERT INTO mysql_query_rules (rule_id, active, match_pattern, destination_hostgroup)
VALUES (2, 1, '^SELECT', 1);  -- All other SELECT to replicas

-- Everything else (INSERT/UPDATE/DELETE) to master (hostgroup 0)
INSERT INTO mysql_query_rules (rule_id, active, match_pattern, destination_hostgroup)
VALUES (3, 1, '.*', 0);

LOAD MYSQL QUERY RULES TO RUNTIME;
SAVE MYSQL QUERY RULES TO DISK;
```

### Multiple Clusters

```sql
-- Cluster 1: users database
INSERT INTO mysql_servers VALUES (10, 'users-master', 3306);
INSERT INTO mysql_servers VALUES (11, 'users-replica1', 3306);
INSERT INTO mysql_servers VALUES (11, 'users-replica2', 3306);

-- Cluster 2: orders database
INSERT INTO mysql_servers VALUES (20, 'orders-master', 3306);
INSERT INTO mysql_servers VALUES (21, 'orders-replica1', 3306);
```

**ProxySQL handles:**
- Load balancing between replicas
- Health checks
- Automatic failover
- Connection multiplexing

---

## 3. Use Case: Splitting Monolith Database

**Example: Ecommerce**

**Split strategy:**
```
Monolith (ecommerce_db) → catalog_db + orders_db

catalog_db: users, products, inventory
orders_db: orders, payments, shipping
```

**Setup hostgroups**
```sql
-- Catalog: hostgroup 10=master, 11=replicas
INSERT INTO mysql_servers (hostgroup_id, hostname, port) VALUES
  (10, 'catalog-master', 3306),
  (11, 'catalog-replica1', 3306),
  (11, 'catalog-replica2', 3306);

-- Orders: hostgroup 20=master, 21=replicas
INSERT INTO mysql_servers (hostgroup_id, hostname, port) VALUES
  (20, 'orders-master', 3306),
  (21, 'orders-replica1', 3306);

LOAD MYSQL SERVERS TO RUNTIME;
```

**Route by table name**
```sql
-- Catalog tables
INSERT INTO mysql_query_rules (rule_id, active, match_pattern, destination_hostgroup) VALUES
  (10, 1, '^SELECT.*FROM\\s+(users|products|inventory)\\b', 11),  -- reads
  (11, 1, '.*(users|products|inventory)', 10);                     -- writes

-- Orders tables  
INSERT INTO mysql_query_rules (rule_id, active, match_pattern, destination_hostgroup) VALUES
  (20, 1, '^SELECT.*FROM\\s+(orders|payments)\\b', 21),  -- reads
  (21, 1, '.*(orders|payments)', 20);                     -- writes

LOAD MYSQL QUERY RULES TO RUNTIME;
```

**App code unchanged:**
```go
db, _ := sql.Open("mysql", "user:pass@tcp(proxysql:6033)/")
db.Query("SELECT * FROM products WHERE id = ?", id)  // → catalog_db
db.Query("SELECT * FROM orders WHERE user_id = ?", uid)  // → orders_db
```

**Note:** ProxySQL routes by query pattern (table names), not by database selected at connection time. You don't need to specify a database in the connection string—ProxySQL inspects the SQL query itself.

**Benefits:** Independent scaling, different backup strategies, no app changes

---

## 4. Scaling ProxySQL

### Vertical Scaling
- Add more CPU/RAM to ProxySQL server
- Good first step, simplest approach
- Limited by single server capacity

### Horizontal Scaling (Multiple Instances)

**Architecture:**
```
App Servers → Load Balancer → ProxySQL-1
                            → ProxySQL-2
                            → ProxySQL-3
```

**Config Synchronization**

Instances don't talk to each other, need config sync:

**Option A: Central config management**
```bash
#!/bin/bash
for proxy in proxysql-1 proxysql-2 proxysql-3; do
  mysql -h $proxy -P6032 < config.sql
done
```

**Option B: ProxySQL Cluster (Native)**
```sql
-- On each ProxySQL instance
INSERT INTO proxysql_servers (hostname, port, weight) VALUES
  ('proxysql-1', 6032, 1),
  ('proxysql-2', 6032, 1),
  ('proxysql-3', 6032, 1);

LOAD PROXYSQL SERVERS TO RUNTIME;
-- Config changes sync to other nodes (not instant, eventual consistency)
-- Still need to LOAD ... TO RUNTIME on each node for changes to take effect
```

**Option C: Orchestration (Ansible/Kubernetes)**
```yaml
# Ansible example
- hosts: proxysql_servers
  tasks:
    - name: Deploy ProxySQL config
      template:
        src: proxysql.cnf.j2
        dest: /etc/proxysql.cnf
    - name: Restart ProxySQL
      service: name=proxysql state=restarted
```

### Load Balancer Setup

**Important: Use Layer 4 (TCP), not Layer 7 (HTTP)**

MySQL protocol is TCP-based, not HTTP
- Layer 4 (Transport): TCP/UDP - treats traffic as byte stream
- Layer 7 (Application): HTTP/HTTPS - understands application protocol

**Recommendation: Layer 4 (TCP Load Balancing)**
- Faster (no protocol parsing)
- Lower latency
- Doesn't need to understand MySQL protocol

**Option 1: HAProxy (Recommended)**

```haproxy
# Layer 4 TCP load balancing
frontend proxysql_frontend
    bind *:3306
    mode tcp
    option tcplog
    default_backend proxysql_cluster

backend proxysql_cluster
    mode tcp
    balance roundrobin
    option tcp-check
    tcp-check connect
    tcp-check send-binary 0a  # Validate MySQL protocol handshake
    
    server proxysql-1 proxysql-1:6033 check inter 5s fall 3 rise 2
    server proxysql-2 proxysql-2:6033 check inter 5s fall 3 rise 2
    server proxysql-3 proxysql-3:6033 check inter 5s fall 3 rise 2
```

**Option 2: Nginx (needs stream module)**

```nginx
# Nginx requires stream module (not http module)
stream {
    upstream proxysql_cluster {
        least_conn;  # or: round_robin, hash
        server proxysql-1:6033 max_fails=3 fail_timeout=10s;
        server proxysql-2:6033 max_fails=3 fail_timeout=10s;
        server proxysql-3:6033 max_fails=3 fail_timeout=10s;
    }

    server {
        listen 3306;
        proxy_pass proxysql_cluster;
        proxy_connect_timeout 5s;
    }
}
```

**Option 3: Cloud Load Balancers (Best for production)**
- AWS: Network Load Balancer (NLB) - Layer 4
- GCP: TCP/UDP Load Balancer
- Azure: Standard Load Balancer (TCP mode)

Benefits: Managed, highly available, auto-scaling

**Benefits:**
- High availability (one ProxySQL fails, others serve)
- Horizontal scaling for connection load
- Rolling updates without downtime

**Trade-offs:**
- Config sync complexity
- Extra hop (load balancer layer)
- Need monitoring for all instances

---
**Trade-offs of App-Layer Routing vs ProxySQL/External Load Balancer**

- **Application-aware routing:** Instead of relying on ProxySQL or load balancers, you can make your application aware of which MySQL host is the master and which are replicas. The application would then route reads/writes directly—this reduces one network hop and can provide more control.
    - **Pros:** 
        - Potentially lower latency (no ProxySQL hop)
        - Fine-grained per-query routing logic (at app layer)
        - Custom failover handling
        - No need to sync ProxySQL config
    - **Cons:** 
        - Complexity moves to every application/service
        - Harder to manage when you have many clients/services
        - Every application needs to handle topology and failover logic
        - Adding/removing DB nodes requires application config changes/redeployment

- **When to consider app-layer routing:** 
    - Suitable if you have a small number of applications/services and want maximum control, or if your app requires custom routing logic that ProxySQL cannot express.
    - Less ideal as your environment/app count grows.

- **Typical approach:** 
    - Use a library (e.g., in Java: HikariCP with custom routing, Python: SQLAlchemy with custom engines) that supports connection routing based on query type.
    - Topology information (master/replica IPs, health state) must be distributed—usually by configuration management or a discovery service.

**Summary:**  
App-layer routing can optimize performance for advanced use-cases but increases operational complexity. Most organizations centralize this logic in ProxySQL or similar tools for easier management, especially at scale.

