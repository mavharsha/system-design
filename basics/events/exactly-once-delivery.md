# Exactly-Once Delivery in Distributed Systems

## The Three Delivery Semantics

When one service sends a message to another, there are three guarantees you can aim for:

### At-Most-Once

Fire and forget. You send the message and move on. If it gets lost, tough luck.

```
Producer ---> [message] ---> Broker ---> Consumer
                 ^
                 |
            might be lost, no retry
```

**Where this is fine:** logging, metrics, analytics events. If you lose a page-view event,
nobody notices.

**How it works in practice:** the producer sends a message without waiting for an
acknowledgment, or it receives a failure and does not retry.

### At-Least-Once

You keep retrying until the consumer acknowledges. The message definitely arrives, but it
might arrive more than once.

```
Producer ---> [message] ---> Broker ---> Consumer
   |                                        |
   +-------- retry if no ACK <-------------+
```

**Where this causes problems:** charging a credit card, decrementing inventory, sending an
email. Getting any of these twice is bad.

**This is the default for most systems.** RabbitMQ, SQS, Kafka (without extra config) -- they
all give you at-least-once out of the box.

### Exactly-Once

The message is processed once and only once. This is what everyone wants. The catch is that
achieving it requires careful coordination between the producer, broker, and consumer.

---

## Why Exactly-Once Is Hard

### The Two Generals Problem

Imagine two generals on opposite sides of a valley need to coordinate an attack. They can only
communicate by sending messengers through the valley, but messengers can be captured.

- General A sends: "Attack at dawn." Did General B get it?
- General B sends back: "Got it." Did General A receive the confirmation?
- General A sends: "I got your confirmation." Did General B get *that*?

This regress never ends. You cannot achieve guaranteed agreement over an unreliable channel
with a finite number of messages. This was proven impossible in 1975 and is the theoretical
foundation for why distributed consensus is hard.

### Network Partitions in Practice

Consider a producer writing to Kafka:

```
1. Producer sends message to broker
2. Broker writes message to disk
3. Broker sends ACK back to producer
4. ACK is lost in the network
5. Producer assumes failure, retries
6. Broker now has TWO copies of the same message
```

The producer did everything right. The broker did everything right. The network broke the
contract. This is not a theoretical edge case -- it happens regularly at scale.

### The Acknowledgment Gap

The fundamental issue is the gap between "the work was done" and "I know the work was done."
A consumer can:

1. Process the message, then commit the offset -- if it crashes between steps 1 and 2, it
   reprocesses on restart (duplicate).
2. Commit the offset, then process -- if it crashes between steps 1 and 2, the message is
   lost (at-most-once).

There is no atomic operation that spans "do the work" and "tell the broker I did the work"
when these are on different machines.

---

## How Kafka Achieves Exactly-Once

Kafka introduced exactly-once semantics (EOS) in version 0.11. It works through three
mechanisms that cooperate together.

### 1. Idempotent Producers

Each producer gets a Producer ID (PID) and attaches a monotonically increasing sequence
number to every message. The broker tracks the last sequence number per PID per partition.

```
Message { PID: 42, SeqNum: 7, partition: 3, payload: "..." }
```

If the broker receives SeqNum 7 twice (because the producer retried after a lost ACK),
it silently discards the duplicate. This solves the producer-side duplication problem.

```java
Properties props = new Properties();
props.put("bootstrap.servers", "localhost:9092");
props.put("enable.idempotence", "true");  // This is the key setting
props.put("acks", "all");
// max.in.flight.requests.per.connection is automatically set to 5
// retries is automatically set to Integer.MAX_VALUE

KafkaProducer<String, String> producer = new KafkaProducer<>(props);
```

Enabling idempotence is a one-line config change. There is minimal performance overhead.

### 2. Transactional API

Idempotent producers prevent duplicates within a single session, but what if you need to
write to multiple partitions atomically? Or read from one topic, process, and write to
another topic as a single unit?

That is what transactions are for.

```java
Properties props = new Properties();
props.put("bootstrap.servers", "localhost:9092");
props.put("enable.idempotence", "true");
props.put("transactional.id", "payment-processor-1");  // Must be stable across restarts

KafkaProducer<String, String> producer = new KafkaProducer<>(props);
producer.initTransactions();

try {
    producer.beginTransaction();

    // Read from input topic, process, write to output topic -- all atomic
    producer.send(new ProducerRecord<>("output-topic", key, processedValue));
    producer.send(new ProducerRecord<>("audit-topic", key, auditRecord));

    // Commit consumer offsets as part of the same transaction
    producer.sendOffsetsToTransaction(offsets, consumerGroupId);

    producer.commitTransaction();
} catch (Exception e) {
    producer.abortTransaction();
}
```

Under the hood, Kafka uses a transaction coordinator and a two-phase commit protocol.
The `transactional.id` survives producer restarts -- if a new producer instance starts
with the same transactional ID, the coordinator fences the old one (zombie fencing).

### 3. Consumer Offset Management

On the consumer side, you set `isolation.level=read_committed` so consumers only see
messages from committed transactions:

```java
props.put("isolation.level", "read_committed");
```

Without this, consumers see all messages including ones from aborted transactions.

### The Scope of Kafka's Exactly-Once

This is critical to understand: **Kafka's exactly-once only covers Kafka-to-Kafka
processing.** The moment you involve an external system (a database, an HTTP call, a
third-party API), you are back to needing application-level guarantees. Kafka cannot
make your database write and your offset commit atomic -- those are different systems.

---

## Idempotency: The Practical Alternative

Since true end-to-end exactly-once is impractical in most real systems, the standard approach
is: **use at-least-once delivery and make your consumers idempotent.**

An operation is idempotent if performing it multiple times has the same effect as performing it
once. `SET balance = 500` is idempotent. `SET balance = balance - 50` is not.

### Idempotency Keys

The producer assigns each message a unique ID. The consumer checks whether it has already
processed that ID before doing any work.

```java
@Transactional
public String processPayment(PaymentEvent event) {
    String idempotencyKey = event.getIdempotencyKey();

    // Check if we already processed this
    Optional<ProcessedEvent> existing = processedEventRepository
            .findByIdempotencyKey(idempotencyKey);

    if (existing.isPresent()) {
        return existing.get().getStatus();  // Already handled, return cached result
    }

    // Process the payment
    PaymentResult result = paymentGateway.charge(event.getAmount(), event.getCardToken());

    // Record that we processed it (in the same DB transaction as any state change)
    processedEventRepository.save(new ProcessedEvent(idempotencyKey, result.getStatus()));

    return result.getStatus();
}
```

The check-and-insert must be atomic. If they are not in the same transaction, you have a
race condition where two workers both see "not processed" and both proceed.

### Deduplication at the Database Level

You can also use unique constraints as a deduplication mechanism:

```sql
CREATE TABLE inventory_updates (
    event_id UUID PRIMARY KEY,   -- natural deduplication key
    product_id BIGINT NOT NULL,
    quantity_change INT NOT NULL,
    applied_at TIMESTAMP DEFAULT NOW()
);

-- This INSERT will fail (silently with ON CONFLICT) if the event was already applied
INSERT INTO inventory_updates (event_id, product_id, quantity_change)
VALUES ($1, $2, $3)
ON CONFLICT (event_id) DO NOTHING;
```

The database constraint does the deduplication for you. No race conditions.

---

## Real-World Examples

### Payment Processing: The Double Charge Problem

A user clicks "Pay" on a checkout page. The request hits your server, which charges the
card and then tries to record the charge. If the server crashes after the charge but
before recording it, a retry will charge the card again.

**Solution:**

```
1. Client generates an idempotency key (UUID) before the first request
2. Client sends: POST /payments { idempotency_key: "abc-123", amount: 99.00 }
3. Server checks: have I seen "abc-123" before?
   - Yes: return the cached result
   - No: process the payment, store the result keyed by "abc-123", return it
4. If the client retries (network timeout, crash), step 3 catches the duplicate
```

Stripe does exactly this. Their API accepts an `Idempotency-Key` header. If you send the
same key within 24 hours, you get the cached response back.

### Inventory Updates

An order service publishes an "order placed" event. The inventory service consumes it and
decrements stock. If the event is delivered twice, you decrement twice and oversell.

**Solution:** store the event ID alongside the inventory change in a single transaction.

```java
@Transactional
public void handleOrderPlaced(OrderPlacedEvent event) {
    // Attempt to record this event -- fails if already processed
    try {
        jdbcTemplate.update(
            "INSERT INTO processed_order_events (event_id) VALUES (?)",
            event.getId()
        );
    } catch (DuplicateKeyException e) {
        return;  // Already processed, skip
    }

    // Safe to apply the change
    for (OrderItem item : event.getItems()) {
        jdbcTemplate.update(
            "UPDATE products SET stock = stock - ? WHERE id = ?",
            item.getQuantity(), item.getProductId()
        );
    }
}
```

The unique constraint on `event_id` and the inventory update happen in the same database
transaction. Either both succeed or neither does.

---

## Patterns

### The Outbox Pattern

Problem: you need to update your database AND publish an event, and these need to be
consistent. If you update the DB and then publish, the publish might fail. If you publish
first, the DB write might fail.

Solution: write the event to an "outbox" table in the same database transaction as your
state change. A separate process reads the outbox and publishes to the message broker.

```
Service                          Database                    Message Broker
   |                                |                              |
   |-- BEGIN TX ------------------>|                              |
   |-- UPDATE orders SET ... ----->|                              |
   |-- INSERT INTO outbox -------->|                              |
   |-- COMMIT -------------------->|                              |
   |                                |                              |
   |        Outbox Publisher (polls or uses CDC)                   |
   |                                |-- read outbox entries ------>|
   |                                |          publish ----------->|
   |                                |-- mark as published -------->|
```

```sql
-- In the same transaction as your business logic
BEGIN;

UPDATE orders SET status = 'confirmed' WHERE id = 42;

INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)
VALUES ('Order', 42, 'OrderConfirmed', '{"order_id": 42, "total": 99.00}');

COMMIT;
```

The outbox publisher can safely retry because the broker-side consumer is idempotent
(using the outbox entry ID as the idempotency key).

### Transactional Outbox with Change Data Capture (CDC)

Polling the outbox table works but adds load to your database. A better approach is
Change Data Capture: a tool like Debezium tails the database's transaction log (WAL in
Postgres, binlog in MySQL) and publishes new outbox entries to Kafka automatically.

```
Database WAL/Binlog --> Debezium --> Kafka --> Consumers
```

Advantages over polling:
- Near real-time (no polling interval)
- No extra queries on your database
- Guaranteed ordering (events appear in commit order)
- The outbox table can be periodically truncated since Debezium reads the log, not the table

This is the approach used at LinkedIn, Uber, and many other companies running large
event-driven architectures.

---

## Common Pitfalls and Misconceptions

### "Exactly-once is impossible"

You will hear this a lot. It is both true and misleading.

**True in the theoretical sense:** the Two Generals Problem proves you cannot guarantee
exactly-once delivery over an unreliable network with finite messages. No protocol can.

**Misleading in the practical sense:** you can achieve exactly-once *processing* (as
opposed to *delivery*) by combining at-least-once delivery with idempotent consumers.
The message might be delivered multiple times, but the side effects happen exactly once.
For all practical purposes, this is what people mean when they say "exactly-once."

Kafka's documentation calls its feature "exactly-once semantics" and they are technically
correct within the Kafka-to-Kafka boundary. The debate is mostly about terminology.

### Pitfall: Idempotency key expiration

If you expire idempotency keys too soon (say, after 1 hour), a very delayed retry can
slip through. Stripe keeps keys for 24 hours. Think about your system's maximum retry
window and add margin.

### Pitfall: Side effects outside the transaction

```java
@Transactional
public void processOrder(OrderEvent event) {
    jdbcTemplate.update("INSERT INTO processed (event_id) VALUES (?)", event.getId());
    jdbcTemplate.update("UPDATE inventory SET stock = stock - 1 WHERE id = ?", event.getProductId());
}

// THIS IS OUTSIDE THE TRANSACTION
// If this fails and the event is retried, the DB insert will block the retry
// but this email was never sent
public void afterProcessOrder(OrderEvent event) {
    processOrder(event);
    emailService.sendConfirmation(event.getCustomerEmail()); // not transactional!
}
```

Side effects that cannot be made part of the database transaction (sending emails, calling
APIs) need their own idempotency or a separate outbox/queue mechanism.

### Pitfall: Confusing idempotent with safe-to-retry

An operation can be idempotent but still have timing-sensitive behavior. Debiting an
account is idempotent if you use an idempotency key, but if the key is not propagated
through every retry path in your system, you still get double charges.

Make sure the idempotency key is generated at the **source** (the client or the producing
service) and carried through every hop.

### Pitfall: Non-deterministic processing

If your consumer produces different output for the same input (because it uses wall-clock
time, random numbers, or external state that changed), replaying a message produces a
different result. Idempotency keys catch this, but if you are relying on "natural"
idempotency (like `SET x = 5`), non-determinism breaks you.

---

## Summary

| Approach | Guarantee | Complexity | Use When |
|----------|-----------|------------|----------|
| At-most-once | Might lose messages | Low | Metrics, logging |
| At-least-once | Might duplicate messages | Medium | Default for most systems |
| At-least-once + idempotent consumers | Effectively exactly-once | Medium-High | Payments, inventory, anything with side effects |
| Kafka EOS (transactional API) | Exactly-once within Kafka | High | Stream processing pipelines |
| Outbox + CDC | Consistent DB + events | High | Event-driven architectures needing strong consistency |

The practical takeaway: **design for at-least-once delivery and make your consumers
idempotent.** This is simpler, more portable, and more resilient than trying to achieve
true exactly-once delivery across system boundaries.
