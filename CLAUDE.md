# CLAUDE.md

## Project Summary

A system design study repository organized as weekly notes from a System Design Master Class course, supplemented with reference materials, interactive mind maps, and AI prompt templates.

## Structure

- **`basics/`** - Concurrency fundamentals: mutex, semaphore, CPU/IO-bound concepts
- **`db/`** - Database deep dives: ACID properties, WAL, storage engines, indexes, two-phase commit; `db/nosql/` covers Cassandra, LSM trees, SSTables, memtables, graph DBs
- **`data-structures/`** - B+ trees, bloom filters
- **`cache/`** - Single-node and distributed cache design, Redis
- **`week1/`** - Online/offline indicators, DB design, caching, scaling, sharding, async workers, Kafka. Contains a Go connection pool implementation (`main.go`)
- **`week2/`** - ACID, DB locking (pessimistic/optimistic), storage-compute separation, read replica lag, NoSQL, Slack real-time messaging
- **`week3/`** - Load balancers, distributed locks, Zookeeper, ID generation (Twitter Snowflake, Amazon centralized), pagination, hot shard problem
- **`week4/`** - CDN, image upload service, Gravatar, hashtag service, unread indicators
- **`week5/`** - Single-node and distributed cache design, word dictionary without DB, fast KV store
- **`week6/`** - LSM tree optimization, multi-tier datastores, data ingestion, live streaming, S3 design
- **`week7/`** - Recent searches, distributed task scheduler, flash sale design
- **`week8/`** - Impression counting
- **`mind-map-interactive/`** - HTML-based interactive mind map viewer with markdown source files for system design, Java, and Angular topics. Uses Node.js for local preview
- **`prompts/`** - Claude Code prompt templates for code review, debugging, architecture, performance optimization, and technical writing

## Content Format

Notes are markdown files following a pattern: Problem Statement, Requirements, Solutions (with trade-offs), Implementation examples, and Key Learnings.

## Tech Used

- **Go** - Connection pool prototype in `week1/main.go`
- **JavaScript/Node.js** - Mind map preview server
- **HTML** - Interactive mind map visualization
- **GitHub Actions** - Deploys mind map via `.github/workflows/deploy-mindmap.yml`
