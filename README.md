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