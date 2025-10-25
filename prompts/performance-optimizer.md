# Performance Optimization Notes

## Purpose
Quick performance analysis notes with actionable optimization suggestions.

## Prompt

You are creating performance optimization notes. Please analyze the provided code/system and create focused notes on optimization opportunities.

**Performance Analysis Areas:**

1. **Algorithm Complexity**
   - Identify time complexity (Big O notation)
   - Identify space complexity
   - Suggest more efficient algorithms if applicable
   - Analyze nested loops and recursive calls

2. **Data Structure Optimization**
   - Are the right data structures being used?
   - Hash tables vs arrays vs trees
   - Consider cache locality
   - Memory overhead analysis

3. **Database Performance**
   - Query optimization
   - Index recommendations
   - N+1 query problems
   - Bulk operations vs individual queries
   - Connection pooling
   - Query result caching

4. **Network Optimization**
   - Reduce number of API calls
   - Batch requests where possible
   - Payload size optimization
   - Enable compression
   - Use CDN for static assets
   - HTTP/2 or HTTP/3 benefits

5. **Caching Strategy**
   - What should be cached?
   - Cache invalidation strategy
   - Redis, Memcached, or in-memory cache
   - CDN caching
   - Browser caching headers

6. **Concurrency & Parallelism**
   - Opportunities for parallel processing
   - Async/await optimization
   - Thread pool sizing
   - Lock contention issues
   - Deadlock prevention

7. **I/O Optimization**
   - File system operations
   - Buffered vs unbuffered I/O
   - Async I/O opportunities
   - Batch operations
   - Streaming large data

8. **Memory Management**
   - Memory leaks identification
   - Object pooling opportunities
   - Reduce allocations/garbage collection
   - String concatenation optimization
   - Large object handling

9. **Code-Level Optimizations**
   - Remove unnecessary computations
   - Lazy evaluation opportunities
   - Memoization/dynamic programming
   - Loop optimization
   - Short-circuit evaluation
   - Compile-time vs runtime computation

10. **Frontend Performance** (if applicable)
    - Bundle size optimization
    - Code splitting
    - Lazy loading components
    - Image optimization
    - Minimize reflows/repaints
    - Virtual scrolling for large lists

**Provide:**
- Specific bottlenecks with evidence
- Before/after code examples
- Expected performance improvement
- Trade-offs (code complexity vs performance gain)
- Benchmark recommendations
- Profiling guidance

**Performance Metrics to Consider:**
- Latency (p50, p95, p99)
- Throughput (requests per second)
- Resource utilization (CPU, memory, I/O)
- Response time
- Time to first byte (TTFB)
- Database query time

**Optimization Priority:**
1. Profile first - measure before optimizing
2. Focus on bottlenecks (80/20 rule)
3. Consider diminishing returns
4. Balance performance with maintainability
5. Verify improvements with benchmarks

**Provide notes in this format:**

1. **Current Performance** (if known)
   - Key metrics (latency, throughput, etc.)
   - Identified bottlenecks

2. **High-Impact Optimizations** (do these first)
   - Top 3-4 changes with biggest impact
   - Expected improvement for each
   - Brief code example

3. **Quick Wins** (easy improvements)
   - 2-3 simple changes
   - Low effort, decent payoff

4. **Future Optimizations** (if needed)
   - More complex improvements
   - When to consider them

5. **Measurement Plan**
   - What to measure
   - How to benchmark
   - Success criteria

**Notes Style:**
- Prioritize by impact vs effort
- Show before/after code snippets
- Include estimated improvements
- Focus on practical, proven techniques
- Skip micro-optimizations

Keep notes focused on optimizations that matter - aim for 80/20 rule.

