# Bloom Filter

## Quick Summary

- **What:** A space-efficient probabilistic data structure that tests whether an element is a **member of a set**. It can tell you "definitely NOT in the set" or "PROBABLY in the set" — but never gives false negatives.
- **Primary use case:** Avoiding expensive disk reads in databases by quickly ruling out keys that don't exist in a given file or table
- **Key highlight:** A bloom filter for 1 million keys uses only ~1.2 MB of memory (with <1% false positive rate) — compared to storing the keys themselves which would take ~8 MB+

---

## Analogy: The Nightclub Bouncer

Imagine a nightclub bouncer who has a **fuzzy memory** of the guest list:

```
  ┌──────────────────────────────────────────────────────┐
  │  BOUNCER'S MENTAL CHECKLIST                          │
  │                                                      │
  │  "Is Alice on the list?"                             │
  │  Bouncer checks 3 mental notes... all say YES.       │
  │  → "You're PROBABLY on the list. Come in."           │
  │     (might be wrong — false positive possible)       │
  │                                                      │
  │  "Is Mallory on the list?"                           │
  │  Bouncer checks 3 mental notes... one says NO.       │
  │  → "You're DEFINITELY NOT on the list. Go away."    │
  │     (never wrong about this — no false negatives)    │
  │                                                      │
  └──────────────────────────────────────────────────────┘

  The bouncer is FAST (O(1) check) and CHEAP (tiny memory),
  but occasionally lets in someone who wasn't on the list.
  
  That's exactly what a bloom filter does.
```

---

## Core Concepts

### The Bit Array

A bloom filter is fundamentally a **bit array** of `m` bits, all initialized to 0:

```
  Bit array (m = 16 bits):
  
  Index:  0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15
        ┌──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┐
        │ 0│ 0│ 0│ 0│ 0│ 0│ 0│ 0│ 0│ 0│ 0│ 0│ 0│ 0│ 0│ 0│
        └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┘
```

### Hash Functions

A bloom filter uses `k` independent hash functions. Each hash function maps an element to a position in the bit array:

```
  k = 3 hash functions:

  h₁(element) → position in [0, m-1]
  h₂(element) → position in [0, m-1]
  h₃(element) → position in [0, m-1]
```

### Insert Operation

To insert an element, compute all `k` hash positions and set those bits to 1:

```
  Insert "apple" (k = 3 hash functions):
  
  h₁("apple") = 2
  h₂("apple") = 7
  h₃("apple") = 13

  Index:  0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15
        ┌──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┐
        │ 0│ 0│ 1│ 0│ 0│ 0│ 0│ 1│ 0│ 0│ 0│ 0│ 0│ 1│ 0│ 0│
        └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┘
               ↑                 ↑                 ↑
              h₁               h₂                h₃

  Insert "banana":
  h₁("banana") = 4
  h₂("banana") = 7    ← already 1 (collision with "apple"!)
  h₃("banana") = 11

  Index:  0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15
        ┌──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┐
        │ 0│ 0│ 1│ 0│ 1│ 0│ 0│ 1│ 0│ 0│ 0│ 1│ 0│ 1│ 0│ 0│
        └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┘
               ↑     ↑        ↑        ↑     ↑
            apple  banana   shared  banana  apple
```

### Query Operation

To check if an element exists, compute all `k` hash positions and check if ALL bits are 1:

```
  Query "apple":
  h₁ = 2 → bit[2] = 1  ✓
  h₂ = 7 → bit[7] = 1  ✓
  h₃ = 13 → bit[13] = 1 ✓
  All bits are 1 → "PROBABLY in the set" ✅

  Query "cherry":
  h₁ = 3 → bit[3] = 0  ✗  ← STOP! At least one bit is 0
  → "DEFINITELY NOT in the set" ❌

  Query "mango":
  h₁ = 2 → bit[2] = 1  ✓  (set by "apple")
  h₂ = 4 → bit[4] = 1  ✓  (set by "banana")
  h₃ = 11 → bit[11] = 1 ✓ (set by "banana")
  All bits are 1 → "PROBABLY in the set"
  BUT "mango" was NEVER inserted! → FALSE POSITIVE! ⚠️
```

### Why False Positives Happen

```
  ┌─────────────────────────────────────────────────────────┐
  │  FALSE POSITIVE SCENARIO                                │
  │                                                         │
  │  As more elements are inserted, more bits become 1.     │
  │  Eventually, a never-inserted element's hash positions  │
  │  all happen to land on bits that were set by OTHER      │
  │  elements. The bloom filter can't distinguish this      │
  │  from a real member.                                    │
  │                                                         │
  │  Empty (0% full):  [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]  │
  │  → false positives: ~0%                                 │
  │                                                         │
  │  Half full (50%):  [1 0 1 1 0 1 0 1 1 0 1 0 1 1 0 1]  │
  │  → false positives: ~12.5% (with k=3)                  │
  │                                                         │
  │  Nearly full (90%): [1 1 1 1 0 1 1 1 1 1 1 1 1 1 1 1] │
  │  → false positives: ~72.9% (with k=3)                  │
  │                                                         │
  │  This is why sizing the bit array correctly is critical.│
  └─────────────────────────────────────────────────────────┘
```

### Why No Delete?

```
  Why you can't delete from a standard bloom filter:

  "apple"  hashes to positions: 2, 7, 13
  "banana" hashes to positions: 4, 7, 11
                                   ↑
                                 shared!

  If we delete "apple" by clearing bits 2, 7, 13:
  Bit 7 becomes 0 → but "banana" also uses bit 7!
  Now querying "banana" returns FALSE → false negative!

  Bloom filters GUARANTEE no false negatives.
  Allowing deletes would break this guarantee.

  Solution: Use a Counting Bloom Filter (see Variants section).
```

---

## The Math

### False Positive Probability

After inserting `n` elements into a bit array of size `m` using `k` hash functions:

```
  Probability that a specific bit is still 0 after n insertions:
  
  P(bit = 0) = (1 - 1/m)^(kn) ≈ e^(-kn/m)

  Probability of a false positive (all k bits happen to be 1):
  
  p = (1 - e^(-kn/m))^k

  Where:
  • m = number of bits in the array
  • n = number of inserted elements
  • k = number of hash functions
  • p = false positive probability
```

### Optimal Number of Hash Functions

Given `m` and `n`, the optimal `k` that minimizes false positives:

```
  k_optimal = (m/n) × ln(2) ≈ 0.693 × (m/n)
```

### Optimal Bit Array Size

Given the desired false positive rate `p` and expected elements `n`:

```
  m = -(n × ln(p)) / (ln(2))²

  Bits per element = m/n = -ln(p) / (ln(2))² ≈ -1.44 × ln(p)
```

### Worked Example

```
  ┌─────────────────────────────────────────────────────────┐
  │  SIZING A BLOOM FILTER                                  │
  │                                                         │
  │  Requirements:                                          │
  │  • Expected elements: n = 1,000,000 (1 million keys)   │
  │  • Desired false positive rate: p = 1% (0.01)          │
  │                                                         │
  │  Step 1: Calculate bit array size                       │
  │  m = -(1,000,000 × ln(0.01)) / (ln(2))²               │
  │    = -(1,000,000 × (-4.605)) / (0.4805)                │
  │    = 4,605,000 / 0.4805                                 │
  │    = 9,585,058 bits                                      │
  │    ≈ 9.6 million bits                                    │
  │    ≈ 1.2 MB                                              │
  │                                                         │
  │  Step 2: Calculate optimal hash functions                │
  │  k = (9,585,058 / 1,000,000) × ln(2)                   │
  │    = 9.585 × 0.693                                       │
  │    ≈ 6.64 → round to 7 hash functions                   │
  │                                                         │
  │  Result:                                                 │
  │  • 1.2 MB of memory for 1 million keys                  │
  │  • 7 hash functions                                      │
  │  • <1% false positive rate                               │
  │  • Compare: storing 1M 8-byte keys = 8 MB (6.7x more!) │
  └─────────────────────────────────────────────────────────┘
```

### Quick Sizing Reference

| False Positive Rate | Bits per Element | Hash Functions | Memory for 1M keys |
|--------------------:|:----------------:|:--------------:|:-------------------:|
| 10% | 4.8 | 3 | 0.6 MB |
| 5% | 6.2 | 4 | 0.8 MB |
| 1% | 9.6 | 7 | 1.2 MB |
| 0.1% | 14.4 | 10 | 1.8 MB |
| 0.01% | 19.2 | 13 | 2.4 MB |

---

## Java Implementation

### From-Scratch Implementation

```java
import java.util.BitSet;

public class BloomFilter<T> {
    
    private final BitSet bitArray;
    private final int size;        // m: number of bits
    private final int hashCount;   // k: number of hash functions
    private int insertedCount;     // n: elements inserted

    /**
     * Create a bloom filter sized for expected elements and false positive rate.
     */
    public static <T> BloomFilter<T> create(int expectedElements, double falsePositiveRate) {
        // m = -(n * ln(p)) / (ln(2))^2
        int bits = (int) Math.ceil(
            -(expectedElements * Math.log(falsePositiveRate)) / (Math.log(2) * Math.log(2))
        );
        // k = (m/n) * ln(2)
        int hashes = (int) Math.round(
            ((double) bits / expectedElements) * Math.log(2)
        );
        return new BloomFilter<>(bits, hashes);
    }

    private BloomFilter(int size, int hashCount) {
        this.size = size;
        this.hashCount = hashCount;
        this.bitArray = new BitSet(size);
        this.insertedCount = 0;
    }

    /**
     * Insert an element. Sets k bits to 1.
     */
    public void put(T element) {
        for (int i = 0; i < hashCount; i++) {
            int position = getHash(element, i);
            bitArray.set(position);
        }
        insertedCount++;
    }

    /**
     * Check if an element MIGHT be in the set.
     * Returns false  → DEFINITELY not in set (guaranteed)
     * Returns true   → PROBABLY in set (may be false positive)
     */
    public boolean mightContain(T element) {
        for (int i = 0; i < hashCount; i++) {
            int position = getHash(element, i);
            if (!bitArray.get(position)) {
                return false; // at least one bit is 0 → definitely not present
            }
        }
        return true; // all bits are 1 → probably present
    }

    /**
     * Double hashing technique: h(i) = h1 + i * h2
     * Uses two real hash functions to simulate k independent ones.
     * (Kirsch & Mitzenmacher, 2004 — proven to work as well as k independent hashes)
     */
    private int getHash(T element, int i) {
        int hash1 = element.hashCode();
        int hash2 = spread(hash1); // secondary hash
        int combined = hash1 + (i * hash2);
        return Math.floorMod(combined, size); // always non-negative
    }

    private int spread(int hash) {
        // Murmur-style bit mixing for the secondary hash
        hash ^= (hash >>> 16);
        hash *= 0x85ebca6b;
        hash ^= (hash >>> 13);
        hash *= 0xc2b2ae35;
        hash ^= (hash >>> 16);
        return hash == 0 ? 1 : hash; // avoid zero
    }

    public double estimatedFalsePositiveRate() {
        return Math.pow(1 - Math.exp(-(double) hashCount * insertedCount / size), hashCount);
    }
}
```

**Usage:**

```java
// Create bloom filter for 1 million elements, 1% false positive rate
BloomFilter<String> filter = BloomFilter.create(1_000_000, 0.01);

// Insert elements
filter.put("user:1001");
filter.put("user:1002");
filter.put("user:1003");

// Query
filter.mightContain("user:1001");  // true (correct)
filter.mightContain("user:9999");  // false (definitely not present)
filter.mightContain("user:5555");  // true? (might be a false positive!)
```

### Using Guava's BloomFilter (Production)

For production use, Google Guava provides a battle-tested implementation:

```java
import com.google.common.hash.BloomFilter;
import com.google.common.hash.Funnels;

// Create for 1M strings with 1% FP rate
BloomFilter<String> filter = BloomFilter.create(
    Funnels.stringFunnel(Charset.defaultCharset()),
    1_000_000,   // expected insertions
    0.01         // false positive probability
);

// Insert
filter.put("user:1001");
filter.put("user:1002");

// Query
boolean maybePresent = filter.mightContain("user:1001");  // true
boolean definitelyNot = filter.mightContain("user:9999"); // false

// Check current FP rate
double currentFPP = filter.expectedFpp(); // ~0.0 when few elements inserted

// Merge two bloom filters (same size/hash config)
BloomFilter<String> other = BloomFilter.create(
    Funnels.stringFunnel(Charset.defaultCharset()), 1_000_000, 0.01
);
other.put("user:2001");
filter.putAll(other); // union — filter now contains both sets
```

---

## Variants

### Counting Bloom Filter

Replaces each bit with a **counter** (typically 4 bits). Supports deletions.

```
  Standard Bloom Filter:                Counting Bloom Filter:
  ┌──┬──┬──┬──┬──┬──┬──┬──┐            ┌──┬──┬──┬──┬──┬──┬──┬──┐
  │ 0│ 1│ 1│ 0│ 1│ 0│ 1│ 0│  1 bit    │ 0│ 2│ 1│ 0│ 1│ 0│ 3│ 0│  4 bits
  └──┴──┴──┴──┴──┴──┴──┴──┘  each     └──┴──┴──┴──┴──┴──┴──┴──┘  each

  Insert: increment counters at hash positions
  Delete: decrement counters at hash positions
  Query:  all counters > 0 → probably present

  Trade-off: 4x memory (4 bits vs 1 bit per slot)
  Benefit:   supports delete without false negatives

  ⚠️ Counter overflow: if counter exceeds max (15 for 4-bit),
     you must stop decrementing → potential false negatives.
     In practice, overflow is extremely rare.
```

### Scalable Bloom Filter

When you don't know `n` (expected elements) upfront:

```
  Start with a small bloom filter. When it fills up, add another:

  Filter 1: [1 1 0 1 1 0 1 1]  ← FP rate: 1%
  Filter 2: [0 1 0 0 1 0 0 1]  ← FP rate: 0.5% (tighter)
  Filter 3: [0 0 0 1 0 0 0 0]  ← FP rate: 0.25% (tighter still)

  Insert → always into the latest (active) filter
  Query  → check ALL filters. "Not present" only if ALL say no.

  Each new filter uses a stricter FP rate so the overall
  combined FP rate stays bounded.
```

---

## Database Use Cases

### 1. LSM Tree / SSTable Read Path Optimization

This is the single most important database use case. Each SSTable in an LSM tree has its own bloom filter:

```
  Read key = "user:5000" from an LSM tree with 5 SSTables:

  WITHOUT bloom filters:
  ┌──────────┐
  │ MemTable │ → not found
  └────┬─────┘
       ▼
  ┌──────────┐
  │ SSTable 1│ → binary search index → read data block → not found
  └────┬─────┘
       ▼
  ┌──────────┐
  │ SSTable 2│ → binary search index → read data block → not found
  └────┬─────┘
       ▼
  ┌──────────┐
  │ SSTable 3│ → binary search index → read data block → FOUND!
  └────┬─────┘
       ▼
  ┌──────────┐
  │ SSTable 4│ → still checked (might have newer version)
  └────┬─────┘
       ▼
  ┌──────────┐
  │ SSTable 5│ → still checked
  └──────────┘
  Total: 5 binary searches + 5 disk reads = SLOW


  WITH bloom filters (10 bits/key, 7 hashes → 0.82% FP rate):
  ┌──────────┐
  │ MemTable │ → not found
  └────┬─────┘
       ▼
  ┌──────────┐
  │ BF for   │ → mightContain("user:5000") = false → SKIP! ✓
  │ SSTable 1│
  └────┬─────┘
       ▼
  ┌──────────┐
  │ BF for   │ → mightContain("user:5000") = false → SKIP! ✓
  │ SSTable 2│
  └────┬─────┘
       ▼
  ┌──────────┐
  │ BF for   │ → mightContain("user:5000") = true → CHECK
  │ SSTable 3│ → binary search → read data block → FOUND!
  └────┬─────┘
       ▼
  ┌──────────┐
  │ BF for   │ → mightContain("user:5000") = false → SKIP! ✓
  │ SSTable 4│
  └────┬─────┘
       ▼
  ┌──────────┐
  │ BF for   │ → mightContain("user:5000") = false → SKIP! ✓
  │ SSTable 5│
  └──────────┘
  Total: 5 bloom filter checks (in-memory, ~ns) + 1 disk read = FAST!
```

**Enabling bloom filters in RocksDB (Java):**

```java
import org.rocksdb.*;

BlockBasedTableConfig tableConfig = new BlockBasedTableConfig()
    .setFilterPolicy(new BloomFilter(10, false));  // 10 bits per key

Options options = new Options()
    .setCreateIfMissing(true)
    .setTableFormatConfig(tableConfig);

RocksDB db = RocksDB.open(options, "/data/mydb");
// Now every SSTable gets a bloom filter automatically
// ~99% of unnecessary disk reads are eliminated
```

### 2. Cache Penetration Prevention

When attackers or bugs send requests for keys that **don't exist**, every request misses the cache and hits the database:

```
  WITHOUT bloom filter (cache penetration attack):

  Request: GET /user/nonexistent_id_12345
       │
       ▼
  ┌─────────┐
  │  Cache   │ → MISS (key doesn't exist)
  └────┬─────┘
       ▼
  ┌─────────┐
  │ Database │ → MISS (key doesn't exist either!)
  └─────────┘     but we did an expensive query for nothing
  
  Repeat 1 million times → database overloaded!


  WITH bloom filter:

  Request: GET /user/nonexistent_id_12345
       │
       ▼
  ┌──────────────┐
  │ Bloom Filter │ → mightContain("nonexistent_id_12345") = false
  │ (all valid   │ → DEFINITELY NOT in DB → return 404 immediately
  │  keys loaded)│
  └──────────────┘
  
  Database never touched! Attack neutralized.
```

**Java implementation:**

```java
public class CacheWithBloomFilter {
    
    private final BloomFilter<String> validKeys;
    private final Cache<String, User> cache;
    private final UserRepository db;

    public CacheWithBloomFilter(UserRepository db) {
        this.db = db;
        this.cache = new LRUCache<>(10_000);
        
        // Load all valid keys into bloom filter at startup
        List<String> allIds = db.getAllUserIds();
        this.validKeys = BloomFilter.create(
            Funnels.stringFunnel(Charset.defaultCharset()),
            allIds.size(),
            0.01  // 1% false positive rate
        );
        allIds.forEach(validKeys::put);
    }

    public User getUser(String userId) {
        // Step 1: Bloom filter check (nanoseconds)
        if (!validKeys.mightContain(userId)) {
            return null; // definitely not in DB — skip everything
        }

        // Step 2: Cache check
        User cached = cache.get(userId);
        if (cached != null) {
            return cached;
        }

        // Step 3: Database query (only reached for valid keys)
        User user = db.findById(userId);
        if (user != null) {
            cache.put(userId, user);
        }
        return user;
    }
}
```

### 3. Cassandra / HBase Partition-Level Bloom Filters

Cassandra uses bloom filters at the **partition level** within each SSTable:

```
  Cassandra SSTable file structure:

  ┌──────────────────────────────────────────┐
  │  SSTable File                            │
  ├──────────────────────────────────────────┤
  │  Data.db       — actual rows             │
  │  Index.db      — partition key → offset  │
  │  Filter.db     — BLOOM FILTER            │  ← one per SSTable
  │  Summary.db    — sampled index           │
  │  Statistics.db — metadata                │
  │  CompressionInfo.db                      │
  └──────────────────────────────────────────┘

  Read path for partition key "user:42":
  
  1. Check MemTable → not found
  2. For each SSTable (newest first):
     a. Check Filter.db (bloom filter) → "definitely not here" → skip
     b. If "maybe here" → check Summary.db → Index.db → Data.db
  
  Default FP rate in Cassandra: 1% (configurable per table)
```

```sql
-- Cassandra: tune bloom filter per table
CREATE TABLE users (
    user_id UUID PRIMARY KEY,
    name TEXT,
    email TEXT
) WITH bloom_filter_fp_chance = 0.001;  -- 0.1% FP rate (more memory, fewer false reads)

-- For tables with mostly range queries (less useful bloom filter):
ALTER TABLE time_series 
WITH bloom_filter_fp_chance = 0.1;  -- 10% FP rate (save memory)
```

### 4. Redis Bloom Filter

Redis supports bloom filters via the RedisBloom module:

```
  Redis Bloom Filter Commands:

  BF.RESERVE myfilter 0.01 1000000
  │           │        │    │
  │           │        │    └── Expected elements (n)
  │           │        └─────── Error rate (p)
  │           └──────────────── Filter name
  └──────────────────────────── Create with custom sizing

  BF.ADD myfilter "user:1001"        -- Insert one element
  BF.MADD myfilter "a" "b" "c"      -- Insert multiple elements

  BF.EXISTS myfilter "user:1001"     -- Check one → 1 (probably) or 0 (definitely not)
  BF.MEXISTS myfilter "a" "b" "x"   -- Check multiple

  BF.INFO myfilter                   -- Size, capacity, filters count
```

**Java with Jedis:**

```java
import redis.clients.jedis.JedisPooled;

JedisPooled jedis = new JedisPooled("localhost", 6379);

// Reserve a bloom filter: 1M capacity, 1% FP rate
jedis.sendCommand(Protocol.Command.valueOf("BF.RESERVE"), 
    "url-seen", "0.01", "1000000");

// Add URLs as they're crawled
jedis.sendCommand(Protocol.Command.valueOf("BF.ADD"), 
    "url-seen", "https://example.com/page1");

// Check before crawling a new URL
long result = (long) jedis.sendCommand(Protocol.Command.valueOf("BF.EXISTS"), 
    "url-seen", "https://example.com/page2");

if (result == 0) {
    // Definitely not seen — crawl this URL
} else {
    // Probably already seen — skip (might miss ~1% of new URLs)
}
```

**When to use Redis bloom filter vs application-level:**

| Scenario | Redis BF | Application BF |
|----------|----------|-----------------|
| Multiple services share the filter | ✅ | ❌ (each has its own copy) |
| Need persistence across restarts | ✅ (Redis persistence) | ❌ (must rebuild) |
| Latency-critical (ns matters) | ❌ (network hop) | ✅ (in-process) |
| Very large filter (>1GB) | ✅ (dedicated memory) | ❌ (JVM heap pressure) |

---

## Non-Database Use Cases

| Use Case | What's Being Checked | Why Bloom Filter |
|----------|---------------------|------------------|
| **Web crawler** | "Have I seen this URL before?" | Millions of URLs, can't store all in memory |
| **Spell checker** | "Is this word in the dictionary?" | Fast rejection of non-words |
| **Network router** | "Is this IP in the blocklist?" | Line-speed packet classification |
| **Bitcoin SPV** | "Is this transaction in this block?" | Lightweight clients check without full block |
| **CDN** | "Is this content cached on this edge?" | Route to the right edge node quickly |
| **Malware detection** | "Is this file hash known malware?" | Millions of hashes, fast pre-filter |

---

## Trade-offs

| Aspect | Advantage | Cost |
|--------|-----------|------|
| Space efficiency | ~10 bits/element vs 64+ bits for hash set | Cannot store actual elements |
| Query speed | O(k) = O(1) constant time | k hash computations per query |
| No false negatives | "Definitely not present" is guaranteed | False positives are unavoidable |
| No deletion | Simple implementation, no counters | Must rebuild to remove elements |
| Tunable accuracy | Lower FP rate with more bits | More memory and hash computations |
| Union-able | Can merge two filters with OR | Cannot intersect (no AND operation) |

---

## Key Points to Remember

### Gotchas

- **Size it correctly upfront.** If you insert more elements than planned, the false positive rate increases rapidly. A bloom filter sized for 1M elements at 1% FP will have ~10% FP if you insert 2M elements.
- **Cannot enumerate elements.** A bloom filter tells you "maybe yes / definitely no" — it can't tell you WHAT is in the set. You cannot iterate over the elements.
- **Hash function quality matters.** Poor hash functions cause clustering (uneven bit distribution), which increases false positives beyond the theoretical rate. MurmurHash3 and xxHash are good choices.
- **Not useful for range queries.** Bloom filters answer "is X in the set?" — not "is anything between X and Y in the set?" For range queries, use B+ tree indexes.
- **False positive rate is per-query, not per-filter.** A 1% FP rate means each individual query has a 1% chance of being wrong — not that 1% of the filter is wrong.

### When NOT to Use a Bloom Filter

- **When false positives are unacceptable.** If you absolutely need exact set membership, use a hash set.
- **When you need deletion.** Use a counting bloom filter (4x memory) or a cuckoo filter.
- **When the set is small.** Below ~1000 elements, a regular `HashSet` is simpler and faster.
- **When you need the actual values.** Bloom filters don't store elements — only membership info.

---

## Quick Reference

```
┌─────────────────────────────────────────────────────────────────┐
│                  BLOOM FILTER CHEAT SHEET                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  FORMULAS:                                                      │
│  • Bits needed:    m = -(n × ln(p)) / (ln(2))²                │
│  • Optimal hashes: k = (m/n) × ln(2)                          │
│  • FP probability: p = (1 - e^(-kn/m))^k                      │
│                                                                 │
│  QUICK SIZING (1% FP rate):                                     │
│  • ~10 bits per element, 7 hash functions                       │
│  • 1M elements ≈ 1.2 MB                                        │
│  • 10M elements ≈ 12 MB                                        │
│  • 100M elements ≈ 120 MB                                      │
│                                                                 │
│  GUARANTEES:                                                    │
│  • "Definitely not in set" → ALWAYS correct                    │
│  • "Probably in set" → correct (100-p)% of the time            │
│  • No false negatives: GUARANTEED                               │
│  • No false positives: NOT guaranteed                           │
│                                                                 │
│  OPERATIONS: O(k) = O(1)                                       │
│  • Insert: compute k hashes, set k bits                        │
│  • Query:  compute k hashes, check k bits                      │
│  • Delete: NOT supported (use counting BF)                     │
│                                                                 │
│  DATABASE USE:                                                  │
│  • LSM/SSTable: skip SSTables that don't contain the key       │
│  • Cache: block requests for non-existent keys                 │
│  • Cassandra: BF per SSTable (bloom_filter_fp_chance)           │
│  • Redis: BF.RESERVE, BF.ADD, BF.EXISTS                        │
│                                                                 │
│  REMEMBER:                                                      │
│  • More elements than expected → FP rate spikes                │
│  • 10 bits/key is the sweet spot for most use cases            │
│  • Bloom filters are NOT for range queries                     │
│  • Rebuild (don't modify) when data changes significantly      │
└─────────────────────────────────────────────────────────────────┘
```

---

## Related Topics

- **[LSM Tree](../db/nosql/lsm-tree.md)** — How bloom filters optimize the SSTable read path in write-optimized databases
- **[SSTable](../db/nosql/sstable.md)** — The sorted file format that embeds bloom filters for fast key lookups
- **[Caching Strategies](../week1/3-cache.md)** — Cache penetration prevention using bloom filters
- **[Redis](../cache/2-redis.md)** — Redis-native bloom filter commands via RedisBloom module
- **[Database Indexes](../db/db-indexes.md)** — B+ tree indexes: the deterministic alternative for speeding up reads
- **[B+ Tree](b-+-tree.md)** — The other key data structure in database read optimization
