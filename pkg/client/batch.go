package client

import (
	"context"
	"fmt"

	"github.com/azrod/kivigo/pkg/models"
)

func (c Client) BatchGet(ctx context.Context, keys []string, dest map[string]any) error {
	batch, ok := c.KV.(models.KVWithBatch)
	if !ok {
		return fmt.Errorf("BatchGet not supported by backend")
	}

	raws, err := batch.BatchGet(ctx, keys)
	if err != nil {
		return err
	}

	for k, raw := range raws {
		if d, ok := dest[k]; ok {
			if err := c.opts.Encoder.Decode(raw, d); err != nil {
				return err
			}
		}
	}

	return nil
}
