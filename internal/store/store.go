package store

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type ValueWithExpiration struct {
	value     interface{}
	ExpiredAt *time.Time
}

type Store struct {
	store map[string]ValueWithExpiration
	mutex sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		store: make(map[string]ValueWithExpiration),
	}
}

func (s *Store) Set(key string, value interface{}, px *int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var expirationTime *time.Time
	if px != nil {
		t := time.Now().Add(time.Duration(*px) * time.Millisecond)

		expirationTime = &t
	}

	s.store[key] = ValueWithExpiration{
		value:     value,
		ExpiredAt: expirationTime,
	}

	log.Println("Set handler: ", key, value)
}

func (s *Store) Get(key string) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if value, ok := s.store[key]; !ok {
		return nil, errors.New("key does not exists")
	} else {
		log.Println("Get handler: ", key, value.value)
		return value.value, nil
	}
}

func (s *Store) Remove(key string) {
	delete(s.store, key)
	fmt.Println("Remove key: ", key)
}
