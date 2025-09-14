package local

import (
	"context"
	"fmt"
	"strings"

	"go.etcd.io/bbolt"

	"github.com/kivigo/kivigo/pkg/errs"
	"github.com/kivigo/kivigo/pkg/models"
)

var (
	_ models.KV           = (*Client)(nil)
	_ models.KVWithHealth = (*Client)(nil)
	_ models.KVWithBatch  = (*Client)(nil)
)

const dbName = "kivigo"

type (
	Client struct {
		c *bbolt.DB
	}

	Option struct {
		Path     string `default:"./"`
		FileName string `default:"kivigo.db"`
	}
)

func NewOptions() Option {
	return Option{}
}

func DefaultOptions() Option {
	return Option{
		Path:     "./",
		FileName: "kivigo.db",
	}
}

func New(opt Option) (Client, error) {
	if opt.Path[len(opt.Path)-1] != '/' {
		opt.Path += "/"
	}

	db, err := bbolt.Open(fmt.Sprintf("%s%s", opt.Path, opt.FileName), 0o600, nil)
	if err != nil {
		return Client{}, fmt.Errorf("could not open local db: %w", err)
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(dbName))
		if err != nil {
			return fmt.Errorf("could not create local db: %w", err)
		}

		return nil
	}); err != nil {
		return Client{}, err
	}

	return Client{c: db}, nil
}

func (c Client) Close() error {
	return c.c.Close()
}

// Get gets a value from the database.
func (c Client) GetRaw(_ context.Context, key string) ([]byte, error) {
	// Check if key is not empty
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	var value []byte

	err := c.c.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(dbName))
		if b == nil {
			return errs.ErrNotFound
		}

		data := b.Get([]byte(key))
		if data == nil {
			return errs.ErrNotFound
		}

		// Copy the data since BoltDB data is only valid during the transaction
		value = make([]byte, len(data))
		copy(value, data)

		return nil
	})

	return value, err
}

// Set sets a value in the database.
func (c Client) SetRaw(_ context.Context, key string, value []byte) error {
	return c.c.Update(func(tx *bbolt.Tx) error {
		// Check if key is not empty
		if key == "" {
			return errs.ErrEmptyKey
		}

		b := tx.Bucket([]byte(dbName))
		if b == nil {
			return errs.ErrNotFound
		}

		return b.Put([]byte(key), value)
	})
}

// Delete deletes a value from the database.
func (c Client) Delete(_ context.Context, key string) error {
	// Check if key is not empty
	if key == "" {
		return errs.ErrEmptyKey
	}

	return c.c.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(dbName))
		if b == nil {
			return errs.ErrNotFound
		}

		return b.Delete([]byte(key))
	})
}

// List lists all the keys from the database.
func (c Client) List(_ context.Context, prefix string) (keys []string, err error) {
	return keys, c.c.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(dbName))
		if b == nil {
			return errs.ErrNotFound
		}

		c := b.Cursor()

		for k, _ := c.Seek([]byte(prefix)); k != nil && strings.HasPrefix(string(k), prefix); k, _ = c.Next() {
			keys = append(keys, string(k))
		}

		return nil
	})
}

func (c Client) Health(_ context.Context) error {
	if c.c == nil {
		return errs.ErrClientNotInitialized
	}

	return c.c.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(dbName))
		if b == nil {
			return fmt.Errorf("bucket %s not found", dbName)
		}

		return nil
	})
}

// BatchGet retrieves multiple keys from the database.
func (c Client) BatchGetRaw(_ context.Context, keys []string) (map[string][]byte, error) {
	results := make(map[string][]byte, len(keys))

	err := c.c.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(dbName))
		if b == nil {
			return errs.ErrNotFound
		}

		for _, key := range keys {
			results[key] = b.Get([]byte(key))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

// BatchSet sets multiple key-value pairs in the database.
func (c Client) BatchSetRaw(_ context.Context, kv map[string][]byte) error {
	return c.c.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(dbName))
		if b == nil {
			return errs.ErrNotFound
		}

		for key, value := range kv {
			if err := b.Put([]byte(key), value); err != nil {
				return err
			}
		}

		return nil
	})
}

// BatchDelete deletes multiple keys from the database.
func (c Client) BatchDelete(_ context.Context, keys []string) error {
	return c.c.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(dbName))
		if b == nil {
			return errs.ErrNotFound
		}

		for _, key := range keys {
			if err := b.Delete([]byte(key)); err != nil {
				return err
			}
		}

		return nil
	})
}
