package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"dht-server/dht"
	"dht-server/name_mapper"
)

/*
DHT Server - Minimal Distributed Hash Table Web Service

This server provides a simple DHT-like key-value store accessible via HTTP endpoints.
It supports storing and retrieving binary content by either a direct key or a human-readable name.

Key Features:
- Content is stored in a local persistent map (using the dht package).
- Each server instance has a unique node ID derived from its address and port.
- Clients can PUT content by specifying either a 'key' or a 'name'.
  - If 'name' is provided, a key is generated as SHA-1(name) and the mapping is stored using the name-mapper package.
  - The mapping from name to key is persisted and can be queried.
- Clients can GET content by 'key' or by 'name'.
- All content values are base64-encoded in requests and responses for safe transport.
- The server persists both the DHT store and the name-key mapping to disk as JSON files.

HTTP Endpoints:
- POST /put
    Request JSON:  { "key": "...", "name": "...", "value": "<base64>" }
      - Either 'key' or 'name' must be provided. If both are provided, 'name' takes precedence.
    Response JSON: { "key": "..." } (the key used to store the value)

- GET /get?key=... or /get?name=...
    Response JSON: { "key": "...", "value": "<base64>", "found": true/false }
      - If 'name' is provided, it is resolved to a key using the name-mapper.
      - If the key is not found, 'found' is false and 'value' is empty.

Architecture:
- The DHT logic (Put, Get, persistence) is in the dht package.
- The name-to-key mapping and its persistence is handled by the name-mapper package.
- The main server wires these together and exposes the HTTP API.

This code is intended for learning and experimentation with DHT concepts and simple web service design.
*/

// PutRequest represents the JSON body for /put
// Either 'key' or 'name' must be provided. 'value' is base64-encoded.
type PutRequest struct {
	Key   string `json:"key,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value"` // base64 encoded
}

// PutResponse is returned by /put, always containing the key used
type PutResponse struct {
	Key string `json:"key"`
}

// GetResponse is returned by /get, with the key, value (base64), and found flag
type GetResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"` // base64 encoded
	Found bool   `json:"found"`
}

// keyFromName generates a key from a name using SHA-1 and hex encoding
func keyFromName(name string) string {
	h := sha1.Sum([]byte(name))
	return hex.EncodeToString(h[:])
}

// putHandler handles POST /put requests
func putHandler(dhtInst *dht.DHT, nm *name_mapper.NameMapper, nameMapFile string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req PutRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var key string
		// If name is provided, generate key and store mapping
		if req.Name != "" {
			key = keyFromName(req.Name)
			nm.Set(req.Name, key)
			_ = nm.Save(nameMapFile)
		} else if req.Key != "" {
			key = req.Key
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("must provide 'key' or 'name'"))
			return
		}
		val, err := base64.StdEncoding.DecodeString(req.Value)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		dhtInst.Put(key, val)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PutResponse{Key: key})
	}
}

// getHandler handles GET /get requests
func getHandler(dhtInst *dht.DHT, nm *name_mapper.NameMapper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		name := r.URL.Query().Get("name")
		// If name is provided, look up the key
		if name != "" {
			if k, ok := nm.Get(name); ok {
				key = k
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		val, ok := dhtInst.Get(key)
		resp := GetResponse{Key: key, Found: ok}
		if ok {
			resp.Value = base64.StdEncoding.EncodeToString(val)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func main() {
	// Server address (default :8080, can override with first arg)
	addr := ":8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	// File paths for DHT and name mapping persistence
	persistFile := filepath.Join(".", "store.json")
	nameMapFile := filepath.Join(".", "namemap.json")

	// Initialize DHT and NameMapper, loading persisted data if available
	dhtInst := dht.NewDHT(addr, persistFile)
	nm := name_mapper.NewNameMapper()
	_ = nm.Load(nameMapFile)

	fmt.Printf("Node ID: %s\n", dhtInst.NodeID)

	// Register HTTP handlers
	http.HandleFunc("/put", putHandler(dhtInst, nm, nameMapFile))
	http.HandleFunc("/get", getHandler(dhtInst, nm))

	log.Printf("Listening on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
