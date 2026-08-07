

# MIT 6.824 Sistemas Distribuidos

Este repositorio contiene todos los materiales del curso de [sistemas distribuidos](http://nil.csail.mit.edu/6.824/2020/schedule.html) impartido por el MIT, incluyendo

- apuntes de clase
- artículos para cada clase
- soluciones de los laboratorios escritas en Golang

## Resumen

Conceptos y técnicas generales de sistemas distribuidos introducidos por este curso:

1. MapReduce: procesamiento simplificado de datos en grandes clústeres
- estructura: un master + múltiples workers
- job: cada worker procesa una tarea a la vez
	- tarea map: `map(k1, v1) -> list(k2 v2)`
	- tarea reduce: `reduce(k2, list(v2)) -> list(v2)`
- tolerancia a fallos: el master reinicia las tareas fallidas
- ejemplo
	- conteo de palabras
	- grep distribuido

2. Go, RPC, hilos (threads)

3. GFS: Google File System
- estructura: un master + chunk servers
	- cada archivo se divide en fragmentos de 64MB y se distribuye entre los chunk servers para tolerancia a fallos y rendimiento paralelo
- pasos de lectura
- pasos de *record append* (apendizado de registros)
- consistencia débil: no hay garantía de orden para un mismo registro, un append fallido podría ser visible

4. VM-FT: replicación primaria primario/backup para la tolerancia a fallos de la máquina virtual de VMware
- tipo de fallo: fail-stop
- dos enfoques principales de replicación:
	- state transfer: más simple
	- máquina de estados replicada: menor uso de ancho de banda de red
- entrada de log: número de instrucción, tipo, datos
	- determinista: `val = x + y`
	- no determinista: interrupciones de temporizador, llegada de paquetes de red, etc., pueden causar divergencia
- regla de salida: antes de que el primario envíe una salida, debe esperar a que el backup confirme todas las entradas de log anteriores

5. Control de concurrencia en Go
- channel (canal)
- mutex
- time (tiempo)
- condition variable (variable de condición)
- go memory model (modelo de memoria de Go)

6. Raft: algoritmo de consenso distribuido entendible
- problema de *split brain* (división cerebral)
- idea:
	- election de líder (leader election)
	- replicación de log
	- compactación de log (snapshot)

8. ZooKeeper: coordinación wait-free para sistemas a escala de Internet
- idea: un sistema similar a Raft que permite lecturas desde réplicas locales y puede devolver datos obsoletos (stale data)
- escrituras linearizables: las escrituras de los clientes son serializadas por el líder
- orden FIFO para clientes: cada cliente especifica un orden para sus operaciones (lecturas y escrituras)
- servicio de coordinación de propósito general: un árbol similar a un sistema de archivos de znodes y un conjunto de APIs
	- bloqueo distribuido sin el Herd Effect (Efecto Rebaño)
	- configuration master (maestro de configuración)
	- test-and-set para sistemas distribuidos

9. CRAQ: chain replication con apportioned queries
- *Chain Replication (CR)*: head para escrituras, tail para lecturas. Bueno para la simplicidad y el balance de carga en comparación con Raft
- apportioned queries: lectura desde cualquier nodo intermedio
	- cada réplica almacena una lista de versiones por objeto, una clean (limpia) + otras dirty (sucias)
	- write: crea una nueva versión dirty a medida que la escritura pasa, y se convierte en clean cuando el ACK pasa de vuelta
	- read: responde si la última versión es clean; si es dirty, pregunta al tail por el número de la última versión

10. Aurora: base de datos en la nube de Amazon

11. Frangipani: un sistema de archivos distribuido escalable 
- Petal: un servicio de almacenamiento en disco virtual tolerante a fallos
- rendimiento: caché con *cache coherence protocol* (protocolo de coherencia de caché)
	- scheme write-back (escritura diferida)
	- lock server (LS): `request grant revoke release`
- crash recovery (recuperación ante fallos):
	- write-ahead log en Petal que otras workstations (WS) pueden leer y recuperar
	- lock leases: el LS solo recupera el lock y el log después de que el lease expire

12. Transacciones Distribuidas: control de concurrencia + commit atómico
- ACID: comportamiento correcto de una transacción: Atomicity, Consistency, Isolation (serializable), Durability
- Control de Concurrencia Pesimista: adquirir locks antes del acceso
- Two-phase Locking (2PL): RW locks
- Two-phase Commit (2PC): transaction coordinator, `PREPARE`, `COMMIT`

13. Spanner: base de datos global replicada de Google
- External Consistency: consistencia fuerte, generalmenteusualmente linearizability
- transacción:
	- RW transaction: two-phase commit con participantes replicados mediante Paxos
	- RO transactions: *Snapshot Isolation*, es decir, asignar a cada transacción un time-stamp, y cada réplica almacena múltiples versiones con time-stamp de cada registro. Sin locking ni 2pc, mucho más rápido
- sincronización de relojes
	- sistema de referencia temporal de Google
	- TrueTime: devuelve un intervalo = [earliest, latest]
	- dos reglas: start rule, commit wait

14. FaRM: sistema de almacenamiento distribuido de alto rendimiento con RDMA y OCC 
- Fast RPC:
	- Kernel Bypass: la aplicación interactúa directamente con la NIC, se proporciona el toolkit `DPDK`
	- RDMA (Remote Direct Memory Access) NIC: maneja todas las R/W de las aplicaciones, tiene su propio reliable transfer protocol
- Optimistic Concurrency Control: read - bufferd write - validation - commit / abort
	- execution phase: read - buffered write
	- commit phase: (write) lock - (read) validation - commit backup - commit primary 

15. Spark: una abstracción tolerante a fallos para computación en clúster en memoria
- RDD (Resilient Distributed Datasets): particiones en memory de read-only data
- execution:
	- *transformations*: `map filter join` (lazy/perezosas)
	- *actions*: `count collect save` (instantáneas)
- fault-tolerance: el lineage graph mantiene la dependancy de los RDDs, se re-ejecuta si falla
- application:
	- PageRank
	- ML tasks
	- Generalized MapReduce

16. Memcached en Facebook: la lucha entre performance y consistency con la técnica de `memcached` en la estructura del servidor
- evolution de la architecture de servicios Web
- trade-off entre performance y consistency
- memcache: data cache entre los servidores front-end y back-end
	- look-aside (comparado con el look-through cache)
	- on write: invalidate en lugar de update
	- lease: resalta 2 problemas
		- stale data: memcache mantiene datos obsoletos indefinidamente
		- thundering herd: intensive write y read a la misma key 
	- cold cache warmup (calentamiento de caché fría)

17. COPS: causal consistency escalable para wide-area storage
- ALPS system: Availability, low Latency, Partition-tolerance, high Scalability
- CAP Theorem: no se pueden lograr simultáneamente strong consistency, availability y partition-tolerance
- consistency models: 
```
							    > causal > FIFO
linearizability (strong consistency) > sequential > causal+  
							    > per-key sequential > eventual
```
- *causal+ consistency*: *causal consistency* + *convergent conflict handling* (manejo de conflictos convergentes)
- implementation:
	- overview: client library y data storage nodes por data center
	- *version*: Lamport Logical Clock
	- *dependency*: registra *potential causality*
	- conflict handling: default last-writer-wins 

18. Certificate Transparency (CT)
- man-in-the-middle attack: antes de 1995
- digital signature
- digital certificate: SSL HTTPS TLS
- certificate transparency
	- motivation: algunas certificate authorities (CA) se vuelven malicious
	- idea: permitir que los certificates sean public para audit a través de un public CT log system
	- implementation:
		- Merkle Tree: inclusion proof, consistency proof
		- *gossip*: monitors y browsers comparan la sign tree head (STH) para verificar consistency en caso de fork attack

19. Bitcoin
- decentralized ledger
- block chain
- [la mejor explicación en Youtube](https://www.youtube.com/watch?v=bBC-nXj3Ng4)

20. Blockstack: un esfuerzo para re-decentralizar Internet, p. ej. DNS, PKI, storage backend
- decentralized apps: transferir la ownership de data a las manos del usuario en lugar del service provider
- Public Key Infrastructure (PKI): map names to data locations y public keys, essential system para secure internet apps. Sin embargo, aún no se ha implementado una global PKI
- Zooko's triangle: es difícil lograrachieve una unique-global, human-readable y decentralized PKI simultáneamente
