package models

import (
	"context"
)

type KV interface {
	Close() error

	List(ctx context.Context, prefix string) (keys []string, err error)
	GetRaw(ctx context.Context, key string) (value []byte, err error)
	SetRaw(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error

	// Health checks the health of the KV store.
	Health(ctx context.Context) error
}
