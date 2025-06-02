package main

import (
	"encoding/hex"
	"log"
	"sort"
	"sync"
)

// PeerInfo holds information about a peer node
type PeerInfo struct {
	NodeID  string `json:"node_id"`
	Address string `json:"address"`
}

// PeerList manages a thread-safe list of peers
type PeerList struct {
	mu    sync.RWMutex
	peers map[string]PeerInfo // key: NodeID
}

func NewPeerList() *PeerList {
	return &PeerList{peers: make(map[string]PeerInfo)}
}

func (pl *PeerList) Add(peer PeerInfo) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	if peer.NodeID != "" && peer.Address != "" {
		if _, exists := pl.peers[peer.NodeID]; !exists {
			log.Printf("Discovered new peer: %s at %s", peer.NodeID, peer.Address)
		} else {
			log.Printf("Peer already known: %s at %s", peer.NodeID, peer.Address)
		}
		pl.peers[peer.NodeID] = peer
	}
}

func (pl *PeerList) All() []PeerInfo {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	result := make([]PeerInfo, 0, len(pl.peers))
	for _, p := range pl.peers {
		result = append(result, p)
	}
	return result
}

// Print the current routing table (peer list)
func logPeerList(pl *PeerList, context string) {
	log.Printf("[%s] Current routing table:", context)
	for _, p := range pl.All() {
		log.Printf("  Peer: %s at %s", p.NodeID, p.Address)
	}
}

// xorDistance computes the XOR distance between two node IDs (as hex strings)
func xorDistance(a, b string) uint64 {
	ba, _ := hex.DecodeString(a)
	bb, _ := hex.DecodeString(b)
	var dist uint64
	for i := 0; i < len(ba) && i < len(bb) && i < 8; i++ {
		dist = (dist << 8) | uint64(ba[i]^bb[i])
	}
	return dist
}

// closestPeers returns up to k peers closest to the target ID
func (pl *PeerList) closestPeers(target string, k int, selfID string) []PeerInfo {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	peers := make([]PeerInfo, 0, len(pl.peers))
	for _, p := range pl.peers {
		if p.NodeID != selfID {
			peers = append(peers, p)
		}
	}
	// Sort by XOR distance
	sort.Slice(peers, func(i, j int) bool {
		return xorDistance(peers[i].NodeID, target) < xorDistance(peers[j].NodeID, target)
	})
	if len(peers) > k {
		peers = peers[:k]
	}
	return peers
}
