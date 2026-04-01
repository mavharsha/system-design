# Apache Kafka — A Technical Deep Dive

## What Kafka Actually Is

Kafka is a **distributed event streaming platform**. That distinction matters. A traditional
message queue (like RabbitMQ) delivers a message to one consumer, then the message is gone.
Kafka keeps an immutable, ordered log of events that multiple consumers can read independently,
at their own pace, and even replay from the past.

Think of it as a distributed commit log you can subscribe to. It is designed for:
- High-throughput event streaming (millions of messages/sec on commodity hardware)
- Durable storage of event streams (days, weeks, forever if you want)
- Decoupling producers and consumers in both time and space

If you only need simple request-reply or task distribution, Kafka is overkill. But if you need
an ordered, replayable, high-throughput event backbone, it is the standard.

---

## Core Architecture

```
                         Kafka Cluster
  +----------------------------------------------------------+
  |                                                          |
  |   Broker 0            Broker 1            Broker 2       |
  |  +-----------+      +-----------+      +-----------+     |
  |  | Topic A   |      | Topic A   |      | Topic A   |     |
  |  | Part 0 *  |      | Part 1 *  |      | Part 2 *  |     |
  |  | Part 1 R  |      | Part 2 R  |      | Part 0 R  |     |
  |  +-----------+      +-----------+      +-----------+     |
  |                                                          |
  |  * = leader replica    R = follower replica              |
  +----------------------------------------------------------+
            |                                  ^
        Producers                          Consumers
```

### Brokers
A Kafka cluster is a set of broker servers. Each broker holds a subset of partition data. There
is no single master — leadership is per-partition. Metadata coordination uses KRaft (Kafka Raft)
in modern versions (ZooKeeper is being phased out).

### Topics
A topic is a named stream of events. Think of it as a category or feed name. Topics are purely
logical — the physical unit is the partition.

### Partitions
Each topic is split into one or more partitions. A partition is an ordered, append-only log.
Ordering is guaranteed only within a single partition. This is the unit of parallelism: you
cannot have more active consumers in a group than you have partitions.

### Segments
Each partition is stored on disk as a series of segment files. When a segment reaches a
configured size or age, Kafka rolls to a new one. Old segments are what get deleted or compacted
by retention policies.

### Replicas and ISR
Each partition has a configurable replication factor (typically 3 in production). One replica is
the **leader** (handles all reads and writes); the rest are **followers**. The set of replicas
that are fully caught up is called the **ISR (In-Sync Replica set)**. If a leader dies, a new
leader is elected from the ISR. If a follower falls behind by more than `replica.lag.time.max.ms`,
it is removed from the ISR.

**Production rule:** Set `min.insync.replicas=2` with `replication.factor=3` and producer
`acks=all`. This ensures you can tolerate one broker failure without data loss.

---

## Producer Internals

```
  Your Application
       |
       v
  +-------------------+
  | KafkaProducer     |
  |  Serializer       |   -- key + value serialized
  |  Partitioner      |   -- decides target partition
  |  RecordAccumulator|   -- batches records per partition
  |  Compression      |   -- compresses each batch (lz4, snappy, zstd)
  |  Sender Thread    |   -- sends batches to brokers
  +-------------------+
       |
       v
    Broker (partition leader)
```

### Batching
The producer does not send each record immediately. It accumulates records into batches
(`batch.size` and `linger.ms`). Larger batches = higher throughput but slightly higher latency.
Start with `linger.ms=5` and `batch.size=32768` and tune from there.

### Compression
Set `compression.type=lz4` or `zstd`. Compression happens at the batch level, so bigger
batches compress better. In production, always enable compression — it reduces network and
disk I/O significantly.

### Partitioner
By default, if you provide a key, Kafka hashes it (`murmur2`) to determine the partition.
Same key always goes to the same partition, which guarantees ordering for that key. If no key
is provided, the sticky partitioner fills up one batch before moving to the next partition.

**Always key your messages** when ordering matters. If you send order events, key by `orderId`.

### Acks
| Setting    | Meaning                                        | Risk                |
|------------|------------------------------------------------|---------------------|
| `acks=0`   | Fire and forget. Don't wait for broker.        | Data loss possible  |
| `acks=1`   | Wait for leader to write to its local log.     | Loss if leader dies before replication |
| `acks=all` | Wait for all ISR replicas to acknowledge.      | Safest. Use this.   |

### Retries and Idempotent Producer
Set `enable.idempotence=true`. This automatically sets `acks=all`, enables retries, and
guarantees exactly-once delivery to a partition (the broker de-duplicates by producer ID and
sequence number). There is no reason not to enable this in any modern Kafka setup.

---

## Consumer Internals

### Consumer Groups

```
  Topic "orders" (3 partitions)

  +----------+  +----------+  +----------+
  | Part 0   |  | Part 1   |  | Part 2   |
  +----+-----+  +----+-----+  +----+-----+
       |              |              |
       v              v              v
  Consumer A     Consumer B     Consumer C
  <----------- Consumer Group "order-service" ----------->
```

Each partition is assigned to exactly one consumer in the group. If you have 3 partitions and
4 consumers, one consumer will be idle. If you have 3 partitions and 2 consumers, one consumer
gets 2 partitions.

### Partition Assignment Strategies
- **Range** (default): Assigns contiguous partitions per topic to each consumer. Can cause
  uneven load across multiple topics.
- **RoundRobin**: Distributes partitions evenly across consumers, across all subscribed topics.
- **Sticky**: Like round-robin but tries to minimize partition movement during rebalancing.
  **Use this one.** Set `partition.assignment.strategy=
  org.apache.kafka.clients.consumer.StickyAssignor` or better yet, the
  `CooperativeStickyAssignor` which avoids stop-the-world rebalances.

### Offset Management
Kafka tracks each consumer group's position (offset) per partition. Two modes:

- **Auto commit** (`enable.auto.commit=true`): Offsets are committed every
  `auto.commit.interval.ms`. Simple, but you can lose or duplicate messages on crashes.
- **Manual commit**: You call `commitSync()` or `commitAsync()` after processing. Use this
  whenever at-least-once or exactly-once semantics matter — which is almost always.

### Rebalancing
When consumers join, leave, or crash, Kafka reassigns partitions. During a rebalance, the
affected partitions stop being consumed. This is why rebalancing storms are dangerous (see
pitfalls section). The `CooperativeStickyAssignor` uses incremental rebalancing which only
moves the partitions that need moving, instead of revoking everything first.

---

## Kafka's Storage Model

```
  Partition 0 directory on disk:
  +-----------------------------------------------+
  | 00000000000000000000.log   (segment 0)        |
  | 00000000000000000000.index                     |
  | 00000000000000000000.timeindex                 |
  | 00000000000000345210.log   (segment 1)        |
  | 00000000000000345210.index                     |
  | 00000000000000345210.timeindex                 |
  +-----------------------------------------------+
```

### Append-Only Log
Every write appends to the end of the active segment. There are no random writes. This is why
Kafka is fast — sequential disk I/O on modern SSDs (or even spinning disks) is extremely
efficient, and Kafka uses the OS page cache aggressively.

### Retention Policies
- **Time-based** (`retention.ms`): Delete segments older than N ms. Default is 7 days.
- **Size-based** (`retention.bytes`): Delete oldest segments when partition exceeds N bytes.
- **Compact** (`cleanup.policy=compact`): Keep only the latest value for each key. This turns
  a topic into a key-value table. Great for changelogs, configuration, or state snapshots.

Log compaction does not happen in real time — it runs in the background and only affects closed
segments.

---

## Spring Boot Integration

### Dependencies (Maven)

```xml
<dependency>
    <groupId>org.springframework.kafka</groupId>
    <artifactId>spring-kafka</artifactId>
</dependency>
```

### Producer Configuration and KafkaTemplate

```java
@Configuration
public class KafkaProducerConfig {

    @Bean
    public ProducerFactory<String, Object> producerFactory() {
        Map<String, Object> props = new HashMap<>();
        props.put(ProducerConfig.BOOTSTRAP_SERVERS_CONFIG, "localhost:9092");
        props.put(ProducerConfig.KEY_SERIALIZER_CLASS_CONFIG, StringSerializer.class);
        props.put(ProducerConfig.VALUE_SERIALIZER_CLASS_CONFIG, JsonSerializer.class);
        props.put(ProducerConfig.ACKS_CONFIG, "all");
        props.put(ProducerConfig.ENABLE_IDEMPOTENCE_CONFIG, true);
        props.put(ProducerConfig.COMPRESSION_TYPE_CONFIG, "lz4");
        props.put(ProducerConfig.LINGER_MS_CONFIG, 5);
        return new DefaultKafkaProducerFactory<>(props);
    }

    @Bean
    public KafkaTemplate<String, Object> kafkaTemplate() {
        return new KafkaTemplate<>(producerFactory());
    }
}
```

```java
@Service
@RequiredArgsConstructor
public class OrderEventPublisher {

    private final KafkaTemplate<String, Object> kafkaTemplate;

    public void publishOrderCreated(OrderCreatedEvent event) {
        // Key by orderId so all events for the same order hit the same partition
        kafkaTemplate.send("order-events", event.getOrderId(), event)
            .whenComplete((result, ex) -> {
                if (ex != null) {
                    log.error("Failed to publish event for order {}", event.getOrderId(), ex);
                } else {
                    log.debug("Published to partition {} offset {}",
                        result.getRecordMetadata().partition(),
                        result.getRecordMetadata().offset());
                }
            });
    }
}
```

### Consumer with @KafkaListener

**Single record processing:**

```java
@Component
public class OrderEventConsumer {

    @KafkaListener(
        topics = "order-events",
        groupId = "order-service",
        properties = {
            "auto.offset.reset=earliest",
            "enable.auto.commit=false"
        }
    )
    public void handleOrderEvent(
            @Payload OrderCreatedEvent event,
            @Header(KafkaHeaders.RECEIVED_PARTITION) int partition,
            @Header(KafkaHeaders.OFFSET) long offset,
            Acknowledgment ack) {

        log.info("Processing order {} from partition {} offset {}",
            event.getOrderId(), partition, offset);

        orderService.process(event);
        ack.acknowledge();  // manual commit after successful processing
    }
}
```

**Batch processing (higher throughput):**

```java
@KafkaListener(
    topics = "clickstream",
    groupId = "analytics-service",
    batch = "true"
)
public void handleClickBatch(List<ClickEvent> events, Acknowledgment ack) {
    analyticsService.bulkInsert(events);  // write to DB in one batch
    ack.acknowledge();
}
```

### Error Handling and Retry

```java
@Configuration
@EnableKafka
public class KafkaConsumerConfig {

    @Bean
    public ConcurrentKafkaListenerContainerFactory<String, Object> kafkaListenerContainerFactory(
            ConsumerFactory<String, Object> consumerFactory) {

        ConcurrentKafkaListenerContainerFactory<String, Object> factory =
            new ConcurrentKafkaListenerContainerFactory<>();
        factory.setConsumerFactory(consumerFactory);
        factory.getContainerProperties().setAckMode(AckMode.MANUAL);

        // Retry 3 times with backoff, then send to DLT (dead letter topic)
        factory.setCommonErrorHandler(new DefaultErrorHandler(
            new DeadLetterPublishingRecoverer(kafkaTemplate(),
                (record, ex) -> new TopicPartition(
                    record.topic() + ".DLT", record.partition())),
            new FixedBackOff(1000L, 3)  // 1 sec interval, 3 attempts
        ));

        return factory;
    }
}
```

The pattern: retry a few times, then dump the poison message to a dead letter topic where you
can inspect and replay it later. Never let a bad message block your consumer indefinitely.

### Serialization

Use `JsonSerializer`/`JsonDeserializer` from Spring Kafka for most cases. For schema evolution
in large organizations, use **Apache Avro** with a **Schema Registry** (Confluent). Avro gives
you compact binary encoding and backward/forward compatibility guarantees.

```yaml
# application.yml — consumer deserialization
spring:
  kafka:
    consumer:
      properties:
        spring.json.trusted.packages: "com.yourcompany.events.*"
      key-deserializer: org.apache.kafka.common.serialization.StringDeserializer
      value-deserializer: org.springframework.kafka.support.serializer.JsonDeserializer
```

---

## Kafka Streams (Brief Intro)

Kafka Streams is a **client library** (not a separate cluster) for building stream processing
applications that read from and write to Kafka topics. It gives you:

- Stateful operations (aggregations, joins, windowing) backed by local RocksDB state stores
- Exactly-once processing semantics
- Automatic parallelism based on partition count
- No need for a separate processing cluster (unlike Flink or Spark Streaming)

**When to use it:** When your input is Kafka and your output is Kafka, and you need
transformations, aggregations, or joins between streams. If your output is a database or an
HTTP call, a plain consumer is simpler.

### Simple Word Count Example

```java
@Bean
public KStream<String, String> wordCountStream(StreamsBuilder builder) {
    KStream<String, String> textLines = builder.stream("text-input");

    KTable<String, Long> wordCounts = textLines
        .flatMapValues(line -> Arrays.asList(line.toLowerCase().split("\\W+")))
        .groupBy((key, word) -> word)
        .count(Materialized.as("word-counts-store"));

    // Write results to an output topic
    wordCounts.toStream()
        .mapValues(Object::toString)
        .to("word-count-output");

    return textLines;
}
```

---

## Kafka Connect (Brief Mention)

Kafka Connect is a framework for streaming data between Kafka and external systems without
writing code. It runs connectors:

- **Source connectors**: Pull data into Kafka (e.g., Debezium for CDC from Postgres/MySQL,
  JDBC source connector)
- **Sink connectors**: Push data from Kafka to external systems (e.g., Elasticsearch sink,
  S3 sink, JDBC sink)

Use Connect when a well-tested connector exists for your use case. Do not write a custom
consumer to dump data into S3 when the S3 sink connector does it better and handles offset
tracking, exactly-once, and fault tolerance for you.

---

## Operational Essentials

### Partition Count Selection
- More partitions = more parallelism = more consumers can work in parallel.
- Start with `max(expected throughput / throughput per consumer, expected throughput / throughput per producer)`.
- A practical starting point: **6 to 12 partitions** for most topics. Scale up if needed.
- You can increase partitions later, but you **cannot decrease** them. And increasing breaks
  key-based ordering guarantees for existing keys.

### Replication Factor
- **Always 3** in production. This tolerates one broker failure with `min.insync.replicas=2`.
- In dev/staging, 1 or 2 is fine.

### Consumer Lag Monitoring
Consumer lag = latest offset in partition minus the consumer group's committed offset. This is
the single most important Kafka operational metric. Monitor it with:
- Kafka's built-in `kafka-consumer-groups.sh --describe`
- Burrow (LinkedIn's lag monitoring tool)
- Your metrics stack (Prometheus + the JMX exporter)

If lag grows continuously, your consumers are too slow. Add more consumers (up to partition
count) or optimize processing.

---

## Common Pitfalls

### Too Few Partitions
You create a topic with 1 partition. You deploy 10 consumer instances. 9 of them sit idle.
You cannot parallelize consumption beyond the partition count. Plan ahead.

### Consumer Group Rebalancing Storms
A consumer takes too long to process a batch and exceeds `max.poll.interval.ms` (default 5
min). Kafka thinks it is dead and triggers a rebalance. The rebalance causes other consumers
to pause. They also exceed the interval. Cascading rebalances. Fix:
- Increase `max.poll.interval.ms` if processing is legitimately slow.
- Decrease `max.poll.records` so each poll returns less work.
- Use `CooperativeStickyAssignor` to reduce rebalance impact.

### Large Messages
Kafka's default max message size is 1 MB. Sending 10 MB payloads requires tuning
`message.max.bytes` on the broker, `max.request.size` on the producer, and
`max.partition.fetch.bytes` on the consumer. Instead, store large payloads in S3/blob storage
and send a reference (URL/key) through Kafka.

### Not Keying Messages
Without a key, messages are distributed across partitions with no ordering guarantee. If you
later need "all events for user X in order," you are stuck. Always define a meaningful key.

### Treating Kafka Like a Database
Kafka is not a database. Do not query it by arbitrary fields. Do not expect fast random lookups.
If you need that, consume into a database and query there.

---

## When NOT to Use Kafka

- **Simple task queues**: If you need "process this job exactly once and forget it," RabbitMQ
  or Redis Streams are simpler and have less operational overhead.
- **Low message volume**: If you handle a few hundred messages per day, Kafka's operational
  complexity is not justified. A simple database-backed queue or SQS will do.
- **Request-reply patterns**: Kafka is designed for async event streaming, not synchronous
  request-response. It can be forced into this pattern, but it is awkward.
- **Very small teams with no Kafka expertise**: Kafka requires understanding of partitions,
  consumer groups, offsets, replication, and retention to operate well. If your team is 3
  people and nobody has run Kafka before, consider a managed service (Confluent Cloud, AWS
  MSK) or a simpler alternative.
- **Strict message ordering across all messages**: Kafka only orders within a partition. If
  you need total global ordering, you need a single partition, which limits throughput to one
  consumer. This is rarely the right tradeoff.

---

## Summary Cheat Sheet

| Decision                  | Recommendation                                       |
|---------------------------|------------------------------------------------------|
| Acks                      | `all` (always)                                       |
| Idempotent producer       | `true` (always)                                      |
| Compression               | `lz4` or `zstd`                                     |
| Replication factor        | 3 in prod                                            |
| `min.insync.replicas`     | 2 in prod                                            |
| Assignment strategy       | `CooperativeStickyAssignor`                          |
| Offset commit             | Manual after processing                              |
| Error handling            | Retry with backoff, then dead letter topic           |
| Partition count           | 6-12 to start, scale up based on throughput needs    |
| Message key               | Always set when ordering matters                     |
| Large payloads            | Store in blob storage, send reference through Kafka  |
| Serialization             | JSON for simplicity, Avro for schema evolution       |
