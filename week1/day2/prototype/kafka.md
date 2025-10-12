# Kafka Listener - Spring Boot Notes

## Dependencies (pom.xml)
```xml
<dependency>
    <groupId>org.springframework.kafka</groupId>
    <artifactId>spring-kafka</artifactId>
</dependency>
```

## application.yml
```yaml
spring:
  kafka:
    bootstrap-servers: localhost:9092
    consumer:
      group-id: default-group
      auto-offset-reset: earliest
      key-deserializer: org.apache.kafka.common.serialization.StringDeserializer
      value-deserializer: org.apache.kafka.common.serialization.StringDeserializer
      enable-auto-commit: true
```

## Kafka Config
```java
// Configuration class for setting up multiple Kafka consumers with different settings
@Configuration
@EnableKafka
public class KafkaConfig {

    @Value("${spring.kafka.bootstrap-servers}")
    private String bootstrapServers;

    // === ORDER CONSUMER SETUP (3 partitions) ===
    // Creates consumer factory for order-events topic with "order-group" consumer group
    @Bean
    public ConsumerFactory<String, String> orderConsumerFactory() {
        Map<String, Object> props = new HashMap<>();
        props.put(ConsumerConfig.BOOTSTRAP_SERVERS_CONFIG, bootstrapServers);
        props.put(ConsumerConfig.GROUP_ID_CONFIG, "order-group");
        props.put(ConsumerConfig.KEY_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class);
        props.put(ConsumerConfig.VALUE_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class);
        return new DefaultKafkaConsumerFactory<>(props);
    }

    // Listener container factory with concurrency=3 to match 3 partitions
    // Each partition will be handled by a separate consumer thread
    @Bean
    public ConcurrentKafkaListenerContainerFactory<String, String> orderListenerFactory() {
        ConcurrentKafkaListenerContainerFactory<String, String> factory = 
            new ConcurrentKafkaListenerContainerFactory<>();
        factory.setConsumerFactory(orderConsumerFactory());
        factory.setConcurrency(3); // 3 threads for 3 partitions
        return factory;
    }

    // === USER CONSUMER SETUP (5 partitions, manual commit) ===
    // Creates consumer with manual acknowledgment (enable-auto-commit=false)
    @Bean
    public ConsumerFactory<String, String> userConsumerFactory() {
        Map<String, Object> props = new HashMap<>();
        props.put(ConsumerConfig.BOOTSTRAP_SERVERS_CONFIG, bootstrapServers);
        props.put(ConsumerConfig.GROUP_ID_CONFIG, "user-group");
        props.put(ConsumerConfig.KEY_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class);
        props.put(ConsumerConfig.VALUE_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class);
        props.put(ConsumerConfig.ENABLE_AUTO_COMMIT_CONFIG, false); // manual commit for reliability
        return new DefaultKafkaConsumerFactory<>(props);
    }

    // Listener container with concurrency=5 to match 5 partitions
    @Bean
    public ConcurrentKafkaListenerContainerFactory<String, String> userListenerFactory() {
        ConcurrentKafkaListenerContainerFactory<String, String> factory = 
            new ConcurrentKafkaListenerContainerFactory<>();
        factory.setConsumerFactory(userConsumerFactory());
        factory.setConcurrency(5); // 5 threads for 5 partitions
        return factory;
    }

    // === NOTIFICATION CONSUMER SETUP (2 partitions) ===
    // Simple auto-commit consumer for notifications
    @Bean
    public ConsumerFactory<String, String> notifConsumerFactory() {
        Map<String, Object> props = new HashMap<>();
        props.put(ConsumerConfig.BOOTSTRAP_SERVERS_CONFIG, bootstrapServers);
        props.put(ConsumerConfig.GROUP_ID_CONFIG, "notif-group");
        props.put(ConsumerConfig.KEY_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class);
        props.put(ConsumerConfig.VALUE_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class);
        return new DefaultKafkaConsumerFactory<>(props);
    }

    // Listener container with concurrency=2 to match 2 partitions
    @Bean
    public ConcurrentKafkaListenerContainerFactory<String, String> notifListenerFactory() {
        ConcurrentKafkaListenerContainerFactory<String, String> factory = 
            new ConcurrentKafkaListenerContainerFactory<>();
        factory.setConsumerFactory(notifConsumerFactory());
        factory.setConcurrency(2); // 2 threads for 2 partitions
        return factory;
    }
}
```

## Kafka Listeners
```java
// Service containing all Kafka listener methods for different topics
@Slf4j
@Service
public class KafkaListeners {

    // === Listener 1: ORDER EVENTS (3 partitions, auto-commit) ===
    // Listens to order-events topic, extracts partition, offset, key metadata
    @KafkaListener(topics = "order-events", groupId = "order-group", 
                   containerFactory = "orderListenerFactory")
    public void listenOrders(@Payload String message,
                             @Header(KafkaHeaders.RECEIVED_PARTITION) int partition,
                             @Header(KafkaHeaders.OFFSET) long offset,
                             @Header(KafkaHeaders.RECEIVED_KEY) String key) {
        log.info("Order - P:{} O:{} K:{} M:{}", partition, offset, key, message);
        // Process order event here
    }

    // === Listener 2: USER EVENTS (5 partitions, manual commit) ===
    // Uses ConsumerRecord for full control, manually acknowledges after processing
    // Guarantees at-least-once delivery (won't commit if processing fails)
    @KafkaListener(topics = "user-events", groupId = "user-group",
                   containerFactory = "userListenerFactory")
    public void listenUsers(ConsumerRecord<String, String> record, 
                            Acknowledgment ack) {
        log.info("User - P:{} O:{} M:{}", record.partition(), record.offset(), record.value());
        try {
            // Process user event
            if (ack != null) ack.acknowledge(); // Commit offset only on success
        } catch (Exception e) {
            log.error("Error: {}", e.getMessage());
            // Don't acknowledge - message will be reprocessed
        }
    }

    // === Listener 3: NOTIFICATION EVENTS (2 partitions, auto-commit) ===
    // Simple listener with just message payload, no metadata needed
    @KafkaListener(topics = "notif-events", groupId = "notif-group",
                   containerFactory = "notifListenerFactory")
    public void listenNotifs(@Payload String message) {
        log.info("Notif - M:{}", message);
        // Process notification
    }

    // === Listener 4: MULTIPLE TOPICS (analytics & metrics) ===
    // Single listener handles multiple topics, uses topic name to route logic
    @KafkaListener(topics = {"analytics", "metrics"}, groupId = "analytics-group")
    public void listenAnalytics(@Payload String message,
                                @Header(KafkaHeaders.RECEIVED_TOPIC) String topic) {
        log.info("Analytics - T:{} M:{}", topic, message);
        // Route logic based on topic name
    }

    // === Listener 5: SPECIFIC PARTITIONS ONLY ===
    // Listens ONLY to partitions 0 and 1 of order-events (not all 3)
    // Useful when you want dedicated processing for certain partitions
    @KafkaListener(topicPartitions = @TopicPartition(topic = "order-events", 
                   partitions = {"0", "1"}), groupId = "order-specific-group")
    public void listenSpecific(@Payload String message) {
        log.info("Specific - M:{}", message);
        // Process only messages from partitions 0 and 1
    }

    // === Listener 6: BATCH PROCESSING ===
    // Receives multiple messages at once for efficient bulk processing
    // Reduces overhead compared to processing one message at a time
    @KafkaListener(topics = "batch-events", groupId = "batch-group")
    public void listenBatch(List<String> messages) {
        log.info("Batch - Count:{}", messages.size());
        messages.forEach(msg -> log.debug("Msg: {}", msg));
        // Process all messages in bulk
    }
}
```

---

## Kafka Commands

### Create Topics
```bash
# order-events: 3 partitions
kafka-topics.sh --create --topic order-events --bootstrap-server localhost:9092 \
  --partitions 3 --replication-factor 1

# user-events: 5 partitions
kafka-topics.sh --create --topic user-events --bootstrap-server localhost:9092 \
  --partitions 5 --replication-factor 1

# notif-events: 2 partitions
kafka-topics.sh --create --topic notif-events --bootstrap-server localhost:9092 \
  --partitions 2 --replication-factor 1
```

### List & Describe Topics
```bash
# List all
kafka-topics.sh --list --bootstrap-server localhost:9092

# Describe specific
kafka-topics.sh --describe --topic order-events --bootstrap-server localhost:9092
```

### Consumer Groups
```bash
# List groups
kafka-consumer-groups.sh --list --bootstrap-server localhost:9092

# Describe group (shows lag, offsets)
kafka-consumer-groups.sh --describe --group order-group --bootstrap-server localhost:9092

# Check all groups
kafka-consumer-groups.sh --describe --all-groups --bootstrap-server localhost:9092

# See members
kafka-consumer-groups.sh --describe --group order-group --bootstrap-server localhost:9092 --members
```

### Reset Offsets
```bash
# Reset to earliest
kafka-consumer-groups.sh --reset-offsets --group order-group --topic order-events \
  --to-earliest --bootstrap-server localhost:9092 --execute

# Reset to latest
kafka-consumer-groups.sh --reset-offsets --group order-group --topic order-events \
  --to-latest --bootstrap-server localhost:9092 --execute

# Reset to specific offset
kafka-consumer-groups.sh --reset-offsets --group order-group --topic order-events:0 \
  --to-offset 100 --bootstrap-server localhost:9092 --execute
```

### Produce Messages (testing)
```bash
kafka-console-producer.sh --topic order-events --bootstrap-server localhost:9092 \
  --property "key.separator=:" --property "parse.key=true"
# Then type: order-1:{"orderId": 1, "amount": 100}
```

### Consume Messages (testing)
```bash
# From beginning
kafka-console-consumer.sh --topic order-events --bootstrap-server localhost:9092 \
  --from-beginning --property print.key=true --property print.partition=true

# With consumer group
kafka-console-consumer.sh --topic order-events --bootstrap-server localhost:9092 \
  --group test-group --from-beginning
```

---

## Java Run Commands

### Maven
```bash
# Build
mvn clean install

# Run
mvn spring-boot:run

# Run with profile
mvn spring-boot:run -Dspring-boot.run.profiles=dev

# Run with properties
mvn spring-boot:run -Dspring-boot.run.arguments="--spring.kafka.bootstrap-servers=localhost:9092"
```

---

## Key Config Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `group.id` | Consumer group ID | Required |
| `auto.offset.reset` | earliest/latest/none | latest |
| `enable.auto.commit` | Auto-commit offsets | true |
| `max.poll.records` | Max records per poll | 500 |
| `concurrency` | Consumer threads | 1 |

## Quick Notes
- Match concurrency with partition count for best performance
- Use manual commit for critical data (user-events example)
- One consumer group = load balanced across consumers
- Multiple consumer groups = all get same messages
- Check lag: `kafka-consumer-groups.sh --describe --group <group-id>`


-----
Prompt:
In this notes file, can you please write java kafka listener that would listen to couple of topics.  with different number of partitions. and java commands to run the listers also configuring consumer-group.

Assumptions, springboot with kafka dependency is used. 
This is notes, with code and not complete project

Keep it concise if possible. just as notes and not documentation

Response:
✅ Core code examples without verbose explanations
✅ Essential Kafka commands in quick reference format
✅ Java run commands condensed
✅ Quick notes at the end instead of long documentation
Now it's more like study notes you can quickly scan through rather than full documentation. The code examples cover:
3 different listeners with different partition counts (3, 5, 2)
Manual acknowledgment example
Multiple topics, specific partitions, and batch processing
All the key Kafka commands for topic/consumer group management
-----
