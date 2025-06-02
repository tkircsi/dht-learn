package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// logRequest logs the method and path of every incoming HTTP request.
func logRequest(handlerName string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HANDLER] %s called: %s %s", handlerName, r.Method, r.URL.Path)
		next(w, r)
	}
}

// pingHandler responds with this node's ID and address.
func pingHandler(nodeID, address string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := PeerInfo{NodeID: nodeID, Address: address}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// peersHandler returns the list of known peers.
func peersHandler(pl *PeerList) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pl.All())
	}
}

// registerHandler allows a peer to announce itself.
func registerHandler(pl *PeerList) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HANDLER] /register called: %s %s", r.Method, r.URL.Path)
		var peer PeerInfo
		if err := json.NewDecoder(r.Body).Decode(&peer); err == nil {
			log.Printf("[HANDLER] /register received peer: %+v", peer)
			pl.Add(peer)
			logPeerList(pl, "/register END")
			w.WriteHeader(http.StatusOK)
		} else {
			log.Printf("[HANDLER] /register decode error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

// findNodeHandler returns the k closest peers to the target node ID.
func findNodeHandler(pl *PeerList, selfID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HANDLER] /find_node called: %s %s", r.Method, r.URL.Path)
		target := r.URL.Query().Get("target")
		if target == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		closest := pl.closestPeers(target, 3, selfID)
		log.Printf("[HANDLER] /find_node for target %s: returning %d peers", target, len(closest))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(closest)
	}
}
