## LSM Tree
A Log-Structured Merge-tree (LSM tree) is a specialized data structure commonly used in modern databases and storage systems to efficiently manage large volumes of write operations. Unlike traditional B-trees, LSM trees optimize write performance by batching updates in memory (using a structure like a MemTable) and periodically flushing them to disk in immutable, sorted files known as SSTables. This log-structured approach reduces write amplification and supports high-throughput ingestion workloads.

When data is queried, the LSM tree compares the in-memory contents and on-disk tables to retrieve the latest value. Compaction processes in the background further merge and reorganize these on-disk files to maintain read efficiency and recover storage space. This balance of fast writes and manageable reads makes LSM trees the underpinning of many NoSQL systems like Apache Cassandra, RocksDB, and LevelDB.


- [Link to LSM](./../db/nosql/lsm-tree.md)
- [B+tree vs LSM tree](./../db/b+trees-vs-lsm-tree-db.md)

As part of LSM tree
- [MemTable](./../db/nosql/memtable.md)
- [SSTable](./../db/nosql/sstable.md)
