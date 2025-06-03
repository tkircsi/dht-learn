package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func logRequest(handlerName string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HANDLER] %s called: %s %s", handlerName, r.Method, r.URL.Path)
		next(w, r)
	}
}

func pingHandler(nodeID, address string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := PeerInfo{NodeID: nodeID, Address: address}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func peersHandler(pl *PeerList) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pl.All())
	}
}

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

type PutRequest struct {
	Key   string `json:"key,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value"`
}

type PutResponse struct {
	Key string `json:"key"`
}

type GetResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Found bool   `json:"found"`
}

// putContentHandler handles POST /put for storing content in the DHT.
func putContentHandler(store *Store, pl *PeerList, selfID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		key := req.Key
		if key == "" && req.Name != "" {
			h := sha1.Sum([]byte(req.Name))
			key = hex.EncodeToString(h[:])
		}
		val, err := base64.StdEncoding.DecodeString(req.Value)
		if err != nil || key == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Find the closest peer to the key (including self)
		closest := pl.closestPeers(key, 1, "")
		isSelfClosest := len(closest) == 0 || closest[0].NodeID == selfID
		if isSelfClosest {
			log.Printf("[DHT] Storing key %s locally (self is closest)", key)
			if err := store.Put(key, val); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PutResponse{Key: key})
			return
		}
		// Forward to closest peer
		peer := closest[0]
		log.Printf("[DHT] Forwarding PUT for key %s to peer %s at %s", key, peer.NodeID, peer.Address)
		forwardReq := PutRequest{Key: req.Key, Name: req.Name, Value: req.Value}
		buf, _ := json.Marshal(forwardReq)
		url := fmt.Sprintf("http://%s/put", peer.Address)
		resp, err := http.Post(url, "application/json", bytes.NewReader(buf))
		if err != nil {
			log.Printf("[DHT] Failed to forward PUT to %s: %v", peer.Address, err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

// getContentHandler handles GET /get for retrieving content from the DHT.
func getContentHandler(store *Store, pl *PeerList, selfID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		name := r.URL.Query().Get("name")
		if key == "" && name != "" {
			h := sha1.Sum([]byte(name))
			key = hex.EncodeToString(h[:])
		}
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		val, ok := store.Get(key)
		if ok {
			log.Printf("[DHT] GET key %s found locally", key)
			resp := GetResponse{Key: key, Value: base64.StdEncoding.EncodeToString(val), Found: true}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		// Not found locally: find closest peer and forward
		closest := pl.closestPeers(key, 1, "")
		isSelfClosest := len(closest) == 0 || closest[0].NodeID == selfID
		if isSelfClosest {
			log.Printf("[DHT] GET key %s not found locally and self is closest", key)
			resp := GetResponse{Key: key, Found: false}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		peer := closest[0]
		log.Printf("[DHT] Forwarding GET for key %s to peer %s at %s", key, peer.NodeID, peer.Address)
		url := fmt.Sprintf("http://%s/get?key=%s", peer.Address, key)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("[DHT] Failed to forward GET to %s: %v", peer.Address, err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
