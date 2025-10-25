## Loadbalancer (paritially generated needs to be reviewed)

```
Theme: High availability
```

L4 and L7
- L4 - TCP/UDP
    - Balances traffic based on IP address and TCP/UDP port.
    - Network level loadbalancer
- L7 - HTTP
    - Balances traffic based on content of the request (e.g., URL path, headers, cookies, etc.)
    - Application level balancer

| Feature         | NLB (L4)      | ALB (L7)                      |
| --------------- | ------------- | ----------------------------- |
| Layer           | 4 (Transport) | 7 (Application)               |
| Protocols       | TCP, UDP, TLS | HTTP, HTTPS, gRPC, WebSocket  |
| Routing         | By IP & port  | By URL path, headers, methods |
| Performance     | Very high     | Slightly lower                |
| SSL termination | Optional      | Yes                           |
| Health checks   | TCP-based     | HTTP-based                    |



#### L7 loadbalancer
- Within one region/vpc
- Can inspect HTTP headers, URL path, cookies
- Use cases:
    - Route /api/* to backend servers, /static/* to CDN origin
    - A/B testing - route based on cookie
    - Canary deployments - send 5% traffic to new version
- More CPU intensive than L4 (needs to parse HTTP)
- Can modify requests/responses
- Session persistence via cookies



#### Scaling LB
- I need to know when I need to add an additional loadbalancer
- LB servers publishes telemetery to a system like prometheus
    - If orchestrational manager sees high networking and CPU utilization
        - It might bootup another loadbalancer
- Metrics to watch:
    - Network throughput (bytes/sec)
    - Active connections count
    - CPU usage on LB instances
    - Request queue depth
    - Error rate spikes
- Scaling LB
    - Can handle gradual surge (15-20% growth over time)
    - But are bad at sudden spikes (double traffic in seconds)
        - There is a need to pre-warm LB's
        - Contact AWS support to pre-warm ALB/NLB before expected traffic
- Auto-scaling LBs:
    - Most cloud LBs auto-scale but takes 3-7 minutes
    - Not instant - can't react to flash crowds
    - Plan capacity ahead for known events (product launches, sales)

#### LB Algorithms
- Round robin - distribute evenly to each server in order
- Weighted round robin - some servers get more (based on capacity)
- Least connections - send to server with fewest active connections
    - Good when request processing time varies
- IP hash - same client IP always goes to same server
    - Helps with caching
    - But bad if client IP changes (mobile users)
- Least response time - route to server with fastest response
- Random - just pick random server (simple, works well)

#### Health Checks
- LB needs to know which backends are alive/healthy
- Active probes:
    - L4: TCP connection check every N seconds
    - L7: HTTP GET to /health endpoint
    - Typical interval: 10-30 seconds
- Mark unhealthy after X consecutive failures (usually 2-3)
- Mark healthy after Y consecutive successes (usually 2)
- Unhealthy servers removed from pool
- Passive health checks:
    - Monitor actual request failures
    - Faster detection but needs real traffic

#### Sticky Sessions (Session Affinity)
- Problem: user state stored on server, next req goes to different server
- Solutions:
    - Cookie-based sticky - LB sets cookie, routes future reqs to same backend
    - IP hash - same IP always goes to same server
    - Session store (Redis) - all servers share session data (better)
- Tradeoffs:
    - Sticky = worse distribution, can't easily remove servers
    - No sticky = need external session storage

#### Connection Draining (Deregistration Delay)
- Taking server offline for deploy/maintenance
- Stop sending NEW connections to that server
- Wait for existing connections to finish (timeout: 30-300 sec)
- Then fully remove from pool
- Prevents users from seeing errors mid-request

#### LB High Availability
- Single LB = single point of failure
- Solutions:
    - Multiple LBs with VIP (virtual IP) using VRRP
    - Active-passive with heartbeat (keepalived)
    - Active-active (both serve traffic)
    - Cross-zone redundancy (different AZs)
- Cloud LBs usually handle this automatically

#### DNS-based Load Balancing
- DNS returns multiple IPs for same domain
- Client picks one (usually first)
- Pros:
    - Simple, no extra infrastructure
    - Geographic distribution
- Cons:
    - DNS caching means can't quickly remove bad servers
    - No health checks
    - No request-level balancing
    - TTL problems
- Use case: coarse-grained region selection

#### Global Server Load Balancing (GSLB)
- Route users to nearest datacenter
- GeoDNS - return different IPs based on client location
- Anycast - same IP announced from multiple locations, BGP routes
- Latency-based routing
- Failover between regions
- Need to handle:
    - Data replication lag
    - Session consistency cross-region
    - Cost of cross-region bandwidth

#### SSL/TLS at LB
- SSL termination at LB:
    - LB decrypts, talks HTTP to backends
    - Pros: less CPU on app servers, easier cert management
    - Cons: unencrypted in VPC (usually ok)
- SSL passthrough:
    - LB doesn't decrypt, just forwards
    - Must use L4 (can't do L7 routing)
- End-to-end encryption:
    - LB decrypts then re-encrypts to backend
    - Most secure but most CPU intensive

#### WebSockets and Long Connections
- WebSocket upgrade happens over HTTP
- Then becomes persistent TCP connection
- L7 LB must support upgrade header
- Can't load balance per-message, only per-connection
- Need connection affinity
- Watch connection limits (max concurrent connections)

#### Rate Limiting and DDoS
- Can do basic rate limiting at LB:
    - Per IP: block single client sending too much
    - Global: total RPS cap
- Better to use dedicated WAF/API gateway for this
- DDoS protection:
    - SYN flood protection (L4)
    - HTTP flood detection (L7)
    - Cloud providers (Cloudflare, AWS Shield) handle better

#### Timeouts
- Connection timeout - how long to wait for connection to backend
- Idle timeout - close inactive connections after X time
- Request timeout - max time for backend to respond
- Important: LB timeout > backend timeout
- If LB timeout < backend, client sees 504 gateway timeout even though backend succeeded

#### Observability and Monitoring
- Key metrics:
    - Request count, RPS
    - Latency (p50, p95, p99)
    - Error rates (4xx, 5xx)
    - Active connections
    - Backend health status
    - Queue depth
- Logs:
    - Access logs for debugging
    - Usually disabled in prod (expensive)
    - Enable sampling if needed
- Distributed tracing:
    - Add trace ID header
    - Track request through LB -> backend

#### Common Mistakes
- Not setting proper timeouts
- Forgetting to enable cross-zone load balancing (uneven distribution)
- Using sticky sessions when not needed (bad scaling)
- No health checks or bad health check endpoints
- Not monitoring queue depth (requests backing up)
- Underestimating LB scaling time
- HTTP keep-alive causing connection reuse imbalance

#### When to Use What
- L4 (NLB):
    - Need extreme performance
    - Non-HTTP protocols
    - Static IP requirement
    - Preserve client IP
- L7 (ALB):
    - HTTP/HTTPS traffic
    - Need content-based routing
    - WebSocket support
    - Want WAF integration
- DNS:
    - Multi-region routing
    - Coarse distribution
    - Simple failover

#### Real World Tools
- Cloud:
    - AWS: ALB (L7), NLB (L4), CLB (classic/deprecated), Route53 (DNS)
    - GCP: Cloud Load Balancing (global, anycast)
    - Azure: Application Gateway (L7), Load Balancer (L4)
- Open Source:
    - Nginx - popular, lightweight, high performance
    - HAProxy - very mature, feature-rich, TCP and HTTP
    - Envoy - modern, service mesh, dynamic config
    - Traefik - container-native, auto-discovery
- Hardware:
    - F5 BIG-IP - enterprise, expensive, powerful
    - Citrix NetScaler - app delivery controller

#### Quick Comparison Notes
- Nginx vs HAProxy:
    - Nginx: easier config, also web server, better for static content
    - HAProxy: more LB features, better health checks, pure LB
- Cloud vs Self-hosted:
    - Cloud: managed, auto-scaling, more expensive
    - Self-hosted: more control, cheaper at scale, need to manage
- L4 vs L7 decision:
    - L4 if: need speed, non-HTTP, static routing
    - L7 if: need routing logic, HTTP-specific features, easier debugging


-----
Questions:
1. Today, client requests go to LB1. Now for high availability, and scalability i add an new LB2. How would the clients know about LB2?
2. 