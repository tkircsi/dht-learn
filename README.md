# DHT Workspace

This workspace is for learning and experimenting with Distributed Hash Table (DHT) concepts in Go. It contains two main modules:

- **dht-app**: A minimal command-line DHT application for storing and retrieving content by key or human-readable name.
- **dht-server**: A simple web server exposing DHT functionality over HTTP, supporting both key-based and name-based storage and retrieval.

## DHT Learning Goals
- Understand DHT fundamentals: key hashing, node IDs, and content addressing.
- Implement local key-value storage and name-to-key mapping.
- Explore persistence, simple APIs, and modular Go design.

---

## dht-app (CLI)
A minimal Go CLI app for storing and retrieving content using a DHT-like approach.

### Build
```sh
cd dht-app
go build
```

### Usage
```sh
# Store content from a file with a name
./dht-app put mymovie movie.txt

# Store content from stdin with a name
./dht-app put mydesc
(type or paste, then Ctrl+D)

# Retrieve by key
./dht-app get <key>

# Retrieve by name
./dht-app get mymovie
```

- Content and name mappings are persisted locally as JSON files.
- Node ID is generated from the environment or process ID.

---

## dht-server (HTTP API)
A simple web server exposing DHT storage and retrieval via HTTP endpoints.

### Build & Run
```sh
cd dht-server
# Run on default :8080
go run main.go
# Or specify address/port
go run main.go :9090
```

### Endpoints
- **POST /put**
  - Request JSON: `{ "key": "...", "name": "...", "value": "<base64>" }`
    - Either 'key' or 'name' must be provided. If both, 'name' takes precedence.
  - Response JSON: `{ "key": "..." }`

- **GET /get?key=...** or **/get?name=...**
  - Response JSON: `{ "key": "...", "value": "<base64>", "found": true/false }`

#### Example (using curl)
```sh
# Store by name
curl -X POST -d '{"name":"my-movie","value":"bXl2YWx1ZQ=="}' localhost:8080/put

# Retrieve by name
curl 'localhost:8080/get?name=my-movie'
```

- Content and name mappings are persisted as JSON files.
- Node ID is derived from the server address and port.

---

## Project Structure
- `dht-app/` - CLI application
- `dht-server/` - HTTP server, DHT and name-mapper packages
- `dht-learn.md` - DHT learning notes and summary
- `go.work` - Go workspace file

---

This workspace is intended for learning, experimentation, and as a foundation for more advanced DHT projects. 