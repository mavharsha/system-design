# B+ Tree

## Quick Summary

- **What:** A variant of the B-tree where **all data lives in leaf nodes** and internal nodes store only keys for navigation. Leaf nodes are connected via a **doubly linked list** for fast sequential access.
- **Primary use case:** The default index structure in nearly every relational database (PostgreSQL, MySQL InnoDB, Oracle, SQL Server)
- **Key highlight:** A 3-level B+ tree with 16KB pages can index over **1 billion rows** — and range queries like `BETWEEN` or `ORDER BY` just follow the leaf linked list without touching internal nodes again

---

## Analogy: The Encyclopedia

Think of a B+ tree like an **encyclopedia**:

```
  The encyclopedia's TABLE OF CONTENTS:
  ┌──────────────────────────────────────────────┐
  │  Volume A-F ──→ Volume G-L ──→ Volume M-Z   │  ← Internal nodes
  │  (just letters, no actual articles)          │     (keys only, NO data)
  └──────────────────────────────────────────────┘
                      │
                      ▼
  The actual PAGES with articles:
  ┌────────┐   ┌────────┐   ┌────────┐   ┌────────┐
  │Aardvark│──→│ Cat    │──→│ Dog    │──→│ Zebra  │  ← Leaf nodes
  │Apple   │   │ Cow    │   │ Duck   │   │ Zoo    │     (ALL data here)
  │Ant     │   │ Cherry │   │ Deer   │   │ Zinc   │
  └────────┘   └────────┘   └────────┘   └────────┘
      ↑                                       ↑
      └───────── linked together ─────────────┘
                (for browsing A→Z)
```

**The table of contents** (internal nodes) has just enough info to route you to the right volume — no full articles. **The pages** (leaf nodes) have all the actual content, and they're **linked** so you can flip forward page by page.

---

## What Makes B+ Trees Different from B-Trees

```
  B-TREE:                                B+ TREE:
  ┌─────────┐                            ┌─────────┐
  │ [20]    │ ← has data                 │  [20]   │ ← keys ONLY
  │ data=R1 │                            │ no data │
  └──┬───┬──┘                            └──┬───┬──┘
     │   │                                  │   │
  ┌──▼┐ ┌▼──┐                           ┌──▼┐ ┌▼──────┐
  │[10]│ │[30]│ ← has data              │[10]│ │[20|30]│ ← data HERE
  │d=R2│ │d=R3│                          │d=R2│ │d=R1 R3│   (20 repeated!)
  └────┘ └────┘                          └──┬─┘ └───────┘
                                            └──→──→──┘
                                            linked list!
```

| Feature | B-Tree | B+ Tree |
|---------|--------|---------|
| Data location | Every node | Leaf nodes only |
| Internal node contents | Keys + data + child pointers | Keys + child pointers only |
| Leaf linked list | No | Yes (doubly linked) |
| Duplicate keys | No duplicates | Keys may appear in internal nodes AND leaves |
| Point query path | May stop early at internal node | Always reaches a leaf |
| Range query | Must do tree traversal | Follow leaf linked list |
| Internal node fanout | Lower (data takes space) | **Higher** (more keys fit per page) |

**The trade-off that matters:** By removing data from internal nodes, B+ trees fit **far more keys per page**, which means higher fanout, shorter trees, and fewer disk reads.

---

## Fanout and Height Math

This is the math that explains why B+ trees dominate databases. Real numbers with a real page size:

```
  Page size: 16 KB (16,384 bytes)
  Key size: 8 bytes (BIGINT)
  Child pointer: 8 bytes
  Row pointer: 8 bytes (in leaf)
  Page header: 96 bytes (typical)

  ─── Internal Node ───
  Available space: 16,384 - 96 = 16,288 bytes
  Each entry: key (8) + child pointer (8) = 16 bytes
  Plus one extra child pointer: 8 bytes

  Fanout = floor((16,288 - 8) / 16) + 1 = floor(16,280 / 16) + 1
         = 1,017 + 1 = 1,018 children per internal node

  ─── Leaf Node ───
  Available space: 16,288 bytes (minus ~16 bytes for prev/next pointers)
  Usable: 16,272 bytes
  Each entry: key (8) + row pointer (8) = 16 bytes

  Records per leaf = floor(16,272 / 16) = 1,017 records per leaf

  ─── Tree Capacity by Height ───

  Height 1 (root only):      1,017 records
  Height 2 (root + leaves):  1,018 × 1,017 = ~1,035,000 records      (~1M)
  Height 3:                  1,018² × 1,017 = ~1,053,000,000 records  (~1B)
  Height 4:                  1,018³ × 1,017 = ~1,072,000,000,000      (~1T)

  ┌────────────────────────────────────────────────────────────┐
  │  A 3-level B+ tree can index 1 BILLION rows.              │
  │  That's 3 disk reads to find ANY row.                      │
  │  The root is always cached in memory → effectively 2 reads.│
  └────────────────────────────────────────────────────────────┘
```

**Compare with a B-tree:** If each internal node also stores an 100-byte data payload per key, fanout drops to ~140. A 3-level tree only holds ~2.7M keys instead of 1B. That's a **370x** difference.

---

## Disk Page Layout

How a real database organizes a B+ tree leaf page on disk:

```
  16 KB Leaf Page:
  ┌──────────────────────────────────────────────────────────┐
  │  PAGE HEADER (96 bytes)                                  │
  │  ┌────────────────────────────────────────────────────┐  │
  │  │ page_id: 4827                                      │  │
  │  │ page_type: LEAF                                    │  │
  │  │ num_records: 487                                   │  │
  │  │ free_space_offset: 8940                            │  │
  │  │ prev_page: 4826  ←──── doubly linked list          │  │
  │  │ next_page: 4828  ────→                             │  │
  │  │ checksum: 0xA3F2                                   │  │
  │  └────────────────────────────────────────────────────┘  │
  ├──────────────────────────────────────────────────────────┤
  │  SLOT DIRECTORY (grows downward ↓)                       │
  │  ┌──────┬──────┬──────┬──────┬───────┐                  │
  │  │ S[0] │ S[1] │ S[2] │ S[3] │  ...  │  Offset to each │
  │  │ =100 │ =132 │ =164 │ =196 │       │  record          │
  │  └──────┴──────┴──────┴──────┴───────┘                  │
  ├──────────────────────────────────────────────────────────┤
  │  RECORDS (grow upward ↑ from bottom)                     │
  │                                                          │
  │  Offset 100: [key=1001 | ptr → row in heap page 52]     │
  │  Offset 132: [key=1002 | ptr → row in heap page 52]     │
  │  Offset 164: [key=1005 | ptr → row in heap page 78]     │
  │  Offset 196: [key=1008 | ptr → row in heap page 91]     │
  │  ...                                                     │
  │  Offset 8908: [key=9542 | ptr → row in heap page 203]   │
  │                                                          │
  ├──────────────────────────────────────────────────────────┤
  │  FREE SPACE (between slot directory and records)         │
  │  Available for new records                                │
  ├──────────────────────────────────────────────────────────┤
  │  PAGE TRAILER                                            │
  │  checksum, LSN (for WAL recovery)                        │
  └──────────────────────────────────────────────────────────┘

  The slot directory allows records to be added/removed without
  shifting all other records — just update the slot pointer.
```

### Fill Factor

Databases don't fill pages to 100%. PostgreSQL uses a default **fill factor of 90%** for B+ tree index pages:

```
  Fill factor = 90%:
  ┌─────────────────────────────────────────────┐
  │ ████████████████████████████████████░░░░░░░ │
  │ ←──── 90% records ─────→ ←── 10% free ──→  │
  └─────────────────────────────────────────────┘

  Why leave 10% free?
  → Future inserts might land on this page
  → Without free space, insert = page split = expensive
  → With free space, insert = just add to this page = fast

  Trade-off:
  • Lower fill factor → fewer splits, more space wasted
  • Higher fill factor → more splits, less space wasted
  • For read-heavy tables: fill factor 100% (no updates expected)
  • For write-heavy tables: fill factor 50-70%
```

---

## Operations

### Search (Point Query)

```
  Find key = 42 in this B+ tree (t = 2):

                         ┌──────────┐
                         │   [30]   │              Root (internal)
                         └──┬────┬──┘
                      ┌─────┘    └─────┐
                      ▼                ▼
               ┌───────────┐    ┌───────────┐
               │ [10 | 20] │    │ [40 | 50] │   Internal nodes
               └─┬──┬──┬───┘    └─┬──┬──┬───┘
                 │  │  │          │  │  │
                 ▼  ▼  ▼          ▼  ▼  ▼
  Leaves:   [3,8]→[12,15]→[22,28]→[35,38]→[42,47]→[55,60]
                                            ^^^^^^^^
                                            FOUND!

  Step 1: Root [30]. 42 > 30 → go right child.
  Step 2: Node [40|50]. 40 ≤ 42 < 50 → go middle child.
  Step 3: Leaf [42,47]. Found key 42 → return data.

  Disk reads: 3 (root often cached → effectively 2)
```

### Range Scan

This is where B+ trees truly shine. Once you find the start key, you **follow the linked list** — no more tree traversal.

```
  Range query: SELECT * FROM users WHERE id BETWEEN 15 AND 42

  Step 1: Search for start key (15)
          Root → Internal → Leaf [12,15] — found start!

  Step 2: Scan forward through linked list:
  
  [12,15] ──→ [22,28] ──→ [35,38] ──→ [42,47]
      ↑           ↑           ↑           ↑
     15          22,28       35,38        42     ← All returned!
   (start)                              (stop)

  Results: 15, 22, 28, 35, 38, 42

  Disk reads for scan: 4 sequential leaf page reads
  (Sequential I/O — the OS can prefetch these pages!)

  Compare with a B-tree (no linked list):
  → Must do a new tree traversal for EACH range boundary
  → Or do an in-order traversal, going up and down the tree
  → Much more random I/O
```

### Insertion with Split

B+ tree splits have a critical distinction: **leaf split copies the key up**, while **internal node split pushes the key up**.

```
  ═══ LEAF SPLIT (key is COPIED up) ═══

  Insert 25 into full leaf [20,22,28]:

  BEFORE:                           AFTER:
  Parent: [30]                      Parent: [25|30]      ← 25 COPIED up
            │                              /    \
  Leaf: [20,22,28] (full!)        [20,22]→[25,28]       ← 25 stays in leaf too!
                                  
  Why copy? Because the leaf MUST contain the actual data for key 25.
  The parent just needs the key for routing.


  ═══ INTERNAL NODE SPLIT (key is PUSHED up) ═══

  Insert causes internal node [10|20|30|40] to overflow:

  BEFORE:                           AFTER:
  Parent: [50]                      Parent: [30|50]      ← 30 PUSHED up
            │                              /    \
  Internal: [10|20|30|40] (full!)  [10|20] [40]          ← 30 is GONE from here

  Why push? Internal nodes don't store data. The median key (30) is only
  needed in one place — the parent — for routing.


  ┌────────────────────────────────────────────────────────┐
  │  SPLIT RULES:                                          │
  │  • Leaf split:     key COPIED up (stays in both)       │
  │  • Internal split: key PUSHED up (moves to parent)     │
  │  This is a classic interview question!                  │
  └────────────────────────────────────────────────────────┘
```

**Full insertion example:**

```
  Insert 25 into this B+ tree (max 3 keys per node):

  BEFORE:
                    ┌────────┐
                    │  [30]  │
                    └──┬──┬──┘
               ┌──────┘  └──────┐
               ▼                ▼
        ┌────────────┐   ┌────────────┐
        │[10|20|28]  │   │[35|40|50]  │
        └────────────┘   └────────────┘
              ↑ full!

  STEP 1: Navigate to the correct leaf: [10|20|28].
          It's full! Must split.

  STEP 2: Split leaf [10|20|28] at median.
          Left: [10|20], Right: [25|28] (insert 25 into right half)
          COPY median key 25 up to parent.

  AFTER:
                 ┌──────────┐
                 │ [25|30]  │
                 └┬───┬───┬─┘
            ┌─────┘   │   └─────┐
            ▼         ▼         ▼
        ┌───────┐┌───────┐┌──────────┐
        │[10|20]││[25|28]││[35|40|50]│
        └───┬───┘└───┬───┘└──────────┘
            └──→─────┘──→──────┘
              linked list maintained!
```

### Deletion with Merge/Redistribute

```
  Delete 20 from leaf [10|20]. Node has min keys after deletion.

  BEFORE:
              ┌──────────┐
              │  [25|30]  │
              └┬───┬───┬──┘
         ┌─────┘   │   └─────┐
         ▼         ▼         ▼
     ┌───────┐┌───────┐┌──────────┐
     │[10|20]││[25|28]││[35|40|50]│
     └───────┘└───────┘└──────────┘

  Remove 20 → leaf becomes [10] — has only 1 key (underflow with t=2).
  Right sibling [25|28] has spare keys → REDISTRIBUTE.

  Borrow 25 from sibling:
  - Move 25 to left leaf: [10|25]
  - Update parent key from 25 to 28 (new boundary)

  AFTER:
              ┌──────────┐
              │  [28|30]  │         ← parent key updated
              └┬───┬───┬──┘
         ┌─────┘   │   └─────┐
         ▼         ▼         ▼
     ┌───────┐┌──────┐┌──────────┐
     │[10|25]││ [28] ││[35|40|50]│
     └───────┘└──────┘└──────────┘

  If sibling also had minimum keys → MERGE instead:
  - Combine [10] + [25|28] into [10|25|28]
  - Remove separator key 25 from parent
  - Parent may underflow → cascade merge upward
```

### Bulk Loading

When creating an index on an existing table (`CREATE INDEX`), inserting rows one-by-one would cause many splits. **Bulk loading** is much faster:

```
  Bulk Load Process:

  Step 1: Sort all keys
  [3, 8, 12, 15, 22, 28, 35, 38, 42, 47, 55, 60]

  Step 2: Fill leaf pages left-to-right at target fill factor (e.g., 90%)
  ┌──────┐   ┌────────┐   ┌────────┐   ┌────────┐
  │[3,8] │──→│[12,15] │──→│[22,28] │──→│[35,38] │──→ ...
  └──────┘   └────────┘   └────────┘   └────────┘

  Step 3: Build internal nodes bottom-up from leaf boundaries
               ┌──────────┐
               │ [22|42]  │
               └┬───┬───┬─┘
          ┌─────┘   │   └─────┐
          ▼         ▼         ▼
  [3,8]→[12,15]  [22,28]→[35,38]  [42,47]→[55,60]

  Why this is faster:
  • No splits — pages are pre-sized
  • Sequential I/O — write pages in order
  • Bottom-up — internal nodes built once, not incrementally
  • 10-100x faster than individual inserts
```

---

## Java Implementation

```java
public class BPlusTree<K extends Comparable<K>, V> {

    private final int order; // max children per internal node
    private Node root;

    public BPlusTree(int order) {
        this.order = order;
        this.root = new LeafNode();
    }

    // ─── Node hierarchy ──────────────────────────────────────

    abstract class Node {
        abstract V search(K key);
        abstract List<V> rangeQuery(K start, K end);
    }

    class InternalNode extends Node {
        List<K> keys = new ArrayList<>();
        List<Node> children = new ArrayList<>();

        @Override
        V search(K key) {
            // Find the child to descend into
            int i = 0;
            while (i < keys.size() && key.compareTo(keys.get(i)) >= 0) {
                i++;
            }
            return children.get(i).search(key);
        }

        @Override
        List<V> rangeQuery(K start, K end) {
            // Route to the correct child for the start key
            int i = 0;
            while (i < keys.size() && start.compareTo(keys.get(i)) >= 0) {
                i++;
            }
            return children.get(i).rangeQuery(start, end);
        }
    }

    class LeafNode extends Node {
        List<K> keys = new ArrayList<>();
        List<V> values = new ArrayList<>();   // ALL data is here
        LeafNode next;                         // linked list pointer
        LeafNode prev;                         // doubly linked

        @Override
        V search(K key) {
            int idx = Collections.binarySearch(keys, key);
            return idx >= 0 ? values.get(idx) : null;
        }

        @Override
        List<V> rangeQuery(K start, K end) {
            List<V> results = new ArrayList<>();
            LeafNode current = this;

            while (current != null) {
                for (int i = 0; i < current.keys.size(); i++) {
                    K k = current.keys.get(i);
                    if (k.compareTo(start) >= 0 && k.compareTo(end) <= 0) {
                        results.add(current.values.get(i));
                    }
                    if (k.compareTo(end) > 0) {
                        return results; // past the end, stop
                    }
                }
                current = current.next; // follow the linked list!
            }
            return results;
        }
    }

    // ─── Public API ──────────────────────────────────────────

    public V search(K key) {
        return root.search(key);
    }

    /**
     * Range query: returns all values where start <= key <= end.
     * This traverses the leaf linked list — O(log n + result count).
     */
    public List<V> rangeQuery(K start, K end) {
        return root.rangeQuery(start, end);
    }

    public void insert(K key, V value) {
        // Find the target leaf
        LeafNode leaf = findLeaf(key);
        leaf.insertSorted(key, value);

        // If leaf overflows, split upward
        if (leaf.keys.size() >= order) {
            splitLeaf(leaf);
        }
    }

    private LeafNode findLeaf(K key) {
        Node current = root;
        while (current instanceof BPlusTree.InternalNode) {
            InternalNode internal = (InternalNode) current;
            int i = 0;
            while (i < internal.keys.size() && key.compareTo(internal.keys.get(i)) >= 0) {
                i++;
            }
            current = internal.children.get(i);
        }
        return (LeafNode) current;
    }
}
```

**Key implementation details:**
- `InternalNode` has keys + child pointers — no data
- `LeafNode` has keys + values + `next`/`prev` pointers
- `rangeQuery` follows the leaf linked list — this is why `ORDER BY` and `BETWEEN` are fast
- Real databases use a **buffer pool** instead of direct memory access — each `children.get(i)` would be a page fetch from the buffer pool

---

## Concurrency: Latch Crabbing

When multiple threads read/write a B+ tree concurrently, you need **latches** (lightweight locks) to prevent corruption. The standard algorithm is **latch crabbing** (also called **lock coupling**):

```
  Latch Crabbing Protocol:

  ─── READ ───
  1. Acquire shared latch on root
  2. Acquire shared latch on child
  3. Release latch on parent           ← "crab" forward
  4. Repeat until you reach the leaf
  
  Multiple readers can hold shared latches simultaneously.

  ─── WRITE (Insert/Delete) ───
  1. Acquire exclusive latch on root
  2. Acquire exclusive latch on child
  3. If child is "safe" (won't split/merge), release ALL ancestor latches
  4. Continue down to leaf
  
  "Safe" means:
  • For insert: node has room (< max keys)
  • For delete: node has spare keys (> min keys)

  ┌─────────────────────────────────────────────────────┐
  │  Why this works:                                     │
  │  • If a child is safe, any split/merge stays local  │
  │  • Ancestors won't be modified → safe to unlatch    │
  │  • Reduces contention — most nodes ARE safe          │
  │  • Root latch is the bottleneck (briefly held)       │
  └─────────────────────────────────────────────────────┘

  Optimization: OPTIMISTIC latch crabbing
  1. Take shared latches all the way down (assume no split)
  2. At leaf, try exclusive latch for the write
  3. If split IS needed, restart with exclusive latches
  → Works well because splits are rare (~1% of inserts)
```

---

## MVCC and B+ Trees

Modern databases (PostgreSQL, InnoDB) support **Multi-Version Concurrency Control**. Here's how it interacts with B+ trees:

```
  PostgreSQL approach:
  ┌────────────────────────────────────────────┐
  │  Leaf entry for key = 42:                  │
  │  ┌──────┬──────────┬──────────┬─────────┐ │
  │  │ key  │ xmin     │ xmax     │ row ptr │ │
  │  │ 42   │ txn:100  │ txn:150  │ → heap  │ │  ← old version (deleted by txn 150)
  │  │ 42   │ txn:150  │ ∞        │ → heap  │ │  ← current version
  │  └──────┴──────────┴──────────┴─────────┘ │
  │                                            │
  │  A reader at txn:120 sees the first row.   │
  │  A reader at txn:200 sees the second row.  │
  └────────────────────────────────────────────┘

  InnoDB approach:
  ┌────────────────────────────────────────────┐
  │  Leaf entry for key = 42:                  │
  │  ┌──────┬─────────┬──────────────────────┐│
  │  │ key  │ current  │ rollback ptr         ││
  │  │ 42   │ value=B  │ → undo log (value=A)││
  │  └──────┴─────────┴──────────────────────┘│
  │                                            │
  │  Leaf stores only the LATEST version.      │
  │  Older versions are in the undo log.       │
  │  Readers follow the rollback chain to find │
  │  the version visible to their transaction. │
  └────────────────────────────────────────────┘
```

---

## Trade-offs

| Aspect | Advantage | Cost |
|--------|-----------|------|
| High fanout | Fewer levels → fewer disk reads | Larger nodes → more data to read per I/O |
| Leaf linked list | Fast range scans, ORDER BY | Extra pointers, more complex split/merge |
| Keys-only internal nodes | Maximum fanout | Point queries always go to leaf (no early exit) |
| Fill factor < 100% | Fewer page splits | Wasted space |
| Bulk loading | 10-100x faster index creation | Must sort data first |
| Latch crabbing | Concurrent access | Complexity, root latch contention |

---

## Key Points to Remember

### Gotchas

- **Page splits are expensive.** A split writes 3 pages (original, new sibling, parent) and must be logged to WAL. In a high-write workload, frequent splits degrade throughput. Tune fill factor to reduce them.
- **Index bloat.** After many deletes, leaf pages become sparsely filled but aren't returned to the OS. PostgreSQL's `REINDEX` or `pg_repack` rebuild the tree. InnoDB's `OPTIMIZE TABLE` does the same.
- **Random I/O on non-clustered indexes.** Leaf nodes contain row pointers to the heap. If the rows are scattered across many heap pages, each lookup is a random read. A **clustered index** avoids this by storing rows in leaf-order.
- **The copy-up vs push-up distinction matters.** Leaf splits COPY the median key to the parent (data must stay in the leaf). Internal splits PUSH the median key up (data only needed in one place). Getting this wrong corrupts the tree.
- **Too many indexes slow writes.** Every `INSERT`/`UPDATE`/`DELETE` must update ALL B+ tree indexes on the table. A table with 10 indexes means each write triggers 10 tree modifications.

### When B+ Trees Struggle

- **Write-heavy workloads:** Each write is a random I/O (navigate to the right leaf). LSM trees convert this to sequential I/O. See [B+ Trees vs LSM Trees](../db/b+trees-vs-lsm-tree-db.md).
- **Very large range scans over non-clustered indexes:** Each row pointer may point to a different heap page, causing random I/O. The query planner may choose a sequential scan instead.
- **Extremely high concurrency writes to the same index range:** Latch contention on a hot leaf page. Techniques like reverse indexes or hash-partitioned indexes help.

---

## Quick Reference

```
┌─────────────────────────────────────────────────────────────────┐
│                   B+ TREE CHEAT SHEET                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Internal nodes: keys + child pointers (NO data)                │
│  Leaf nodes: keys + data + prev/next pointers                   │
│  All data lives in leaves: ALWAYS                               │
│  Leaf linked list: doubly linked                                │
│                                                                 │
│  SPLIT RULES:                                                   │
│  • Leaf split:     median key COPIED up to parent               │
│  • Internal split: median key PUSHED up to parent               │
│                                                                 │
│  FANOUT (16KB page, 8B key):                                    │
│  • ~1000 children per internal node                             │
│  • Height 3 = ~1 billion rows                                   │
│  • Root cached in memory → 2 disk reads for any lookup          │
│                                                                 │
│  FILL FACTOR:                                                   │
│  • Default 90% (PostgreSQL)                                     │
│  • Lower = fewer splits, more space                             │
│  • Higher = more splits, less space                             │
│                                                                 │
│  CONCURRENCY:                                                   │
│  • Latch crabbing: hold parent latch until child is safe        │
│  • Optimistic: assume no split, retry if wrong                  │
│                                                                 │
│  REMEMBER:                                                      │
│  • Range queries follow leaf linked list (no tree traversal)    │
│  • Bulk load is 10-100x faster than individual inserts          │
│  • Page splits write 3 pages + WAL entry                        │
│  • REINDEX to fix bloat after heavy deletes                     │
└─────────────────────────────────────────────────────────────────┘
```

---

## Related Topics

- **[B-Tree](b%20tree.md)** — The classic variant where data lives in all nodes
- **[Database Indexes](../db/db-indexes.md)** — Index types (clustered, composite, covering, partial) built on B+ trees
- **[B+ Trees in SQL](../week1/db-sql-b+tree.md)** — Practical SQL index usage patterns with B+ trees
- **[B+ Trees vs LSM Trees](../db/b+trees-vs-lsm-tree-db.md)** — When to choose each storage approach
- **[Storage Engines](../db/db-storage-engines.md)** — The role of B+ trees in the database storage layer
- **[WAL](../db/db-wal.md)** — How B+ tree modifications are made durable through write-ahead logging
- **[Bloom Filter](bloom-filter.md)** — The probabilistic data structure used in LSM trees as an alternative read-path optimization
