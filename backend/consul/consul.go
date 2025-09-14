package consul

import (
	"context"
	"errors"

	"github.com/hashicorp/consul/api"

	"github.com/kivigo/kivigo/pkg/errs"
	"github.com/kivigo/kivigo/pkg/models"
)

var (
	_ models.KV           = (*Client)(nil)
	_ models.KVWithBatch  = (*Client)(nil)
	_ models.KVWithHealth = (*Client)(nil)
)

type Client struct {
	cli *api.Client
}

func NewOptions() *api.Config {
	return &api.Config{}
}

func DefaultOptions() *api.Config {
	return api.DefaultConfig()
}

func New(opt *api.Config) (*Client, error) {
	client, err := api.NewClient(opt)
	if err != nil {
		return nil, err
	}

	return &Client{cli: client}, nil
}

func (c *Client) SetRaw(_ context.Context, key string, value []byte) error {
	_, err := c.cli.KV().Put(&api.KVPair{Key: key, Value: value}, nil)

	return err
}

func (c *Client) GetRaw(_ context.Context, key string) ([]byte, error) {
	kvp, _, err := c.cli.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	if kvp == nil {
		return nil, errs.ErrNotFound
	}

	return kvp.Value, nil
}

func (c *Client) Delete(_ context.Context, key string) error {
	_, err := c.cli.KV().Delete(key, nil)

	return err
}

func (c *Client) List(_ context.Context, prefix string) ([]string, error) {
	kvps, _, err := c.cli.KV().List(prefix, nil)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(kvps))
	for _, kv := range kvps {
		keys = append(keys, kv.Key)
	}

	return keys, nil
}

func (c *Client) Close() error {
	// Consul's client does not require explicit close
	return nil
}

// BatchGetRaw retrieves multiple keys from Consul.
func (c *Client) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
	results := make(map[string][]byte, len(keys))

	for _, key := range keys {
		val, err := c.GetRaw(ctx, key)
		if err != nil {
			results[key] = nil

			continue
		}

		results[key] = val
	}

	return results, nil
}

// BatchSetRaw sets multiple key-value pairs in Consul.
func (c *Client) BatchSetRaw(ctx context.Context, kv map[string][]byte) error {
	for key, value := range kv {
		if err := c.SetRaw(ctx, key, value); err != nil {
			return err
		}
	}

	return nil
}

// BatchDelete deletes multiple keys from Consul.
func (c *Client) BatchDelete(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := c.Delete(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Health(_ context.Context) error {
	if c.cli == nil {
		return errors.New("consul client is not initialized")
	}

	// Try to get the status leader as a simple health check
	_, err := c.cli.Status().Leader()
	if err != nil {
		return err
	}

	return nil
}
