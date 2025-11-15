## CPU Bound vs I/O Bound

**CPU Bound:**
- Performance limited by CPU speed/processing power
- CPU is the bottleneck - always busy computing
- Examples: video encoding, image processing, mathematical calculations, encryption, data compression, machine learning training
- Characteristics: High CPU usage (near 100%), low wait time

**I/O Bound:**
- Performance limited by input/output operations (disk, network, database)
- CPU spends most time waiting for I/O to complete
- Examples: file operations, database queries, network requests, web scraping, API calls
- Characteristics: Low CPU usage, high wait time, processes are often blocked waiting

## Quick Comparison

| Aspect | CPU Bound | I/O Bound |
|--------|-----------|-----------|
| Bottleneck | Processor speed | Disk/Network speed |
| CPU Usage | High (90-100%) | Low (often <20%) |
| Benefits from | More/faster cores | Async I/O, caching, faster storage |
| Concurrency helps? | Only with multiple cores | Yes, greatly! |
| Example improvement | Upgrade CPU | Use SSD, add cache, async operations |

## Optimization Strategies

**For CPU Bound:**
- Use more cores/threads (parallel processing)
- Optimize algorithms
- Use faster CPU
- Cache computed results

**For I/O Bound:**
- Use asynchronous I/O
- Implement caching
- Use connection pooling
- Batch operations
- Upgrade to faster storage (HDD → SSD)

**Real world:** Most applications are **I/O bound** - they spend more time waiting for databases, APIs, or file systems than doing actual computation.


### How Different Databases Can Be CPU or I/O Bound

#### Redis
- **Usually I/O Bound**: For most workloads, Redis (an in-memory cache/database) is so fast at computation that network or disk persistence (for backups) are the bottlenecks.
- **Can Become CPU Bound**: If you're doing heavy operations (e.g., running large Lua scripts, computing set intersections on massive data), Redis can become CPU bound, especially since it uses a single-threaded event loop for commands.
- **Optimization**: Run multiple Redis instances (one per CPU core), optimize data structures, or offload computations to clients.

#### Postgres (or MySQL)
- **Often I/O Bound**: Traditional relational databases like Postgres and MySQL spend much of their time waiting on disk or network I/O – for example, reading/writing files, waiting for client queries, etc.
- **Can Be CPU Bound**: On analytic or complex queries (e.g., big joins, aggregates, stored procedures, large sorts), or when CPU-intensive functions (encryption, compression) are used.
- **Optimization**: Use faster SSDs, more RAM for caching, proper indexes, run queries asynchronously, or parallelize execution.

#### Summary Table

| DB        | CPU Bound Example                         | I/O Bound Example                            |
|-----------|------------------------------------------|----------------------------------------------|
| Redis     | Heavy script, big set ops                | Network latency, RDB/AOF background saves    |
| Postgres  | Big join/aggregate, JSON functions       | Reading from slow disk, writing query result |
| MySQL     | Encryption, complex stored procedures    | Large result set fetch, slow disk            |

**Takeaway:**  
The same database may be CPU bound in some situations and I/O bound in others — always profile your workload.
