# MIT 6.824 Distributed Systems

This repository contains all course materials for [distributed systems](http://nil.csail.mit.edu/6.824/2020/schedule.html) given by MIT, including

- lecture notes
- papers for each lecture
- lab solutions written in Golang

## Summary

Concepts and general techniques of distributed systems introduced by this course

1. MapReduce: simplified data processing on large clusters
- structure: single master + multiple workers
- job: each worker processes a task at a time
	- map task: `map(k1, v1) -> list(k2 v2)`
	- reduce task: `reduce(k2, list(v2)) -> list(v2)`
- fault tolerance: master restarts failed task
- example
	- word count
	- distributed grep

2. Go, RPC, threads

3. GFS: Google File System
- structure: single master + chunk servers
	- each file is split into 64MB chunks and spread over chunk servers for fault-tolerance and parallel performance
- read steps
- *record append* steps
- weak consistency: no guarantee for a same record order, failed append could be visible

4. VM-FT: primary/backup replication for fault tolerance of VMware virtual machine
- failure kind: fail-stop
- two main replication approaches:
	- state tranfer: simpler
	- replicated state machine: less network bandwith
- log entry: instruction #, type, data
	- deterministic: `val = x + y`
	- non-deterministic: timer interrupt, network package arrival etc. may cause divergence
- output rule: before primary sends output, must wait for backup to acknowledge all previous log entries


5. Go concurrency control
- channel
- mutex
- time
- condition variable
- go memory model

6. 7.Raft: understandable distributed consensus algorithm
- *split brain* problem
- idea:
	- leader election
	- log replication
	- log compaction (snapshot)

8. ZooKeeper: wait-free coordination for internet-scale systems
- idea: a system like Raft but allows read from local replicas and to yield stale data
- linearizable writes: clients' writes are serialized by the leader
- FIFO client order: each client specifies an order for its operations (reads and writes)
- general-purpose coordination service: a file-system-like tree of znodes and a set of API
	- distributed lock without Herd Effect
	- configuration master
	- test-and-set for distributed systems

9. CRAQ: chain replication with apportioned queries
- *Chain Replication (CR)*: head for write, tail for read. Good for simplicity, load-balancing compared to Raft
-  apportioned queries: read from any intermidiate node
	- each replica stores a list of versions per object, one clean + other dirty
	- write: create new dirty version as write passes through, turn to clean as ACK passes back
	- read: reply if the latest version is clean; if dirty, ask tail for lastest version number

10. Aurora: Amazon's cloud database

11. Frangipani: a scalable distributed file system 
- Petal: a fault-tolerant virtual disk storage service
- performance: caching with *cache coherence protocol*
	- write-back scheme
	- lock server (LS): `request grant revoke release`
- crash recovery:
	- write-ahead log in Petal that other workstations (WS) can read and recover
	- lock leases: LS only takes back lock and recovers log after lease expires
12. Distributed Transaction: concurrency control + atomic commit
- ACID: correct behavior of a transaction: Atomic, Consistent, Isolated (serializable), Durable
- Pessimistic Concurrency Control: acquire locks before access
- Two-phase Locking (2PL): RW locks
- Two-phase Commit (2PC): transaction coordinator, `PREPARE`, `COMMIT`

13. Spanner: Google's global replicated database
- External Consistency: strong consistency, usually linearizability
- transaction:
	- RW transaction: two-phase commit with Paxos-replicated participants
	- RO transactions: *Snapshot Isolation*, i.e. assign every transaction a time-stamp, and each replica stores multiple time-stamped versions of each record. No locking or 2pc, much faster
- clock synchronization
	- Goolge's time reference system
	- TrueTime: return an interval = [earliest, latest]
	- two rules: start rule, commit wait

14. FaRM: high performance distributed storage system with RDMA and OCC 
- Fast RPC:
	- Kernel Bypass: application directly interacts with NIC, toolkit `DPDK` is provided
	- RDMA (Remote Direct Memory Access) NIC: handles all R/W of applications, has its own reliable transfer protocal
- Optimistic Concurrency Control: read - bufferd write - validation - commit / abort
	- execution phase: read - buffered write
	- commit phase: (write) lock - (read) validation - commit backup - commit primary 
15. Spark: a fault-tolerant abstraction for in-memory cluster computing
- RDD (Resilient Distributed Datasets): in-memory partitions of read-only data
- execution:
	- *transformations*: `map filter join` lazy
	- *actions*: `count collect save` instant
- fault-tolerance: lineage graph holds dependancy of RDDs, re-execute if fails
- application:
	- PageRank
	- ML tasks
	- Generalized MapReduce

16. Memcached at Facebook: struggle betwwen performance and consisency with `memcached` technique in server structure
- architecture evolution of Web service
- performance consistency trade-off
- memcache: data cache between front end and back end servers
	- look-aside (compared to look-through cache)
	- on write: invalidate instead of update
	- lease: stress 2 problems
		- stale data: memcache holds stale data indefinitely
		- thundering herd: intensive write and read to the same key 
	- cold cache warmup

17. COPS: scalable causal consistency for wide-area storage
- ALPS system: Availability, low Latency, Partition-tolerance, high Scalability
- CAP Theorem: strong consistency, availability and partition-tolerance can not all be achieved
- consistency models: 
```
							    > causal > FIFO
linearizability (strong consistency) > sequential > causal+  
							    > per-key sequential > eventual
```
- *causal+ consistency*: *causal consistency* + *convergent conflict handling*
- implementation:
	- overview: client library and data storage nodes per data center
	- *version*: Lamport Logical Clock
	- *dependency*: records *potential causality*
	- conflict handling: default last-writer-wins 

18. Certificate Transparency (CT)
- man-in-the-middle attack: before 1995
- digital signature
- digital certificate: SSL HTTPS TLS
- certificate transparency
	- motivation: some certificate authorities (CA) becomes mallicious
	- idea: allow certificates to be public for audit via a public CT log system
	- implementation:
		- Merkle Tree: inclusion proof,consistency proof
		- *gossip*: monitors and browsers compare sign tree head (STH) to check consistency in case of fork attack

19. Bitcoin
- decentralized ledger
- block chain
- [best explaination on Youtube](https://www.youtube.com/watch?v=bBC-nXj3Ng4)

20. Blockstack: an effort to re-decentralize the internet, e.g. DNS, PKI, storage backend
- decentralized apps: move ownership of data into user's hands instead of service provider
- Public Key Infrastructure (PKI): map names to data locations and public keys, essensial system for secure internet apps. Yet no global PKI has been implemented
- Zooko's triangle: unique-global, human-readable, decentralized PKI is hard to achieve 