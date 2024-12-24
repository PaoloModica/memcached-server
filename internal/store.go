package store

import (
	"fmt"
	"time"
)

type MapEntry struct {
	Flags      uint16
	InsertDt   time.Time
	Expiration time.Duration
	Data       []byte
}

func (m *MapEntry) IsExpired() bool {
	if time.Since(m.InsertDt) > m.Expiration {
		return true
	} else {
		return false
	}
}

type KeyNotFoundError string

func (e KeyNotFoundError) Error() string {
	return string(e)
}

type Store interface {
	Add(string, []byte, uint16, time.Duration) error
	Get(string) (*MapEntry, error)
	Remove(string) error
}

type InMemoryStore struct {
	data map[string]MapEntry
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{data: make(map[string]MapEntry)}
}

func (s *InMemoryStore) Add(key string, data []byte, flags uint16, expiration time.Duration) error {
	s.data[key] = MapEntry{Flags: flags, Data: data, InsertDt: time.Now(), Expiration: expiration * time.Second}
	return nil
}

func (s *InMemoryStore) Get(key string) (*MapEntry, error) {
	value, found := s.data[key]
	if !found {
		return nil, KeyNotFoundError(fmt.Sprintf("%s not found", key))
	}
	return &value, nil
}

func (s *InMemoryStore) Remove(key string) error {
	_, found := s.data[key]
	if !found {
		return KeyNotFoundError(fmt.Sprintf("%s not found", key))
	}
	delete(s.data, key)
	return nil
}
