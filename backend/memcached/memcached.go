package memcached

import (
	"context"
	"errors"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

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
		c *memcache.Client
	}

	Option struct {
		Servers []string      `default:"localhost:11211"`
		Timeout time.Duration `default:"100ms"`
	}
)

// NewOptions returns a new set of options for the Memcached client.
func NewOptions() Option {
	return Option{}
}

// DefaultOptions returns the default options for the Memcached client.
// Servers: ["localhost:11211"]
// Timeout: 100ms
func DefaultOptions() Option {
	return Option{
		Servers: []string{"localhost:11211"},
		Timeout: 100 * time.Millisecond,
	}
}

// New returns a new Memcached client.
func New(opt Option) (Client, error) {
	client := memcache.New(opt.Servers...)
	client.Timeout = opt.Timeout

	return Client{c: client}, nil
}

func (c Client) SetRaw(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	item := &memcache.Item{
		Key:   key,
		Value: value,
	}

	return c.c.Set(item)
}

// GetRaw gets a value from Memcached.
func (c Client) GetRaw(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	item, err := c.c.Get(key)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return nil, errs.ErrNotFound
		}

		return nil, err
	}

	return item.Value, nil
}

// List lists all the keys from Memcached that match the given prefix.
// Note: Memcached doesn't natively support key listing, so this is a limitation.
// This implementation returns an error indicating the operation is not supported.
func (c Client) List(ctx context.Context, prefix string) ([]string, error) {
	// Memcached doesn't support listing keys natively
	// This is a known limitation of Memcached
	return nil, errs.ErrOperationNotSupported
}

// Delete deletes a key from Memcached.
func (c Client) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	err := c.c.Delete(key)
	if err != nil && errors.Is(err, memcache.ErrCacheMiss) {
		// Memcached returns error when key doesn't exist, but we treat this as success
		return nil
	}

	return err
}

func (c Client) Close() error {
	// Memcached client doesn't have a Close method
	return nil
}

// Health checks the health of the Memcached connection.
func (c Client) Health(ctx context.Context) error {
	// Try to ping the server by setting and getting a temporary key
	testKey := "health_check_kivigo"
	testValue := []byte("ping")

	if err := c.SetRaw(ctx, testKey, testValue); err != nil {
		return err
	}

	// Clean up the test key
	defer func() {
		_ = c.Delete(ctx, testKey)
	}()

	_, err := c.GetRaw(ctx, testKey)
	if err != nil {
		return err
	}

	return nil
}

// BatchGetRaw retrieves multiple keys from Memcached.
func (c Client) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	results := make(map[string][]byte, len(keys))

	items, err := c.c.GetMulti(keys)
	if err != nil {
		return nil, err
	}

	// Fill results map - memcache.GetMulti only returns existing keys
	for _, key := range keys {
		if item, exists := items[key]; exists {
			results[key] = item.Value
		} else {
			results[key] = nil // Key not found
		}
	}

	return results, nil
}

// BatchSetRaw stores multiple key-value pairs in Memcached.
func (c Client) BatchSetRaw(ctx context.Context, kv map[string][]byte) error {
	if len(kv) == 0 {
		return nil
	}

	// Memcached doesn't have native batch set, so we do individual sets
	// This is not atomic, but it's the best we can do with Memcached
	for key, value := range kv {
		if err := c.SetRaw(ctx, key, value); err != nil {
			return err
		}
	}

	return nil
}

// BatchDelete removes multiple keys from Memcached.
func (c Client) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// Memcached doesn't have native batch delete, so we do individual deletes
	var lastError error

	for _, key := range keys {
		if err := c.Delete(ctx, key); err != nil {
			lastError = err
		}
	}

	return lastError
}
