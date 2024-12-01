package store

import "fmt"

type MapEntry struct {
	flags uint16
	data  []byte
}

type KeyNotFoundError string

func (e KeyNotFoundError) Error() string {
	return string(e)
}

type Store interface {
	Add(string, []byte, uint16) error
	Get(string) (*MapEntry, error)
}

type InMemoryStore struct {
	data map[string]MapEntry
}

func (s *InMemoryStore) Add(key string, data []byte, flags uint16) error {
	s.data[key] = MapEntry{flags: flags, data: data}
	return nil
}

func (s *InMemoryStore) Get(key string) (*MapEntry, error) {
	value, found := s.data[key]
	if !found {
		return nil, KeyNotFoundError(fmt.Sprintf("%s not found", key))
	}
	return &value, nil
}
