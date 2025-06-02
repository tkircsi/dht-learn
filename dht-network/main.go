package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Command-line flags
	var bootstrapAddr string
	flag.StringVar(&bootstrapAddr, "bootstrap", "", "Bootstrap node address (host:port)")
	flag.Parse()

	// Server address (default :8080, can override with first arg)
	addr := ":8080"
	if flag.NArg() > 0 {
		addr = flag.Arg(0)
	}

	// Peer management
	selfNodeID := generateNodeID(addr)
	selfAddr := addr
	if addr[0] == ':' {
		selfAddr = "127.0.0.1" + addr
	}
	pl := NewPeerList()
	pl.Add(PeerInfo{NodeID: selfNodeID, Address: selfAddr})

	// If bootstrap address is provided, join the network
	if bootstrapAddr != "" {
		joinNetwork(bootstrapAddr, selfAddr, selfNodeID, pl)
	}

	fmt.Printf("Node ID: %s\n", selfNodeID)

	// Register HTTP handlers with logging
	http.HandleFunc("/ping", logRequest("/ping", pingHandler(selfNodeID, selfAddr)))
	http.HandleFunc("/peers", logRequest("/peers", peersHandler(pl)))
	http.HandleFunc("/register", registerHandler(pl))
	http.HandleFunc("/find_node", logRequest("/find_node", findNodeHandler(pl, selfNodeID)))

	log.Printf("Listening on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// generateNodeID creates a node ID from the address using SHA-1
func generateNodeID(addr string) string {
	h := sha1.Sum([]byte(addr))
	return hex.EncodeToString(h[:8])
}
