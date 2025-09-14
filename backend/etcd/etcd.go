package etcd

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/kivigo/kivigo/pkg/errs"
	"github.com/kivigo/kivigo/pkg/models"
)

type Option clientv3.Config

func NewOptions() Option {
	return Option{}
}

func DefaultOptions(endpoints ...string) Option {
	return Option{
		Endpoints: endpoints,
	}
}

type Client struct {
	cli *clientv3.Client
}

var (
	_ models.KV           = (*Client)(nil)
	_ models.KVWithBatch  = (*Client)(nil)
	_ models.KVWithHealth = (*Client)(nil)
)

func New(opt Option) (*Client, error) {
	cli, err := clientv3.New(clientv3.Config(opt))
	if err != nil {
		return nil, err
	}

	return &Client{cli: cli}, nil
}

func (c *Client) SetRaw(ctx context.Context, key string, value []byte) error {
	_, err := c.cli.Put(ctx, key, string(value))

	return err
}

func (c *Client) GetRaw(ctx context.Context, key string) ([]byte, error) {
	resp, err := c.cli.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, errs.ErrNotFound
	}

	return resp.Kvs[0].Value, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.cli.Delete(ctx, key)

	return err
}

func (c *Client) List(ctx context.Context, prefix string) ([]string, error) {
	resp, err := c.cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		keys = append(keys, string(kv.Key))
	}

	return keys, nil
}

func (c *Client) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return nil, errs.ErrEmptyKey
	}

	ops := make([]clientv3.Op, 0, len(keys))
	for _, key := range keys {
		ops = append(ops, clientv3.OpGet(key))
	}

	txnResp, err := c.cli.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return nil, err
	}

	results := make(map[string][]byte, len(keys))

	for i, resp := range txnResp.Responses {
		getResp := resp.GetResponseRange()
		if getResp == nil || len(getResp.Kvs) == 0 {
			results[keys[i]] = nil

			continue
		}

		results[keys[i]] = getResp.Kvs[0].Value
	}

	return results, nil
}

func (c *Client) BatchSetRaw(ctx context.Context, kv map[string][]byte) error {
	if len(kv) == 0 {
		return errs.ErrEmptyKey
	}

	ops := make([]clientv3.Op, 0, len(kv))
	for key, value := range kv {
		ops = append(ops, clientv3.OpPut(key, string(value)))
	}

	txn := c.cli.Txn(ctx).Then(ops...)
	_, err := txn.Commit()

	return err
}

func (c *Client) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return errs.ErrEmptyKey
	}

	ops := make([]clientv3.Op, 0, len(keys))
	for _, key := range keys {
		ops = append(ops, clientv3.OpDelete(key))
	}

	txn := c.cli.Txn(ctx).Then(ops...)
	_, err := txn.Commit()

	return err
}

func (c *Client) Health(ctx context.Context) error {
	_, err := c.cli.Status(ctx, c.cli.Endpoints()[0])

	return err
}

func (c *Client) Close() error {
	return c.cli.Close()
}
