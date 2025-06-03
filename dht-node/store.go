package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string][]byte
	file string
}

func NewStore(nodeID string) *Store {
	file := fmt.Sprintf("store_%s.json", nodeID)
	return &Store{
		data: make(map[string][]byte),
		file: file,
	}
}

func (s *Store) Put(key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return s.save()
}

func (s *Store) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

func (s *Store) save() error {
	tmp := make(map[string]string, len(s.data))
	for k, v := range s.data {
		tmp[k] = string(v)
	}
	f, err := os.Create(s.file)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(tmp)
}

func (s *Store) Load() error {
	f, err := os.Open(s.file)
	if err != nil {
		return err
	}
	defer f.Close()
	var tmp map[string]string
	if err := json.NewDecoder(f).Decode(&tmp); err != nil {
		return err
	}
	for k, v := range tmp {
		s.data[k] = []byte(v)
	}
	return nil
}
