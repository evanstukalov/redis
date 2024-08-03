package store

import (
	"errors"
	"log"
	"sync"
)

type Store struct {
	store map[string]interface{}
	mutex sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		store: make(map[string]interface{}),
	}
}

func (s *Store) Set(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.store[key] = value

	log.Println("Set handler: ", key, value)
}

func (s *Store) Get(key string) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if value, ok := s.store[key]; !ok {
		return nil, errors.New("key does not exists")
	} else {
		log.Println("Get handler: ", key, value)
		return value, nil
	}
}
