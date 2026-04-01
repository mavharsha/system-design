# Dead Letter Queues (DLQ)

## What Is a Dead Letter Queue?

A Dead Letter Queue is a holding area for messages that a system could not process successfully. Think of it as the "undeliverable mail" bin at a post office. Instead of losing a message or letting it block everything behind it, you move it aside so the main pipeline keeps flowing and an engineer can inspect the failure later.

Without a DLQ, you face two bad options:
1. **Drop the message** -- you lose data silently.
2. **Retry forever** -- one bad message blocks all subsequent messages (head-of-line blocking).

A DLQ gives you a third option: acknowledge the message failed, park it somewhere safe, alert someone, and keep processing the rest.

---

## When Messages End Up in a DLQ

### 1. Poison Pills
A message that crashes or confuses your consumer every single time it is delivered. No amount of retrying will fix it.

```
// A consumer expects JSON but receives this:
"<xml>not what you expected</xml>"
```

### 2. Schema Mismatches
A producer upgrades and starts sending v2 events while the consumer still only understands v1. The consumer deserializes the message, hits an unknown field or missing required field, and fails.

### 3. Transient Processing Failures That Exhaust Retries
The consumer calls a downstream service that is down. You retry 3 times with backoff, but it never recovers in time. After exhausting retries, the message goes to the DLQ.

### 4. TTL (Time-To-Live) Expiry
Some brokers let you set a per-message or per-queue TTL. If a message sits unprocessed longer than the TTL, the broker moves it to the DLQ automatically. This is common in RabbitMQ.

---

## DLQ in RabbitMQ

RabbitMQ has first-class support through **dead letter exchanges (DLX)**. When a message is rejected, expires, or exceeds a queue's max length, RabbitMQ routes it to the configured DLX.

### Setup

```java
// Spring AMQP configuration
@Configuration
public class RabbitDLQConfig {

    @Bean
    public DirectExchange deadLetterExchange() {
        return new DirectExchange("dlx.orders");
    }

    @Bean
    public Queue deadLetterQueue() {
        return QueueBuilder.durable("orders.dlq").build();
    }

    @Bean
    public Binding dlqBinding() {
        return BindingBuilder.bind(deadLetterQueue())
                .to(deadLetterExchange())
                .with("orders.failed");
    }

    @Bean
    public Queue mainQueue() {
        return QueueBuilder.durable("orders.main")
                .withArgument("x-dead-letter-exchange", "dlx.orders")
                .withArgument("x-dead-letter-routing-key", "orders.failed")
                .withArgument("x-message-ttl", 60000)       // messages expire after 60s
                .withArgument("x-max-length", 10000)        // overflow goes to DLX too
                .build();
    }
}
```

### What Triggers Dead Lettering

| Trigger | How It Happens |
|---|---|
| `basic.reject` or `basic.nack` with `requeue=False` | Consumer explicitly rejects |
| TTL expiry | Message sits in queue past `x-message-ttl` |
| Queue length exceeded | Queue hits `x-max-length` or `x-max-length-bytes` |

### Consumer That Rejects to DLQ

```java
@RabbitListener(queues = "orders.main", ackMode = "MANUAL")
public void handleOrder(Message message, Channel channel) throws IOException {
    long deliveryTag = message.getMessageProperties().getDeliveryTag();
    try {
        Order order = objectMapper.readValue(message.getBody(), Order.class);
        processOrder(order);
        channel.basicAck(deliveryTag, false);
    } catch (Exception e) {
        log.error("Failed to process order: {}", e.getMessage());
        // requeue=false sends it to the DLX
        channel.basicNack(deliveryTag, false, false);
    }
}
```

RabbitMQ attaches `x-death` headers to dead-lettered messages, telling you the original queue, reason, and how many times it was dead-lettered. This metadata is invaluable for debugging.

---

## DLQ in Kafka

Kafka does **not** have a native DLQ mechanism. You build it yourself. The standard pattern is to publish failed messages to a separate "error topic."

### Basic Pattern

```java
@KafkaListener(topics = "orders")
public void consume(ConsumerRecord<String, String> record) {
    try {
        Order order = objectMapper.readValue(record.value(), Order.class);
        orderService.process(order);
    } catch (Exception e) {
        log.error("Processing failed for offset {}: {}", record.offset(), e.getMessage());
        // Send to DLQ topic
        kafkaTemplate.send("orders.dlq", record.key(), record.value());
    }
}
```

### Retry Topics (Graduated Backoff)

A common Kafka pattern uses multiple retry topics with increasing delays before the message finally lands in the DLQ.

```
orders  -->  orders.retry-1  -->  orders.retry-2  -->  orders.dlq
              (1 min delay)       (10 min delay)       (permanent)
```

Spring Kafka has built-in support for this:

```java
@Configuration
public class KafkaRetryConfig {

    @Bean
    public RetryTopicConfiguration retryTopicConfig(KafkaTemplate<String, String> template) {
        return RetryTopicConfigurationBuilder
                .newInstance()
                .maxRetryAttempts(3)
                .fixedBackOff(5000)              // 5s between retries
                .retryTopicSuffix(".retry")
                .dltSuffix(".dlq")
                .create(template);
    }
}
```

### Key Difference from RabbitMQ

In Kafka, the consumer must handle DLQ routing in application code. The broker will not do it for you. This gives you more control (you decide exactly which exceptions are retryable) but more responsibility.

---

## DLQ in AWS SQS

SQS has the simplest DLQ setup of the three. You configure a **redrive policy** on the source queue.

### Setup via AWS CLI

```bash
# Create the DLQ
aws sqs create-queue --queue-name orders-dlq

# Create the main queue with a redrive policy
aws sqs create-queue \
  --queue-name orders-main \
  --attributes '{
    "RedrivePolicy": "{\"deadLetterTargetArn\":\"arn:aws:sqs:us-east-1:123456789:orders-dlq\",\"maxReceiveCount\":\"3\"}"
  }'
```

### How It Works

1. A consumer receives a message from `orders-main`.
2. If the consumer does not delete the message (i.e., processing failed), the message becomes visible again after the visibility timeout.
3. SQS tracks the `ApproximateReceiveCount` for each message.
4. When `ApproximateReceiveCount` exceeds `maxReceiveCount`, SQS moves the message to `orders-dlq`.

### CloudFormation

```yaml
OrdersDLQ:
  Type: AWS::SQS::Queue
  Properties:
    QueueName: orders-dlq
    MessageRetentionPeriod: 1209600  # 14 days (max)

OrdersQueue:
  Type: AWS::SQS::Queue
  Properties:
    QueueName: orders-main
    VisibilityTimeout: 30
    RedrivePolicy:
      deadLetterTargetArn: !GetAtt OrdersDLQ.Arn
      maxReceiveCount: 3
```

SQS also supports **redrive allow policies** on the DLQ side, letting you control which source queues are permitted to use it as a DLQ.

---

## Retry Strategies Before the DLQ

The goal is to avoid the DLQ when the failure is transient. Only permanent or repeatedly-failing messages should end up there.

### Exponential Backoff

```java
private static final int MAX_RETRIES = 4;

public boolean processWithRetry(String message) {
    for (int attempt = 0; attempt < MAX_RETRIES; attempt++) {
        try {
            process(message);
            return true;
        } catch (TransientException e) {
            long wait = (long) Math.pow(2, attempt) * 1000 + ThreadLocalRandom.current().nextLong(1000);
            log.info("Retry {}, waiting {}ms", attempt + 1, wait);
            Thread.sleep(wait);
        }
    }
    // All retries exhausted -- send to DLQ
    sendToDlq(message);
    return false;
}
```

### Classify Your Errors

Not every error deserves a retry. A good rule of thumb:

| Error Type | Action |
|---|---|
| Deserialization failure | Send to DLQ immediately (retrying won't help) |
| HTTP 500 from downstream | Retry with backoff |
| HTTP 400 from downstream | Send to DLQ immediately (bad data) |
| Database connection timeout | Retry with backoff |
| Missing required field | Send to DLQ immediately |

---

## Monitoring and Alerting

A DLQ that nobody watches is just a slower way to lose data. You need:

### Metrics to Track
- **DLQ depth** -- number of messages sitting in the DLQ right now.
- **DLQ ingress rate** -- how fast messages are arriving. A spike means something broke.
- **Age of oldest message** -- messages sitting for days are a sign nobody is looking.

### AWS CloudWatch Alarm Example

```yaml
DLQAlarm:
  Type: AWS::CloudWatch::Alarm
  Properties:
    AlarmName: orders-dlq-not-empty
    MetricName: ApproximateNumberOfMessagesVisible
    Namespace: AWS/SQS
    Dimensions:
      - Name: QueueName
        Value: orders-dlq
    Statistic: Sum
    Period: 300
    EvaluationPeriods: 1
    Threshold: 0
    ComparisonOperator: GreaterThanThreshold
    AlarmActions:
      - !Ref OpsAlertSNSTopic
```

### Grafana / Prometheus for Kafka

```yaml
# Prometheus alert rule
groups:
  - name: kafka-dlq
    rules:
      - alert: DLQMessagesAccumulating
        expr: kafka_consumergroup_lag{topic="orders.dlq"} > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Messages in orders DLQ"
```

The alert should page someone or post to a Slack channel. Treat DLQ messages as incidents that need human attention.

---

## Reprocessing Messages from DLQ

Once you fix the root cause, you need to replay the failed messages.

### AWS SQS: Redrive to Source

AWS added a native "start message move task" feature:

```bash
aws sqs start-message-move-task \
  --source-arn arn:aws:sqs:us-east-1:123456789:orders-dlq \
  --destination-arn arn:aws:sqs:us-east-1:123456789:orders-main \
  --max-number-of-messages-per-second 10
```

The rate limit prevents the replayed messages from overwhelming your consumer.

### Kafka: Replay with a Consumer Script

```java
Properties consumerProps = new Properties();
consumerProps.put("bootstrap.servers", "localhost:9092");
consumerProps.put("group.id", "dlq-replay");
consumerProps.put("auto.offset.reset", "earliest");
consumerProps.put("key.deserializer", StringDeserializer.class.getName());
consumerProps.put("value.deserializer", StringDeserializer.class.getName());

KafkaConsumer<String, String> consumer = new KafkaConsumer<>(consumerProps);
KafkaProducer<String, String> producer = new KafkaProducer<>(producerProps);
consumer.subscribe(List.of("orders.dlq"));

while (true) {
    ConsumerRecords<String, String> records = consumer.poll(Duration.ofMillis(1000));
    for (ConsumerRecord<String, String> msg : records) {
        // Optionally transform or filter before replaying
        producer.send(new ProducerRecord<>("orders", msg.key(), msg.value()));
    }
    producer.flush();
    consumer.commitSync();
}
```

### RabbitMQ: Shovel Plugin

```bash
rabbitmqctl set_parameter shovel replay-orders \
  '{"src-uri": "amqp://", "src-queue": "orders.dlq",
    "dest-uri": "amqp://", "dest-queue": "orders.main"}'
```

### Best Practices for Replay
- **Fix the bug first.** Replaying into the same broken consumer just fills the DLQ again.
- **Replay at a controlled rate.** Do not dump thousands of messages back at once.
- **Log replayed messages.** You want an audit trail showing which messages were reprocessed.
- **Test with one message first.** Pull a single message, process it manually, verify success, then replay the batch.

---

## Real-World Example: E-Commerce Order Pipeline

Consider an order processing system where a customer places an order and it flows through several services.

```
[Web App] --> [orders topic] --> [Order Service] --> [Payment Service]
                                       |
                                       v (on failure)
                                 [orders.dlq topic]
```

### The Happy Path

```json
{
  "order_id": "ORD-9182",
  "customer_id": "C-441",
  "items": [
    {"sku": "WIDGET-01", "qty": 2, "price_cents": 1500}
  ],
  "shipping_address": {
    "street": "123 Main St",
    "city": "Portland",
    "state": "OR",
    "zip": "97201"
  }
}
```

This message deserializes fine, passes validation, and goes to payment.

### The Malformed Order (Poison Pill)

A frontend bug sends this:

```json
{
  "order_id": "ORD-9183",
  "customer_id": "C-442",
  "items": "WIDGET-01",
  "shipping_address": null
}
```

Problems: `items` is a string instead of an array, `shipping_address` is null.

### What Happens

```java
@KafkaListener(topics = "orders")
public void handleOrder(ConsumerRecord<String, String> record) {
    try {
        Order order = objectMapper.readValue(record.value(), Order.class);
        validate(order);              // throws if address is null
        paymentService.charge(order);
        inventoryService.reserve(order);
    } catch (JsonProcessingException e) {
        // Schema/deserialization problem -- never retryable
        log.error("Malformed order at offset {}: {}", record.offset(), e.getMessage());
        dlqProducer.send("orders.dlq", record.key(), record.value());
    } catch (PaymentException e) {
        // Downstream failure -- might be transient, retry
        throw e;  // let the retry mechanism handle it
    }
}
```

The malformed order goes straight to `orders.dlq`. The payment failure gets retried. Neither one blocks the hundreds of valid orders behind them.

### Resolution

1. The monitoring alert fires: "1 message in orders.dlq."
2. An engineer inspects the message, sees the malformed `items` field.
3. They trace it to the frontend bug and deploy a fix.
4. They decide whether to manually correct and replay the order or contact the customer.

---

## Key Takeaways

- Every message queue in production needs a DLQ strategy. It is not optional.
- Classify errors as retryable vs. permanent before deciding what to do with a failed message.
- RabbitMQ gives you DLQ for free via dead letter exchanges. Kafka requires you to build it. SQS sits in the middle with a simple redrive policy.
- A DLQ without monitoring is just a graveyard. Alert on it.
- Have a tested replay runbook before you need it, not after.
