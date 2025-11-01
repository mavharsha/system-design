# Two-Phase Commit (2PC)

## 1. Quick Summary

**What it does:** Two-Phase Commit is a distributed algorithm that ensures all participants in a distributed transaction either commit or rollback together, maintaining ACID properties across multiple databases or services.

**Primary use case:** Coordinating transactions across multiple databases/services where all operations must succeed or fail atomically.

**Key features:**
- **Atomic commitment** across distributed systems
- **Coordinator-participant model** for transaction management
- **Blocking protocol** that ensures consistency

---

## 2. What Problem Does It Solve?

### The Distributed Transaction Problem

Imagine you're building a banking system where:
- **Account Service** manages user balances (Database A)
- **Transaction Service** logs all transactions (Database B)
- **Audit Service** records compliance data (Database C)

**Scenario:** Transfer $100 from Alice to Bob

```
Operation 1: Deduct $100 from Alice's account (Database A)
Operation 2: Add $100 to Bob's account (Database A)
Operation 3: Log transaction (Database B)
Operation 4: Create audit record (Database C)
```

**What could go wrong without 2PC:**
- ✅ Operation 1 succeeds → Alice loses $100
- ✅ Operation 2 succeeds → Bob gains $100
- ✅ Operation 3 succeeds → Transaction logged
- ❌ Operation 4 fails → Network issue!

**Result:** Money transferred but no audit trail exists - compliance violation! 💥

### Visual: The Problem

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Database A  │     │  Database B  │     │  Database C  │
│   Accounts   │     │    Logs      │     │    Audit     │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │
       ▼ SUCCESS            ▼ SUCCESS            ▼ FAILURE
   [-$100 Alice]        [Log created]         [FAILED!]
   [+$100 Bob]

Result: INCONSISTENT STATE! Some operations committed, some failed.
```

**2PC Solution:** All databases agree to commit before any actually commit. If any database can't commit, ALL rollback.

---

## 3. Core Concepts

### Architecture Components

```
                    ┌─────────────────┐
                    │   COORDINATOR   │
                    │  (Transaction   │
                    │    Manager)     │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
      ┌──────────┐    ┌──────────┐   ┌──────────┐
      │Participant│    │Participant│   │Participant│
      │    #1     │    │    #2     │   │    #3     │
      │(Database A)│   │(Database B)│  │(Database C)│
      └───────────┘    └───────────┘   └───────────┘
```

### Key Components

1. **Coordinator (Transaction Manager)**
   - Initiates and manages the 2PC protocol
   - Tracks all participants
   - Makes final commit/abort decision
   - Maintains transaction log

2. **Participants (Resource Managers)**
   - Individual databases/services in the transaction
   - Execute local operations
   - Vote on ability to commit
   - Follow coordinator's final decision

### The Two Phases

#### **Phase 1: PREPARE (Voting Phase)**

```
Coordinator                          Participants
     │                                    │
     │───────── PREPARE ────────────────>│
     │                                    │ Execute operations
     │                                    │ Write undo/redo logs
     │                                    │ Lock resources
     │                                    │ Vote: Can I commit?
     │<────── YES/NO (Vote) ─────────────│
     │                                    │
```

**What happens:**
- Coordinator sends PREPARE message to all participants
- Each participant:
  - Executes the transaction operations locally
  - Writes changes to **persistent log** (not yet committed)
  - **Locks all affected resources**
  - Responds with:
    - **YES** (ready to commit) or
    - **NO** (cannot commit - abort)

#### **Phase 2: COMMIT/ABORT (Decision Phase)**

```
Scenario A: All voted YES
────────────────────────────
Coordinator                          Participants
     │                                    │
     │──────── COMMIT ──────────────────>│
     │                                    │ Commit local transaction
     │                                    │ Release locks
     │<────── ACK ───────────────────────│
     │                                    │


Scenario B: Any voted NO
────────────────────────────
Coordinator                          Participants
     │                                    │
     │──────── ABORT ───────────────────>│
     │                                    │ Rollback changes
     │                                    │ Release locks
     │<────── ACK ───────────────────────│
     │                                    │
```

**What happens:**
- If **ALL participants voted YES**:
  - Coordinator sends COMMIT to all participants
  - Each participant commits and releases locks
  - Sends ACK back to coordinator
  
- If **ANY participant voted NO** (or timeout):
  - Coordinator sends ABORT to all participants
  - Each participant rolls back and releases locks
  - Sends ACK back to coordinator

---

## 4. Usage Examples

### Example 1: Basic 2PC Implementation (Java)

```java
public class TwoPhaseCommitCoordinator {
    private List<Participant> participants;
    private TransactionLog log;
    
    public boolean executeTransaction(String transactionId) {
        // PHASE 1: PREPARE
        log.write("BEGIN " + transactionId);
        
        List<Vote> votes = new ArrayList<>();
        for (Participant p : participants) {
            Vote vote = p.prepare(transactionId);
            votes.add(vote);
            
            if (vote == Vote.NO) {
                // Early abort - at least one participant can't commit
                abortTransaction(transactionId);
                return false;
            }
        }
        
        // All voted YES
        log.write("COMMIT " + transactionId);
        
        // PHASE 2: COMMIT
        for (Participant p : participants) {
            p.commit(transactionId);
        }
        
        log.write("COMPLETED " + transactionId);
        return true;
    }
    
    private void abortTransaction(String transactionId) {
        log.write("ABORT " + transactionId);
        
        for (Participant p : participants) {
            p.abort(transactionId);
        }
        
        log.write("ABORTED " + transactionId);
    }
}

enum Vote { YES, NO }
```

### Example 2: Participant Implementation

```java
public class DatabaseParticipant implements Participant {
    private Connection connection;
    private Map<String, Savepoint> savepoints;
    
    @Override
    public Vote prepare(String transactionId) {
        try {
            // Start local transaction
            connection.setAutoCommit(false);
            
            // Execute operations (example)
            PreparedStatement stmt = connection.prepareStatement(
                "UPDATE accounts SET balance = balance - ? WHERE id = ?"
            );
            stmt.setDouble(1, 100.0);
            stmt.setString(2, "Alice");
            stmt.executeUpdate();
            
            // Create savepoint for potential rollback
            Savepoint savepoint = connection.setSavepoint(transactionId);
            savepoints.put(transactionId, savepoint);
            
            // Write to persistent log (WAL)
            writeToLog(transactionId, "PREPARED");
            
            // Vote YES - ready to commit
            return Vote.YES;
            
        } catch (SQLException e) {
            // Cannot complete transaction
            return Vote.NO;
        }
    }
    
    @Override
    public void commit(String transactionId) {
        try {
            connection.commit();
            savepoints.remove(transactionId);
            writeToLog(transactionId, "COMMITTED");
        } catch (SQLException e) {
            // Recovery needed
            handleCommitFailure(transactionId);
        }
    }
    
    @Override
    public void abort(String transactionId) {
        try {
            Savepoint savepoint = savepoints.get(transactionId);
            connection.rollback(savepoint);
            savepoints.remove(transactionId);
            writeToLog(transactionId, "ABORTED");
        } catch (SQLException e) {
            handleRollbackFailure(transactionId);
        }
    }
}
```

### Example 3: Banking Transfer with 2PC

```java
public class BankingService {
    private TwoPhaseCommitCoordinator coordinator;
    
    public boolean transferMoney(String fromAccount, String toAccount, 
                                 double amount) {
        String txnId = UUID.randomUUID().toString();
        
        // Register participants
        List<Participant> participants = Arrays.asList(
            new AccountDatabase(),      // Participant 1
            new TransactionLogDB(),     // Participant 2
            new AuditDatabase()         // Participant 3
        );
        
        coordinator = new TwoPhaseCommitCoordinator(participants);
        
        // Prepare transaction context
        TransactionContext ctx = new TransactionContext(txnId);
        ctx.addOperation(() -> deductFrom(fromAccount, amount));
        ctx.addOperation(() -> creditTo(toAccount, amount));
        ctx.addOperation(() -> logTransaction(txnId, fromAccount, toAccount));
        ctx.addOperation(() -> createAuditRecord(txnId));
        
        // Execute with 2PC
        boolean success = coordinator.executeTransaction(txnId);
        
        if (success) {
            System.out.println("Transfer completed successfully!");
        } else {
            System.out.println("Transfer failed - all changes rolled back");
        }
        
        return success;
    }
}
```

**Expected Output (Success):**
```
[2024-11-01 10:15:23] Transaction tx_12345 - Phase 1: PREPARE
[2024-11-01 10:15:24] Participant AccountDB voted: YES
[2024-11-01 10:15:24] Participant TransactionLog voted: YES
[2024-11-01 10:15:24] Participant AuditDB voted: YES
[2024-11-01 10:15:25] Transaction tx_12345 - Phase 2: COMMIT
[2024-11-01 10:15:26] All participants committed successfully
Transfer completed successfully!
```

**Expected Output (Failure):**
```
[2024-11-01 10:15:23] Transaction tx_12346 - Phase 1: PREPARE
[2024-11-01 10:15:24] Participant AccountDB voted: YES
[2024-11-01 10:15:24] Participant TransactionLog voted: YES
[2024-11-01 10:15:24] Participant AuditDB voted: NO (connection timeout)
[2024-11-01 10:15:25] Transaction tx_12346 - ABORT initiated
[2024-11-01 10:15:26] All participants rolled back
Transfer failed - all changes rolled back
```

---

## 5. Key Points to Remember

### Critical Considerations

⚠️ **Blocking Protocol**
- Resources are **locked during both phases**
- If coordinator crashes, participants remain blocked
- Can lead to **resource starvation** in high-traffic systems

⚠️ **Single Point of Failure**
- Coordinator failure can block entire system
- **Mitigation:** Use persistent logs and recovery protocols

⚠️ **Performance Impact**
- Two round trips of network communication
- Resource locks held longer than single-database transactions
- **Not suitable for high-throughput systems**

### Common Mistakes to Avoid

1. **Not handling timeout properly**
   ```java
   // BAD: Infinite wait
   Vote vote = participant.prepare(txnId);
   
   // GOOD: With timeout
   Vote vote = participant.prepare(txnId, TIMEOUT_MS);
   if (vote == null) {
       // Treat timeout as NO vote
       vote = Vote.NO;
   }
   ```

2. **Forgetting persistent logging**
   - Coordinator MUST log decisions before Phase 2
   - Participants MUST log PREPARE state
   - Enables recovery after crashes

3. **Not implementing recovery protocol**
   ```java
   public void recoverAfterCrash() {
       List<String> uncommittedTxns = log.getUncommittedTransactions();
       
       for (String txnId : uncommittedTxns) {
           // Query participants about their state
           List<State> states = queryParticipantStates(txnId);
           
           if (allPrepared(states)) {
               // Complete the commit
               commitTransaction(txnId);
           } else {
               // Abort the transaction
               abortTransaction(txnId);
           }
       }
   }
   ```

4. **Using 2PC when not necessary**
   - Consider eventual consistency models
   - Use Saga pattern for long-running transactions
   - Explore alternatives like 3PC or Paxos

### Performance Considerations

| Aspect | Impact | Mitigation |
|--------|--------|------------|
| **Latency** | 2 network round trips | Use faster networks, colocate services |
| **Lock duration** | Locks held across network calls | Use shorter timeouts, optimize operations |
| **Coordinator overhead** | Single bottleneck | Use coordinator pools, sharding |
| **Scalability** | Limited by coordinator | Consider eventual consistency instead |

### When to Use 2PC

✅ **Good fit:**
- Financial transactions (transfers, payments)
- Inventory management across warehouses
- Booking systems (flights + hotels + cars)
- Any operation requiring **strict consistency**

❌ **Poor fit:**
- High-throughput systems (e.g., social media likes)
- Geographically distributed systems (high latency)
- Microservices with eventual consistency tolerance
- Long-running business processes (use Sagas instead)

---

## 6. Quick Reference

### State Machine

```
Coordinator States:
─────────────────────
INIT → PREPARING → COMMITTING → COMMITTED
                  ↘ ABORTING  → ABORTED

Participant States:
──────────────────
INIT → PREPARED → COMMITTED
               ↘ ABORTED
```

### Decision Matrix

| Votes Received | Coordinator Decision |
|---------------|---------------------|
| All YES | COMMIT |
| Any NO | ABORT |
| Timeout | ABORT |
| Participant crash | ABORT |

### Recovery Rules

| Coordinator State | Recovery Action |
|------------------|-----------------|
| No log entry | Abort (transaction never started) |
| PREPARE logged | Query participants, abort if uncertain |
| COMMIT logged | Complete commit across all participants |
| ABORT logged | Complete abort across all participants |

| Participant State | Recovery Action |
|------------------|-----------------|
| No log entry | Respond "ABORT" to coordinator query |
| PREPARED logged | Wait for coordinator decision |
| COMMITTED logged | Already committed |
| ABORTED logged | Already aborted |

### Timeout Settings (Example)

```java
// Typical timeout configuration
public class TwoPhaseCommitConfig {
    public static final long PREPARE_TIMEOUT_MS = 5000;    // 5 seconds
    public static final long COMMIT_TIMEOUT_MS = 10000;    // 10 seconds
    public static final long COORDINATOR_RETRY_MS = 3000;  // 3 seconds
    public static final int MAX_RETRY_ATTEMPTS = 3;
}
```

---

## 7. Related Topics

### Alternatives to 2PC

1. **Three-Phase Commit (3PC)**
   - Adds a "pre-commit" phase to reduce blocking
   - More complex, still has limitations
   - Better availability but more overhead

2. **Saga Pattern**
   - Long-running transactions broken into steps
   - Each step has a compensating transaction
   - Better for microservices architecture
   - **Use when:** Eventual consistency acceptable

3. **Distributed Consensus Algorithms**
   - **Paxos:** Complex but proven consensus protocol
   - **Raft:** More understandable alternative to Paxos
   - **Use when:** Need agreement without coordinator

4. **Event Sourcing + CQRS**
   - Store events instead of state
   - Eventual consistency with better scalability
   - **Use when:** High throughput needed

### Implementation in Real Systems

- **MySQL XA Transactions:** Built-in 2PC support
- **PostgreSQL:** Supports prepared transactions (2PC)
- **Java Transaction API (JTA):** Standard for distributed transactions
- **Microsoft DTC:** Distributed Transaction Coordinator for Windows

### Further Reading

- **ACID Properties:** Understanding database consistency
- **Write-Ahead Logging (WAL):** How persistent logs work
- **CAP Theorem:** Trade-offs in distributed systems
- **Distributed Transactions:** Broader context of 2PC
- **Microservices Patterns:** Saga pattern as alternative

---

## Visual Summary

```
┌─────────────────────────────────────────────────────────────┐
│                    TWO-PHASE COMMIT                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Problem: Keep multiple databases consistent                │
│  Solution: All agree before anyone commits                  │
│                                                             │
│  PHASE 1 (PREPARE)          PHASE 2 (COMMIT/ABORT)        │
│  ─────────────────          ──────────────────────         │
│  • Execute operations       • If all YES → COMMIT          │
│  • Write logs              • If any NO → ABORT             │
│  • Lock resources          • Release locks                 │
│  • Vote YES/NO            • Send ACK                       │
│                                                             │
│  Pros:                      Cons:                          │
│  ✓ Strong consistency      ✗ Blocking protocol            │
│  ✓ ACID across databases   ✗ Single point of failure      │
│  ✓ Simple concept          ✗ Performance overhead         │
│                            ✗ Limited scalability          │
│                                                             │
│  Best for: Financial transactions, booking systems         │
│  Avoid for: High-throughput, geo-distributed systems      │
└─────────────────────────────────────────────────────────────┘
```

---

## Complete Flow Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                    SUCCESSFUL 2PC EXECUTION                      │
└──────────────────────────────────────────────────────────────────┘

Time
 │
 │    COORDINATOR              PARTICIPANT A         PARTICIPANT B
 │         │                        │                     │
 ▼         │                        │                     │
 │    ┌────┴────┐                  │                     │
 │    │  BEGIN  │                  │                     │
 │    │   TXN   │                  │                     │
 │    └────┬────┘                  │                     │
 │         │                        │                     │
 │         │──── PREPARE ──────────>│                     │
 │         │──── PREPARE ───────────┼────────────────────>│
 │         │                        │                     │
 │         │                    [Execute]             [Execute]
 │         │                    [Write Log]           [Write Log]
 │         │                    [Lock Data]           [Lock Data]
 │         │                        │                     │
 │         │<──── YES ──────────────│                     │
 │         │<──── YES ───────────────┼─────────────────────│
 │         │                        │                     │
 │    ┌────┴────┐                  │                     │
 │    │Log:COMMIT│                 │                     │
 │    └────┬────┘                  │                     │
 │         │                        │                     │
 │         │──── COMMIT ────────────>│                     │
 │         │──── COMMIT ─────────────┼────────────────────>│
 │         │                        │                     │
 │         │                   [Commit]              [Commit]
 │         │                   [Unlock]              [Unlock]
 │         │                        │                     │
 │         │<──── ACK ──────────────│                     │
 │         │<──── ACK ───────────────┼─────────────────────│
 │         │                        │                     │
 │    ┌────┴────┐                  │                     │
 │    │  DONE   │                  │                     │
 │    └─────────┘                  │                     │
 │                                  │                     │
 │                        ✓ COMMITTED           ✓ COMMITTED
 │

┌──────────────────────────────────────────────────────────────────┐
│                      FAILED 2PC EXECUTION                        │
└──────────────────────────────────────────────────────────────────┘

Time
 │
 │    COORDINATOR              PARTICIPANT A         PARTICIPANT B
 │         │                        │                     │
 ▼         │                        │                     │
 │    ┌────┴────┐                  │                     │
 │    │  BEGIN  │                  │                     │
 │    │   TXN   │                  │                     │
 │    └────┬────┘                  │                     │
 │         │                        │                     │
 │         │──── PREPARE ──────────>│                     │
 │         │──── PREPARE ───────────┼────────────────────>│
 │         │                        │                     │
 │         │                    [Execute]             [FAILED!]
 │         │                    [Write Log]           [Cannot proceed]
 │         │                    [Lock Data]               │
 │         │                        │                     │
 │         │<──── YES ──────────────│                     │
 │         │<──── NO ────────────────┼─────────────────────│
 │         │                        │                     │
 │    ┌────┴────┐                  │                     │
 │    │Log:ABORT│                  │                     │
 │    └────┬────┘                  │                     │
 │         │                        │                     │
 │         │──── ABORT ─────────────>│                     │
 │         │──── ABORT ──────────────┼────────────────────>│
 │         │                        │                     │
 │         │                  [Rollback]            [Cleanup]
 │         │                   [Unlock]              [Unlock]
 │         │                        │                     │
 │         │<──── ACK ──────────────│                     │
 │         │<──── ACK ───────────────┼─────────────────────│
 │         │                        │                     │
 │    ┌────┴────┐                  │                     │
 │    │ ABORTED │                  │                     │
 │    └─────────┘                  │                     │
 │                                  │                     │
 │                        ✗ ROLLED BACK      ✗ ROLLED BACK
 │
```

---

**Remember:** 2PC guarantees consistency but at the cost of availability and performance. Choose wisely based on your system's requirements!

