## Mutex

```
Analogy: A single bathroom accessed by multiple people one after the other.
```

Also known as mutual exclusion, is a synchronization primitive that prevents multiple threads from accessing a shared resource or
"critical section" of code simultaneously. It acts as a lock ensuring that only one thread can "own" the mutex at a time, thereby
protecting shared data from race conditions and other access issues.


How mutex works.
- A thread locks the mutex before entering the critical section of the code
- While the mutex is locked, any other thread attempting to access the same critical section will be blocked and must wait
- Once the thread finishes its work in the critical section, it unlocks the mutex, allowing another waiting thread to acquire the lock


Key properties
- Exlusive access: only one thread can hold the lock at any give time.
Synchronization: They are used to coordinate the activities of multople threads to ensure data integrity
Critical sections: The part of the code that is protected by the mutex is called critical section

### Real-world Example: Bank Account Transfer

Suppose you have multiple threads handling money transfers between bank accounts in a multi-threaded application. To prevent inconsistent or incorrect balances (for example, two simultaneous withdrawals that both succeed because each thread reads the old balance before either writes the new value), a mutex is used to protect access to account balances. Only one thread can modify an account's balance at a time, ensuring correct and predictable results.

**Example in Python:**
```python
import threading

balance = 1000
balance_mutex = threading.Lock()

def withdraw(amount):
    global balance
    with balance_mutex:  # Acquire the lock before accessing/modifying the balance
        if balance >= amount:
            balance -= amount
            print(f"Withdrawal successful. New balance: {balance}")
        else:
            print("Insufficient funds.")

# Multiple threads calling withdraw() will be synchronized via the mutex
```
This prevents race conditions by ensuring only one thread can modify the balance at any moment.


**Example in Java:**
```java
public class BankAccount {
    private int balance = 1000;
    private final Object mutex = new Object();

    public void withdraw(int amount) {
        synchronized (mutex) {
            if (balance >= amount) {
                balance -= amount;
                System.out.println("Withdrawal successful. New balance: " + balance);
            } else {
                System.out.println("Insufficient funds.");
            }
        }
    }

    // For demonstration: multiple threads withdrawing
    public static void main(String[] args) {
        BankAccount account = new BankAccount();

        Runnable withdrawTask = () -> account.withdraw(600);

        Thread t1 = new Thread(withdrawTask);
        Thread t2 = new Thread(withdrawTask);

        t1.start();
        t2.start();
    }
}
```

### Notes: Possible ways to implement mutexes in Java

Java provides multiple ways to achieve mutual exclusion (mutex behavior):

1. **`synchronized` keyword**  
   - You can use `synchronized` blocks or methods as shown above (`synchronized(mutex) {...}`).
   - You can also use `synchronized` methods:  
     ```java
     public synchronized void withdraw(int amount) { ... }
     ```
   - Both approaches use the intrinsic lock (monitor) of the specified object.

2. **Explicit Lock Classes (`java.util.concurrent.locks.Lock`)**  
   - Java's concurrent package offers explicit lock classes, such as `ReentrantLock`, which often provide more features (try-lock, lock fairness, etc.).
   - Example:
     ```java
     import java.util.concurrent.locks.Lock;
     import java.util.concurrent.locks.ReentrantLock;

     public class BankAccount {
         private int balance = 1000;
         private final Lock lock = new ReentrantLock();

         public void withdraw(int amount) {
             lock.lock();
             try {
                 if (balance >= amount) {
                     balance -= amount;
                     System.out.println("Withdrawal successful. New balance: " + balance);
                 } else {
                     System.out.println("Insufficient funds.");
                 }
             } finally {
                 lock.unlock();
             }
         }
     }
     ```

3. **Other locking/synchronization mechanisms**  
   - **`java.util.concurrent` atomic classes**: For certain use cases, classes like `AtomicInteger` can be used for lock-free thread-safe access to variables.
   - **Semaphores**: `java.util.concurrent.Semaphore` can be used for more complex locking requirements, but for mutual exclusion, its use resembles a mutex.
   - **Synchronizers:** Advanced utilities, e.g., `ReadWriteLock`, `StampedLock`, can provide mutex-like or more nuanced locking control.
   - **`volatile` keyword:** *Note:* While `volatile` ensures visibility, it does not provide mutual exclusion/mutex behavior.

For most typical mutex needs, `synchronized` and `ReentrantLock` are most commonly used in Java.

---

