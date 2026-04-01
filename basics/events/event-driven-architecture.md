# Event-Driven Architecture (EDA)

## What Is Event-Driven Architecture?

Think of the difference between polling and interrupts at the hardware level. Polling
(request-driven) means you keep asking "did something happen?" over and over.
Interrupts (event-driven) mean the system *tells you* when something happens.

EDA applies the same idea at the application and system level. Instead of Service A
calling Service B directly and waiting for a response, Service A announces "something
happened" and any interested service reacts to it independently.

### Request-Driven vs Event-Driven

**Request-driven (synchronous):**
```
OrderService --HTTP POST--> InventoryService --HTTP POST--> PaymentService
     |                            |                              |
     |<-------- 200 OK ----------|<---------- 200 OK -----------|
```
OrderService knows about InventoryService. InventoryService knows about PaymentService.
If PaymentService is down, the entire chain fails. Every service waits for the next one.

**Event-driven (asynchronous):**
```
OrderService --publishes "OrderPlaced"--> [Event Broker]
                                              |
                          +-------------------+-------------------+
                          |                   |                   |
                   InventoryService    PaymentService    NotificationService
                   (subscribes)        (subscribes)      (subscribes)
```
OrderService knows nothing about who consumes its events. Services fail independently.
New consumers can be added without touching the producer.

---

## Core Concepts

### Events

An event is a record of something that already happened. Past tense matters here.
Not "CreateOrder" (a command) but "OrderPlaced" (a fact). Events are immutable. You
cannot un-happen something.

A typical event looks like:

```json
{
  "eventId": "evt-8a3f-4b2c",
  "eventType": "OrderPlaced",
  "timestamp": "2026-03-15T10:30:00Z",
  "source": "order-service",
  "data": {
    "orderId": "ord-1234",
    "customerId": "cust-567",
    "items": [
      { "sku": "WIDGET-A", "quantity": 2, "price": 29.99 }
    ],
    "totalAmount": 59.98
  }
}
```

### Producers (Publishers)

The service where the event originates. The OrderService creates orders and publishes
`OrderPlaced` events. It does not care who is listening or what they do with it.

### Consumers (Subscribers)

Services that react to events. The InventoryService subscribes to `OrderPlaced` to
reserve stock. The AnalyticsService subscribes to the same event to update dashboards.
Neither knows the other exists.

### Event Channels / Brokers

The infrastructure that routes events from producers to consumers. This is the
middleman -- a message broker or streaming platform like Kafka, RabbitMQ, or AWS
EventBridge. It handles delivery, buffering, and (depending on the tool) persistence.

Key broker responsibilities:
- **Routing**: getting events to the right consumers
- **Buffering**: holding events when consumers are slow or temporarily down
- **Persistence**: storing events for replay (in streaming platforms like Kafka)
- **Ordering**: maintaining event sequence (usually per partition/key)

---

## Types of Events

Not all events serve the same purpose. Martin Fowler draws a useful distinction
between three patterns:

### 1. Event Notification

A thin event that says "something happened" but carries minimal data. Consumers
must call back to the source if they need details.

```json
{
  "eventType": "OrderPlaced",
  "data": { "orderId": "ord-1234" }
}
```

Consumers that care about order details call the OrderService API to fetch them.

**Pros**: Small payloads, source stays the authority.
**Cons**: Creates runtime coupling (the callback). If OrderService is down, consumers
are stuck.

### 2. Event-Carried State Transfer

The event carries all the data a consumer might need. No callbacks required.

```json
{
  "eventType": "OrderPlaced",
  "data": {
    "orderId": "ord-1234",
    "customerId": "cust-567",
    "customerEmail": "alice@example.com",
    "items": [...],
    "shippingAddress": { ... },
    "totalAmount": 59.98
  }
}
```

**Pros**: True decoupling at runtime. Consumers work even if the producer is down.
**Cons**: Larger payloads. Data can go stale (consumer has a snapshot, not live data).

### 3. Domain Events vs Integration Events

- **Domain events** are internal to a bounded context. `OrderItemAdded` might be
  used inside the order service to trigger business rules. These use your domain
  language and internal data structures.

- **Integration events** cross service boundaries. `OrderPlaced` is published to
  the broker for external consumers. These need a stable, versioned contract because
  you do not control who consumes them.

Rule of thumb: domain events can change freely. Integration events need backward
compatibility and schema management.

---

## Patterns

### Pub/Sub (Publish-Subscribe)

The most common pattern. Producers publish to a topic. Consumers subscribe to topics
they care about. The broker handles fan-out.

```
Producer --> Topic: "orders.placed" --> Consumer Group A (InventoryService)
                                    --> Consumer Group B (NotificationService)
                                    --> Consumer Group C (AnalyticsService)
```

Each consumer group gets every message. Within a group, messages are distributed
across instances for parallel processing.

Use pub/sub when multiple independent services need to react to the same event.

### Event Streaming

Like pub/sub, but the broker *retains* events. Consumers can replay from any point
in the stream. Kafka is the canonical example.

This is powerful because:
- A new service can "catch up" by reading the full history
- You can reprocess events after fixing a bug in a consumer
- The event log becomes a source of truth

### Event Sourcing (Brief)

Instead of storing current state (like a row in a database), you store the sequence
of events that produced that state.

```
Account-123 events:
  1. AccountOpened { balance: 0 }
  2. MoneyDeposited { amount: 500 }
  3. MoneyWithdrawn { amount: 200 }
  4. MoneyDeposited { amount: 100 }
  
Current state: balance = 400 (derived by replaying events)
```

Event sourcing gives you a complete audit trail and the ability to reconstruct state
at any point in time. But it adds significant complexity -- you need snapshots for
performance, and your domain logic must work with event replay. Use it when the
history itself is valuable (financial systems, collaborative editing, audit-heavy
domains), not as a default.

### Choreography vs Orchestration

Two ways to coordinate multi-step workflows across services:

**Choreography**: each service listens for events and reacts independently. No
central coordinator. Services only know about events, not about each other.

```
OrderPlaced --> InventoryService reserves stock, emits "InventoryReserved"
InventoryReserved --> PaymentService charges card, emits "PaymentCharged"
PaymentCharged --> ShippingService creates shipment, emits "ShipmentCreated"
ShipmentCreated --> NotificationService emails customer
```

**Orchestration**: a central orchestrator (saga coordinator) tells each service
what to do and tracks the overall workflow state.

```
OrderSaga:
  1. Tell InventoryService to reserve stock
  2. Wait for confirmation
  3. Tell PaymentService to charge
  4. Wait for confirmation
  5. Tell ShippingService to ship
  6. Notify customer
```

| Aspect | Choreography | Orchestration |
|--------|-------------|---------------|
| Coupling | Very loose | Orchestrator depends on all services |
| Visibility | Hard to see full flow | Flow is explicit in orchestrator |
| Adding steps | Add a new subscriber | Modify the orchestrator |
| Error handling | Compensating events | Centralized rollback logic |
| Debugging | Challenging | Easier (single place to look) |

In practice, many systems use both. Choreography between bounded contexts,
orchestration within complex workflows inside a context.

---

## Benefits

### Loose Coupling

Producers and consumers evolve independently. You can replace, upgrade, or add
services without touching existing ones. The OrderService team ships on their
schedule; the InventoryService team ships on theirs.

### Scalability

Consumers scale independently based on their own load. If notification sending
is the bottleneck, scale NotificationService without touching anything else.
The broker absorbs spikes -- producers can publish faster than consumers process,
and the backlog is handled gracefully.

### Auditability

Events form a natural audit log. "What happened and when?" is answered by the
event stream itself. In regulated industries (finance, healthcare), this is
often a hard requirement, not just a nice-to-have.

### Extensibility

Need to add fraud detection to the order flow? Subscribe a new FraudService to
`OrderPlaced`. Zero changes to existing services. This is the open-closed
principle applied at the system level.

---

## Challenges

### Eventual Consistency

In a request-driven system, after the API returns 200, you know the write is done.
In EDA, when `OrderPlaced` is published, inventory has not been reserved yet. The
system is *eventually* consistent.

This means your UI and APIs must handle intermediate states. "Order received,
processing..." is a real state your users will see. If your business cannot
tolerate a window of inconsistency (e.g., double-selling limited inventory),
you need careful design (pessimistic reservations, idempotency, etc.).

### Debugging Distributed Flows

When something goes wrong, there is no single call stack to inspect. The order
was placed, inventory was reserved, but payment never happened. Where did it
break?

Mitigations:
- **Correlation IDs**: attach a unique ID to the initial event and propagate it
  through every downstream event. Now you can search logs across services.
- **Distributed tracing**: tools like Jaeger or OpenTelemetry visualize the
  full event flow.
- **Dead letter queues (DLQs)**: events that fail processing land here for
  inspection instead of being lost.

### Event Ordering

If `OrderPlaced` and `OrderCancelled` arrive out of order, you have a problem.
Brokers like Kafka guarantee ordering *within a partition*. Design your partition
key carefully (e.g., partition by orderId so all events for an order arrive in
sequence).

Across partitions or across topics, you generally cannot guarantee ordering. Design
consumers to handle this (timestamps, version numbers, idempotent operations).

### Schema Evolution

Your `OrderPlaced` event version 1 has 5 fields. Version 2 adds a `couponCode`
field. Version 3 renames `totalAmount` to `total`. Consumers on older versions
must not break.

Strategies:
- **Always add, never remove or rename** (backward compatible changes)
- **Schema registry** (Confluent Schema Registry for Kafka, for example)
  enforces compatibility rules at the broker level
- **Use flexible formats** like Avro or Protobuf with explicit versioning
- **Consumer tolerance**: ignore unknown fields, use sensible defaults for
  missing fields

---

## Real-World Example: E-Commerce Order Flow (Choreography)

Here is a concrete choreography-based flow for placing an order:

```
Customer places order via API
         |
         v
  [OrderService]
    - Validates order
    - Saves order (status: PENDING)
    - Publishes: OrderPlaced { orderId, items, customerId, totalAmount }
         |
         v
  [Event Broker - "orders.placed" topic]
         |
    +----+----+--------------------+
    |         |                    |
    v         v                    v
[Inventory] [Fraud]          [Notification]
 Service    Service            Service
    |         |                    |
 Reserves   Checks risk        Sends "order
 stock      score              received" email
    |         |
    v         v
 Publishes  Publishes
 Inventory  FraudCheckPassed
 Reserved   (or FraudCheckFailed)
    |         |
    +----+----+
         |
         v
  [PaymentService]
    - Waits for BOTH InventoryReserved AND FraudCheckPassed
    - Charges payment method
    - Publishes: PaymentCharged { orderId, transactionId }
         |
         v
  [ShippingService]
    - Creates shipping label
    - Publishes: ShipmentCreated { orderId, trackingNumber }
         |
         v
  [NotificationService]
    - Sends shipping confirmation with tracking number

  [OrderService]
    - Subscribes to all events for its orders
    - Updates order status: PENDING -> CONFIRMED -> SHIPPED
```

**What happens when payment fails?**

```
PaymentService publishes: PaymentFailed { orderId, reason }
    |
    +--------> InventoryService: releases reserved stock
    +--------> NotificationService: sends "payment failed" email
    +--------> OrderService: updates status to PAYMENT_FAILED
```

Each service publishes compensating events. No central coordinator needed, but
notice how the full flow is now spread across multiple services and event
handlers. Tracing a single order requires following events across topics
and services -- this is the trade-off.

---

## When NOT to Use EDA

EDA is not universally better. Avoid it when:

- **Simple CRUD applications**: a monolithic app with a REST API and a database
  is simpler, faster to build, and easier to debug. Do not add a message broker
  to a TODO app.

- **Strong consistency is required**: if "read your own write" semantics are
  critical and users cannot tolerate stale reads, synchronous calls or
  distributed transactions may be more appropriate. Example: checking account
  balance before an ATM withdrawal.

- **Low latency, request-response flows**: if the caller needs an immediate
  answer (e.g., "is this username available?"), going through a broker adds
  latency and complexity for no benefit.

- **Small teams / early-stage products**: the operational overhead of brokers,
  DLQs, schema registries, tracing, and eventual consistency handling is real.
  Start simple and introduce EDA when you have a concrete scaling or coupling
  problem.

- **Tight coordination with rollback**: when you need all-or-nothing across
  multiple steps and compensating events are impractical, an orchestrated saga
  or even a two-phase commit might be clearer.

---

## Tools Landscape

| Tool | Type | Best For |
|------|------|----------|
| **Apache Kafka** | Event streaming platform | High-throughput, durable event streams, replay capability. The go-to for large-scale event-driven systems. |
| **RabbitMQ** | Message broker | Traditional pub/sub and work queues. Mature, flexible routing, good for task distribution. |
| **AWS EventBridge** | Serverless event bus | AWS-native event routing with filtering rules. Great for connecting AWS services and SaaS integrations. |
| **NATS** | Lightweight messaging | Low-latency, simple pub/sub. Popular in cloud-native / Kubernetes environments. JetStream adds persistence. |
| **Pulsar** | Event streaming | Multi-tenancy, tiered storage, geo-replication out of the box. Alternative to Kafka with different trade-offs. |
| **Redis Streams** | Lightweight streaming | When you already have Redis and need simple event streaming without a dedicated broker. |

**Choosing between them**: if you need durable event streams with replay, look at
Kafka or Pulsar. If you need flexible routing and traditional messaging, RabbitMQ.
If you are in AWS and want minimal ops, EventBridge. If you want lightweight and
fast, NATS. There is no single right answer -- it depends on your durability,
throughput, and operational requirements.

---

## Key Takeaways

1. EDA decouples services in time (asynchronous) and knowledge (producers do not
   know consumers). This is its superpower and its primary source of complexity.

2. Choose the right event type for your situation. Thin notifications create
   runtime coupling; fat state-transfer events create data staleness risk.

3. Choreography keeps things loosely coupled but makes flows hard to follow.
   Orchestration centralizes logic but creates a coordination bottleneck. Use both.

4. Invest in observability *before* you need it: correlation IDs, distributed
   tracing, dead letter queues. You will thank yourself during the first
   production incident.

5. EDA is a tool, not a religion. Use it where the trade-offs make sense. Many
   successful systems are hybrids -- synchronous for queries and simple writes,
   event-driven for cross-service workflows and reactive processing.
