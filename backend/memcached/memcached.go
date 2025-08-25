package memcached

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/azrod/kivigo/pkg/errs"
	"github.com/azrod/kivigo/pkg/models"
)

var (
	_ models.KV           = (*Client)(nil)
	_ models.KVWithHealth = (*Client)(nil)
	_ models.KVWithBatch  = (*Client)(nil)
)

type (
	Client struct {
		c        *memcache.Client
		keyIndex map[string]bool // Track keys for listing functionality
		mu       sync.RWMutex    // Protect keyIndex
	}

	Option struct {
		Servers []string `default:"localhost:11211"`
	}
)

// NewOptions returns a new set of options for the Memcached client.
func NewOptions() Option {
	return Option{}
}

// DefaultOptions returns the default options for the Memcached client.
// Servers: []string{"localhost:11211"}
func DefaultOptions() Option {
	return Option{
		Servers: []string{"localhost:11211"},
	}
}

// New returns a new Memcached client.
func New(opt Option) (*Client, error) {
	if len(opt.Servers) == 0 {
		opt.Servers = []string{"localhost:11211"}
	}

	client := memcache.New(opt.Servers...)

	return &Client{
		c:        client,
		keyIndex: make(map[string]bool),
	}, nil
}

func (c *Client) SetRaw(ctx context.Context, key string, value []byte) error {
	// Check if key is not empty
	if key == "" {
		return errs.ErrEmptyKey
	}

	item := &memcache.Item{
		Key:   key,
		Value: value,
	}

	if err := c.c.Set(item); err != nil {
		return err
	}

	// Track the key for listing
	c.mu.Lock()
	c.keyIndex[key] = true
	c.mu.Unlock()

	return nil
}

func (c *Client) GetRaw(ctx context.Context, key string) ([]byte, error) {
	// Check if key is not empty
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	item, err := c.c.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return item.Value, nil
}

// List returns all keys with the given prefix.
// Note: Memcached doesn't natively support listing keys, so we maintain a local index.
func (c *Client) List(ctx context.Context, prefix string) ([]string, error) {
	// Check if prefix is not empty
	if prefix == "" {
		return nil, errs.ErrEmptyPrefix
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []string
	for key := range c.keyIndex {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// Delete removes the value associated with the specified key.
func (c *Client) Delete(ctx context.Context, key string) error {
	// Check if key is not empty
	if key == "" {
		return errs.ErrEmptyKey
	}

	if err := c.c.Delete(key); err != nil {
		if err == memcache.ErrCacheMiss {
			// Key doesn't exist, but that's not an error for delete operation
			return nil
		}
		return err
	}

	// Remove from key index
	c.mu.Lock()
	delete(c.keyIndex, key)
	c.mu.Unlock()

	return nil
}

// Close closes the database connection.
func (c *Client) Close() error {
	// Memcache client doesn't have a Close method, so we just clear our index
	c.mu.Lock()
	c.keyIndex = make(map[string]bool)
	c.mu.Unlock()
	return nil
}

func (c *Client) Health(ctx context.Context) error {
	// Check if the client is nil
	if c.c == nil {
		return errs.ErrClientNotInitialized
	}

	// Try to ping the memcached server by attempting to get a non-existent key
	_, err := c.c.Get("__health_check__")
	if err != nil && err != memcache.ErrCacheMiss {
		return errs.ErrHealthCheckFailed(err)
	}

	return nil
}

// BatchGetRaw retrieves multiple keys from the database.
func (c *Client) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
	// Check if keys slice is not empty
	if len(keys) == 0 {
		return nil, errs.ErrEmptyBatch
	}

	results := make(map[string][]byte, len(keys))

	items, err := c.c.GetMulti(keys)
	if err != nil {
		return nil, err
	}

	// Initialize all keys to nil first
	for _, key := range keys {
		results[key] = nil
	}

	// Fill in the found values
	for key, item := range items {
		results[key] = item.Value
	}

	return results, nil
}

// BatchSetRaw stores multiple key-value pairs in the database.
// Note: This is not atomic in memcached, each operation is independent.
func (c *Client) BatchSetRaw(ctx context.Context, kv map[string][]byte) error {
	if len(kv) == 0 {
		return errs.ErrEmptyBatch
	}

	// Prepare items for batch set
	items := make([]*memcache.Item, 0, len(kv))
	for key, value := range kv {
		items = append(items, &memcache.Item{
			Key:   key,
			Value: value,
		})
	}

	// Set all items (not atomic, but efficient)
	for _, item := range items {
		if err := c.c.Set(item); err != nil {
			return fmt.Errorf("failed to set key %s: %w", item.Key, err)
		}

		// Track the key for listing
		c.mu.Lock()
		c.keyIndex[item.Key] = true
		c.mu.Unlock()
	}

	return nil
}

// BatchDelete removes multiple keys from the database.
// Note: This is not atomic in memcached, each operation is independent.
func (c *Client) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return errs.ErrEmptyBatch
	}

	for _, key := range keys {
		if err := c.c.Delete(key); err != nil && err != memcache.ErrCacheMiss {
			return fmt.Errorf("failed to delete key %s: %w", key, err)
		}

		// Remove from key index
		c.mu.Lock()
		delete(c.keyIndex, key)
		c.mu.Unlock()
	}

	return nil
}
