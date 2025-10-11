# Database ACID Properties

## Overview

ACID is a set of properties that guarantee reliable processing of database transactions. The acronym stands for:

- **A**tomicity
- **C**onsistency
- **I**solation
- **D**urability

These properties ensure data integrity and reliability even in the face of errors, power failures, or other unexpected issues.

---

## Quick Summary

### Atomicity
**"All or Nothing"**
- A transaction is treated as a single unit
- Either all operations succeed, or none do
- No partial updates allowed

### Consistency
**"Valid State to Valid State"**
- Database moves from one valid state to another
- All constraints and rules are satisfied
- Data integrity is maintained

### Isolation
**"Transactions Don't Interfere"**
- Concurrent transactions execute independently
- Intermediate states are not visible to other transactions
- Prevents race conditions and conflicts

### Durability
**"Changes Are Permanent"**
- Once committed, changes persist
- Survives system crashes and power failures
- Data is safely stored on non-volatile storage

---

## Why ACID Matters

1. **Data Integrity**: Ensures your data remains accurate and consistent
2. **Reliability**: Guarantees that operations complete successfully or fail safely
3. **Concurrency**: Allows multiple users to access the database simultaneously without conflicts
4. **Recovery**: Enables system recovery after crashes without data loss

---

## ACID vs BASE

### ACID (Traditional RDBMS)
- Strong consistency
- Immediate consistency
- Better for financial systems, inventory management
- Examples: PostgreSQL, MySQL, Oracle

### BASE (NoSQL Systems)
- **B**asically **A**vailable
- **S**oft state
- **E**ventual consistency
- Better for high-scale, distributed systems
- Examples: Cassandra, MongoDB, DynamoDB

---

## Real-World Example: Bank Transfer

Consider transferring $100 from Account A to Account B:

```sql
BEGIN TRANSACTION;
  UPDATE accounts SET balance = balance - 100 WHERE account_id = 'A';
  UPDATE accounts SET balance = balance + 100 WHERE account_id = 'B';
COMMIT;
```

**How ACID Applies:**

- **Atomicity**: Both updates happen or neither happens (no money lost or created)
- **Consistency**: Total money in system remains the same (if A had sufficient balance)
- **Isolation**: Other transactions see either the old state or new state, not an in-between state
- **Durability**: Once committed, the transfer is permanent even if the system crashes

---

## Detailed Documentation

For in-depth understanding of each property, see the dedicated documents:

1. [**Atomicity**](./db-acid-atomicity.md) - Transaction indivisibility and rollback mechanisms
2. [**Consistency**](./db-acid-consistency.md) - Data integrity constraints and validation
3. [**Isolation**](./db-acid-isolation.md) - Concurrency control and isolation levels
4. [**Durability**](./db-acid-durability.md) - Persistence mechanisms and recovery

---

## Trade-offs

### Performance vs Guarantees
- Stricter ACID compliance = Lower performance
- Relaxed ACID = Higher throughput but weaker guarantees

### CAP Theorem
In distributed systems, you can only guarantee 2 of 3:
- **C**onsistency
- **A**vailability
- **P**artition tolerance

Most distributed databases choose AP or CP, sacrificing some ACID properties for scalability.

---

## Common Questions

**Q: Do all databases support ACID?**
A: No. Traditional RDBMS (PostgreSQL, MySQL) fully support ACID. NoSQL databases often trade ACID for scalability.

**Q: Can ACID impact performance?**
A: Yes. Strict ACID compliance requires additional overhead (locks, logs, etc.) which can reduce throughput.

**Q: Is ACID always necessary?**
A: No. For some applications (social media feeds, analytics), eventual consistency is acceptable and offers better performance.

---

## Further Reading

- [Database Write Sequence](./db-write-sequence.md)
- [SQL B+ Tree](../week1/db-sql-b+tree.md)

