## Semaphore

```
Analogy: Ticket counter with limited number of seats.
```

Semaphore is a variable used in computer science to corntrol the access to a shre resource by multiple threads or processes, preventing conflicts in a concurrent system. It works by using two atomic opertions, typically
- wait
- signal
to increment or decrement a counter, which manages how many process can access a resource at one time.
Terms comes from a visual siginlaing system, such as those used in railroads or by flags to send messages.

Functional Syncrhonization primitive that acts like a counter, ensuring only a specified number of process can access a shared resource simultaniously

Operations:
- Wait: Decrements the semaphore count. If the count is greater than zero, the process continues.
- signal: increments the semaphore count. If the process are waiting, one of them is woken up

Used to solve critical section problems, such as producer consumer problem, to prevent race conditions and ensure the integrity of the shared data.

Types of semaphores
- Binary: only allowing one process at a time, like a lock
- Counting: allowing a specific number of process to access



### Real-world Example: Parking Lot (Semaphore in Java)

Imagine a parking lot with a limited number of spaces. Multiple cars (threads) may try to enter (park) at the same time. A semaphore ensures that no more than the allowed number of cars are inside the parking lot.

**Counting Semaphore Example:**
Allows up to a fixed number of threads to access a resource (e.g. 3 parking spots).

```java
import java.util.concurrent.Semaphore;

public class ParkingLot {
    // 3 spots available
    private final Semaphore semaphore = new Semaphore(3);

    public void parkCar(String carName) {
        try {
            System.out.println(carName + " is trying to park.");
            semaphore.acquire(); // Wait (decrement)
            System.out.println(carName + " has parked. Spots left: " + semaphore.availablePermits());
            // Simulate parking
            Thread.sleep(2000);
            System.out.println(carName + " is leaving the parking lot.");
        } catch (InterruptedException e) {
            e.printStackTrace();
        } finally {
            semaphore.release(); // Signal (increment)
        }
    }

    public static void main(String[] args) {
        ParkingLot lot = new ParkingLot();

        // Simulate 5 cars trying to park
        for (int i = 1; i <= 5; i++) {
            String carName = "Car" + i;
            new Thread(() -> lot.parkCar(carName)).start();
        }
    }
}
```
This ensures that only 3 cars can be in the parking lot at any time. Others wait for a spot.

---

**Binary Semaphore Example:**
A binary semaphore (permit = 1) is similar to a mutex; only one thread can access the critical section at a time.

```java
import java.util.concurrent.Semaphore;

public class OneAtATimeResouce {
    private final Semaphore binarySemaphore = new Semaphore(1); // Binary semaphore

    public void accessResource(String threadName) {
        try {
            System.out.println(threadName + " is trying to access the resource.");
            binarySemaphore.acquire();
            System.out.println(threadName + " has entered the critical section.");
            // Simulate work
            Thread.sleep(1500);
            System.out.println(threadName + " is leaving the critical section.");
        } catch (InterruptedException e) {
            e.printStackTrace();
        } finally {
            binarySemaphore.release();
        }
    }

    public static void main(String[] args) {
        OneAtATimeResouce demo = new OneAtATimeResouce();
        for (int i = 1; i <= 3; i++) {
            String threadName = "Thread" + i;
            new Thread(() -> demo.accessResource(threadName)).start();
        }
    }
}
```

---

### All the Ways Java Supports Semaphores

1. **`java.util.concurrent.Semaphore`**  
   - Supports both counting and binary semaphores (by specifying number of permits).
   - Can be created as *fair* (FIFO) or *non-fair* (default), e.g. `new Semaphore(permits, true)`.

2. **Binary semaphore vs counting:**  
   - *Binary semaphore:* `new Semaphore(1)`
   - *Counting semaphore:* `new Semaphore(N)`, where N > 1

3. **Alternatives:**  
   - For simple mutual exclusion, Java’s `synchronized`, `Lock`, and `ReentrantLock` are alternatives (see the "Mutex" section), but true counting semaphores are best handled with `Semaphore`.
   - You can create semaphore-like behavior with lower-level thread or wait/notify constructs, but this is strongly discouraged—use `Semaphore` for clarity and safety.

##### Summary Table

| Mechanism             | Type           | Usage                                                     |
|-----------------------|----------------|-----------------------------------------------------------|
| `Semaphore(1)`        | Binary         | One-at-a-time access (mutex/semaphore)                    |
| `Semaphore(N)`        | Counting       | N-at-a-time access (e.g. database pool, parking lot)      |
| `Semaphore(permits,true)` | Fair      | FIFO order (first come first served)                      |
| `Semaphore(permits)`  | Default (unfair) | May not guarantee FIFO                            |

---


