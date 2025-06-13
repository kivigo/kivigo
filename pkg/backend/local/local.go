package local

import (
	"context"
	"fmt"
	"strings"

	"go.etcd.io/bbolt"

	"github.com/azrod/kivigo/pkg/client"
	"github.com/azrod/kivigo/pkg/errs"
	"github.com/azrod/kivigo/pkg/models"
)

var _ models.KV = (*Client)(nil)

const dbName = "kivigo"

type (
	Client struct {
		_ models.Backend
		c *bbolt.DB
	}

	Option struct {
		Path     string `default:"./"`
		FileName string `default:"kivigo.db"`
	}
)

func New(opt Option, _ client.Option) (Client, error) {
	if opt.Path == "" {
		opt.Path = "./"
	}
	if opt.FileName == "" {
		opt.FileName = "kivigo.db"
	}

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
	value := make([]byte, 0)

	return value, c.c.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(dbName))
		if b == nil {
			return errs.ErrNotFound
		}

		value = b.Get([]byte(key))
		if value == nil {
			return errs.ErrNotFound
		}

		return nil
	})
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
			return errs.ErrHealthCheckFailed(fmt.Errorf("bucket %s not found", dbName))
		}
		return nil
	})
}
