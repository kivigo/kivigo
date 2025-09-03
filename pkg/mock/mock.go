package mock

import (
	"context"
	"strings"
	"sync"

	"github.com/azrod/kivigo/pkg/errs"
	"github.com/azrod/kivigo/pkg/models"
)

var (
	_ models.KV           = (*MockKV)(nil)
	_ models.KVWithHealth = (*MockKV)(nil)
	_ models.KVWithBatch  = (*MockKV)(nil)
)

type MockKV struct { //nolint:revive
	Data map[string][]byte
	mu   sync.RWMutex
}

func (m *MockKV) GetRaw(_ context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	m.mu.RLock()
	val, ok := m.Data[key]
	m.mu.RUnlock()

	if !ok {
		return nil, errs.ErrNotFound
	}

	return val, nil
}

func (m *MockKV) SetRaw(_ context.Context, key string, value []byte) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	m.mu.Lock()
	m.Data[key] = value
	m.mu.Unlock()

	return nil
}

func (m *MockKV) Delete(_ context.Context, key string) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	if _, ok := m.Data[key]; !ok {
		return errs.ErrNotFound
	}

	m.mu.Lock()
	delete(m.Data, key)
	m.mu.Unlock()

	return nil
}

func (m *MockKV) List(_ context.Context, prefix string) ([]string, error) {
	m.mu.RLock()
	var keys []string

	for k := range m.Data {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	m.mu.RUnlock()

	return keys, nil
}

func (m *MockKV) Close() error { return nil }

// Health implements models.KVWithHealth.
// Always returns nil to indicate the mock is healthy.
func (m *MockKV) Health(_ context.Context) error {
	return nil
}

// BatchGet implements models.KVWithBatch.
// Returns a map of found keys to their values.
func (m *MockKV) BatchGetRaw(_ context.Context, keys []string) (map[string][]byte, error) {
	m.mu.RLock()
	result := make(map[string][]byte)

	for _, k := range keys {
		if v, ok := m.Data[k]; ok {
			result[k] = v
		} else {
			m.mu.RUnlock()
			return nil, errs.ErrNotFound
		}
	}
	m.mu.RUnlock()

	return result, nil
}

// BatchSet implements models.KVWithBatch.
// Sets all key-value pairs provided.
func (m *MockKV) BatchSetRaw(_ context.Context, kv map[string][]byte) error {
	m.mu.Lock()
	for k, v := range kv {
		m.Data[k] = v
	}
	m.mu.Unlock()

	return nil
}

// BatchDelete implements models.KVWithBatch.
// Deletes all keys provided.
func (m *MockKV) BatchDelete(_ context.Context, keys []string) error {
	if len(keys) == 0 {
		return errs.ErrEmptyBatch
	}

	m.mu.Lock()
	for _, k := range keys {
		if _, ok := m.Data[k]; !ok {
			m.mu.Unlock()
			return errs.ErrNotFound
		}
		delete(m.Data, k)
	}
	m.mu.Unlock()

	return nil
}
