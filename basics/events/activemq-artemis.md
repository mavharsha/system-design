# ActiveMQ Artemis

## What Is ActiveMQ Artemis?

ActiveMQ Artemis is a high-performance, embeddable, JMS-compliant message broker built by
Red Hat (originally as HornetQ) and donated to the Apache ActiveMQ project. It is the
**successor** to ActiveMQ "Classic" -- the original ActiveMQ that has been around since 2004.

This distinction matters. When someone says "ActiveMQ" in a job listing or architecture doc,
you need to ask: **Classic or Artemis?** They are different codebases with different
internals. Artemis was designed from scratch with modern performance and correctness in mind.
Classic is in maintenance mode. New projects should use Artemis.

```
ActiveMQ "Classic" (5.x)          ActiveMQ Artemis (2.x)
  - KahaDB persistence              - Journal-based persistence
  - Mature but aging                 - Modern async IO (libaio / NIO)
  - OpenWire native                  - Multi-protocol native
  - Maintenance mode                 - Active development
```

---

## Core Architecture

### Journal-Based Persistence

Artemis writes messages to an append-only journal on disk. This journal is optimized for
sequential I/O, which is dramatically faster than random I/O on both spinning disks and SSDs.

```
Producer ---> [ In-memory ring buffer ] ---> [ Journal (append-only files) ]
                                                   |
                                                   v
                                            [ Compaction / cleanup ]
```

The journal uses one of two I/O implementations:
- **Linux AIO (libaio)** -- true asynchronous I/O on Linux. Best performance.
- **NIO (Java NIO)** -- fallback for macOS, Windows, or when libaio is unavailable.

This is why Artemis benchmarks so well: it treats the disk like a write-ahead log, not a
random-access database.

### Addresses, Queues, and Routing Types

This is the single most important concept in Artemis. The addressing model has three pieces:

| Concept      | What it is                                                        |
|--------------|-------------------------------------------------------------------|
| **Address**  | A logical name / destination. Messages are sent to an address.    |
| **Queue**    | A storage container bound to an address. Consumers read from it.  |
| **Routing**  | Determines how messages flow from address to queue(s).            |

There are two routing types:

**Anycast** -- point-to-point. The address has one queue and multiple consumers compete for
messages. Each message is delivered to exactly one consumer. This is your classic work queue.

**Multicast** -- pub/sub. Each subscription creates its own queue on the address. Every queue
gets a copy of every message. This is your classic topic.

```
Anycast (point-to-point):

  Producer ---> [ Address: "orders" ]
                        |
                        v
                  [ Queue: "orders" ]
                   /          \
            Consumer A    Consumer B
            (gets msg 1)  (gets msg 2)    <-- competing consumers


Multicast (pub/sub):

  Producer ---> [ Address: "events.order.created" ]
                      /                    \
                     v                      v
          [ Queue: sub-email ]    [ Queue: sub-analytics ]
                |                          |
          Consumer A                 Consumer B
          (gets ALL msgs)            (gets ALL msgs)
```

An address can support BOTH routing types simultaneously. This means you can have a single
address that behaves as a queue for some consumers and as a topic for others. In practice,
pick one per address to keep things simple.

---

## Protocols

Artemis natively supports six wire protocols through its pluggable acceptor architecture:

| Protocol    | Port (default) | Use case                                      |
|-------------|----------------|-----------------------------------------------|
| Core        | 61616          | Artemis native. Best performance.              |
| AMQP 1.0   | 5672           | Cross-platform standard. Use with non-Java.    |
| STOMP       | 61613          | Simple text protocol. Good for scripting/debug. |
| MQTT        | 1883           | IoT devices, lightweight pub/sub.              |
| OpenWire    | 61616          | Backward compat with ActiveMQ Classic clients. |
| HornetQ     | 5445           | Migration from legacy HornetQ installs.        |

All protocols land on the same internal addressing model. A message sent via MQTT can be
consumed via AMQP. This multi-protocol support is one of Artemis's strongest differentiators.

---

## Key Features

### Message Grouping

All messages with the same group ID are delivered to the same consumer. Useful when you need
ordering guarantees within a group (e.g., all events for a given customer go to one worker).

```java
jmsTemplate.convertAndSend("orders", order, message -> {
    message.setStringProperty("JMSXGroupID", order.getCustomerId());
    return message;
});
```

### Scheduled Delivery

Send a message now, but have it delivered later.

```java
jmsTemplate.convertAndSend("reminders", reminder, message -> {
    // Deliver 30 minutes from now
    long delay = System.currentTimeMillis() + Duration.ofMinutes(30).toMillis();
    message.setLongProperty("_AMQ_SCHED_DELIVERY", delay);
    return message;
});
```

### Last-Value Queues

Only the latest message with a given key is kept. Older messages with the same key are
discarded. Perfect for "current state" scenarios like price feeds or status updates.

Configure in `broker.xml`:
```xml
<address name="stock.prices">
  <anycast>
    <queue name="stock.prices" last-value-key="symbol" />
  </anycast>
</address>
```

### Large Message Support

Messages exceeding a configurable threshold (default 100 KB) are streamed to disk instead of
held in memory. This prevents one large payload from blowing up your heap.

### Paging

When a destination accumulates more messages than fit in memory, Artemis pages them to disk
transparently. This is a safety net -- not a performance feature. If you are paging
regularly, you have a consumer that cannot keep up and need to address that.

---

## Spring Boot Integration

### Dependencies

```xml
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-artemis</artifactId>
</dependency>
```

For an embedded broker (useful in dev/test), also add:

```xml
<dependency>
    <groupId>org.apache.activemq</groupId>
    <artifactId>artemis-jakarta-server</artifactId>
</dependency>
```

### application.yml -- Embedded Broker

```yaml
spring:
  artemis:
    mode: embedded
    embedded:
      enabled: true
      queues: orders,notifications    # auto-create these queues
```

### application.yml -- Remote Broker

```yaml
spring:
  artemis:
    mode: native
    broker-url: tcp://artemis-host:61616
    user: admin
    password: secret
```

### Sending Messages with JmsTemplate

```java
@Service
@RequiredArgsConstructor
public class OrderPublisher {

    private final JmsTemplate jmsTemplate;

    public void publishOrder(Order order) {
        jmsTemplate.convertAndSend("orders", order);
    }

    public void publishWithPriority(Order order, int priority) {
        jmsTemplate.convertAndSend("orders", order, message -> {
            message.setJMSPriority(priority);
            return message;
        });
    }
}
```

### Consuming Messages with @JmsListener

```java
@Component
@Slf4j
public class OrderConsumer {

    @JmsListener(destination = "orders")
    public void handleOrder(Order order) {
        log.info("Received order: {}", order.getId());
        // process the order
    }
}
```

You can also access raw JMS headers:

```java
@JmsListener(destination = "orders")
public void handleOrder(Order order,
                        @Header("JMSXGroupID") String groupId,
                        @Header("JMSMessageID") String messageId) {
    log.info("Order {} from group {}", order.getId(), groupId);
}
```

### Request-Reply Pattern

When you need a synchronous response over an async transport:

```java
// --- Sender side ---
@Service
@RequiredArgsConstructor
public class OrderService {

    private final JmsTemplate jmsTemplate;

    public OrderConfirmation placeOrder(Order order) {
        // sendAndReceive blocks until a reply arrives (or times out)
        Message reply = jmsTemplate.sendAndReceive("orders.request",
            session -> {
                ObjectMessage msg = session.createObjectMessage(order);
                return msg;
            });

        // extract the reply payload
        return (OrderConfirmation) ((ObjectMessage) reply).getObject();
    }
}

// --- Responder side ---
@Component
public class OrderResponder {

    @JmsListener(destination = "orders.request")
    @SendTo    // replies to the JMSReplyTo address automatically
    public OrderConfirmation handleOrderRequest(Order order) {
        // process and return confirmation
        return new OrderConfirmation(order.getId(), "CONFIRMED");
    }
}
```

The `@SendTo` annotation without a value tells Spring to use the `JMSReplyTo` header set by
`sendAndReceive`. You can also hardcode a reply destination: `@SendTo("orders.reply")`.

### JSON Message Conversion

By default, JmsTemplate uses Java serialization. You almost certainly want JSON instead.

```java
@Configuration
public class JmsConfig {

    @Bean
    public MessageConverter jacksonJmsMessageConverter() {
        MappingJackson2MessageConverter converter =
            new MappingJackson2MessageConverter();
        converter.setTargetType(MessageType.TEXT);
        converter.setTypeIdPropertyName("_type");  // required for deserialization
        return converter;
    }
}
```

With this in place, `jmsTemplate.convertAndSend("orders", order)` sends a JSON text message
with a `_type` header containing the fully-qualified class name. On the consumer side, the
`@JmsListener` method parameter is deserialized from JSON automatically.

Important: the `typeIdPropertyName` is a JMS string property set on the message. Both
producer and consumer must agree on the property name. If you are consuming messages from a
non-Spring producer that does not set this header, you will need a custom
`MessageConverter` or manual deserialization.

---

## Dead Letter and Expiry Addresses

Artemis has built-in DLQ and expiry support configured per address in `broker.xml`:

```xml
<address-settings>
    <!-- Catch-all defaults -->
    <address-setting match="#">
        <dead-letter-address>DLA</dead-letter-address>
        <expiry-address>ExpiryQueue</expiry-address>
        <max-delivery-attempts>5</max-delivery-attempts>
        <redelivery-delay>2000</redelivery-delay>        <!-- ms -->
        <redelivery-multiplier>2.0</redelivery-multiplier>
        <max-redelivery-delay>30000</max-redelivery-delay>
        <expiry-delay>86400000</expiry-delay>             <!-- 24 hours -->
    </address-setting>

    <!-- Override for a specific address -->
    <address-setting match="orders.#">
        <max-delivery-attempts>10</max-delivery-attempts>
        <redelivery-delay>5000</redelivery-delay>
    </address-setting>
</address-settings>
```

```
Message flow on failure:

  Consumer fails to process (throws exception)
       |
       v
  Broker redelivers (up to max-delivery-attempts, with backoff)
       |
       v  (all attempts exhausted)
  Message moved to dead-letter-address (DLA)
       |
       v
  [ DLA Queue ] <-- engineer inspects / replays from here
```

The `match` attribute supports wildcards:
- `#` matches any sequence of words separated by `.`
- `*` matches a single word

This is how you set different retry policies per destination. Always configure a catch-all
`#` as a safety net, then add specific overrides as needed.

---

## Clustering

Artemis supports clustering for high availability and horizontal scaling.

### Live-Backup Pairs

The fundamental HA unit is a live-backup pair. The live server handles traffic. The backup
server replicates its state and takes over if the live fails.

```
                 +------------------+          +------------------+
  Clients  --->  |  Live Server     |  <---->  |  Backup Server   |
                 |  (active)        |  (sync)  |  (passive)       |
                 +------------------+          +------------------+
                                                      |
                                                      v
                                               Takes over on failure
```

Two replication strategies:

| Strategy         | How it works                                     | Trade-off                       |
|------------------|--------------------------------------------------|---------------------------------|
| **Shared Store** | Live and backup share a filesystem (NFS, SAN)    | Simpler. Shared storage = SPOF. |
| **Replication**  | Live replicates journal to backup over network   | No shared storage. More network.|

For most teams, **replication** is the better choice unless you already have reliable shared
storage infrastructure.

### Scaling Out with Broker Clusters

Multiple live-backup pairs can form a cluster. Messages are load-balanced across cluster
members, and consumers connected to any member can receive messages from any queue in the
cluster.

```
  +--------------------+      +--------------------+      +--------------------+
  | Live A / Backup A' | <--> | Live B / Backup B' | <--> | Live C / Backup C' |
  +--------------------+      +--------------------+      +--------------------+
         ^                           ^                            ^
         |                           |                            |
    Producers /                Producers /                  Producers /
    Consumers                  Consumers                   Consumers
```

Cluster members discover each other via UDP multicast or a static list of connectors.
Message redistribution can be configured so that if a consumer is connected to node A but
the message arrives on node B, the broker forwards it automatically.

---

## Artemis vs RabbitMQ vs Kafka

This is not an exhaustive comparison -- see `rabbitmq-vs-kafka.md` for that deeper dive.
The goal here is to help you decide when Artemis belongs in the conversation.

| Dimension              | Artemis                          | RabbitMQ                     | Kafka                          |
|------------------------|----------------------------------|------------------------------|--------------------------------|
| **Primary model**      | Message broker (JMS)             | Message broker (AMQP)        | Event log / streaming          |
| **Protocol**           | 6 protocols native               | AMQP 0.9.1 + plugins        | Kafka protocol                 |
| **JMS support**        | Native, first-class              | Via plugin (limited)         | No                             |
| **Embeddable**         | Yes, in-process                  | No (Erlang runtime)          | No (JVM but not designed for it)|
| **Ordering**           | Per-queue + message groups       | Per-queue                    | Per-partition                  |
| **Message replay**     | No (consumed = gone)             | No                           | Yes (offset-based)             |
| **Throughput ceiling** | Very high (100k+ msg/s)          | High (50-80k msg/s)         | Extreme (millions msg/s)       |
| **Ecosystem**          | Java / Jakarta EE                | Polyglot                     | Polyglot + stream processing   |
| **Operations**         | JVM. Familiar to Java teams.     | Erlang. OTP knowledge helps. | JVM + ZK/KRaft. Heavy infra.   |

### When Artemis is the right choice

- **JMS is a requirement.** Your organization mandates JMS, or you are migrating from a
  legacy JMS broker (WebLogic JMS, IBM MQ, old ActiveMQ Classic). Artemis gives you a
  compliant, modern JMS implementation.
- **Java-heavy / Spring Boot shop.** The Spring Boot starter works out of the box. The
  broker is a single JVM process your team already knows how to operate.
- **Embedded broker for testing or lightweight deployments.** Spin up a real broker
  inside your application with zero external dependencies. No Docker, no separate process.
- **Multi-protocol gateway.** You have IoT devices speaking MQTT, web clients on STOMP,
  and backend services on AMQP or Core. Artemis bridges all of these natively.
- **You do not need event replay.** If your use case is "process this message and move on"
  rather than "replay the last 7 days of events," a traditional broker like Artemis is
  simpler than Kafka.

### When Artemis is NOT the right choice

- **Event sourcing / stream processing.** You need consumers to replay events from an
  arbitrary offset. Use Kafka.
- **Massive horizontal throughput.** You need millions of messages per second across
  hundreds of partitions. Kafka is purpose-built for this.
- **Polyglot ecosystem with no Java.** If nobody on the team writes Java and you have no
  JMS requirement, RabbitMQ is a more natural fit for Python/Ruby/Go shops.

---

## Common Pitfalls

### 1. Not Configuring Address Settings

Out of the box, Artemis auto-creates addresses and queues, which is convenient in dev but
dangerous in production. Messages to an address with no matching settings get the broker
defaults, which may have no DLQ, no expiry, and unlimited paging.

**Fix:** Always define explicit `<address-setting>` blocks in `broker.xml` for every address
pattern your application uses, plus a restrictive catch-all `#` default.

### 2. Journal Disk Performance

The journal is the single biggest determinant of Artemis throughput. Putting the journal on
a slow or shared NFS mount will cripple performance.

**Fix:** Use local SSDs for the journal directory. On Linux, use `libaio` (install the
`libaio` package). Monitor disk I/O latency -- if journal sync times exceed a few
milliseconds consistently, your storage is the bottleneck.

### 3. Large Message Threshold

The default large message threshold is 100 KB. Messages above this size are streamed to a
separate large-messages directory instead of being held in memory. If you routinely send
payloads close to this threshold, you may see unexpected performance characteristics as
messages flip between in-memory and streamed handling.

**Fix:** Set the threshold deliberately based on your actual payload sizes. If all your
messages are under 10 KB, the default is fine. If you regularly send 200 KB payloads and
accept the memory cost, raise the threshold. Do not ignore it.

### 4. Forgetting to Set Type ID for JSON Conversion

When using `MappingJackson2MessageConverter`, you must set `typeIdPropertyName`. Without it,
the consumer cannot determine which class to deserialize into and you get cryptic errors.

```java
converter.setTypeIdPropertyName("_type");  // do not forget this
```

### 5. Blocking in @JmsListener Methods

`@JmsListener` methods run on a thread pool. If your listener makes slow HTTP calls or
blocks on a database, you starve the pool and message consumption stalls.

**Fix:** Configure the listener container concurrency to match your workload:

```yaml
spring:
  jms:
    listener:
      concurrency: 5        # min threads
      max-concurrency: 20   # max threads
```

Or offload slow work to a separate async executor.

---

## When to Use Artemis -- Summary

```
Do you need JMS compliance?
  |
  +-- Yes --> Artemis is your best modern option.
  |
  +-- No
       |
       Do you need event replay / stream processing?
         |
         +-- Yes --> Kafka.
         |
         +-- No
              |
              Embedded broker or multi-protocol needed?
                |
                +-- Yes --> Artemis.
                |
                +-- No
                     |
                     Java shop?
                       |
                       +-- Yes --> Artemis or RabbitMQ. Both work well with Spring Boot.
                       |
                       +-- No  --> RabbitMQ (broader polyglot ecosystem).
```

Artemis occupies a practical sweet spot: it is a serious, high-performance broker that runs
on the JVM, speaks every protocol you are likely to need, embeds trivially in Spring Boot for
testing, and gives you JMS when you need it. It does not try to be Kafka. It does not try to
be a streaming platform. It is a message broker, and a very good one.
