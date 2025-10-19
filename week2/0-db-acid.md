## DB ACID properties

**Atomicity**
 - All or nothing
 - All opertions in a transaction or nothing.
 - All operations in a transactions are considered as a single logical unit

 ```sql
    BEGIN TRANSACTION

        INSERT INTO orders ...

        UPATE inventory ...

        INSERT INTO order_items ...

    COMMIT;
 ```

When all the operations in the transaction are successful, changes are commited
If any of the operation in the transaction fail, changes are rolled back 

Most DB's use Write ahead logging


----


**Consistency**

Ensures that a transcation bring sthe database fromone valid state to another valid state, maintaining all defined rules, constraints and relations. After a transaction completes, all the data integrity constraints must be satisfied.

```
    Valid state to valid state
```

----

**Durability**

Onces transactions are commited, its changes are permanent and will survive any subsequenet system failures, crashes, power outages. 

All changes are permanent.

Gives Persistance gurantees.


Primary mechanism is supported by WAL.


---- 

**Isolation**

Concurent transactions execute independently without interfering with each other. Each transaction should be unaware of other transactions running simulatneously, and intemendiate states of tranactions should be visible to other tranascrtions.

