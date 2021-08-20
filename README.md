# MIT 6.824 Distributed Systems

This repository contains all course materials for [distributed systems](http://nil.csail.mit.edu/6.824/2020/schedule.html) given by MIT, including

- lecture notes
- papers for each lecture
- lab solutions written in Golang

## Terminology

Concepts and general techniques of distributed systems introduced by this course

12. Distributed Transaction
- Pessimistic Concurrency Control: acquire locks before access
- Two-phase Locking: RW locks
- Two-phase Commit: transaction coordinator, `PREPARE`, `COMMIT`
13. Spanner
- External Consistency: strong consistency, usually linearizability
- Snapshot Isolation: assign every transaction a time-stamp
14. FaRM 
- Fast RPC:
	- Kernel Bypass: application directly interacts with NIC, toolkit `DPDK` is provided
	- RDMA (Remote Direct Memory Access) NIC: handles all R/W of applications, has its own reliable transfer protocal
- Optimistic Concurrency Control: read - bufferd write - validation - commit / abort
	- execution phase: read - buffered write
	- commit phase: (write) lock - (read) validation - commit backup - commit primary 
15. Spark
- RDD (Resilient Distributed Datasets): in-memory partitions of read-only data
- execution:
	- *transformations*: `map filter join` lazy
	- *actions*: `count collect save` instant
- fault-tolerance: lineage graph holds dependancy of RDDs, re-execute if fails
- application:
	- PageRank
	- ML tasks
	- Generalized MapReduce

16. Memcached at Facebook
- architecture evolution of Web service
- performance consistency trade-off
- memcache: data cache between front end and back end servers
	- look-aside (compared to look-through cache)
	- on write: invalidate instead of update
	- lease: stress 2 problems
		- stale data: memcache holds stale data indefinitely
		- thundering herd: intensive write and read to the same key 
	- cold cache warmup

17. COPS
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