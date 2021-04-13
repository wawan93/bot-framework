package storage

import (
	"fmt"
	"sync"
)

type InMemoryStorage struct {
	handlers map[string]map[string]map[int64]Command
	mu       sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	s := new(InMemoryStorage)
	s.handlers = make(map[string]map[string]map[int64]Command)
	return s
}

func (s *InMemoryStorage) Set(kind string, name string, chatID int64, handler Command) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.handlers[kind]; !ok {
		s.handlers[kind] = make(map[string]map[int64]Command)
	}
	if _, ok := s.handlers[kind][name]; !ok {
		s.handlers[kind][name] = make(map[int64]Command)
	}

	s.handlers[kind][name][chatID] = handler
}

func (s *InMemoryStorage) Get(kind string, name string, chatID int64) (Command, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if f, ok := s.handlers[kind][name][chatID]; ok {
		return f, nil
	}
	if f, ok := s.handlers[kind][name][0]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("not found")
}

func (s *InMemoryStorage) Unset(kind string, name string, chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.handlers[kind][name], chatID)
}
