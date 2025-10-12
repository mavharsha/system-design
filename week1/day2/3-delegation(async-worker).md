# Delegation (Async Processing)

Offload work from main request/response cycle to background workers

**Philosophy:** If it can be async, make it async

**Benefits:** Faster response, fault tolerance, decoupling

---

## Message Queues vs Message Streams

### 1. Message Queues (Task Queues)

**Examples:** SQS, RabbitMQ, Redis Queue

**How it works:**
- One consumer per message → message deleted after consumption
- FIFO or priority-based
- Fire and forget

**Ordering:**
- **No strict ordering** (unless using FIFO queues like SQS FIFO)
- Multiple consumers → order not guaranteed
- Use when order doesn't matter

**Blog Publishing Example:**
```
POST /publish → Queue task → Return 202 Accepted (fast)
Worker picks up task → Process (HTML, thumbnails, CDN, cache, notifications)
```

**Use cases:** Any of these tasks could be handled—either each message is a different task type (email, image processing, reports, order processing), or a single worker/consumer might process any/all of them depending on your implementation.

---

### 2. Message Streams (Event Streams)

**Examples:** Kafka, Kinesis, Pulsar

**How it works:**
- Multiple consumers read same message (consumer groups)
- Messages persist (configurable retention)
- Can replay events

**Ordering:**
- **Guaranteed within a partition, NOT across partitions**
- Messages with same key → same partition → ordered
- Key-based partitioning ensures related events stay ordered

**Example:**
```
Topic: "user-events" (3 partitions)
├─ Partition 0: user123 events → msg1, msg2, msg3 (ordered)
├─ Partition 1: user456 events → msg1, msg2, msg3 (ordered)
└─ Partition 2: user789 events → msg1, msg2, msg3 (ordered)

All events for user123 go to partition 0 → strict ordering
```

**User Activity Example:**
```
User purchases → Event stream
├─ Analytics group: Calculate metrics
├─ Recommendations group: Update profile
└─ Notifications group: Send confirmation

Same message consumed by all groups independently
```

**When ordering matters:** User state changes, financial transactions, audit logs

**Parallelism:**
- **Max consumers = number of partitions** (per consumer group)
- 3 partitions → max 3 parallel consumers
- More consumers than partitions → some sit idle
- To scale: add more partitions

**Example:**
```
Topic with 3 partitions:
├─ Consumer 1 → Partition 0
├─ Consumer 2 → Partition 1
└─ Consumer 3 → Partition 2

Add Consumer 4? → sits idle (no partition to assign)
Want more parallelism? → increase partitions
```

**Best practice:** 
- Use userId (or relevant key) as partition key to maintain order per user
- Set partitions based on expected parallelism needs
- Can't reduce partitions later (only increase)

**Use cases:** Event sourcing, real-time analytics, audit logs, CDC (change data capture)

---

## Comparison

| Feature | Message Queue | Message Stream |
|---------|---------------|----------------|
| Consumption | One consumer, then deleted | Multiple consumer groups |
| Persistence | Short-lived | Long-lived (configurable) |
| Ordering | No strict order | Order guaranteed per partition |
| Use Case | Task distribution | Event broadcasting |
| Examples | SQS, RabbitMQ | Kafka, Kinesis |

---

## Implementation Examples

**Message Queue (SQS):**
```java
// Producer
SendMessageRequest request = new SendMessageRequest()
    .withQueueUrl(queueUrl)
    .withMessageBody("{\"task\":\"send_email\",\"to\":\"user@example.com\"}");
sqs.sendMessage(request);

// Consumer (worker)
List<Message> messages = sqs.receiveMessage(queueUrl).getMessages();
for (Message msg : messages) {
    processTask(msg.getBody());
    sqs.deleteMessage(queueUrl, msg.getReceiptHandle()); // Delete after processing
}
```

**Message Stream (Kafka):**
```java
// Producer - userId as partition key ensures ordering per user
ProducerRecord<String, String> record = 
    new ProducerRecord<>("user-events", userId, eventJson); // key = userId
producer.send(record); // Same userId → same partition → ordered

// Consumer (Analytics group)
consumer.subscribe(Arrays.asList("user-events"));
while (true) {
    ConsumerRecords<String, String> records = consumer.poll(Duration.ofMillis(100));
    for (ConsumerRecord<String, String> record : records) {
        updateAnalytics(record.value()); // Message stays in stream
    }
}
```