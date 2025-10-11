# ACID: Atomicity

## Definition

**Atomicity** ensures that a transaction is treated as a single, indivisible unit of work. Either all operations within the transaction succeed and are committed, or all operations fail and the database is rolled back to its original state.

> **Key Principle**: "All or Nothing"

---

## Core Concepts

### Transaction as an Atomic Unit

A transaction consists of one or more database operations that are treated as a single logical unit:

```sql
BEGIN TRANSACTION;
  -- Operation 1
  INSERT INTO orders (order_id, customer_id, total) VALUES (1001, 42, 150.00);
  
  -- Operation 2
  UPDATE inventory SET quantity = quantity - 1 WHERE product_id = 5;
  
  -- Operation 3
  INSERT INTO order_items (order_id, product_id, quantity) VALUES (1001, 5, 1);
COMMIT;
```

**Atomicity guarantees:**
- If all operations succeed → COMMIT (all changes are saved)
- If any operation fails → ROLLBACK (all changes are undone)
- No partial state exists

---

## How Atomicity Works

### 1. Transaction States

A transaction goes through several states:

```
[BEGIN] → [ACTIVE] → [PARTIALLY COMMITTED] → [COMMITTED]
                ↓
            [FAILED] → [ABORTED]
```

- **Active**: Transaction is executing
- **Partially Committed**: All operations completed but not yet written to disk
- **Committed**: Changes permanently saved
- **Failed**: Error occurred during execution
- **Aborted**: Transaction rolled back, database restored

### 2. Write-Ahead Logging (WAL)

Most databases use WAL to ensure atomicity:

```
1. Write changes to transaction log (WAL)
2. Apply changes to database
3. Mark transaction as committed in log
```

**Benefits:**
- Changes are logged before being applied
- In case of crash, log can be used to redo or undo transactions
- Enables rollback capability

### 3. Shadow Paging

Alternative approach used by some databases:

```
1. Create a copy (shadow) of affected pages
2. Make changes to the shadow pages
3. On COMMIT, atomically swap pointers
4. On ROLLBACK, discard shadow pages
```

---

## Implementation Mechanisms

### Undo Logs

Records the old values of data before modification:

```
Transaction T1:
  BEGIN TRANSACTION
  [Log: <T1, start>]
  
  UPDATE accounts SET balance = 500 WHERE id = 1;
  [Log: <T1, accounts.id=1, old_balance=1000>]
  
  UPDATE accounts SET balance = 1500 WHERE id = 2;
  [Log: <T1, accounts.id=2, old_balance=1000>]
  
  COMMIT
  [Log: <T1, commit>]
```

**On Rollback:**
- Read undo log in reverse order
- Restore old values
- Example: Set accounts.id=1.balance back to 1000

### Redo Logs

Records the new values after modification:

```
Transaction T1:
  BEGIN TRANSACTION
  [Log: <T1, start>]
  
  UPDATE accounts SET balance = 500 WHERE id = 1;
  [Log: <T1, accounts.id=1, new_balance=500>]
  
  COMMIT
  [Log: <T1, commit>]
```

**On Recovery:**
- Replay redo log
- Reapply committed changes that weren't written to disk

---

## Real-World Examples

### Example 1: Bank Transfer

```sql
BEGIN TRANSACTION;
  -- Deduct from account A
  UPDATE accounts SET balance = balance - 100 WHERE account_id = 'A';
  
  -- Check if balance is negative (business rule)
  SELECT balance FROM accounts WHERE account_id = 'A';
  -- If balance < 0, ROLLBACK
  
  -- Add to account B
  UPDATE accounts SET balance = balance + 100 WHERE account_id = 'B';
COMMIT;
```

**Scenarios:**

| Scenario | Result | Atomicity in Action |
|----------|--------|---------------------|
| Both updates succeed | COMMIT | All changes saved |
| Account A doesn't exist | ROLLBACK | No changes made |
| Account A has insufficient funds | ROLLBACK | No changes made |
| System crashes after first update | ROLLBACK | Changes undone on recovery |

### Example 2: E-commerce Order

```sql
BEGIN TRANSACTION;
  -- Create order
  INSERT INTO orders (order_id, user_id, total) VALUES (1234, 42, 299.99);
  
  -- Reserve inventory
  UPDATE products SET stock = stock - 1 WHERE product_id = 567;
  
  -- Process payment
  INSERT INTO payments (order_id, amount, status) VALUES (1234, 299.99, 'pending');
  
  -- If payment processing fails
  IF payment_error THEN
    ROLLBACK;  -- Order not created, inventory not reserved
  ELSE
    COMMIT;    -- Everything succeeds together
  END IF;
```

### Example 3: Multi-Table Update

```sql
BEGIN TRANSACTION;
  -- Update employee department
  UPDATE employees SET dept_id = 5 WHERE emp_id = 101;
  
  -- Update department head count
  UPDATE departments SET employee_count = employee_count + 1 WHERE dept_id = 5;
  UPDATE departments SET employee_count = employee_count - 1 WHERE dept_id = 3;
  
  -- Update payroll system
  INSERT INTO payroll_changes (emp_id, old_dept, new_dept, date) 
    VALUES (101, 3, 5, NOW());
COMMIT;
```

**Atomicity ensures:** All four tables are updated together or none are updated.

---

## Failure Scenarios and Recovery

### 1. Application Crash

```
BEGIN TRANSACTION;
  UPDATE accounts SET balance = balance - 100 WHERE id = 1;
  -- APPLICATION CRASHES HERE --
  UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;
```

**Recovery:**
- Database detects uncommitted transaction
- Automatically rolls back using undo log
- Account 1 balance is restored

### 2. Database Crash

```
BEGIN TRANSACTION;
  UPDATE accounts SET balance = balance - 100 WHERE id = 1;
  UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;  -- Committed but not flushed to disk
-- DATABASE CRASHES HERE --
```

**Recovery:**
- On restart, database reads WAL
- Replays committed transactions (redo)
- Rolls back uncommitted transactions (undo)

### 3. Disk Write Failure

```
BEGIN TRANSACTION;
  UPDATE large_table SET status = 'processed' WHERE status = 'pending';
  -- Affects 1 million rows
  -- Disk fills up after 500,000 rows
COMMIT;
```

**Recovery:**
- Transaction fails
- All 500,000 partial updates are rolled back
- Database state is restored to before transaction

---

## Atomicity Violations

### Without Atomicity (Hypothetical)

```sql
-- BAD: No transaction, no atomicity
UPDATE accounts SET balance = balance - 100 WHERE account_id = 'A';
-- System crashes here
UPDATE accounts SET balance = balance + 100 WHERE account_id = 'B';
```

**Problem:** $100 deducted from A but never added to B (money disappeared!)

### With Atomicity

```sql
-- GOOD: With transaction
BEGIN TRANSACTION;
  UPDATE accounts SET balance = balance - 100 WHERE account_id = 'A';
  -- System crashes here
  UPDATE accounts SET balance = balance + 100 WHERE account_id = 'B';
COMMIT;
```

**Solution:** Entire transaction is rolled back, both accounts unchanged (money preserved!)

---

## Programming Examples

### Python (psycopg2)

```python
import psycopg2

conn = psycopg2.connect("dbname=bank user=postgres")
cur = conn.cursor()

try:
    # Start transaction (implicit with psycopg2)
    cur.execute("UPDATE accounts SET balance = balance - 100 WHERE id = %s", (1,))
    cur.execute("UPDATE accounts SET balance = balance + 100 WHERE id = %s", (2,))
    
    # Commit if all successful
    conn.commit()
    print("Transfer successful")
    
except Exception as e:
    # Rollback on any error
    conn.rollback()
    print(f"Transfer failed: {e}")
    
finally:
    cur.close()
    conn.close()
```

### Java (JDBC)

```java
Connection conn = DriverManager.getConnection("jdbc:postgresql://localhost/bank");

try {
    // Disable auto-commit to control transaction manually
    conn.setAutoCommit(false);
    
    // Execute operations
    PreparedStatement ps1 = conn.prepareStatement(
        "UPDATE accounts SET balance = balance - ? WHERE id = ?");
    ps1.setDouble(1, 100.0);
    ps1.setInt(2, 1);
    ps1.executeUpdate();
    
    PreparedStatement ps2 = conn.prepareStatement(
        "UPDATE accounts SET balance = balance + ? WHERE id = ?");
    ps2.setDouble(1, 100.0);
    ps2.setInt(2, 2);
    ps2.executeUpdate();
    
    // Commit transaction
    conn.commit();
    System.out.println("Transfer successful");
    
} catch (SQLException e) {
    // Rollback on error
    conn.rollback();
    System.out.println("Transfer failed: " + e.getMessage());
    
} finally {
    conn.setAutoCommit(true);
    conn.close();
}
```

### Node.js (pg)

```javascript
const { Client } = require('pg');
const client = new Client({ database: 'bank' });

async function transfer(fromId, toId, amount) {
    await client.connect();
    
    try {
        // Begin transaction
        await client.query('BEGIN');
        
        // Deduct from sender
        await client.query(
            'UPDATE accounts SET balance = balance - $1 WHERE id = $2',
            [amount, fromId]
        );
        
        // Add to receiver
        await client.query(
            'UPDATE accounts SET balance = balance + $1 WHERE id = $2',
            [amount, toId]
        );
        
        // Commit transaction
        await client.query('COMMIT');
        console.log('Transfer successful');
        
    } catch (e) {
        // Rollback on error
        await client.query('ROLLBACK');
        console.log('Transfer failed:', e.message);
        
    } finally {
        await client.end();
    }
}
```

---

## Savepoints

Advanced feature for partial rollback within a transaction:

```sql
BEGIN TRANSACTION;
  -- Step 1: Create order
  INSERT INTO orders (id, total) VALUES (1, 100);
  
  SAVEPOINT order_created;
  
  -- Step 2: Try to apply discount
  UPDATE orders SET total = total * 0.9 WHERE id = 1;
  
  -- Step 3: Validate discount
  IF discount_invalid THEN
    ROLLBACK TO SAVEPOINT order_created;  -- Undo discount only
  END IF;
  
  -- Step 4: Process payment
  INSERT INTO payments (order_id, amount) VALUES (1, total);
  
COMMIT;  -- Commit everything
```

**Benefits:**
- Fine-grained control
- Partial rollback without aborting entire transaction
- Useful for complex business logic

---

## Best Practices

### 1. Keep Transactions Short

```sql
-- BAD: Long-running transaction
BEGIN TRANSACTION;
  SELECT * FROM large_table;  -- Millions of rows
  -- Process in application for 10 minutes
  UPDATE large_table SET processed = true;
COMMIT;

-- GOOD: Short transaction
BEGIN TRANSACTION;
  UPDATE large_table SET processed = true WHERE id IN (batch_ids);
COMMIT;
```

**Why:** Long transactions hold locks and resources, blocking other users.

### 2. Handle Errors Properly

```python
# BAD: No error handling
conn.execute("UPDATE ...")
conn.execute("UPDATE ...")
conn.commit()

# GOOD: Proper error handling
try:
    conn.execute("UPDATE ...")
    conn.execute("UPDATE ...")
    conn.commit()
except Exception:
    conn.rollback()
    raise
```

### 3. Avoid User Interaction in Transactions

```sql
-- BAD: Waiting for user input
BEGIN TRANSACTION;
  UPDATE inventory SET reserved = true WHERE product_id = 5;
  -- Wait for user to confirm payment (30 seconds)
  UPDATE orders SET status = 'completed' WHERE order_id = 1;
COMMIT;

-- GOOD: User interaction outside transaction
-- 1. User confirms payment
-- 2. Then start transaction:
BEGIN TRANSACTION;
  UPDATE inventory SET reserved = false WHERE product_id = 5;
  UPDATE orders SET status = 'completed' WHERE order_id = 1;
COMMIT;
```

### 4. Use Appropriate Isolation Levels

```sql
-- For simple reads
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;

-- For critical financial transactions
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
```

---

## Testing Atomicity

### Test 1: Force Rollback

```sql
BEGIN TRANSACTION;
  INSERT INTO test_table VALUES (1, 'test');
  SELECT * FROM test_table WHERE id = 1;  -- Should see the row
ROLLBACK;

SELECT * FROM test_table WHERE id = 1;  -- Should NOT see the row
```

### Test 2: Constraint Violation

```sql
BEGIN TRANSACTION;
  INSERT INTO accounts VALUES (1, 'Alice', 1000);
  INSERT INTO accounts VALUES (1, 'Bob', 2000);  -- Duplicate ID, fails
COMMIT;

-- Check: Neither row should exist
SELECT * FROM accounts WHERE id = 1;  -- Returns nothing
```

### Test 3: Crash Simulation

```python
import psycopg2
import os

conn = psycopg2.connect("dbname=test")
cur = conn.cursor()

try:
    cur.execute("BEGIN")
    cur.execute("UPDATE accounts SET balance = balance - 100 WHERE id = 1")
    cur.execute("UPDATE accounts SET balance = balance + 100 WHERE id = 2")
    
    # Simulate crash (kill process)
    os._exit(1)  # Abrupt termination
    
    cur.execute("COMMIT")  # Never reached
except:
    pass

# On restart, check balances - should be unchanged
```

---

## Database-Specific Notes

### PostgreSQL
- Uses WAL (Write-Ahead Logging)
- Excellent atomicity guarantees
- `pg_wal` directory contains transaction logs

### MySQL (InnoDB)
- Uses redo/undo logs
- InnoDB storage engine fully supports atomicity
- MyISAM does NOT support transactions (no atomicity)

### Oracle
- Uses redo logs and undo tablespaces
- Automatic undo management
- System Change Number (SCN) for transaction ordering

### SQL Server
- Transaction log file (.ldf)
- Supports savepoints
- Distributed transactions via MS DTC

---

## Summary

| Aspect | Description |
|--------|-------------|
| **Definition** | All operations in a transaction succeed or all fail |
| **Key Mechanism** | Write-Ahead Logging (WAL) |
| **Benefits** | Prevents partial updates, ensures data integrity |
| **Implementation** | Undo/Redo logs, shadow paging |
| **Commands** | BEGIN, COMMIT, ROLLBACK, SAVEPOINT |
| **Best Practice** | Keep transactions short, handle errors properly |

---

## Related Topics

- [Consistency](./db-acid-consistency.md) - Maintaining valid database states
- [Isolation](./db-acid-isolation.md) - Concurrent transaction handling
- [Durability](./db-acid-durability.md) - Persistence after crashes
- [Database Write Sequence](./db-write-sequence.md) - How writes are processed

