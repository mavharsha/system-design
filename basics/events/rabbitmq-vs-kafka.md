# RabbitMQ vs Kafka

## Core Philosophy

RabbitMQ and Kafka solve overlapping but fundamentally different problems. Understanding
their design philosophy is the key to choosing correctly.

**RabbitMQ** is a **message broker**. It routes messages from producers to consumers,
then discards them. Think of it as a post office -- it delivers mail and moves on.

**Kafka** is an **event streaming platform**. It stores an ordered, immutable log of
events that multiple consumers can read independently. Think of it as a newspaper --
everyone reads the same edition at their own pace, and back issues are kept on file.

```
RabbitMQ mental model:            Kafka mental model:

  Producer --> [Broker] --> Consumer     Producer --> [ Append-only Log ] --> Consumer A (offset 4)
               (deletes msg                          [ 0 | 1 | 2 | 3 | 4 ]   Consumer B (offset 2)
                after ack)                           (retains events)
```

---

## Architecture

### RabbitMQ: Exchanges, Bindings, Queues

RabbitMQ uses AMQP (Advanced Message Queuing Protocol). The core components are:

- **Producer** -- publishes a message to an **exchange**.
- **Exchange** -- routes the message to one or more **queues** based on **bindings** and a **routing key**.
- **Queue** -- holds messages until a consumer acknowledges them.
- **Consumer** -- subscribes to a queue and processes messages.

```
                        Binding (routing key = "order.created")
                       +------------------------------------+
                       |                                    v
Producer ---> [ Exchange (topic) ] ---+---> [ Queue: order-service ]  ---> Consumer A
                       |              |
                       |              +--> [ Queue: email-service ]   ---> Consumer B
                       |
                       +--- Binding (routing key = "order.*")
```

Exchange types:
| Type    | Routing behavior                            |
|---------|---------------------------------------------|
| direct  | Exact match on routing key                  |
| topic   | Wildcard pattern match (`order.*`, `#.log`) |
| fanout  | Broadcast to all bound queues               |
| headers | Match on message header attributes           |

### Kafka: Topics, Partitions, Consumer Groups

Kafka organizes data into:

- **Topic** -- a named feed of events (e.g., `orders`).
- **Partition** -- a topic is split into ordered, append-only logs. Partitions are the unit of parallelism.
- **Consumer Group** -- a set of consumers that share the work of reading a topic. Each partition is assigned to exactly one consumer in the group.
- **Offset** -- a consumer's position in a partition.

```
Topic: "orders"  (3 partitions)

  Partition 0:  [ msg0 | msg1 | msg2 | msg3 ]  --->  Consumer A  \
  Partition 1:  [ msg0 | msg1 | msg2 ]          --->  Consumer B   |-- Consumer Group "order-svc"
  Partition 2:  [ msg0 | msg1 ]                 --->  Consumer C  /

  Partition 0:  [ msg0 | msg1 | msg2 | msg3 ]  --->  Consumer X  \
  Partition 1:  [ msg0 | msg1 | msg2 ]          --->  Consumer X   |-- Consumer Group "analytics"
  Partition 2:  [ msg0 | msg1 ]                 --->  Consumer X  /
```

Key insight: multiple consumer groups each get **the full stream** independently.
Within a group, partitions are divided among members for parallel processing.

---

## Message Delivery Model: Push vs Pull

### RabbitMQ -- Push (broker pushes to consumers)

The broker pushes messages to connected consumers. You control flow with `prefetch`:

```java
// Java (Spring AMQP)
@RabbitListener(queues = "orders")
public void handleMessage(Message message, Channel channel) throws IOException {
    try {
        process(message.getBody());
        channel.basicAck(message.getMessageProperties().getDeliveryTag(), false);
    } catch (Exception e) {
        channel.basicNack(message.getMessageProperties().getDeliveryTag(), false, true);
    }
}

// Prefetch configuration
@Bean
public SimpleRabbitListenerContainerFactory rabbitListenerContainerFactory(
        ConnectionFactory connectionFactory) {
    SimpleRabbitListenerContainerFactory factory = new SimpleRabbitListenerContainerFactory();
    factory.setConnectionFactory(connectionFactory);
    factory.setPrefetchCount(10);  // receive at most 10 unacked messages
    return factory;
}
```

Pros: low latency -- message arrives the instant it is available.
Cons: a slow consumer can get overwhelmed; you must tune prefetch carefully.

### Kafka -- Pull (consumers poll the broker)

Consumers fetch batches of records at their own pace:

```java
// Java (Kafka consumer API)
while (true) {
    ConsumerRecords<String, String> records = consumer.poll(Duration.ofMillis(100));
    for (ConsumerRecord<String, String> record : records) {
        process(record.value());
    }
    consumer.commitSync();
}
```

Pros: consumer controls throughput; easy to batch; no risk of overwhelming the consumer.
Cons: introduces a small polling delay (typically < 100 ms in practice).

---

## Ordering Guarantees

### RabbitMQ

Messages in a **single queue** are delivered in FIFO order **to a single consumer**.
Once you add competing consumers on the same queue, ordering across consumers is
**not guaranteed** because different consumers process at different speeds.

```
Queue: [ A | B | C | D ]
  Consumer 1 gets A, C    (A finishes first)
  Consumer 2 gets B, D    (D finishes before B)

Processing order could be: A, D, C, B  -- not FIFO.
```

### Kafka

Ordering is guaranteed **within a partition**. Messages with the same key always land
in the same partition (via hash of key mod partition count), so you get per-key ordering.

```java
// Producing with a key ensures ordering for that key
producer.send(new ProducerRecord<>("orders", orderId, orderJson));
//                                  topic     key      value
```

All events for `orderId=42` go to the same partition and are read in order.

**Rule of thumb:** if you need strict global ordering, use a single partition (sacrificing
parallelism) or use RabbitMQ with a single consumer. If per-entity ordering is enough
(the common case), Kafka with keyed messages is ideal.

---

## Message Retention

This is one of the sharpest differences.

### RabbitMQ -- Consume and Delete

A message is removed from the queue once the consumer acknowledges it. The broker's job
is delivery, not storage. You can configure dead-letter queues for failed messages, and
you can set TTLs, but the default intent is: deliver once, then forget.

```
Time -->
Queue state:  [ A B C D ]  ->  consumer acks A  ->  [ B C D ]  ->  consumer acks B  -> [ C D ]
```

### Kafka -- Log-Based Retention

Messages are retained based on a configurable policy (time or size), regardless of whether
anyone has consumed them.

```properties
# server.properties or topic-level override
log.retention.hours=168        # keep for 7 days
log.retention.bytes=1073741824 # or keep up to 1 GB per partition
```

This means:
- A new consumer group can start reading from the **beginning** of the log.
- A consumer that crashes can resume from its last committed offset.
- You can replay events for debugging, reprocessing, or building new derived views.

```
Time -->
Partition:  [ 0 | 1 | 2 | 3 | 4 | 5 | 6 ]
                        ^               ^
                  Consumer A          Consumer B
                  (slow, offset 2)    (caught up, offset 6)

Both consumers read from the same immutable log.
Old segments are deleted only when retention policy expires.
```

---

## Throughput and Latency

| Dimension          | RabbitMQ                              | Kafka                                     |
|--------------------|---------------------------------------|-------------------------------------------|
| **Throughput**     | ~20K-50K msg/s per node (typical)     | ~100K-1M+ msg/s per broker (typical)      |
| **Latency**        | Sub-millisecond (push model)          | Low ms (pull/batch model)                 |
| **Scaling model**  | Add queues, cluster nodes, shovels    | Add partitions and brokers                |
| **Bottleneck**     | Per-queue is single-threaded          | Per-partition is single-consumer-in-group |

**Why Kafka is faster for throughput:** Messages are appended sequentially to disk
(sequential I/O is fast), batched in transit, and compressed. The broker does minimal
per-message work.

**Why RabbitMQ has lower latency for small volumes:** Push delivery avoids polling delay.
For request-reply patterns or low-volume command routing, RabbitMQ often wins.

---

## Acknowledgment and Delivery Semantics

### RabbitMQ

- **At-most-once:** auto-ack enabled -- message is removed before processing completes.
- **At-least-once:** manual ack after processing. If the consumer crashes, the message is redelivered.
- **Exactly-once:** not natively supported. Use idempotent consumers.

```java
// At-least-once: ack only after successful processing
@RabbitListener(queues = "payments", ackMode = "MANUAL")
public void handle(Message message, Channel channel) throws IOException {
    processPayment(message.getBody());
    channel.basicAck(message.getMessageProperties().getDeliveryTag(), false);
}
```

### Kafka

- **At-most-once:** commit offset before processing.
- **At-least-once:** commit offset after processing (default with `enable.auto.commit=false`).
- **Exactly-once:** supported via idempotent producers + transactional API (EOS).

```java
// Exactly-once setup
props.put("enable.idempotence", "true");
props.put("transactional.id", "order-processor-1");

producer.initTransactions();
producer.beginTransaction();
producer.send(record);
producer.commitTransaction();
```

---

## When to Pick Which

### Choose RabbitMQ when:

1. **Task/work queues** -- You need to distribute tasks among workers and each task should
   be processed by exactly one worker.
   - Example: image thumbnail generation, PDF rendering, email sending.

2. **Request-reply / RPC patterns** -- You need a response back from the consumer.
   - Example: a gateway sends a request to an auth service and waits for a reply on
     a temporary reply queue.

3. **Complex routing logic** -- You need header-based routing, priority queues, or
   delayed message scheduling.
   - Example: route audit events to different queues based on severity headers.

4. **Low-volume microservice communication** -- Simple pub/sub among a handful of services
   where operational simplicity matters more than throughput.

### Choose Kafka when:

1. **Event sourcing / event-driven architecture** -- You need a durable, replayable log
   of everything that happened.
   - Example: an e-commerce platform stores every order state change; a new reporting
     service replays the full history to build its own read model.

2. **High-throughput data pipelines** -- You need to move large volumes of data between
   systems reliably.
   - Example: clickstream data from a website --> Kafka --> Spark/Flink for real-time
     analytics, and also --> S3 for batch analytics.

3. **Multiple independent consumers** -- Many downstream systems need the same events.
   - Example: an `order.placed` event is consumed by inventory, billing, shipping,
     analytics, and fraud detection -- each as a separate consumer group.

4. **Stream processing** -- You need to join, aggregate, or transform streams of events
   in real time.
   - Example: join a `page-views` stream with a `users` stream to compute real-time
     per-user engagement metrics using Kafka Streams or ksqlDB.

5. **Audit log / compliance** -- You need an immutable, time-ordered record.

---

## Real-World Scenario Walkthrough

### Scenario: E-Commerce Order Processing

```
                            +-------------------+
                            |  Order Service    |
                            +---------+---------+
                                      |
                         publishes "order.placed"
                                      |
              +-----------+-----------+-----------+-----------+
              |           |           |           |           |
         Inventory    Billing    Shipping    Analytics    Fraud
         Service      Service    Service     Pipeline     Engine
```

**With RabbitMQ:**
You would create a fanout exchange and bind five queues (one per service). Each queue
gets a copy. This works, but if you add a sixth service later, you must create a new
queue and binding. If that service needs to process historical orders, you are out of luck
-- the messages are gone.

**With Kafka:**
You create a topic `order.placed`. Each service runs its own consumer group. Adding
a new service means adding a new consumer group -- no broker config change needed.
The new service can set `auto.offset.reset=earliest` to replay from the beginning.

Kafka is the better fit here because of multiple independent consumers and the need
for replay.

### Scenario: Background Job Queue (thumbnail generation)

```
  Web Server ---> [ Queue: generate-thumbnails ] ---> Worker 1
                                                  ---> Worker 2
                                                  ---> Worker 3
```

**With RabbitMQ:**
A natural fit. Push one task per image, workers compete for tasks, and each task is
processed exactly once. Failed tasks are requeued or sent to a dead-letter queue.
Prefetch ensures workers are not overwhelmed.

**With Kafka:**
This works but is awkward. You would need to manage partition assignment carefully.
If a worker is slow on one partition, other messages in that partition wait. There is
no built-in dead-letter queue. You are also retaining thumbnails tasks on disk for
days when you do not need them.

RabbitMQ is the better fit here.

---

## Summary Comparison

| Aspect                  | RabbitMQ                            | Kafka                                  |
|-------------------------|-------------------------------------|----------------------------------------|
| **Model**               | Message broker (smart broker)       | Event log (smart consumer)             |
| **Protocol**            | AMQP, MQTT, STOMP                   | Custom binary protocol                 |
| **Delivery**            | Push                                | Pull (consumer polls)                  |
| **Retention**           | Delete after ack                    | Retain by time/size policy             |
| **Ordering**            | Per-queue FIFO (single consumer)    | Per-partition FIFO                     |
| **Replay**              | Not possible                        | Built-in (reset offset)               |
| **Throughput**          | Tens of thousands msg/s             | Millions msg/s                         |
| **Latency**             | Sub-ms possible                     | Low ms (batching)                      |
| **Consumer groups**     | Competing consumers on one queue    | Independent groups, each gets full log |
| **Routing**             | Rich (exchange types, bindings)     | Topic-based only                       |
| **Exactly-once**        | No (use idempotent consumers)       | Yes (transactional API)                |
| **Dead letter support** | Built-in (DLX)                      | Manual (publish to error topic)        |
| **Best for**            | Task queues, RPC, routing           | Event streaming, pipelines, replay     |
| **Operational weight**  | Lighter (Erlang VM, single binary)  | Heavier (JVM, ZooKeeper/KRaft)         |

---

## One-Liner Decision Rule

> If the message is a **command** ("do this thing") and you want it processed once
> and forgotten, use **RabbitMQ**.
>
> If the message is an **event** ("this thing happened") and you want it stored,
> replayed, and consumed by many, use **Kafka**.

Both are production-grade, battle-tested tools. The wrong choice is not picking the one
that fits your access pattern.
