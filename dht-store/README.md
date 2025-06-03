# Minimal DHT CLI Application

This is a minimal Distributed Hash Table (DHT) application for learning purposes.

## Features
- Hashes and stores content using SHA-1
- Each instance has a unique node_id (from DHT_NODE_NAME env or process ID)
- Supports storing and retrieving content by key or human-readable name

## Usage

Build:

```
go build
```

Run:

```
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

## Node ID

The node ID is generated at startup from the `DHT_NODE_NAME` environment variable (if set), or from the process ID. It is shown on each run. 