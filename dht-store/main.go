package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

var (
	store   = make(map[string][]byte) // key: hash, value: content
	nameMap = make(map[string]string) // key: name, value: hash
	nodeID  string
)

const (
	storeFile   = "store.json"
	nameMapFile = "namemap.json"
)

func hashContent(content []byte) string {
	h := sha1.Sum(content)
	return hex.EncodeToString(h[:])
}

func nodeIDFromEnvOrRandom() string {
	name := os.Getenv("DHT_NODE_NAME")
	if name == "" {
		name = fmt.Sprintf("node-%d", os.Getpid())
	}
	h := sha1.Sum([]byte(name))
	id := binary.BigEndian.Uint64(h[:8])
	return fmt.Sprintf("%016x", id)
}

func loadStore() {
	f, err := os.Open(storeFile)
	if err == nil {
		defer f.Close()
		dec := json.NewDecoder(f)
		var tmp map[string]string
		if err := dec.Decode(&tmp); err == nil {
			for k, v := range tmp {
				store[k], _ = hex.DecodeString(v)
			}
		}
	}
}

func saveStore() {
	tmp := make(map[string]string, len(store))
	for k, v := range store {
		tmp[k] = hex.EncodeToString(v)
	}
	f, err := os.Create(storeFile)
	if err == nil {
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.Encode(tmp)
	}
}

func loadNameMap() {
	f, err := os.Open(nameMapFile)
	if err == nil {
		defer f.Close()
		dec := json.NewDecoder(f)
		dec.Decode(&nameMap)
	}
}

func saveNameMap() {
	f, err := os.Create(nameMapFile)
	if err == nil {
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.Encode(nameMap)
	}
}

func put(name, contentPath string) {
	var content []byte
	var err error
	if contentPath != "" {
		content, err = os.ReadFile(contentPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Enter content (end with Ctrl+D):")
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read stdin: %v\n", err)
			os.Exit(1)
		}
	}
	hash := hashContent(content)
	store[hash] = content
	if name != "" {
		nameMap[name] = hash
	}
	saveStore()
	saveNameMap()
	fmt.Printf("Stored. Key: %s\n", hash)
	if name != "" {
		fmt.Printf("Name: %s\n", name)
	}
}

func get(keyOrName string) {
	hash := keyOrName
	if v, ok := nameMap[keyOrName]; ok {
		hash = v
	}
	if content, ok := store[hash]; ok {
		os.Stdout.Write(content)
	} else {
		fmt.Fprintf(os.Stderr, "Not found: %s\n", keyOrName)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  put <name> [file]")
	fmt.Println("  get <key|name>")
	os.Exit(1)
}

func main() {
	nodeID = nodeIDFromEnvOrRandom()
	loadStore()
	loadNameMap()
	fmt.Printf("Node ID: %s\n", nodeID)
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "put":
		if len(os.Args) == 4 {
			put(os.Args[2], os.Args[3])
		} else if len(os.Args) == 3 {
			put(os.Args[2], "")
		} else {
			usage()
		}
	case "get":
		if len(os.Args) == 3 {
			get(os.Args[2])
		} else {
			usage()
		}
	default:
		usage()
	}
}
