package redis

import (
	"context"
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"

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
		_ models.Backend
		c *redis.Client
	}
	Option redis.Options
)

func New(opt Option) (Client, error) {
	client := redis.NewClient((*redis.Options)(&opt))

	return Client{c: client}, nil
}

func (c Client) SetRaw(ctx context.Context, key string, value []byte) error {
	// Check if key is not empty
	if key == "" {
		return errs.ErrEmptyKey
	}

	return c.c.Set(ctx, key, string(value), 0).Err()
}

func (c Client) GetRaw(ctx context.Context, key string) ([]byte, error) {
	// Check if key is not empty
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	v, err := c.c.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errs.ErrNotFound
		}

		return nil, err
	}

	return []byte(v), nil
}

// List lists all the keys from the database.
func (c Client) List(ctx context.Context, prefix string) (keys []string, err error) {
	// Check if prefix is not empty
	if prefix == "" {
		return nil, errs.ErrEmptyPrefix
	}

	pattern := func() string {
		if strings.HasSuffix(prefix, ":") {
			return prefix + "*"
		}

		return prefix + ":*"
	}()

	return c.c.Keys(ctx, pattern).Result()
}

// Delete deletes the value for the given key.
func (c Client) Delete(ctx context.Context, key string) error {
	// Check if key is not empty
	if key == "" {
		return errs.ErrEmptyKey
	}

	return c.c.Del(ctx, key).Err()
}

// Close closes the database connection.
func (c Client) Close() error {
	return c.c.Close()
}

func (c Client) Health(ctx context.Context) error {
	// Check if the client is nil
	if c.c == nil {
		return errs.ErrClientNotInitialized
	}

	// Ping the Redis server to check health
	if err := c.c.Ping(ctx).Err(); err != nil {
		return errs.ErrHealthCheckFailed(err)
	}

	return nil
}

// BatchGet retrieves multiple keys from the database.
func (c Client) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
	// Check if keys slice is not empty
	if len(keys) == 0 {
		return nil, errs.ErrEmptyBatch
	}

	results := make(map[string][]byte, len(keys))

	v, err := c.c.MGet(ctx, keys...).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errs.ErrNotFound
		}

		return nil, err
	}

	for i, key := range keys {
		if v[i] == nil {
			results[key] = nil

			continue
		}

		results[key] = []byte(v[i].(string))
	}

	return results, nil
}

// BatchSet sets multiple key-value pairs in the database.
func (c Client) BatchSetRaw(ctx context.Context, kv map[string][]byte) error {
	if len(kv) == 0 {
		return errs.ErrEmptyBatch
	}

	pipe := c.c.Pipeline()
	for key, value := range kv {
		if err := pipe.Set(ctx, key, string(value), 0).Err(); err != nil {
			return err
		}
	}

	_, err := pipe.Exec(ctx)

	return err
}

// BatchDelete deletes multiple keys from the database.
func (c Client) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return errs.ErrEmptyBatch
	}

	pipe := c.c.Pipeline()
	for _, key := range keys {
		if err := pipe.Del(ctx, key).Err(); err != nil {
			return err
		}
	}

	_, err := pipe.Exec(ctx)

	return err
}
