# Graph Databases Notes (Generated need to review)

## What are Graph Databases?
- NoSQL database type that uses graph structures for semantic queries
- Data stored as nodes (entities), edges (relationships), and properties
- Unlike relational DBs that use tables/joins, graph DBs store relationships as first-class citizens
- Examples: Neo4j, Amazon Neptune, ArangoDB, OrientDB, JanusGraph

## Why Use Graph Databases?
- **Social networks** - friends, followers, connections (think Twitter, LinkedIn)
- **Recommendation engines** - "people you may know", "similar products"
- **Fraud detection** - finding patterns in transactions
- **Knowledge graphs** - Google's knowledge graph, Wikidata
- **Network/IT operations** - infrastructure dependencies

When you have highly connected data and need to traverse relationships quickly, graph DBs excel.

---

## Data Model

### Nodes (Vertices)
- Represent entities (users, products, locations)
- Can have labels (Person, Company, Product)
- Store properties as key-value pairs

```
Node: User
- id: 123
- name: "Alice"
- age: 28
- city: "SF"
```

### Edges (Relationships)
- Connect nodes together
- Have direction (A → B)
- Have types (FOLLOWS, LIKES, WORKS_AT)
- Can also store properties

```
(Alice)-[FOLLOWS {since: 2020}]->(Bob)
(Alice)-[WORKS_AT {role: "Engineer"}]->(Company)
```

### Properties
- Both nodes and edges can have properties
- Schema-flexible (different nodes can have different properties)

---

## Reading Data

### Graph Traversal
- Start at one or more nodes
- Follow edges based on patterns
- Can go multiple hops deep
- Much faster than SQL joins for connected data

### Example: Find Friends of Friends (Neo4j/Cypher)
```cypher
// Find all friends of Alice
MATCH (alice:User {name: "Alice"})-[:FRIENDS_WITH]->(friend)
RETURN friend.name

// Find friends of friends (2 hops)
MATCH (alice:User {name: "Alice"})-[:FRIENDS_WITH*2]->(fof)
RETURN DISTINCT fof.name

// Find shortest path between two users
MATCH path = shortestPath(
  (alice:User {name: "Alice"})-[:FRIENDS_WITH*]-(bob:User {name: "Bob"})
)
RETURN path
```

### Pattern Matching
Graph DBs excel at pattern matching - finding specific structures in data

```cypher
// Find users who like the same products
MATCH (user1:User)-[:LIKES]->(product:Product)<-[:LIKES]-(user2:User)
WHERE user1.id = 123
RETURN user2, product

// Find influencers (users with >10k followers who follow <100 people)
MATCH (user:User)
WHERE size((user)<-[:FOLLOWS]-()) > 10000
  AND size((user)-[:FOLLOWS]->()) < 100
RETURN user
```

### Time Complexity
- Traversing relationships: **O(1)** per edge (index lookup)
- SQL joins would be O(n log n) or worse
- This is why graph DBs are fast for connected data queries

---

## Writing Data

### Creating Nodes
```cypher
// Create a single user
CREATE (alice:User {name: "Alice", age: 28})
RETURN alice

// Create multiple nodes
CREATE (bob:User {name: "Bob"}),
       (company:Company {name: "TechCorp"})
```

### Creating Relationships
```cypher
// Create relationship between existing nodes
MATCH (alice:User {name: "Alice"}),
      (bob:User {name: "Bob"})
CREATE (alice)-[:FOLLOWS {since: date()}]->(bob)

// Create nodes and relationships in one go
CREATE (alice:User {name: "Alice"})
       -[:WORKS_AT {role: "Engineer"}]->
       (company:Company {name: "TechCorp"})
```

### Updating Data
```cypher
// Update node properties
MATCH (user:User {name: "Alice"})
SET user.age = 29, user.lastLogin = timestamp()

// Add new relationship
MATCH (user:User {name: "Alice"}),
      (product:Product {id: 456})
MERGE (user)-[:PURCHASED {date: date()}]->(product)
```

### Deleting Data
```cypher
// Delete a node (must delete relationships first!)
MATCH (user:User {name: "Alice"})
DELETE user  // ERROR if relationships exist!

// Delete node and all its relationships
MATCH (user:User {name: "Alice"})
DETACH DELETE user

// Delete only specific relationships
MATCH (alice:User {name: "Alice"})-[r:FOLLOWS]->(bob:User {name: "Bob"})
DELETE r
```

---

## Storage/How Data is Physically Stored

### Native vs Non-Native Storage

**Native Graph Storage** (Neo4j, Amazon Neptune)
- Each node/edge stored with direct pointers to adjacent nodes
- Nodes contain pointers to their first relationship
- Relationships contain pointers to start node, end node, previous/next relationships
- Physical storage mirrors logical graph structure

```
Node Record:
[ID | Labels | Properties | FirstRelationship*]

Relationship Record:
[ID | Type | Properties | StartNode* | EndNode* | PrevRel* | NextRel*]
```

**Non-Native** (some use existing storage engines)
- Store graph data in traditional databases
- Use indexes to find relationships
- Not as fast for traversals

### Index-Free Adjacency
- Connected nodes physically point to each other
- Don't need index lookups to traverse relationships
- Each node "knows" its neighbors
- This is what makes graph traversal O(1) per hop

### On-Disk Layout (Neo4j example)
```
graph.db/
  ├── neostore.nodestore.db       # Node records
  ├── neostore.relationshipstore.db # Relationship records
  ├── neostore.propertystore.db   # Properties
  ├── neostore.labeltokenstore.db # Label definitions
  └── schema/                      # Indexes and constraints
```

- Fixed-size records for fast random access
- Variable-length data (strings, arrays) stored separately
- Memory-mapped files for performance

### Caching Strategy
- Keep "hot" nodes/relationships in memory
- LRU cache for frequently accessed paths
- Can configure page cache size
- Transaction logs for durability (WAL-like)

---

## Graph Database Examples

### Neo4j (Most Popular)
```cypher
// Create movie database
CREATE (matrix:Movie {title: "The Matrix", year: 1999})
CREATE (keanu:Actor {name: "Keanu Reeves"})
CREATE (keanu)-[:ACTED_IN {role: "Neo"}]->(matrix)

// Query
MATCH (actor:Actor)-[r:ACTED_IN]->(movie:Movie)
WHERE movie.year > 1995
RETURN actor.name, movie.title, r.role
```

### Amazon Neptune (AWS managed)
- Supports both Gremlin and Cypher query languages
- Fully managed, automated backups
- Read replicas for scaling reads

```gremlin
// Gremlin query example
g.V().hasLabel('User').has('name', 'Alice')
  .out('FOLLOWS')
  .values('name')
```

### Property Graph vs RDF
**Property Graph** (Neo4j, Neptune)
- Nodes, edges, properties
- More intuitive for developers

**RDF/Triple Stores** (Blazegraph, Stardog)
- Subject-Predicate-Object triples
- More for semantic web, ontologies
- SPARQL query language

---

## Performance Characteristics

### Strengths
- **Traversals**: O(1) per edge - insanely fast
- **Pattern matching**: Natural and efficient
- **Variable-depth queries**: "friends of friends of friends..."
- **Recommendation queries**: Find similar users/items

### Weaknesses
- **Aggregate queries** (SUM, AVG over millions of nodes) - slower than columnar DBs
- **Tabular scans** - not optimized for scanning all data
- **Sharding** - harder to partition graphs (where to cut?)
- **Not great for** simple CRUD operations without relationships

### When to Use
✅ Social graphs (Twitter, LinkedIn)
✅ Recommendation engines (Netflix, Amazon)
✅ Fraud detection (financial transactions)
✅ Network topology (infrastructure, IT)
✅ Access control (who can access what)

❌ Simple key-value lookups (use Redis)
❌ Large analytical queries (use columnar DB)
❌ Document storage (use MongoDB)
❌ Financial transactions (use SQL for ACID)

---

## Real-World Example: Twitter's Social Graph

Twitter needs to:
- Store follower/following relationships
- Find "who to follow" recommendations
- Detect spam/fake accounts (graph patterns)
- Show timeline (posts from people you follow)

### Traditional SQL Approach
```sql
-- Find people Alice follows
SELECT u2.* FROM users u1
JOIN follows f ON u1.id = f.follower_id
JOIN users u2 ON f.following_id = u2.id
WHERE u1.username = 'alice';

-- Find friends of friends (NOT followed yet)
SELECT DISTINCT u3.* FROM users u1
JOIN follows f1 ON u1.id = f1.follower_id
JOIN follows f2 ON f1.following_id = f2.follower_id
JOIN users u3 ON f2.following_id = u3.id
WHERE u1.username = 'alice'
  AND u3.id NOT IN (
    SELECT following_id FROM follows WHERE follower_id = u1.id
  );
-- ^^^ This gets SLOW with millions of users!
```

### Graph Database Approach
```cypher
// Find people Alice follows
MATCH (alice:User {username: 'alice'})-[:FOLLOWS]->(following)
RETURN following

// Find friends of friends (NOT followed yet)
MATCH (alice:User {username: 'alice'})-[:FOLLOWS]->()-[:FOLLOWS]->(recommendation)
WHERE NOT (alice)-[:FOLLOWS]->(recommendation)
  AND alice <> recommendation
RETURN recommendation
LIMIT 10
```

Much simpler query, and executes in milliseconds even with millions of users.

---

# Twitter's WTF (Who To Follow) Paper

## Background
- Paper: "WTF: The Who to Follow Service at Twitter" (2013)
- Authors: Pankaj Gupta, Ashish Goel, Jimmy Lin, et al.
- Problem: Recommend new users to follow to increase engagement
- This is a **graph-based recommendation system** at massive scale

## The Problem
- Twitter has hundreds of millions of users
- Need to recommend relevant accounts in real-time
- Users who follow more people → more engagement → stay on platform longer
- Challenge: Social graph is MASSIVE (billions of edges)

## Why This Matters
- Recommendation systems are core to all social platforms
- Graph algorithms at scale is hard
- Shows how theory (graph algorithms) meets practice (engineering constraints)

## Algorithm Overview

### Approach: Personalized PageRank (SALSA)
Twitter uses a variation of PageRank called SALSA (Stochastic Approach for Link Structure Analysis)

### Key Idea
1. Start from accounts user already follows (seed set)
2. Do random walks on the social graph
3. Nodes visited most often = good recommendations
4. Filter out already-followed accounts

### Why Not Simple "Followers of Followers"?
- Too shallow - recommends only 2-hop neighbors
- Misses important accounts (celebrities, news)
- Personalized PageRank goes deeper

### The SALSA Algorithm
```
1. Start with users you follow (seeds)
2. Do random walk:
   - From seed, jump to one of their followers
   - Then jump to someone that person follows
   - Repeat many times
3. Count how often you visit each node
4. Top-visited nodes (excluding already followed) = recommendations
```

### Why This Works
- Captures "who influential people in your network follow"
- Surfaces accounts similar to your interests
- Handles different communities (tech, sports, music)

### Circle of Trust
- Limits random walks to within "circles of trust"
- Prevents recommendations from completely unrelated communities
- E.g., if you follow tech people, won't recommend random celebrity gossip accounts

## System Architecture

### Components
1. **Social Graph Storage** - massive graph database
2. **Candidate Generation** - run SALSA to get candidates
3. **Ranking** - ML model to rank candidates
4. **Filtering** - remove spam, bots, inappropriate accounts
5. **Serving** - real-time API to serve recommendations

### Scale Challenges
- **Billions of edges** in the social graph
- **Millions of users** need recommendations simultaneously
- Need **<100ms latency** for good UX

### Engineering Solutions

**1. Pre-computation**
- Can't run PageRank in real-time for every user
- Pre-compute recommendations offline (batch processing)
- Update periodically (e.g., daily)
- Store in fast lookup cache (Redis, memcached)

**2. Graph Partitioning**
- Split social graph across multiple machines
- Use consistent hashing by user ID
- Challenge: random walks cross partition boundaries

**3. Hybrid Approach**
- **Offline**: Compute base recommendations for all users (batch)
- **Online**: Refine with recent activity (streaming)
- **Real-time signals**: Recent follows, likes, retweets

**4. Caching Layers**
```
User Request
    ↓
CDN Cache (most recently recommended)
    ↓
Application Cache (Redis)
    ↓
Recommendation Service
    ↓
Graph Database
```

## Data Flow Example

```
Alice requests "Who to Follow"
    ↓
1. Get Alice's follows: [Bob, Carol, David]
    ↓
2. Run SALSA random walks starting from {Bob, Carol, David}
    ↓
3. Visit counts: {Eve: 150, Frank: 120, Grace: 80, ...}
    ↓
4. Filter: Remove already followed, spam, etc.
    ↓
5. Rank: Use ML model (engagement prediction)
    ↓
6. Return top 10: [Eve, Frank, Grace, ...]
```

## Machine Learning Layer

### Beyond Graph Structure
SALSA gives candidates, but needs ranking:

**Features Used:**
- Graph features: SALSA score, mutual followers count, distance
- User features: Tweet frequency, follower count, verified status
- Similarity: Tweet topics, hashtags, bio keywords
- Engagement: Predicted probability user will follow this account

**Model:**
- Logistic regression or gradient boosted trees
- Trained on historical follow/unfollow data
- Predict: P(Alice will follow Eve | features)

### Feedback Loop
- User follows recommendation → positive signal → update model
- User ignores/dismisses → negative signal
- Continuous A/B testing to improve

## Key Insights from Paper

### What Works
✅ Personalized PageRank outperforms simple heuristics
✅ Combining graph + ML features improves quality
✅ Pre-computation + caching makes it scalable
✅ Fresh recommendations (update frequently) increase engagement

### Challenges
- Cold start problem (new users have no follows)
  - Solution: Recommend popular accounts in user's region/language
- Filter bubble (only recommend similar accounts)
  - Solution: Add diversity/exploration term
- Spam and fake accounts
  - Solution: Trust/reputation scores

### Impact
- Increased user engagement significantly
- More follows → more content → more time on platform
- Became template for other social platforms

## Graph Database Connection

This is WHY graph databases exist!

**Without Graph DB:**
- Store follows in SQL table
- JOIN queries for traversal
- Scales poorly for multi-hop queries
- Can't do real-time recommendations

**With Graph DB:**
- Natural representation of social graph
- Fast traversals (SALSA random walks)
- Pattern matching for similar users
- Can handle billions of edges

**Twitter's Implementation:**
- Built custom graph storage (FlockDB, later Manhattan)
- Optimized for read-heavy workload
- Sharded by user ID
- In-memory cache for hot data

## Takeaways for System Design

1. **Graph problems need graph solutions** - don't force into SQL
2. **Pre-compute when possible** - real-time is hard at scale
3. **Hybrid offline/online** - batch + streaming
4. **Cache aggressively** - recommendations don't change every second
5. **ML on top of algorithms** - graph algorithms find candidates, ML ranks them
6. **Measure impact** - A/B test everything

---

## Comparison: Graph Query Languages

### Cypher (Neo4j)
```cypher
MATCH (u:User)-[:FOLLOWS]->(friend)-[:FOLLOWS]->(fof)
WHERE u.name = 'Alice'
RETURN fof.name, COUNT(*) as mutual_friends
ORDER BY mutual_friends DESC
LIMIT 10
```

### Gremlin (TinkerPop, Neptune)
```groovy
g.V().has('User', 'name', 'Alice')
  .out('FOLLOWS')
  .out('FOLLOWS')
  .groupCount()
  .order(local).by(values, desc)
  .limit(10)
```

### SPARQL (RDF stores)
```sparql
SELECT ?fof (COUNT(?friend) as ?count)
WHERE {
  ?alice rdf:type :User .
  ?alice :name "Alice" .
  ?alice :follows ?friend .
  ?friend :follows ?fof .
}
GROUP BY ?fof
ORDER BY DESC(?count)
LIMIT 10
```

---

## Summary

### Graph Databases
- Store data as nodes and relationships
- Read: Fast traversals, pattern matching (O(1) per edge)
- Write: CREATE nodes/edges, UPDATE properties
- Storage: Index-free adjacency, direct pointers
- Use cases: Social networks, recommendations, fraud detection

### Twitter WTF Paper
- Problem: Recommend who to follow at massive scale
- Solution: Personalized PageRank (SALSA) + ML ranking
- Architecture: Pre-compute + cache + real-time refinement
- Shows why graph databases matter for real-world problems

### Key Insight
Social networks are graphs. Trying to model them in relational databases leads to slow, complex queries. Graph databases make these queries natural and fast. Twitter's WTF service is a perfect example of graph algorithms + engineering at scale.

