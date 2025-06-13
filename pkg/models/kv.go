package models

import (
	"context"
)

type (
	KV interface {
		Close() error

		List(ctx context.Context, prefix string) (keys []string, err error)
		GetRaw(ctx context.Context, key string) (value []byte, err error)
		SetRaw(ctx context.Context, key string, value []byte) error
		Delete(ctx context.Context, key string) error
	}

	KVWithBatch interface {
		BatchGet(ctx context.Context, keys []string) (map[string][]byte, error)
		BatchSet(ctx context.Context, kv map[string][]byte) error
		BatchDelete(ctx context.Context, keys []string) error
	}

	KVWithHealth interface {
		Health(ctx context.Context) error
	}
)
