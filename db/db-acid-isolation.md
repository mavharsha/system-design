# ACID: Isolation

## Definition

**Isolation** ensures that concurrent transactions execute independently without interfering with each other. Each transaction should be unaware of other transactions running simultaneously, and intermediate states of a transaction should not be visible to other transactions.

> **Key Principle**: "Transactions Don't Interfere"

---

## Core Concepts

### Concurrent Execution

Multiple transactions may execute simultaneously:

```
Time →
T1: |-------- BEGIN ---------|------- UPDATE -------|---- COMMIT ----|
T2:          |---- BEGIN ----|---- SELECT ---|------- UPDATE --|---- COMMIT ----|
T3:                    |-- BEGIN --|-- SELECT --|-- COMMIT --|
```

**Isolation ensures:**
- Transactions don't see each other's uncommitted changes
- Final result is as if transactions ran sequentially (serializability)
- No interference between concurrent operations

---

## Isolation Problems

Without proper isolation, several problems can occur:

### 1. Dirty Read

Reading uncommitted data from another transaction:

```sql
-- Initial state: Account A = $1000

-- Transaction 1
BEGIN;
    UPDATE accounts SET balance = balance - 500 WHERE id = 'A';
    -- Balance is now $500 (uncommitted)
    
    -- Transaction 2 reads here
    
ROLLBACK;  -- Transaction 1 rolls back

-- Transaction 2
BEGIN;
    SELECT balance FROM accounts WHERE id = 'A';  -- Reads $500 (DIRTY!)
    -- But Transaction 1 rolled back, actual balance is $1000
COMMIT;
```

**Problem:** Transaction 2 read data that was never committed (doesn't actually exist).

### 2. Non-Repeatable Read

Same query returns different results within the same transaction:

```sql
-- Initial state: Account A = $1000

-- Transaction 1
BEGIN;
    SELECT balance FROM accounts WHERE id = 'A';  -- Reads $1000
    
    -- Transaction 2 modifies the data
    
    SELECT balance FROM accounts WHERE id = 'A';  -- Reads $500 (DIFFERENT!)
COMMIT;

-- Transaction 2
BEGIN;
    UPDATE accounts SET balance = 500 WHERE id = 'A';
COMMIT;
```

**Problem:** Transaction 1 gets different values for the same row within the same transaction.

### 3. Phantom Read

New rows appear or disappear between queries:

```sql
-- Initial state: 5 accounts with balance > $1000

-- Transaction 1
BEGIN;
    SELECT COUNT(*) FROM accounts WHERE balance > 1000;  -- Returns 5
    
    -- Transaction 2 inserts new account
    
    SELECT COUNT(*) FROM accounts WHERE balance > 1000;  -- Returns 6 (PHANTOM!)
COMMIT;

-- Transaction 2
BEGIN;
    INSERT INTO accounts VALUES (6, 'Account F', 2000);
COMMIT;
```

**Problem:** New rows appeared that match the query criteria (or existing rows disappeared).

### 4. Lost Update

Two transactions update the same data, and one update is lost:

```sql
-- Initial state: balance = $1000

-- Transaction 1
BEGIN;
    SELECT balance FROM accounts WHERE id = 'A';  -- Reads $1000
    -- Calculate new balance: $1000 + $100 = $1100
    
    -- Transaction 2 also runs
    
    UPDATE accounts SET balance = 1100 WHERE id = 'A';
COMMIT;

-- Transaction 2
BEGIN;
    SELECT balance FROM accounts WHERE id = 'A';  -- Reads $1000
    -- Calculate new balance: $1000 + $50 = $1050
    UPDATE accounts SET balance = 1050 WHERE id = 'A';  -- OVERWRITES T1's update!
COMMIT;

-- Final balance: $1050 (Transaction 1's $100 deposit is LOST!)
```

**Problem:** Transaction 2 overwrote Transaction 1's changes.

---

## Isolation Levels

SQL standard defines four isolation levels, each preventing different problems:

| Isolation Level | Dirty Read | Non-Repeatable Read | Phantom Read | Lost Update |
|----------------|------------|---------------------|--------------|-------------|
| **Read Uncommitted** | ❌ Possible | ❌ Possible | ❌ Possible | ❌ Possible |
| **Read Committed** | ✅ Prevented | ❌ Possible | ❌ Possible | ❌ Possible |
| **Repeatable Read** | ✅ Prevented | ✅ Prevented | ❌ Possible | ✅ Prevented |
| **Serializable** | ✅ Prevented | ✅ Prevented | ✅ Prevented | ✅ Prevented |

### 1. Read Uncommitted

**Lowest isolation level** - Allows reading uncommitted data from other transactions.

```sql
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

BEGIN;
    SELECT * FROM accounts WHERE id = 'A';
    -- May read uncommitted changes from other transactions
COMMIT;
```

**Characteristics:**
- No read locks
- Fastest performance
- Dirty reads possible
- Rarely used (except for analytics/reporting on approximate data)

**Use cases:**
- Reporting systems where approximate data is acceptable
- Read-only analytics queries
- Dashboard statistics

### 2. Read Committed

**Default for most databases** - Only reads committed data.

```sql
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;

BEGIN;
    SELECT * FROM accounts WHERE id = 'A';  -- Only sees committed data
    -- But might see different data if queried again
COMMIT;
```

**Characteristics:**
- Read locks released immediately after read
- Prevents dirty reads
- Non-repeatable reads possible
- Good balance of consistency and performance

**Use cases:**
- Most OLTP applications
- E-commerce platforms
- Content management systems

### 3. Repeatable Read

**Stronger isolation** - Guarantees same results for repeated reads.

```sql
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;

BEGIN;
    SELECT * FROM accounts WHERE id = 'A';  -- Reads $1000
    -- Other transactions can't modify this row
    SELECT * FROM accounts WHERE id = 'A';  -- Still reads $1000
COMMIT;
```

**Characteristics:**
- Read locks held until transaction ends
- Prevents dirty reads and non-repeatable reads
- Phantom reads still possible
- Better consistency than Read Committed

**Use cases:**
- Financial reports requiring consistency
- Batch processing
- Data migrations

### 4. Serializable

**Highest isolation level** - Transactions appear to execute sequentially.

```sql
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;

BEGIN;
    SELECT * FROM accounts WHERE balance > 1000;
    -- No other transaction can insert/delete/modify rows that match this condition
    SELECT * FROM accounts WHERE balance > 1000;  -- Identical results
COMMIT;
```

**Characteristics:**
- Range locks on queries
- Prevents all anomalies
- Slowest performance
- May cause more deadlocks

**Use cases:**
- Critical financial transactions
- Regulatory compliance requirements
- Systems requiring absolute consistency

---

## Implementation Mechanisms

### 1. Locks

#### Shared Lock (S-Lock / Read Lock)

Multiple transactions can hold shared locks simultaneously:

```sql
-- Transaction 1
BEGIN;
    SELECT * FROM accounts WHERE id = 'A' FOR SHARE;
    -- Holds shared lock
    
-- Transaction 2 (concurrent)
BEGIN;
    SELECT * FROM accounts WHERE id = 'A' FOR SHARE;
    -- Can also get shared lock (compatible)
    
    UPDATE accounts SET balance = 500 WHERE id = 'A';
    -- BLOCKED! Waits for T1's shared lock to release
```

#### Exclusive Lock (X-Lock / Write Lock)

Only one transaction can hold an exclusive lock:

```sql
-- Transaction 1
BEGIN;
    UPDATE accounts SET balance = 500 WHERE id = 'A';
    -- Holds exclusive lock
    
-- Transaction 2 (concurrent)
BEGIN;
    SELECT * FROM accounts WHERE id = 'A';
    -- BLOCKED! Waits for T1's exclusive lock to release
    
-- Transaction 3 (concurrent)
BEGIN;
    UPDATE accounts SET balance = 600 WHERE id = 'A';
    -- BLOCKED! Waits for T1's exclusive lock to release
```

#### Lock Compatibility Matrix

|         | S-Lock | X-Lock |
|---------|--------|--------|
| **S-Lock** | ✅ Compatible | ❌ Incompatible |
| **X-Lock** | ❌ Incompatible | ❌ Incompatible |

### 2. Two-Phase Locking (2PL)

Protocol to ensure serializability:

```
Phase 1 - Growing Phase:
  - Transaction acquires locks
  - Cannot release any locks

Phase 2 - Shrinking Phase:
  - Transaction releases locks
  - Cannot acquire new locks
```

**Example:**

```sql
BEGIN TRANSACTION;
    -- Growing Phase
    SELECT * FROM accounts WHERE id = 'A';  -- Acquire S-lock on 'A'
    SELECT * FROM accounts WHERE id = 'B';  -- Acquire S-lock on 'B'
    
    UPDATE accounts SET balance = 500 WHERE id = 'A';  -- Upgrade to X-lock on 'A'
    UPDATE accounts SET balance = 1500 WHERE id = 'B';  -- Upgrade to X-lock on 'B'
    
COMMIT;  -- Shrinking Phase - Release all locks
```

### 3. Multi-Version Concurrency Control (MVCC)

Instead of locking, keep multiple versions of data:

```
Version 1: balance = $1000 (created by T1 at timestamp 100)
Version 2: balance = $500  (created by T2 at timestamp 200)
Version 3: balance = $1500 (created by T3 at timestamp 300)

-- Transaction at timestamp 150 sees Version 1 ($1000)
-- Transaction at timestamp 250 sees Version 2 ($500)
-- Transaction at timestamp 350 sees Version 3 ($1500)
```

**Benefits:**
- Readers don't block writers
- Writers don't block readers
- Better concurrency
- Used by PostgreSQL, MySQL InnoDB, Oracle

**How it works:**

```sql
-- Transaction 1 (timestamp 100)
BEGIN;
    SELECT balance FROM accounts WHERE id = 'A';  -- Sees version at timestamp 100
    
    -- Transaction 2 modifies data (timestamp 150)
    
    SELECT balance FROM accounts WHERE id = 'A';  -- Still sees version at timestamp 100
    -- Transaction 1 is isolated from Transaction 2's changes
COMMIT;
```

---

## Real-World Examples

### Example 1: Bank Transfer with Isolation

```sql
-- Without proper isolation (BAD)
-- Transaction 1: Transfer $100 from A to B
BEGIN;
    balance_a = SELECT balance FROM accounts WHERE id = 'A';  -- $1000
    
    -- Transaction 2 reads here
    
    UPDATE accounts SET balance = balance_a - 100 WHERE id = 'A';
    UPDATE accounts SET balance = balance + 100 WHERE id = 'B';
COMMIT;

-- Transaction 2: Check total balance
BEGIN;
    total = SELECT SUM(balance) FROM accounts;
    -- Might see $1900 (dirty read) or $2000 depending on timing
COMMIT;

-- With proper isolation (GOOD)
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;

-- Transaction 1
BEGIN;
    UPDATE accounts SET balance = balance - 100 WHERE id = 'A';
    UPDATE accounts SET balance = balance + 100 WHERE id = 'B';
COMMIT;

-- Transaction 2
BEGIN;
    total = SELECT SUM(balance) FROM accounts;
    -- Always sees $2000 (either before or after T1, never during)
COMMIT;
```

### Example 2: E-commerce Inventory

```sql
-- Scenario: 2 customers buy the last item

-- Product stock = 1

-- Customer 1's transaction
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
BEGIN;
    SELECT stock FROM products WHERE id = 1 FOR UPDATE;  -- Reads 1, locks row
    
    IF stock > 0 THEN
        UPDATE products SET stock = stock - 1 WHERE id = 1;
        INSERT INTO orders VALUES (customer_1, product_1);
    END IF;
COMMIT;  -- stock = 0

-- Customer 2's transaction (concurrent)
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
BEGIN;
    SELECT stock FROM products WHERE id = 1 FOR UPDATE;  
    -- BLOCKED until Customer 1 commits
    -- After Customer 1 commits, reads 0
    
    IF stock > 0 THEN
        -- Condition is false, won't execute
        UPDATE products SET stock = stock - 1 WHERE id = 1;
        INSERT INTO orders VALUES (customer_2, product_1);
    END IF;
COMMIT;

-- Result: Only Customer 1 gets the item (no overselling)
```

### Example 3: Seat Reservation System

```sql
-- Scenario: 2 people try to book the same seat

-- Seat 12A is available

-- Person 1
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
BEGIN;
    SELECT status FROM seats WHERE seat_number = '12A';  -- 'available'
    
    -- Person 2 tries to book concurrently
    
    UPDATE seats SET status = 'reserved', customer_id = 1 WHERE seat_number = '12A';
COMMIT;

-- Person 2
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
BEGIN;
    SELECT status FROM seats WHERE seat_number = '12A';
    -- Either sees 'available' (if before Person 1) or 'reserved' (if after Person 1)
    -- Never sees inconsistent state
    
    UPDATE seats SET status = 'reserved', customer_id = 2 WHERE seat_number = '12A';
    -- Fails if Person 1 already reserved it
COMMIT;
```

---

## Deadlocks

### What is a Deadlock?

Two or more transactions waiting for each other to release locks:

```sql
-- Transaction 1
BEGIN;
    UPDATE accounts SET balance = 900 WHERE id = 'A';  -- Locks row A
    -- T1 now waits for lock on row B
    UPDATE accounts SET balance = 1100 WHERE id = 'B';  -- BLOCKED by T2
    
-- Transaction 2
BEGIN;
    UPDATE accounts SET balance = 1200 WHERE id = 'B';  -- Locks row B
    -- T2 now waits for lock on row A
    UPDATE accounts SET balance = 800 WHERE id = 'A';  -- BLOCKED by T1

-- DEADLOCK! Both transactions waiting for each other
```

**Visual representation:**

```
T1: Has lock on A → Wants lock on B
T2: Has lock on B → Wants lock on A

T1 → B
↑     ↓
A ← T2
```

### Deadlock Detection

Most databases automatically detect and resolve deadlocks:

```
-- PostgreSQL example
ERROR: deadlock detected
DETAIL: Process 12345 waits for ShareLock on transaction 67890;
        Process 67890 waits for ShareLock on transaction 12345.
HINT: See server log for query details.
```

**Resolution:** Database automatically rolls back one transaction (victim).

### Preventing Deadlocks

#### 1. Lock Ordering

Always acquire locks in the same order:

```sql
-- BAD: Different order
-- Transaction 1
UPDATE accounts SET balance = 900 WHERE id = 'A';
UPDATE accounts SET balance = 1100 WHERE id = 'B';

-- Transaction 2
UPDATE accounts SET balance = 1200 WHERE id = 'B';
UPDATE accounts SET balance = 800 WHERE id = 'A';

-- GOOD: Same order
-- Transaction 1
UPDATE accounts SET balance = 900 WHERE id = 'A';
UPDATE accounts SET balance = 1100 WHERE id = 'B';

-- Transaction 2
UPDATE accounts SET balance = 800 WHERE id = 'A';  -- Waits for T1
UPDATE accounts SET balance = 1200 WHERE id = 'B';  -- Then proceeds
```

#### 2. Minimize Lock Hold Time

```sql
-- BAD: Long-held locks
BEGIN;
    SELECT * FROM accounts WHERE id = 'A' FOR UPDATE;
    -- ... complex business logic for 5 seconds ...
    UPDATE accounts SET balance = new_balance WHERE id = 'A';
COMMIT;

-- GOOD: Short locks
-- Do business logic outside transaction
new_balance = calculate_new_balance();

BEGIN;
    UPDATE accounts SET balance = new_balance WHERE id = 'A';
COMMIT;
```

#### 3. Use Timeout

```sql
SET lock_timeout = '5s';

BEGIN;
    SELECT * FROM accounts WHERE id = 'A' FOR UPDATE;
    -- If can't acquire lock within 5 seconds, abort
COMMIT;
```

---

## Programming Examples

### Python: Isolation Levels

```python
import psycopg2
from psycopg2 import sql

# Read Committed (default)
conn = psycopg2.connect("dbname=test")
conn.set_isolation_level(psycopg2.extensions.ISOLATION_LEVEL_READ_COMMITTED)

# Repeatable Read
conn = psycopg2.connect("dbname=test")
conn.set_isolation_level(psycopg2.extensions.ISOLATION_LEVEL_REPEATABLE_READ)

# Serializable
conn = psycopg2.connect("dbname=test")
conn.set_isolation_level(psycopg2.extensions.ISOLATION_LEVEL_SERIALIZABLE)

# Using with transaction
def transfer_money(from_id, to_id, amount):
    conn = psycopg2.connect("dbname=bank")
    conn.set_isolation_level(psycopg2.extensions.ISOLATION_LEVEL_SERIALIZABLE)
    cur = conn.cursor()
    
    try:
        cur.execute("BEGIN")
        
        # Lock rows to prevent lost updates
        cur.execute(
            "SELECT balance FROM accounts WHERE id = %s FOR UPDATE",
            (from_id,)
        )
        balance = cur.fetchone()[0]
        
        if balance < amount:
            raise ValueError("Insufficient funds")
        
        cur.execute(
            "UPDATE accounts SET balance = balance - %s WHERE id = %s",
            (amount, from_id)
        )
        cur.execute(
            "UPDATE accounts SET balance = balance + %s WHERE id = %s",
            (amount, to_id)
        )
        
        cur.execute("COMMIT")
        
    except psycopg2.extensions.TransactionRollbackError:
        # Serialization failure or deadlock
        cur.execute("ROLLBACK")
        print("Transaction failed due to conflict, please retry")
        
    finally:
        cur.close()
        conn.close()
```

### Java: Handling Isolation

```java
import java.sql.*;

public class BankTransfer {
    public void transfer(int fromId, int toId, double amount) {
        Connection conn = null;
        
        try {
            conn = DriverManager.getConnection("jdbc:postgresql://localhost/bank");
            
            // Set isolation level
            conn.setTransactionIsolation(Connection.TRANSACTION_SERIALIZABLE);
            conn.setAutoCommit(false);
            
            // Lock row to prevent lost updates
            PreparedStatement ps1 = conn.prepareStatement(
                "SELECT balance FROM accounts WHERE id = ? FOR UPDATE");
            ps1.setInt(1, fromId);
            ResultSet rs = ps1.executeQuery();
            
            if (rs.next()) {
                double balance = rs.getDouble("balance");
                if (balance < amount) {
                    throw new InsufficientFundsException();
                }
            }
            
            // Perform transfer
            PreparedStatement ps2 = conn.prepareStatement(
                "UPDATE accounts SET balance = balance - ? WHERE id = ?");
            ps2.setDouble(1, amount);
            ps2.setInt(2, fromId);
            ps2.executeUpdate();
            
            PreparedStatement ps3 = conn.prepareStatement(
                "UPDATE accounts SET balance = balance + ? WHERE id = ?");
            ps3.setDouble(1, amount);
            ps3.setInt(2, toId);
            ps3.executeUpdate();
            
            conn.commit();
            
        } catch (SQLException e) {
            if (e.getSQLState().equals("40001")) {
                // Serialization failure - retry
                System.out.println("Conflict detected, retrying...");
            }
            
            try {
                if (conn != null) conn.rollback();
            } catch (SQLException ex) {
                ex.printStackTrace();
            }
            
        } finally {
            try {
                if (conn != null) conn.close();
            } catch (SQLException e) {
                e.printStackTrace();
            }
        }
    }
}
```

### Node.js: Concurrent Transactions

```javascript
const { Pool } = require('pg');
const pool = new Pool({ database: 'bank' });

async function reserveSeat(seatNumber, customerId) {
    const client = await pool.connect();
    
    try {
        // Start transaction with Repeatable Read
        await client.query('BEGIN ISOLATION LEVEL REPEATABLE READ');
        
        // Lock the seat row
        const result = await client.query(
            'SELECT status FROM seats WHERE seat_number = $1 FOR UPDATE',
            [seatNumber]
        );
        
        if (result.rows[0].status !== 'available') {
            throw new Error('Seat already reserved');
        }
        
        // Reserve the seat
        await client.query(
            'UPDATE seats SET status = $1, customer_id = $2 WHERE seat_number = $3',
            ['reserved', customerId, seatNumber]
        );
        
        await client.query('COMMIT');
        console.log(`Seat ${seatNumber} reserved for customer ${customerId}`);
        
    } catch (error) {
        await client.query('ROLLBACK');
        
        if (error.code === '40001') {
            // Serialization failure
            console.log('Conflict detected, please retry');
        } else {
            console.error('Reservation failed:', error.message);
        }
        
    } finally {
        client.release();
    }
}

// Simulate concurrent reservations
Promise.all([
    reserveSeat('12A', 1),
    reserveSeat('12A', 2)  // One will fail
]);
```

---

## Database-Specific Behaviors

### PostgreSQL

```sql
-- Default: Read Committed
-- Supports all four isolation levels
-- Uses MVCC (excellent concurrency)

-- Setting isolation level
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;

-- Or for entire session
SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE;

-- Check current level
SHOW transaction_isolation;
```

### MySQL (InnoDB)

```sql
-- Default: Repeatable Read
-- Supports all four isolation levels
-- Uses MVCC + locking

-- Setting isolation level
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;

-- Or for entire session
SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

-- Check current level
SELECT @@transaction_isolation;
```

### Oracle

```sql
-- Default: Read Committed
-- Supports only Read Committed and Serializable
-- Uses MVCC (undo segments)

-- Setting isolation level
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;

-- Oracle-specific: Read-only transaction
SET TRANSACTION READ ONLY;
```

### SQL Server

```sql
-- Default: Read Committed
-- Supports all four isolation levels
-- Uses locking by default (can enable MVCC)

-- Setting isolation level
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;

-- Check current level
SELECT CASE transaction_isolation_level 
    WHEN 0 THEN 'Unspecified' 
    WHEN 1 THEN 'ReadUncommitted' 
    WHEN 2 THEN 'ReadCommitted' 
    WHEN 3 THEN 'RepeatableRead' 
    WHEN 4 THEN 'Serializable' 
END AS IsolationLevel
FROM sys.dm_exec_sessions 
WHERE session_id = @@SPID;
```

---

## Best Practices

### 1. Choose Appropriate Isolation Level

```sql
-- For reporting/analytics (approximate data OK)
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

-- For most OLTP applications
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;

-- For financial transactions
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
```

### 2. Use SELECT FOR UPDATE

```sql
-- Prevent lost updates
BEGIN;
    SELECT * FROM accounts WHERE id = 'A' FOR UPDATE;
    -- Row is locked, other transactions must wait
    UPDATE accounts SET balance = new_balance WHERE id = 'A';
COMMIT;
```

### 3. Handle Serialization Failures

```python
max_retries = 3
for attempt in range(max_retries):
    try:
        # Execute transaction
        execute_transaction()
        break  # Success
    except SerializationFailure:
        if attempt == max_retries - 1:
            raise  # Give up after max retries
        time.sleep(0.1 * (2 ** attempt))  # Exponential backoff
```

### 4. Keep Transactions Short

```sql
-- BAD: Long transaction
BEGIN;
    SELECT * FROM large_table;
    -- Process for 10 minutes in application
    UPDATE another_table SET ...;
COMMIT;

-- GOOD: Short transaction
-- Process data outside transaction
BEGIN;
    UPDATE another_table SET ...;
COMMIT;
```

---

## Performance Considerations

| Isolation Level | Concurrency | Consistency | Performance |
|----------------|-------------|-------------|-------------|
| Read Uncommitted | Highest | Lowest | Fastest |
| Read Committed | High | Medium | Fast |
| Repeatable Read | Medium | High | Medium |
| Serializable | Lowest | Highest | Slowest |

**Trade-offs:**
- Higher isolation = Better consistency but lower concurrency
- Lower isolation = Better performance but weaker guarantees

---

## Summary

| Aspect | Description |
|--------|-------------|
| **Definition** | Concurrent transactions don't interfere |
| **Problems** | Dirty read, non-repeatable read, phantom read, lost update |
| **Mechanisms** | Locks, MVCC, 2PL |
| **Isolation Levels** | Read Uncommitted, Read Committed, Repeatable Read, Serializable |
| **Trade-offs** | Consistency vs performance/concurrency |
| **Best Practice** | Choose appropriate level for use case |

---

## Related Topics

- [Atomicity](./db-acid-atomicity.md) - All-or-nothing transactions
- [Consistency](./db-acid-consistency.md) - Valid database states
- [Durability](./db-acid-durability.md) - Persistence of committed data
- [Database Locking Strategies](../week1/2-db.md)

