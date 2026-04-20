# B-Tree (Balanced Tree)

## Quick Summary

- **What:** A self-balancing tree data structure that maintains sorted data and allows searches, insertions, and deletions in O(log n) time. Unlike binary trees, each node can hold **multiple keys** and have **multiple children**.
- **Primary use case:** Minimizing disk I/O in databases and file systems by keeping trees short and wide
- **Key highlight:** A B-tree of order 1000 can index **1 billion keys** in just 3 levels — that's 3 disk reads vs ~30 for a binary search tree

---

## Analogy: The Filing Cabinet

Imagine a **filing cabinet** where each drawer has labeled dividers:

```
  Filing Cabinet (B-Tree of order 3):

  ┌─────────────────────────────┐
  │  Drawer: [M]                │  ← Root: one divider splits A-L | N-Z
  └──────┬──────────┬───────────┘
         │          │
    ┌────▼────┐  ┌──▼──────┐
    │ [D | H] │  │ [R | W] │     ← Each sub-drawer has dividers too
    └─┬──┬──┬─┘  └─┬──┬──┬─┘
      │  │  │      │  │  │
      ▼  ▼  ▼      ▼  ▼  ▼       ← Actual folders with documents
    A-C D-G I-L  N-Q R-V X-Z
```

**Key insight:** You don't open every drawer. You read the labels (keys), pick the right drawer, and go deeper. Each level narrows your search dramatically — just like a B-tree node narrows which subtree to visit.

In a **binary search tree**, each "drawer" has just ONE divider (one key, two choices). In a B-tree, each drawer can have **hundreds** of dividers — so the cabinet is much shorter, and you reach your folder in fewer steps.

---

## Why B-Trees Exist: The Disk I/O Problem

### The Real Bottleneck

CPU operations take **nanoseconds**. Disk seeks take **milliseconds**. That's a **1,000,000x** difference.

```
  Operation               Time          Relative
  ─────────────────────────────────────────────────
  CPU register access      ~1 ns        1x
  L1 cache hit             ~1 ns        1x
  L2 cache hit             ~4 ns        4x
  RAM access              ~100 ns       100x
  SSD random read      ~100,000 ns      100,000x
  HDD random read   ~10,000,000 ns      10,000,000x
```

Every tree level = one disk read. **Reducing tree height is everything.**

### Binary Search Tree vs B-Tree: Height Comparison

```
  1,000,000 keys:

  Binary Search Tree (BST):              B-Tree (order 1000):
  ┌──────────────────────┐               ┌──────────────────────┐
  │  Height: ~20 levels  │               │  Height: 2 levels    │
  │  20 disk reads       │               │  2 disk reads        │
  │  per lookup!         │               │  per lookup!         │
  └──────────────────────┘               └──────────────────────┘

  Why?
  BST:    each node has 2 children   → height = log₂(1M) ≈ 20
  B-Tree: each node has 1000 children → height = log₁₀₀₀(1M) ≈ 2

  That's 10x fewer disk reads!
```

### Why Not Just Use a Sorted Array?

| Operation | Sorted Array | B-Tree |
|-----------|-------------|--------|
| **Search** | O(log n) — binary search | O(log n) — tree traversal |
| **Insert** | O(n) — shift elements! | O(log n) — split nodes |
| **Delete** | O(n) — shift elements! | O(log n) — merge nodes |

Sorted arrays are great for reads but terrible for writes. B-trees give you O(log n) for **everything**.

---

## Core Concepts

### B-Tree Properties

A B-tree of **minimum degree t** (also called order) has these invariants:

```
  ┌─────────────────────────────────────────────────────────────┐
  │  B-TREE INVARIANTS (minimum degree t)                       │
  ├─────────────────────────────────────────────────────────────┤
  │                                                             │
  │  1. Every node has at most 2t - 1 keys                     │
  │  2. Every node has at most 2t children                     │
  │  3. Every non-root node has at least t - 1 keys            │
  │  4. Every non-root internal node has at least t children   │
  │  5. Root has at least 1 key (if non-empty)                 │
  │  6. ALL leaves are at the same depth (perfectly balanced)  │
  │  7. Keys within a node are sorted                          │
  │                                                             │
  │  Example with t = 3:                                        │
  │  • Max keys per node: 2(3) - 1 = 5                         │
  │  • Max children per node: 2(3) = 6                         │
  │  • Min keys per non-root node: 3 - 1 = 2                  │
  │  • Min children per non-root internal node: 3              │
  └─────────────────────────────────────────────────────────────┘
```

### Node Structure

**Critical difference from B+ trees:** In a B-tree, **every node** (internal AND leaf) stores actual data alongside keys.

```
  B-Tree Node (t = 3, up to 5 keys):

  ┌──────────────────────────────────────────────────────┐
  │  n = 3 (current number of keys)    leaf = false      │
  ├──────────────────────────────────────────────────────┤
  │                                                      │
  │  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐           │
  │  │ K₁  │ │ K₂  │ │ K₃  │ │     │ │     │  Keys     │
  │  │ =10 │ │ =20 │ │ =30 │ │     │ │     │           │
  │  └─────┘ └─────┘ └─────┘ └─────┘ └─────┘           │
  │                                                      │
  │  ┌─────┐ ┌─────┐ ┌─────┐                            │
  │  │ D₁  │ │ D₂  │ │ D₃  │  Data (values/pointers)   │
  │  │=Row1│ │=Row2│ │=Row3│  ← DATA LIVES HERE!       │
  │  └─────┘ └─────┘ └─────┘                            │
  │                                                      │
  │  ┌──┐  ┌──┐  ┌──┐  ┌──┐                             │
  │  │C₀│  │C₁│  │C₂│  │C₃│   Child pointers            │
  │  │  │  │  │  │  │  │  │   (null if leaf)             │
  │  └──┘  └──┘  └──┘  └──┘                             │
  │   ↓     ↓     ↓     ↓                               │
  │  <10  10-20  20-30  >30   Key ranges for children    │
  └──────────────────────────────────────────────────────┘
```

### A Complete B-Tree Example

```
  B-Tree of minimum degree t = 2 (2-3-4 tree) with 12 keys:

                         ┌───────────┐
                         │    [16]   │              Level 0 (Root)
                         │   data:R1 │
                         └─────┬─────┘
                     ┌─────────┴──────────┐
                     ▼                    ▼
              ┌────────────┐       ┌────────────┐
              │  [4 | 8]   │       │ [20 | 24]  │   Level 1
              │ d:R2  d:R3 │       │ d:R4  d:R5 │
              └──┬──┬──┬───┘       └──┬──┬──┬───┘
           ┌─────┘  │  └────┐    ┌────┘  │  └─────┐
           ▼        ▼       ▼    ▼       ▼        ▼
        ┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐
        │[1|2] ││[5|6] ││[9|12]││[17|  ││[21|  ││[25|28]│  Level 2
        │d  d  ││d  d  ││d  d  ││18   ││22   ││d   d  │  (Leaves)
        └──────┘└──────┘└──────┘│d  d  ││d  d  │└──────┘
                                └──────┘└──────┘

  Note: EVERY node stores data (d = data pointer).
  In a B+ tree, only leaf nodes would store data.
```

---

## Operations

### Search

Multi-way search: at each node, find the right key or the right child to descend into.

```
  Search for key = 21:

  Step 1: Root [16]
          21 > 16 → go to right child
                                    
  Step 2: Node [20 | 24]
          21 > 20 and 21 < 24 → go to middle child

  Step 3: Leaf [21 | 22]
          21 == 21 → FOUND! Return data.

  Total: 3 node accesses = 3 disk reads
```

### Insertion

B-trees use **proactive splitting**: if a node is full, split it BEFORE inserting. This avoids backtracking up the tree.

```
  Insert key = 15 into this B-tree (t = 2, max 3 keys per node):

  BEFORE:
                    ┌─────────┐
                    │ [8 | 16]│
                    └──┬──┬──┬┘
               ┌──────┘  │  └──────┐
               ▼         ▼         ▼
          ┌────────┐┌────────┐┌────────┐
          │[1|4|5] ││[9|12|14]│[17|20|25]│
          └────────┘└────────┘└────────┘

  Node [9|12|14] is full (3 keys = 2t-1). We need to insert 15 here.

  STEP 1: Split the full node [9|12|14] BEFORE inserting.
          Middle key (12) gets PROMOTED to the parent.

                    ┌────────────┐
                    │ [8|12|16]  │         ← 12 promoted to parent
                    └┬──┬──┬──┬─┘
              ┌──────┘  │  │  └──────┐
              ▼         ▼  ▼         ▼
         ┌───────┐┌────┐┌────┐┌─────────┐
         │[1|4|5]││ [9]││[14]││[17|20|25]│
         └───────┘└────┘└────┘└─────────┘
                              ↑
                              Insert 15 here

  STEP 2: Insert 15 into node [14]. It has room.

                    ┌────────────┐
                    │ [8|12|16]  │
                    └┬──┬──┬──┬─┘
              ┌──────┘  │  │  └──────┐
              ▼         ▼  ▼         ▼
         ┌───────┐┌────┐┌──────┐┌─────────┐
         │[1|4|5]││ [9]││[14|15]││[17|20|25]│
         └───────┘└────┘└──────┘└─────────┘

  DONE! Key 15 inserted. The tree grew wider, not taller.
```

**When the root splits**, a new root is created — this is the **only** way a B-tree grows taller:

```
  Root [8|12|16] is full. Inserting causes a split:

  BEFORE:                        AFTER:
  ┌────────────┐                 ┌─────┐        ← New root!
  │ [8|12|16]  │                 │ [12]│
  └────────────┘                 └──┬──┘
                              ┌────┘ └────┐
                              ▼           ▼
                          ┌──────┐   ┌──────┐
                          │ [8]  │   │ [16] │
                          └──────┘   └──────┘

  Tree height increased by 1. All leaves are still at the same depth.
```

### Deletion

Deletion has three cases, each ensuring the B-tree invariants are maintained:

```
  ┌──────────────────────────────────────────────────────────┐
  │  DELETION CASES                                          │
  ├──────────────────────────────────────────────────────────┤
  │                                                          │
  │  Case 1: Key is in a LEAF node                           │
  │  → Simply remove it (if node still has ≥ t-1 keys)      │
  │                                                          │
  │  Case 2: Key is in an INTERNAL node                      │
  │  → Replace with predecessor (largest key in left         │
  │    subtree) or successor (smallest key in right          │
  │    subtree), then delete from leaf                       │
  │                                                          │
  │  Case 3: Deletion causes UNDERFLOW (node has < t-1 keys)│
  │  → Option A: Borrow from a sibling (redistribute)       │
  │  → Option B: Merge with sibling + pull down parent key   │
  │                                                          │
  └──────────────────────────────────────────────────────────┘
```

**Example — Deletion with merge (Case 3B):**

```
  Delete key = 5 from this tree (t = 2, min 1 key per non-root node):

  BEFORE:
              ┌──────┐
              │ [8]  │
              └──┬──┬┘
            ┌────┘  └────┐
            ▼            ▼
        ┌──────┐    ┌──────┐
        │[3|5] │    │ [12] │
        └──────┘    └──────┘

  Remove 5 from [3|5] → node becomes [3] (still has ≥ 1 key, OK!)

  AFTER:
              ┌──────┐
              │ [8]  │
              └──┬──┬┘
            ┌────┘  └────┐
            ▼            ▼
        ┌──────┐    ┌──────┐
        │ [3]  │    │ [12] │
        └──────┘    └──────┘

  Now delete key = 3:

  Remove 3 from [3] → node becomes [] — UNDERFLOW!
  Sibling [12] also has minimum keys → can't borrow.
  MERGE: pull parent key 8 down, merge with sibling.

              ┌──────────┐
              │ [8 | 12] │    ← Merged node becomes the new root
              └──────────┘

  Tree height decreased by 1.
```

---

## Java Implementation

```java
public class BTree<K extends Comparable<K>, V> {

    private final int t; // minimum degree
    private Node<K, V> root;

    public BTree(int minDegree) {
        this.t = minDegree;
        this.root = new Node<>(true);
    }

    static class Node<K, V> {
        int n;                     // current number of keys
        boolean leaf;
        K[] keys;
        V[] values;                // data stored in EVERY node
        Node<K, V>[] children;

        @SuppressWarnings("unchecked")
        Node(boolean leaf) {
            this.leaf = leaf;
            this.n = 0;
            // allocated to max size (2t-1 keys, 2t children)
            // actual sizes set during BTree construction
        }
    }

    // ─── Search ───────────────────────────────────────────────

    public V search(K key) {
        return search(root, key);
    }

    private V search(Node<K, V> node, K key) {
        // Find the first key >= search key
        int i = 0;
        while (i < node.n && key.compareTo(node.keys[i]) > 0) {
            i++;
        }

        // Found exact match
        if (i < node.n && key.compareTo(node.keys[i]) == 0) {
            return node.values[i]; // data lives here, in ANY node
        }

        // Key not found and we're at a leaf
        if (node.leaf) {
            return null;
        }

        // Recurse into the appropriate child
        // In a real DB, this is a DISK READ
        return search(node.children[i], key);
    }

    // ─── Insert ───────────────────────────────────────────────

    public void insert(K key, V value) {
        Node<K, V> r = root;

        // If root is full, split it first (tree grows taller)
        if (r.n == 2 * t - 1) {
            Node<K, V> newRoot = new Node<>(false);
            newRoot.children[0] = r;
            splitChild(newRoot, 0);
            root = newRoot;
            insertNonFull(newRoot, key, value);
        } else {
            insertNonFull(r, key, value);
        }
    }

    private void insertNonFull(Node<K, V> node, K key, V value) {
        int i = node.n - 1;

        if (node.leaf) {
            // Shift keys to make room and insert
            while (i >= 0 && key.compareTo(node.keys[i]) < 0) {
                node.keys[i + 1] = node.keys[i];
                node.values[i + 1] = node.values[i];
                i--;
            }
            node.keys[i + 1] = key;
            node.values[i + 1] = value;
            node.n++;
        } else {
            // Find the child to descend into
            while (i >= 0 && key.compareTo(node.keys[i]) < 0) {
                i--;
            }
            i++;

            // If child is full, split it first (proactive split)
            if (node.children[i].n == 2 * t - 1) {
                splitChild(node, i);
                if (key.compareTo(node.keys[i]) > 0) {
                    i++;
                }
            }
            insertNonFull(node.children[i], key, value);
        }
    }

    private void splitChild(Node<K, V> parent, int i) {
        Node<K, V> fullChild = parent.children[i];
        Node<K, V> newChild = new Node<>(fullChild.leaf);

        // Move the upper half of keys to newChild
        newChild.n = t - 1;
        for (int j = 0; j < t - 1; j++) {
            newChild.keys[j] = fullChild.keys[j + t];
            newChild.values[j] = fullChild.values[j + t];
        }
        if (!fullChild.leaf) {
            for (int j = 0; j < t; j++) {
                newChild.children[j] = fullChild.children[j + t];
            }
        }

        fullChild.n = t - 1;

        // Shift parent's children to make room
        for (int j = parent.n; j > i; j--) {
            parent.children[j + 1] = parent.children[j];
        }
        parent.children[i + 1] = newChild;

        // Promote median key to parent
        for (int j = parent.n - 1; j >= i; j--) {
            parent.keys[j + 1] = parent.keys[j];
            parent.values[j + 1] = parent.values[j];
        }
        parent.keys[i] = fullChild.keys[t - 1];     // median key goes UP
        parent.values[i] = fullChild.values[t - 1];  // median data goes UP too
        parent.n++;
    }
}
```

**Key observation:** When we split a node, the median key AND its data are **promoted** to the parent. In a B+ tree, only the key is copied up (data stays in leaves). This is the fundamental structural difference.

---

## B-Tree vs B+ Tree

| Aspect | B-Tree | B+ Tree |
|--------|--------|---------|
| **Data location** | ALL nodes (internal + leaf) | Leaf nodes ONLY |
| **Internal node fanout** | Lower (keys + data take space) | Higher (keys only → more fit per page) |
| **Point query** | Can finish at any level | Always goes to leaf level |
| **Range query** | Must traverse tree (no leaf links) | Follow leaf linked list (fast!) |
| **Tree height** | Slightly taller (lower fanout) | Slightly shorter (higher fanout) |
| **Disk reads for search** | 0 to h (might find early) | Always h (must reach leaf) |
| **Used by** | File systems, some embedded DBs | Most relational databases |

**Why databases prefer B+ trees:** Internal nodes without data → more keys per page → higher fanout → shorter tree → fewer disk reads for the common case. Plus the leaf linked list makes range queries and `ORDER BY` extremely efficient. See [B+ Tree](b-+-tree.md) for the deep dive.

---

## Database and System Usage

### Where Classic B-Trees Are Used

| System | Component | Why B-Tree (not B+) |
|--------|-----------|---------------------|
| **NTFS** | Master File Table | File metadata found at internal nodes without reaching leaves |
| **ext4** | Extent tree | Small extents fit in internal nodes, avoiding leaf traversal |
| **HFS+** | Catalog file | Mixed read patterns benefit from finding data at any level |
| **MongoDB (WiredTiger)** | Internal page format | Data can be returned from internal pages for point queries |
| **SQLite** | Table storage | `WITHOUT ROWID` tables use B-tree (not B+) for the table itself |

### B-Tree for Disk Page Layout

In practice, the minimum degree `t` is chosen so that a node fits exactly in one **disk page**:

```
  Disk page size: 4096 bytes (4KB)
  Key size: 8 bytes (long)
  Data pointer: 8 bytes
  Child pointer: 8 bytes
  Node header: 16 bytes

  Space per key-data pair: 8 + 8 = 16 bytes
  Space per child pointer: 8 bytes

  For a node with n keys:
  16 (header) + n × 16 (key+data) + (n+1) × 8 (children) = 4096
  16 + 16n + 8n + 8 = 4096
  24n = 4072
  n ≈ 169 keys per node

  → t = 85 (minimum degree)
  → Height for 1 billion keys: log₁₆₉(10⁹) ≈ 4 levels
  → 4 disk reads to find ANY key among 1 billion!
```

---

## Key Points to Remember

### Gotchas

- **B-tree ≠ Binary tree.** The "B" stands for Bayer (inventor) or "balanced" or "broad" — not "binary." Each node can have hundreds of children.
- **Minimum degree vs order:** Some textbooks define "order" as maximum children (= 2t), others as maximum keys (= 2t-1). Always clarify which definition is being used.
- **Data in internal nodes hurts fanout.** Storing data alongside keys in every node means fewer keys fit per page, so the tree is taller. This is why most databases use B+ trees instead.
- **Deletion is complex.** Insertion has one tricky case (split). Deletion has three cases with subcases. Most implementations defer actual deletion using tombstones or lazy cleanup.
- **Not cache-friendly for in-memory use.** B-trees are optimized for disk, not CPU cache. For in-memory data, a regular balanced BST (Red-Black tree, AVL) or even a sorted array may be faster.

### Performance Characteristics

```
  ┌────────────────────────────────────────────────────┐
  │  B-TREE PERFORMANCE (n keys, minimum degree t)     │
  ├────────────────────────────────────────────────────┤
  │  Search:  O(t · log_t(n))  comparisons             │
  │           O(log_t(n))      disk reads               │
  │  Insert:  O(t · log_t(n))  comparisons              │
  │           O(log_t(n))      disk reads               │
  │  Delete:  O(t · log_t(n))  comparisons              │
  │           O(log_t(n))      disk reads               │
  │  Space:   O(n)                                      │
  │                                                     │
  │  In practice, t is 100-1000, so log_t(n) is tiny:  │
  │  log₅₀₀(1,000,000,000) ≈ 3.3 → 4 disk reads      │
  └────────────────────────────────────────────────────┘
```

---

## Quick Reference

```
┌─────────────────────────────────────────────────────────────────┐
│                     B-TREE CHEAT SHEET                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Minimum degree: t                                              │
│  Max keys per node: 2t - 1                                      │
│  Max children per node: 2t                                      │
│  Min keys (non-root): t - 1                                     │
│  Min children (non-root internal): t                            │
│                                                                 │
│  Height for n keys: O(log_t(n))                                 │
│  All leaves at same depth: ALWAYS                               │
│  Data in internal nodes: YES (unlike B+ tree)                   │
│  Leaf linked list: NO (unlike B+ tree)                          │
│                                                                 │
│  Insertion: split full nodes BEFORE descending (proactive)      │
│  Deletion: borrow or merge to fix underflow                     │
│  Root split: only way the tree grows taller                     │
│  Root merge: only way the tree shrinks                          │
│                                                                 │
│  REMEMBER:                                                      │
│  • B-tree ≠ Binary tree                                         │
│  • Node size = disk page size (4KB / 8KB / 16KB)                │
│  • Higher t = shorter tree = fewer disk reads                   │
│  • Databases mostly use B+ tree variant (data only in leaves)   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Related Topics

- **[B+ Tree](b-+-tree.md)** — The variant used by most relational databases, where data lives only in leaf nodes
- **[Database Indexes](../db/db-indexes.md)** — How B-trees/B+ trees are used as index structures in SQL databases
- **[Storage Engines](../db/db-storage-engines.md)** — How B-tree-based engines compare to LSM-tree-based engines
- **[B+ Trees vs LSM Trees](../db/b+trees-vs-lsm-tree-db.md)** — Database-level comparison of the two dominant storage approaches
- **[LSM Tree](../db/nosql/lsm-tree.md)** — The write-optimized alternative to B-trees used in NoSQL databases
