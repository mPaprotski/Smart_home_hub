// Package storage currently provides an explicit nonpersistent simulation store.
// It is not an implementation of SQLite, transactions or migrations.
package storage

import (
	"errors"
	"github.com/mPaprotski/Smart_home_hub/internal/events"
	"sync"
)

type Observations interface {
	Put(events.Envelope) error
	Get(string) (events.Envelope, bool)
}
type Memory struct {
	mu       sync.Mutex
	values   map[string]events.Envelope
	capacity int
}

func NewMemory(capacity int) *Memory {
	return &Memory{values: make(map[string]events.Envelope), capacity: capacity}
}
func (m *Memory) Put(e events.Envelope) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.values[e.Resource.Key()]; !exists && len(m.values) >= m.capacity {
		return errors.New("simulation store full")
	}
	e.Payload = append([]byte(nil), e.Payload...)
	m.values[e.Resource.Key()] = e
	return nil
}
func (m *Memory) Get(key string) (events.Envelope, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.values[key]
	e.Payload = append([]byte(nil), e.Payload...)
	return e, ok
}
