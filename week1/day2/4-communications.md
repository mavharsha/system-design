### Communications

```
Assignment: Deployment log streaming. 
1. Load initial log through regular API
2. Realtime changes streamed
```

**Usual - Client Request Server Response**
- Mostly HTTP 1.1, 2
- Client initiates → Server responds → Connection closes

**Pros:** Simple, stateless, works everywhere
**Cons:** Client must initiate, can't push updates to client



**Short Polling**
- Client refreshes/makes requests every few seconds
- Server responds immediately (with or without new data)

**Example:** Continuously refreshing cricket score, checking if server is ready

**Pros:** 
- Simple to implement, standard HTTP

**Cons:** 
- High HTTP overhead (headers, connections)
- Most requests return "no update"
- Wastes bandwidth and server resources

---

**Long Polling**
- Client makes a request with **long timeout** (e.g., 30-60 seconds)
- Server holds connection open, waits for data
- Server responds when: data is available OR timeout expires
- Client immediately reconnects after response

**Key difference from short polling:** Just implementation - server waits instead of responding immediately

**Example:** Chat notifications, order status updates

**Pros:** 
- Near real-time updates
- Less overhead than short polling
- Standard HTTP

**Cons:** 
- Holds server connections open (resource intensive)
- Still has HTTP overhead on reconnect
- Need timeout handling

---

**Notes:**

Short polling and long polling are **just implementation patterns**, not different protocols. Both use standard HTTP - the difference is in timing/waiting behavior.

---

**WebSockets**
- Full-duplex, persistent bidirectional connection
- Starts as HTTP, then upgrades to WebSocket protocol
- Server can push anytime, client can send anytime

**Example:** Chat apps, live collaboration, gaming, real-time dashboards

**Pros:**
- True real-time bidirectional communication
- Low overhead (no HTTP headers per message)
- Single persistent connection

**Cons:**
- Stateful (harder to scale, need sticky sessions)
- More complex than HTTP
- Load balancers need special config
- Firewall/proxy issues

---

**Server-Sent Events (SSE)**
- Server pushes events to client over HTTP
- One-way: server → client only
- Connection stays open, server sends events as they happen

**Example:** Live feeds, notifications, stock tickers, deployment logs

**Pros:**
- Simple (standard HTTP)
- Auto-reconnect built-in
- Works through proxies/firewalls
- Server can push anytime

**Cons:**
- One-way only (can't send from client to server)
- Limited browser connections per domain
- Text-based only (no binary)

**Use case match for assignment:** SSE perfect for deployment log streaming (server pushes new logs)

---

## WebSockets vs SSE - When to Use What?

| Feature | WebSockets | SSE |
|---------|------------|-----|
| **Direction** | Bidirectional (⇄) | One-way (server → client) |
| **Protocol** | Custom (ws://) | HTTP |
| **Complexity** | Higher | Lower |
| **Scaling** | Harder (stateful) | Easier |
| **Resource overhead** | Higher (memory per connection, CPU for state management) | Lower (simpler, standard HTTP) |
| **Auto-reconnect** | Manual | Built-in |
| **Data format** | Binary + Text | Text only |

**Cost/Resource considerations:**
- **WebSockets:** More expensive to scale - each connection consumes memory, needs sticky sessions (can't easily load balance). More server resources = higher costs.
- **SSE:** Cheaper to scale - simpler connection management, works with standard HTTP load balancers. Less overhead per connection.
- Both hold persistent connections, but WebSockets require more careful resource management at scale.

**Use WebSockets when:**
- Need bidirectional communication (both sides send/receive)
- Chat applications, multiplayer games
- Real-time collaboration (Google Docs style)
- Low latency is critical
- Need binary data transfer

**Use SSE when:**
- Only server needs to push (one-way)
- Live feeds, notifications, dashboards
- Want simplicity (just HTTP)
- Need auto-reconnect
- Text data is sufficient (logs, JSON updates)

**Rule of thumb:** 
- If server pushes AND client needs to send back → **WebSockets**
- If only server pushes → **SSE** (simpler)

**For the assignment (deployment logs):** SSE is better - server pushes logs, client just displays. No need for bidirectional complexity.

---

