## Online offline indicator

Requirement: is a user online or offline
Potential access patterns: Key value

The system design can be broken into
- get online or offline
- write if the user is online or offline


For GET online/offline
- would it make sense to fetch only one user?
- would it make sense to batch? like a group of users

Prefer batching when its possible

GET /status/user?ids=u1,u2,u3

For Update/Insert online/offline
- Brings to pull vs push (opposites framework)
- In this scenario, there isn't a way to pull. Server cannot reaches out to all the clients until there is a persistant connection

- So, push based is the only available (client pushes to the server that they are online)
- Client calls heartbeat at regular basis (maybe every 20 seconds)

POST /heartbeat


DB schema

userId | epoch
u1     | e1
u2     | e2
u3     | e3
u4     | e4

4bytes + 4bytes = 8 bytes per record

Assuming 1 billion users: 1B * 8 bytes = 8GB of data (fits in memory) 


Dense vs Sparse
- Dense: Store every user's status (including offline users). Takes more space but simpler queries.
- Sparse: Only store online users. When user goes offline, delete the record. Saves space, especially if most users are offline.

For this use case, **Sparse is better** because:
- Most users will be offline at any given time
- We only need to store active/online users
- Absence of record = offline status
- Significantly reduces memory footprint



Questions to decide DB

Self managed redis vs Managed Dynamodb?

**Redis (In-Memory Key-Value Store)**
Pros:
- Extremely fast reads/writes (< 1ms latency)
- Perfect for this use case (simple key-value with TTL)
- Can set TTL (Time To Live) on keys - auto-delete after heartbeat timeout
- Fits entirely in memory (8GB for 1B users, but sparse would be much less)

Cons:
- Need to manage infrastructure
- Need to handle persistence/replication
- Need to handle scaling

**Managed DynamoDB**
Pros:
- Fully managed (no ops overhead)
- Auto-scaling
- Built-in replication and availability

Cons:
- Higher latency (single digit ms)
- Single kv reads might be faster, but dynamodb doesn't gurantees similar SLA's for batch reads ()
- More expensive for read-heavy workloads
- No native TTL support (requires manual cleanup or DynamoDB TTL which has lag)

**Recommendation: Redis** with TTL for automatic expiry. Set TTL to 30-40 seconds (slightly longer than heartbeat interval).


In realworld, we would build it with web sockets.
Which is persistent connection between client and server.
- Server knows immediately when connection drops (user offline)
- No need for polling/heartbeat
- More efficient for real-time updates
- Bidirectional communication allows server to push status updates to clients


Writing to DB can also be optimized.
DB's should create a new connection for every read or write.

For every connection, the API and DB does a 3 way handshake (which becomes an over head)
Use Connection pools. Have a pre-stablished list of connections.

```go
// DBConnectionPool with blocking queue (buffered channel)
type DBConnectionPool struct {
	connections chan *sql.DB // Buffered channel acts as blocking queue
	poolSize    int
}

// Initialize pool with 10 connections
func NewDBConnectionPool(dsn string, poolSize int) (*DBConnectionPool, error) {
	pool := &DBConnectionPool{
		connections: make(chan *sql.DB, poolSize), // Buffered channel with size 10
	}
	
	// Create 10 connections and add to pool
	for i := 0; i < poolSize; i++ {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, err
		}
		pool.connections <- db // Add to blocking queue
	}
	
	return pool, nil
}

// GetConnection - BLOCKS if all 10 connections are in use
func (p *DBConnectionPool) GetConnection() *sql.DB {
	return <-p.connections // Blocks until a connection is available
}

// PutConnection - Returns connection back to pool
func (p *DBConnectionPool) PutConnection(conn *sql.DB) {
	p.connections <- conn // Put back in the queue
}
```

**Key Benefits:**
- Avoids overhead of creating new connections for every request
- Eliminates 3-way handshake overhead for each DB operation
- Buffered channel provides natural blocking when all connections are busy
- Reuses existing connections efficiently


---- 
Topics I learnt
- Batching requests
- Deciding schema based on the requirements (sparse vs dense)
	- Do I just need the online offline? If yes, then TTL (with this query can just be if I see a record, that means the user is online)
	- Do I also need to support last seen? 
- Redis vs Dynamodb
	- Key value
	- Dynamodb single reads are faster compared to batch reads (distributed)
- Connection pooling
	- blocking queue
	- min and max connections
	- idle time for connections
	- stale?
- Better implementation of the system would be to use websockets
	- **Need to explore how to do it**