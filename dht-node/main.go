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
	var bootstrapAddr string
	flag.StringVar(&bootstrapAddr, "bootstrap", "", "Bootstrap node address (host:port)")
	flag.Parse()

	addr := ":8080"
	if flag.NArg() > 0 {
		addr = flag.Arg(0)
	}

	selfNodeID := generateNodeID(addr)
	selfAddr := addr
	if addr[0] == ':' {
		selfAddr = "127.0.0.1" + addr
	}
	pl := NewPeerList()
	pl.Add(PeerInfo{NodeID: selfNodeID, Address: selfAddr})

	// Content store setup
	store := NewStore(selfNodeID)
	if err := store.Load(); err != nil {
		log.Printf("[STORE] No existing store loaded: %v", err)
	}

	if bootstrapAddr != "" {
		joinNetwork(bootstrapAddr, selfAddr, selfNodeID, pl)
	}

	fmt.Printf("Node ID: %s\n", selfNodeID)

	http.HandleFunc("/ping", logRequest("/ping", pingHandler(selfNodeID, selfAddr)))
	http.HandleFunc("/peers", logRequest("/peers", peersHandler(pl)))
	http.HandleFunc("/register", registerHandler(pl))
	http.HandleFunc("/find_node", logRequest("/find_node", findNodeHandler(pl, selfNodeID)))
	// Content endpoints
	http.HandleFunc("/put", putContentHandler(store, pl, selfNodeID))
	http.HandleFunc("/get", getContentHandler(store, pl, selfNodeID))

	log.Printf("Listening on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func generateNodeID(addr string) string {
	h := sha1.Sum([]byte(addr))
	return hex.EncodeToString(h[:8])
}
