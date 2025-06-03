package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func pingBootstrap(bootstrapAddr string) (PeerInfo, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/ping", bootstrapAddr))
	if err != nil {
		return PeerInfo{}, err
	}
	defer resp.Body.Close()
	var bootstrap PeerInfo
	if err := json.NewDecoder(resp.Body).Decode(&bootstrap); err != nil {
		return PeerInfo{}, err
	}
	return bootstrap, nil
}

func fetchBootstrapPeers(bootstrapAddr, selfNodeID, selfAddr string, pl *PeerList) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/peers", bootstrapAddr))
	if err != nil {
		log.Printf("[JOIN] Failed to fetch peers from bootstrap: %v", err)
		return
	}
	defer resp.Body.Close()
	var peers []PeerInfo
	if err := json.NewDecoder(resp.Body).Decode(&peers); err == nil {
		log.Printf("[JOIN] Fetched %d peers from bootstrap node", len(peers))
		for _, p := range peers {
			if p.NodeID != selfNodeID && p.Address != selfAddr {
				pl.Add(p)
			}
		}
		log.Printf("[JOIN] Merged %d peers from bootstrap", len(peers))
	} else {
		log.Printf("[JOIN] Failed to decode peers from bootstrap: %v", err)
	}
}

func announceSelf(bootstrapAddr, selfNodeID, selfAddr string, pl *PeerList) {
	client := &http.Client{Timeout: 3 * time.Second}
	selfInfo := PeerInfo{NodeID: selfNodeID, Address: selfAddr}
	buf, _ := json.Marshal(selfInfo)
	log.Println("[JOIN] Announcing self to bootstrap node via /register...")
	logPeerList(pl, "joinNetwork BEFORE REGISTER")
	regResp, err := client.Post(fmt.Sprintf("http://%s/register", bootstrapAddr), "application/json", bytes.NewReader(buf))
	if err == nil {
		log.Printf("[JOIN] Announced self to bootstrap node at %s", bootstrapAddr)
		log.Printf("[JOIN] /register response status: %s", regResp.Status)
		body, _ := io.ReadAll(regResp.Body)
		log.Printf("[JOIN] /register response body: %s", string(body))
		regResp.Body.Close()
	} else {
		log.Printf("[JOIN] Failed to announce self to bootstrap: %v", err)
	}
	logPeerList(pl, "joinNetwork END")
}

func kademliaLookup(bootstrapAddr, selfNodeID, selfAddr string, pl *PeerList) {
	client := &http.Client{Timeout: 3 * time.Second}
	log.Printf("[JOIN] Performing Kademlia-style lookup for own node ID: %s", selfNodeID)
	lookupURL := fmt.Sprintf("http://%s/find_node?target=%s", bootstrapAddr, selfNodeID)
	resp, err := client.Get(lookupURL)
	if err == nil {
		defer resp.Body.Close()
		var foundPeers []PeerInfo
		if err := json.NewDecoder(resp.Body).Decode(&foundPeers); err == nil {
			log.Printf("[JOIN] /find_node returned %d peers", len(foundPeers))
			for _, p := range foundPeers {
				if p.NodeID != selfNodeID && p.Address != selfAddr {
					pl.Add(p)
				}
			}
			logPeerList(pl, "joinNetwork AFTER FIND_NODE")
		} else {
			log.Printf("[JOIN] Failed to decode /find_node response: %v", err)
		}
	} else {
		log.Printf("[JOIN] Failed to call /find_node: %v", err)
	}
}

func joinNetwork(bootstrapAddr, selfAddr, selfNodeID string, pl *PeerList) {
	log.Printf("[JOIN] Attempting to join network via bootstrap node at %s", bootstrapAddr)
	logPeerList(pl, "joinNetwork START")

	bootstrap, err := pingBootstrap(bootstrapAddr)
	if err != nil {
		log.Printf("[JOIN] Failed to ping bootstrap node: %v", err)
		return
	}
	pl.Add(bootstrap)
	log.Printf("[JOIN] Added bootstrap peer: %+v", bootstrap)

	fetchBootstrapPeers(bootstrapAddr, selfNodeID, selfAddr, pl)
	announceSelf(bootstrapAddr, selfNodeID, selfAddr, pl)
	kademliaLookup(bootstrapAddr, selfNodeID, selfAddr, pl)

	log.Printf("[JOIN] Discovery and connection process complete.")
}
