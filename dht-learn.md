# Distributed Hash Table (DHT) Learning Summary

## 1. DHT Basics

* A Distributed Hash Table (DHT) is a decentralized system that maps keys to values.
* Each node and each key is assigned a numeric ID in a fixed-size ring using a hash function (e.g., SHA-1).
* DHT protocols like **Chord** and **Kademlia** allow efficient lookup, insertion, and deletion of data in a distributed network.

## 2. Hashing Fundamentals

* Node and key IDs are calculated as:

  ```
  node_id = hash(node_name) % 2^m
  key_id  = hash(key) % 2^m
  ```
* In binary, `x % 2^m` is equivalent to keeping the last `m` bits of the hashed number.
* In Go, this is often done by taking the first 8 bytes of the SHA-1 hash (truncation):

  ```go
  binary.BigEndian.Uint64(hash[:8])
  ```
* Although not the same as `% 2^m`, it is consistent and efficient.

## 3. Node Responsibilities

* Each node stores a portion of the total key space.
* When a key is hashed, it is stored at the first node clockwise in the ring with `node_id >= key_id` (successor).
* Nodes maintain:

  * A **local key-value store**: `map[keyID]value`
  * A **routing table** (depends on protocol): `map[nodeID]NodeInfo`

## 4. Server and Client Roles

* Every DHT node acts as both a **client** and a **server**:

  * As a server, it handles lookups and stores key-value data.
  * As a client, it initiates requests to store/retrieve data.

## 5. Key Generation from Content

* The thing being hashed to produce the key depends on the application:

  * Raw content (e.g., IPFS, Git): `key = hash(file_contents)`
  * Metadata (e.g., BitTorrent): `key = hash(info_section)`
  * Logical identifier (e.g., "user:123"): `key = hash(name)`

## 6. Human-Friendly Mapping

* Since humans don’t use hashes, DHTs require a **mapping layer**:

  * Maps human-readable names to keys
  * Can be local, distributed (e.g., IPNS), or centralized (e.g., search index)

## 7. Practical Observations

* Production DHTs (e.g., Kademlia, IPFS, Chord):

  * Use consistent key derivation via hash truncation
  * Rarely perform `% 2^m` directly
  * Rely on uniform hash functions for even distribution

---

This summary reflects our step-by-step learning journey through the fundamentals of DHTs, hashing, node responsibilities, and key-based content addressing.
