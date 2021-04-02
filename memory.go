package tgbot

import (
	"fmt"
	"sync"
)

type InMemoryStorage struct {
	handlers map[Kind]map[string]map[int64]CommonHandler
	mu       sync.Mutex
}

func New() *InMemoryStorage {
	s := new(InMemoryStorage)
	s.handlers = make(map[Kind]map[string]map[int64]CommonHandler)
	return s
}

func (s *InMemoryStorage) Set(kind Kind, name string, chatID int64, handler CommonHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.handlers[kind]; !ok {
		s.handlers[kind] = make(map[string]map[int64]CommonHandler)
	}
	if _, ok := s.handlers[kind][name]; !ok {
		s.handlers[kind][name] = make(map[int64]CommonHandler)
	}

	s.handlers[kind][name][chatID] = handler
}

func (s *InMemoryStorage) Get(kind Kind, name string, chatID int64) (CommonHandler, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if f, ok := s.handlers[kind][name][chatID]; ok {
		return f, nil
	}
	if f, ok := s.handlers[kind][name][0]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("not found")
}

func (s *InMemoryStorage) Unset(kind Kind, name string, chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.handlers[kind][name], chatID)
}
