###  Locking

#### Pessimistic Locking
```
Need for DB locking (fill)

Usecase for using locking:
- Fixed inventory + Contention

How will we handle when all 120 people fling in one flight in the same trip check-in at the same time. With different locks
 - FOR UPDATE NOWAIT
 - FOR UPDATE SKIP LOCKED  
Example: IRCTC, Bookmyshow, flash sale
```

Two types of locks
- shared locks (fill)
- Exclusive locks (fill)

- Details about lock manager (fill)
- Deadlock state detection (fill)

How does Exclusive lock work? with an example of two transactions trying to lock rows
 - T1 has lock on row1, other transactions can't update or read
 - Once T1 commits or rolls back, the row is available for other transactions


### Different locking strategies

| Strategy | Behavior | Error on Lock Conflict | Best For |
|----------|----------|------------------------|----------|
| NOWAIT | Returns immediately | Yes (ERROR 3572) | Real-time systems, fail-fast operations |
| SKIP LOCKED | Skips locked rows | No | Queue processing, job systems |
| Default | Waits up to timeout | Yes if timeout exceeded | Standard transactions |
| Custom Timeout | Waits up to custom time | Yes if timeout exceeded | Flexible wait scenarios |



#### Optimisting locking

Optimistic locking assumes conflicts are rare and checks for conflicts only at commit time, rather than locking rows during reads. 



----
Watch vod
    missed need for locking

Reading:
Shared locking (read by your self. read after understanding exclusive locking)
