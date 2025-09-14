package badger

import (
	"context"
	"errors"
	"fmt"

	bd "github.com/dgraph-io/badger/v4"

	"github.com/kivigo/kivigo/pkg/errs"
	"github.com/kivigo/kivigo/pkg/models"
)

var (
	_ models.KV           = (*Client)(nil)
	_ models.KVWithHealth = (*Client)(nil)
	_ models.KVWithBatch  = (*Client)(nil)
)

type (
	Client struct {
		db *bd.DB
	}
)

func NewOptions() bd.Options {
	return bd.Options{}
}

func DefaultOptions(path string) bd.Options {
	return bd.DefaultOptions(path)
}

func New(opt bd.Options) (Client, error) {
	db, err := bd.Open(opt)
	if err != nil {
		return Client{}, fmt.Errorf("could not open badger db: %w", err)
	}

	return Client{db: db}, nil
}

func (c Client) Close() error {
	return c.db.Close()
}

func (c Client) SetRaw(_ context.Context, key string, value []byte) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	return c.db.Update(func(txn *bd.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

func (c Client) GetRaw(_ context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	var valCopy []byte

	err := c.db.View(func(txn *bd.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		valCopy = val

		return nil
	})

	if errors.Is(err, bd.ErrKeyNotFound) {
		return nil, errs.ErrNotFound
	}

	return valCopy, err
}

func (c Client) Delete(_ context.Context, key string) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	return c.db.Update(func(txn *bd.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (c Client) List(_ context.Context, prefix string) (keys []string, err error) {
	if prefix == "" {
		return nil, errs.ErrEmptyPrefix
	}

	err = c.db.View(func(txn *bd.Txn) error {
		it := txn.NewIterator(bd.DefaultIteratorOptions)
		defer it.Close()

		prefixBytes := []byte(prefix)
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			keys = append(keys, string(item.Key()))
		}

		return nil
	})

	return keys, err
}

func (c Client) Health(_ context.Context) error {
	// Simple health check: try a read-only transaction
	return c.db.View(func(_ *bd.Txn) error { return nil })
}

func (c Client) BatchGetRaw(_ context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return nil, errs.ErrEmptyBatch
	}

	results := make(map[string][]byte, len(keys))
	err := c.db.View(func(txn *bd.Txn) error {
		for _, key := range keys {
			item, err := txn.Get([]byte(key))
			if errors.Is(err, bd.ErrKeyNotFound) {
				results[key] = nil

				continue
			}

			if err != nil {
				return err
			}

			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			results[key] = val
		}

		return nil
	})

	return results, err
}

func (c Client) BatchSetRaw(_ context.Context, kv map[string][]byte) error {
	if len(kv) == 0 {
		return errs.ErrEmptyBatch
	}

	return c.db.Update(func(txn *bd.Txn) error {
		for key, value := range kv {
			if err := txn.Set([]byte(key), value); err != nil {
				return err
			}
		}

		return nil
	})
}

func (c Client) BatchDelete(_ context.Context, keys []string) error {
	if len(keys) == 0 {
		return errs.ErrEmptyBatch
	}

	return c.db.Update(func(txn *bd.Txn) error {
		for _, key := range keys {
			if err := txn.Delete([]byte(key)); err != nil {
				return err
			}
		}

		return nil
	})
}
