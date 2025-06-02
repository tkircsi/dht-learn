package name_mapper

import (
	"encoding/json"
	"os"
	"sync"
)

type NameMapper struct {
	m  map[string]string
	mu sync.RWMutex
}

func NewNameMapper() *NameMapper {
	return &NameMapper{m: make(map[string]string)}
}

func (nm *NameMapper) Set(name, key string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.m[name] = key
}

func (nm *NameMapper) Get(name string) (string, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	key, ok := nm.m[name]
	return key, ok
}

func (nm *NameMapper) Save(filename string) error {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(nm.m)
}

func (nm *NameMapper) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&nm.m)
}
