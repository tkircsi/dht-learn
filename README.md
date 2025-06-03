# DHT Workspace Overview

This workspace contains several projects for learning and experimenting with Distributed Hash Table (DHT) concepts in Go. Each project builds on the previous, adding new features and complexity.

---

## dht-store
A minimal, local key-value store with JSON persistence. Used as a foundation for content storage in DHT nodes.
- **Features:**
  - Thread-safe in-memory map
  - Persistent storage to a JSON file
  - Simple `Put` and `Get` API

**Build:**
```sh
cd dht-store
go build
```

**Usage:**
```sh
# Store a value by key
./dht-store put mykey "my value"

# Retrieve a value by key
./dht-store get mykey
```
- Data is persisted to a local JSON file (`store.json`) in the dht-store directory.
- The CLI supports `put <key> <value>` and `get <key>` commands.

---

## dht-server
A simple HTTP server that exposes DHT-like content storage and retrieval endpoints.
- **Features:**
  - `/put` and `/get` endpoints for storing and retrieving content
  - Name-to-key mapping (optional)
  - Local-only storage (no peer discovery or DHT routing)
  - JSON persistence

**Build:**
```sh
cd dht-server
go build
```

**Run:**
```sh
./dht-server :8080
```

**API Usage:**
- Store content:
  ```sh
  curl -X POST -d '{"key":"mykey","value":"bXl2YWx1ZQ=="}' localhost:8080/put
  # (where value is base64-encoded)
  ```
- Retrieve content:
  ```sh
  curl 'localhost:8080/get?key=mykey'
  ```

---

## dht-network
A minimal DHT peer discovery and networking implementation (Kademlia-style), but without content storage.
- **Features:**
  - Peer discovery and routing table (XOR distance)
  - Kademlia-style `/find_node` endpoint
  - Bootstrap and join logic
  - No content storage or retrieval endpoints

**Build:**
```sh
cd dht-network
go build
```

**Run:**
```sh
# Start first node
./dht-network 127.0.0.1:8081
# Start second node, joining the first
./dht-network --bootstrap 127.0.0.1:8081 127.0.0.1:8082
```

**API Usage:**
- Query peers:
  ```sh
  curl localhost:8081/peers
  ```
- Find closest node:
  ```sh
  curl 'localhost:8081/find_node?target=<node_id>'
  ```

---

## dht-node
A full DHT node combining peer discovery, routing, and content storage.
- **Features:**
  - DHT peer discovery and routing (from dht-network)
  - Local key-value store with JSON persistence (from dht-store)
  - `/put` and `/get` endpoints with DHT-based routing: requests are forwarded to the node responsible for the key
  - Each node uses a unique store file (by node ID) for local persistence
  - Foundation for further DHT features (replication, value lookup, etc.)

**Build:**
```sh
cd dht-node
go build
```

**Run:**
```sh
# Start first node
./dht-node 127.0.0.1:8081
# Start second node, joining the first
./dht-node --bootstrap 127.0.0.1:8081 127.0.0.1:8082
```

**API Usage:**
- Store content (DHT-routed):
  ```sh
  # Compute a key (e.g., first 16 hex chars of SHA-1 of content)
  echo -n "hello world" | shasum | awk '{print substr($1,1,16)}'
  # Use the key in the request
  curl -X POST -d '{"key":"2aae6c35c94fcfb4","value":"aGVsbG8gd29ybGQ="}' localhost:8081/put
  ```
- Retrieve content (DHT-routed):
  ```sh
  curl 'localhost:8082/get?key=2aae6c35c94fcfb4'
  ```
- Query peers:
  ```sh
  curl localhost:8081/peers
  ```

---

Each project is self-contained and can be run independently for experimentation and learning. `dht-node` is the most complete, combining all previous features for a realistic DHT node experience.

---

## DHT Learning Goals
- Understand DHT fundamentals: key hashing, node IDs, and content addressing.
- Implement local key-value storage and name-to-key mapping.
- Explore persistence, simple APIs, and modular Go design.

---

## Project Structure
- `dht-store/` - Standalone CLI key-value store
- `dht-server/` - HTTP server, DHT and name-mapper packages
- `dht-network/` - Peer discovery and routing
- `dht-node/` - Full DHT node (networking + storage)
- `dht-learn.md` - DHT learning notes and summary
- `go.work` - Go workspace file

---

This workspace is intended for learning, experimentation, and as a foundation for more advanced DHT projects. 