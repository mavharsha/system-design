# ACID: Consistency

## Definition

**Consistency** ensures that a transaction brings the database from one valid state to another valid state, maintaining all defined rules, constraints, and relationships. After a transaction completes, all data integrity constraints must be satisfied.

> **Key Principle**: "Valid State to Valid State"

---

## Core Concepts

### Database Integrity

Consistency means the database remains in a correct state that satisfies all rules:

```sql
-- Before transaction: Valid state
Account A: $1000
Account B: $1000
Total: $2000

-- During transaction: May be temporarily invalid (not visible to others)
Account A: $900
Account B: $1000
Total: $1900 (temporarily inconsistent)

-- After transaction: Valid state again
Account A: $900
Account B: $1100
Total: $2000 (consistent!)
```

**Key Point:** Intermediate states may violate constraints, but the final committed state must be valid.

---

## Types of Consistency

### 1. Application-Level Consistency

Rules defined by business logic:

```sql
-- Business Rule: Total money in system must remain constant

CREATE FUNCTION check_total_balance()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT SUM(balance) FROM accounts) < 0 THEN
        RAISE EXCEPTION 'Total balance cannot be negative';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER balance_check
    AFTER UPDATE ON accounts
    FOR EACH STATEMENT
    EXECUTE FUNCTION check_total_balance();
```

### 2. Database-Level Consistency

Rules enforced by the database:

```sql
-- Primary Key Constraint
CREATE TABLE users (
    id INT PRIMARY KEY,  -- Must be unique and not null
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL
);

-- Foreign Key Constraint
CREATE TABLE orders (
    order_id INT PRIMARY KEY,
    user_id INT,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Check Constraint
CREATE TABLE products (
    product_id INT PRIMARY KEY,
    price DECIMAL(10,2) CHECK (price >= 0),  -- Price can't be negative
    stock INT CHECK (stock >= 0)  -- Stock can't be negative
);
```

---

## Consistency Constraints

### 1. Entity Integrity

Every table must have a primary key, and the primary key cannot be null:

```sql
-- VALID
CREATE TABLE employees (
    emp_id INT PRIMARY KEY,  -- NOT NULL automatically
    name VARCHAR(100)
);

-- INVALID - Will fail
INSERT INTO employees (emp_id, name) VALUES (NULL, 'John');
-- Error: null value in column "emp_id" violates not-null constraint
```

### 2. Referential Integrity

Foreign keys must reference existing records:

```sql
CREATE TABLE departments (
    dept_id INT PRIMARY KEY,
    dept_name VARCHAR(50)
);

CREATE TABLE employees (
    emp_id INT PRIMARY KEY,
    name VARCHAR(100),
    dept_id INT,
    FOREIGN KEY (dept_id) REFERENCES departments(dept_id)
);

-- VALID
INSERT INTO departments VALUES (1, 'Engineering');
INSERT INTO employees VALUES (101, 'Alice', 1);  -- References existing dept

-- INVALID - Will fail
INSERT INTO employees VALUES (102, 'Bob', 99);  -- Dept 99 doesn't exist
-- Error: insert or update on table "employees" violates foreign key constraint
```

### 3. Domain Integrity

Column values must satisfy defined constraints:

```sql
CREATE TABLE products (
    product_id INT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) CHECK (price > 0),
    category VARCHAR(20) CHECK (category IN ('Electronics', 'Clothing', 'Food')),
    stock INT DEFAULT 0 CHECK (stock >= 0),
    created_at TIMESTAMP DEFAULT NOW()
);

-- VALID
INSERT INTO products (product_id, name, price, category, stock)
VALUES (1, 'Laptop', 999.99, 'Electronics', 10);

-- INVALID - Negative price
INSERT INTO products (product_id, name, price, category)
VALUES (2, 'Shirt', -20.00, 'Clothing');
-- Error: new row violates check constraint "products_price_check"

-- INVALID - Invalid category
INSERT INTO products (product_id, name, price, category)
VALUES (3, 'Book', 15.99, 'Literature');
-- Error: new row violates check constraint "products_category_check"
```

### 4. User-Defined Integrity

Custom business rules:

```sql
-- Rule: Age must be between 18 and 100
CREATE TABLE users (
    user_id INT PRIMARY KEY,
    name VARCHAR(100),
    age INT CHECK (age >= 18 AND age <= 100)
);

-- Rule: End date must be after start date
CREATE TABLE projects (
    project_id INT PRIMARY KEY,
    name VARCHAR(100),
    start_date DATE,
    end_date DATE,
    CHECK (end_date > start_date)
);

-- Rule: Discount cannot exceed price
CREATE TABLE orders (
    order_id INT PRIMARY KEY,
    price DECIMAL(10,2),
    discount DECIMAL(10,2),
    CHECK (discount <= price)
);
```

---

## Real-World Examples

### Example 1: Bank Account Transfer

```sql
-- Constraint: Account balance cannot be negative
CREATE TABLE accounts (
    account_id INT PRIMARY KEY,
    balance DECIMAL(10,2) CHECK (balance >= 0)
);

-- Insert initial data
INSERT INTO accounts VALUES (1, 1000.00), (2, 500.00);

-- Transaction 1: VALID transfer
BEGIN TRANSACTION;
    UPDATE accounts SET balance = balance - 100 WHERE account_id = 1;
    -- Balance: 900 (valid)
    UPDATE accounts SET balance = balance + 100 WHERE account_id = 2;
    -- Balance: 600 (valid)
COMMIT;  -- Success! Database is consistent

-- Transaction 2: INVALID transfer (insufficient funds)
BEGIN TRANSACTION;
    UPDATE accounts SET balance = balance - 2000 WHERE account_id = 1;
    -- Would make balance = -1000 (violates CHECK constraint)
ROLLBACK;  -- Fails! Database remains consistent
-- Error: new row violates check constraint "accounts_balance_check"
```

### Example 2: E-commerce Inventory

```sql
CREATE TABLE products (
    product_id INT PRIMARY KEY,
    name VARCHAR(100),
    stock INT CHECK (stock >= 0)  -- Can't have negative stock
);

CREATE TABLE orders (
    order_id INT PRIMARY KEY,
    product_id INT,
    quantity INT CHECK (quantity > 0),  -- Must order at least 1
    FOREIGN KEY (product_id) REFERENCES products(product_id)
);

-- Initial state
INSERT INTO products VALUES (1, 'Laptop', 10);

-- Scenario 1: Valid order
BEGIN TRANSACTION;
    -- Check stock
    SELECT stock FROM products WHERE product_id = 1;  -- Returns 10
    
    -- Place order for 5 units
    INSERT INTO orders VALUES (1001, 1, 5);
    
    -- Reduce stock
    UPDATE products SET stock = stock - 5 WHERE product_id = 1;
    -- Stock becomes 5 (valid)
COMMIT;  -- Success!

-- Scenario 2: Invalid order (oversell)
BEGIN TRANSACTION;
    -- Place order for 20 units (but only 5 in stock)
    INSERT INTO orders VALUES (1002, 1, 20);
    
    -- Try to reduce stock
    UPDATE products SET stock = stock - 20 WHERE product_id = 1;
    -- Stock becomes -15 (INVALID!)
ROLLBACK;  -- Automatically rolled back
-- Error: new row violates check constraint "products_stock_check"
```

### Example 3: Course Registration

```sql
CREATE TABLE courses (
    course_id INT PRIMARY KEY,
    name VARCHAR(100),
    max_students INT,
    enrolled INT CHECK (enrolled >= 0 AND enrolled <= max_students)
);

-- Initialize course
INSERT INTO courses VALUES (101, 'Database Systems', 30, 0);

-- Valid registration
BEGIN TRANSACTION;
    UPDATE courses 
    SET enrolled = enrolled + 1 
    WHERE course_id = 101 AND enrolled < max_students;
    -- enrolled = 1 (valid)
COMMIT;

-- Invalid registration (course full)
-- After 30 students are enrolled...
BEGIN TRANSACTION;
    UPDATE courses 
    SET enrolled = enrolled + 1 
    WHERE course_id = 101;
    -- enrolled = 31 (exceeds max_students = 30)
ROLLBACK;  -- Violates constraint
```

---

## Consistency Mechanisms

### 1. Triggers

Automatically enforce complex rules:

```sql
-- Ensure order total matches sum of order items
CREATE TABLE orders (
    order_id INT PRIMARY KEY,
    customer_id INT,
    total DECIMAL(10,2)
);

CREATE TABLE order_items (
    item_id INT PRIMARY KEY,
    order_id INT,
    product_id INT,
    quantity INT,
    price DECIMAL(10,2),
    FOREIGN KEY (order_id) REFERENCES orders(order_id)
);

-- Trigger to maintain consistency
CREATE FUNCTION check_order_total()
RETURNS TRIGGER AS $$
DECLARE
    calculated_total DECIMAL(10,2);
BEGIN
    SELECT SUM(quantity * price) INTO calculated_total
    FROM order_items
    WHERE order_id = NEW.order_id;
    
    IF calculated_total != (SELECT total FROM orders WHERE order_id = NEW.order_id) THEN
        RAISE EXCEPTION 'Order total does not match sum of items';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER verify_order_total
    AFTER INSERT OR UPDATE ON order_items
    FOR EACH ROW
    EXECUTE FUNCTION check_order_total();
```

### 2. Stored Procedures

Encapsulate business logic:

```sql
-- Transfer money between accounts with validation
CREATE PROCEDURE transfer_money(
    from_account INT,
    to_account INT,
    amount DECIMAL(10,2)
)
LANGUAGE plpgsql
AS $$
DECLARE
    from_balance DECIMAL(10,2);
BEGIN
    -- Check sender has sufficient balance
    SELECT balance INTO from_balance
    FROM accounts
    WHERE account_id = from_account;
    
    IF from_balance < amount THEN
        RAISE EXCEPTION 'Insufficient funds';
    END IF;
    
    -- Check amount is positive
    IF amount <= 0 THEN
        RAISE EXCEPTION 'Amount must be positive';
    END IF;
    
    -- Perform transfer
    UPDATE accounts SET balance = balance - amount WHERE account_id = from_account;
    UPDATE accounts SET balance = balance + amount WHERE account_id = to_account;
    
    -- Log transaction
    INSERT INTO transaction_log (from_account, to_account, amount, timestamp)
    VALUES (from_account, to_account, amount, NOW());
END;
$$;

-- Usage
CALL transfer_money(1, 2, 100.00);
```

### 3. Database Constraints

Built-in enforcement:

```sql
-- Comprehensive example with multiple constraints
CREATE TABLE employees (
    emp_id INT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    age INT CHECK (age >= 18 AND age <= 65),
    salary DECIMAL(10,2) CHECK (salary > 0),
    dept_id INT,
    manager_id INT,
    hire_date DATE DEFAULT CURRENT_DATE,
    
    -- Foreign key constraints
    FOREIGN KEY (dept_id) REFERENCES departments(dept_id),
    FOREIGN KEY (manager_id) REFERENCES employees(emp_id),
    
    -- Check constraints
    CHECK (hire_date <= CURRENT_DATE),  -- Can't hire in future
    CHECK (manager_id != emp_id)  -- Can't be own manager
);
```

---

## Consistency vs Other ACID Properties

### Consistency + Atomicity

```sql
-- Atomicity ensures all-or-nothing
-- Consistency ensures the "all" is valid

BEGIN TRANSACTION;
    -- Operation 1: Deduct from sender
    UPDATE accounts SET balance = balance - 100 WHERE id = 1;
    
    -- Operation 2: Add to receiver
    UPDATE accounts SET balance = balance + 100 WHERE id = 2;
    
    -- Atomicity: Both operations execute or neither
    -- Consistency: If balance would go negative, transaction fails
COMMIT;
```

### Consistency + Isolation

```sql
-- Isolation: Concurrent transactions don't interfere
-- Consistency: Each transaction sees a consistent state

-- Transaction 1
BEGIN;
    SELECT balance FROM accounts WHERE id = 1;  -- Sees 1000
    UPDATE accounts SET balance = balance - 100 WHERE id = 1;
COMMIT;

-- Transaction 2 (concurrent)
BEGIN;
    SELECT balance FROM accounts WHERE id = 1;  -- Sees either 1000 or 900, not 950
    UPDATE accounts SET balance = balance - 50 WHERE id = 1;
COMMIT;
```

### Consistency + Durability

```sql
-- Durability: Committed changes persist
-- Consistency: Persisted state is always valid

BEGIN TRANSACTION;
    UPDATE accounts SET balance = 500 WHERE id = 1;
    -- Consistency: 500 is valid (>= 0)
COMMIT;

-- System crashes here

-- After restart:
-- Durability: balance is still 500
-- Consistency: 500 is still valid
```

---

## Consistency in Distributed Systems

### CAP Theorem Trade-off

In distributed systems, you can only have 2 of 3:
- **C**onsistency
- **A**vailability
- **P**artition tolerance

```
Traditional RDBMS (Single Node):
  - Full consistency
  - High availability (single point of failure)
  - No partitions

Distributed Database (e.g., Cassandra):
  - Eventual consistency (weak)
  - High availability
  - Partition tolerant
```

### Eventual Consistency

```javascript
// Write to replica 1
db.replica1.update({ id: 1 }, { balance: 900 });

// Read from replica 2 (not yet synchronized)
db.replica2.find({ id: 1 });  // Returns { balance: 1000 } (old value)

// After synchronization (few milliseconds to seconds)
db.replica2.find({ id: 1 });  // Returns { balance: 900 } (new value)
```

**Use cases for eventual consistency:**
- Social media feeds
- Product catalogs
- Analytics data
- Caching systems

**Use cases requiring strong consistency:**
- Financial transactions
- Inventory management
- Booking systems
- Authentication

---

## Testing Consistency

### Test 1: Constraint Violation

```sql
-- Setup
CREATE TABLE accounts (
    id INT PRIMARY KEY,
    balance DECIMAL(10,2) CHECK (balance >= 0)
);

INSERT INTO accounts VALUES (1, 1000);

-- Test: Try to violate constraint
BEGIN;
    UPDATE accounts SET balance = -100 WHERE id = 1;
COMMIT;

-- Expected: Transaction fails with constraint violation error
-- Verify: SELECT balance FROM accounts WHERE id = 1;  -- Should still be 1000
```

### Test 2: Referential Integrity

```sql
-- Setup
CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100));
CREATE TABLE orders (
    id INT PRIMARY KEY, 
    user_id INT, 
    FOREIGN KEY (user_id) REFERENCES users(id)
);

INSERT INTO users VALUES (1, 'Alice');

-- Test: Try to insert order for non-existent user
INSERT INTO orders VALUES (1, 999);

-- Expected: Foreign key violation error
-- Verify: SELECT * FROM orders;  -- Should be empty
```

### Test 3: Complex Business Rule

```sql
-- Business rule: Total order amount must match sum of line items

CREATE FUNCTION verify_order_integrity()
RETURNS BOOLEAN AS $$
DECLARE
    order_rec RECORD;
    calculated_total DECIMAL(10,2);
    is_consistent BOOLEAN := TRUE;
BEGIN
    FOR order_rec IN SELECT order_id, total FROM orders LOOP
        SELECT SUM(quantity * unit_price) INTO calculated_total
        FROM order_items
        WHERE order_id = order_rec.order_id;
        
        IF calculated_total != order_rec.total THEN
            RAISE NOTICE 'Inconsistency found in order %', order_rec.order_id;
            is_consistent := FALSE;
        END IF;
    END LOOP;
    
    RETURN is_consistent;
END;
$$ LANGUAGE plpgsql;

-- Run consistency check
SELECT verify_order_integrity();  -- Should return TRUE
```

---

## Programming Examples

### Python: Enforcing Consistency

```python
import psycopg2

def transfer_money(conn, from_account, to_account, amount):
    """
    Transfer money with consistency checks
    """
    cur = conn.cursor()
    
    try:
        # Start transaction
        cur.execute("BEGIN")
        
        # Check sender has sufficient balance (application-level consistency)
        cur.execute("SELECT balance FROM accounts WHERE id = %s", (from_account,))
        balance = cur.fetchone()[0]
        
        if balance < amount:
            raise ValueError(f"Insufficient funds: {balance} < {amount}")
        
        if amount <= 0:
            raise ValueError("Amount must be positive")
        
        # Perform transfer
        cur.execute(
            "UPDATE accounts SET balance = balance - %s WHERE id = %s",
            (amount, from_account)
        )
        cur.execute(
            "UPDATE accounts SET balance = balance + %s WHERE id = %s",
            (amount, to_account)
        )
        
        # Database will enforce CHECK constraints here
        cur.execute("COMMIT")
        print(f"Transfer successful: ${amount} from {from_account} to {to_account}")
        
    except psycopg2.IntegrityError as e:
        # Database constraint violation
        cur.execute("ROLLBACK")
        print(f"Database constraint violated: {e}")
        
    except ValueError as e:
        # Application-level validation failed
        cur.execute("ROLLBACK")
        print(f"Validation error: {e}")
        
    finally:
        cur.close()
```

### Java: Domain Validation

```java
public class Account {
    private int id;
    private BigDecimal balance;
    
    public void withdraw(BigDecimal amount) throws InvalidOperationException {
        // Application-level consistency check
        if (amount.compareTo(BigDecimal.ZERO) <= 0) {
            throw new InvalidOperationException("Amount must be positive");
        }
        
        if (balance.compareTo(amount) < 0) {
            throw new InvalidOperationException("Insufficient funds");
        }
        
        balance = balance.subtract(amount);
        
        // Additional check to ensure consistency
        if (balance.compareTo(BigDecimal.ZERO) < 0) {
            throw new IllegalStateException("Balance became negative!");
        }
    }
    
    public void deposit(BigDecimal amount) throws InvalidOperationException {
        if (amount.compareTo(BigDecimal.ZERO) <= 0) {
            throw new InvalidOperationException("Amount must be positive");
        }
        
        balance = balance.add(amount);
    }
}
```

---

## Best Practices

### 1. Use Database Constraints

```sql
-- GOOD: Let database enforce rules
CREATE TABLE products (
    id INT PRIMARY KEY,
    price DECIMAL(10,2) CHECK (price > 0),  -- Database enforces
    stock INT CHECK (stock >= 0)
);

-- BAD: Only application-level checks (can be bypassed)
CREATE TABLE products (
    id INT PRIMARY KEY,
    price DECIMAL(10,2),  -- No constraint
    stock INT  -- No constraint
);
-- Then checking in application code only
```

### 2. Validate in Multiple Layers

```
User Input → Application Validation → Database Constraints

Example:
  1. Frontend: Check price format, positive value
  2. Backend: Validate business rules (price within range)
  3. Database: CHECK constraint (price > 0)
```

### 3. Use Transactions for Multi-Step Operations

```sql
-- GOOD: Atomic multi-step operation
BEGIN TRANSACTION;
    INSERT INTO orders (id, total) VALUES (1, 100);
    INSERT INTO order_items VALUES (1, 1, 'Product A', 100);
    UPDATE inventory SET stock = stock - 1 WHERE product = 'Product A';
COMMIT;

-- BAD: Separate operations (can leave inconsistent state)
INSERT INTO orders (id, total) VALUES (1, 100);
-- App crashes here - order created but no items!
INSERT INTO order_items VALUES (1, 1, 'Product A', 100);
```

### 4. Design for Consistency

```sql
-- Include derived data with constraints to maintain consistency
CREATE TABLE orders (
    order_id INT PRIMARY KEY,
    subtotal DECIMAL(10,2),
    tax DECIMAL(10,2),
    total DECIMAL(10,2),
    -- Ensure total = subtotal + tax
    CHECK (total = subtotal + tax)
);
```

---

## Common Pitfalls

### 1. Relying Only on Application Logic

```python
# BAD: Only application checks
def create_user(email):
    if not is_valid_email(email):
        raise ValueError("Invalid email")
    db.execute("INSERT INTO users (email) VALUES (?)", (email,))

# Problem: Direct database access bypasses checks
# Someone could: INSERT INTO users (email) VALUES ('invalid');
```

**Solution:** Add database constraint

```sql
ALTER TABLE users ADD CONSTRAINT email_format 
    CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$');
```

### 2. Ignoring Constraint Violations

```python
# BAD: Silently ignoring errors
try:
    db.execute("UPDATE accounts SET balance = -100 WHERE id = 1")
except:
    pass  # Ignoring the error!

# GOOD: Handle appropriately
try:
    db.execute("UPDATE accounts SET balance = -100 WHERE id = 1")
except IntegrityError as e:
    logger.error(f"Constraint violated: {e}")
    # Take corrective action
```

### 3. Race Conditions with Checks

```sql
-- BAD: Check-then-act (race condition)
-- Thread 1:
SELECT stock FROM products WHERE id = 1;  -- Returns 5
-- Thread 2 also executes here
UPDATE products SET stock = stock - 5 WHERE id = 1;  -- Both succeed, stock = -5!

-- GOOD: Atomic operation with constraint
UPDATE products 
SET stock = stock - 5 
WHERE id = 1 AND stock >= 5;  -- Only succeeds if enough stock
-- Plus: CHECK (stock >= 0) constraint
```

---

## Summary

| Aspect | Description |
|--------|-------------|
| **Definition** | Database remains in valid state after transactions |
| **Key Mechanisms** | Constraints, triggers, stored procedures |
| **Types** | Entity, referential, domain, user-defined integrity |
| **Benefits** | Data accuracy, prevents invalid states |
| **Responsibility** | Shared between application and database |
| **Trade-offs** | Strong consistency vs performance/availability |

---

## Related Topics

- [Atomicity](./db-acid-atomicity.md) - All-or-nothing transactions
- [Isolation](./db-acid-isolation.md) - Concurrent transaction handling
- [Durability](./db-acid-durability.md) - Persistence of committed data
- [Database Constraints and Triggers](../week1/2-db.md)

