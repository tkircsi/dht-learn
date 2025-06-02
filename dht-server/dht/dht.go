package dht

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type DHT struct {
	store       map[string][]byte
	mu          sync.RWMutex
	NodeID      string
	persistFile string
}

func NewDHT(serverURI, persistFile string) *DHT {
	h := sha1.Sum([]byte(serverURI))
	id := binary.BigEndian.Uint64(h[:8])
	d := &DHT{
		store:       make(map[string][]byte),
		NodeID:      fmt.Sprintf("%016x", id),
		persistFile: persistFile,
	}
	d.load()
	return d
}

func (d *DHT) Put(key string, value []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.store[key] = value
	d.save()
}

func (d *DHT) Get(key string) ([]byte, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	v, ok := d.store[key]
	return v, ok
}

func (d *DHT) save() {
	tmp := make(map[string]string, len(d.store))
	for k, v := range d.store {
		tmp[k] = hex.EncodeToString(v)
	}
	f, err := os.Create(d.persistFile)
	if err == nil {
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.Encode(tmp)
	}
}

func (d *DHT) load() {
	f, err := os.Open(d.persistFile)
	if err == nil {
		defer f.Close()
		dec := json.NewDecoder(f)
		var tmp map[string]string
		if err := dec.Decode(&tmp); err == nil {
			for k, v := range tmp {
				if b, err := hex.DecodeString(v); err == nil {
					d.store[k] = b
				}
			}
		}
	}
}
