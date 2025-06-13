package client

import "context"

func (c Client) Get(ctx context.Context, key string, value any) error {
	vV, err := c.GetRaw(ctx, key)
	if err != nil {
		return err
	}

	return c.opts.Encoder.Decode(vV, value)
}

func (c Client) Set(ctx context.Context, key string, value any) error {
	vV, err := c.opts.Encoder.Encode(value)
	if err != nil {
		return err
	}

	return c.SetRaw(ctx, key, vV)
}
