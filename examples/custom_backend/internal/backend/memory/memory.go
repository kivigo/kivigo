package memory

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/azrod/kivigo/pkg/models"
)

var _ models.KV = (*BMemory)(nil)

type BMemory struct {
	mu    sync.RWMutex
	store map[string][]byte
}

func New() *BMemory {
	return &BMemory{
		store: make(map[string][]byte),
	}
}

func (b *BMemory) SetRaw(_ context.Context, key string, value []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.store[key] = value

	return nil
}

func (b *BMemory) GetRaw(_ context.Context, key string) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	val, ok := b.store[key]
	if !ok {
		return nil, errors.New("key not found")
	}

	return val, nil
}

func (b *BMemory) Delete(_ context.Context, key string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.store, key)

	return nil
}

func (b *BMemory) List(_ context.Context, prefix string) (keys []string, err error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	keys = make([]string, 0, len(b.store))

	for k := range b.store {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

func (b *BMemory) Close() error {
	return nil
}
