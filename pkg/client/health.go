package client

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/azrod/kivigo/pkg/models"
)

// Health is a instantaneous check to see if the client is healthy.
// If you want to check regularly, HealthCheck function should be used.
func (c Client) Health(ctx context.Context, additionalChecks []HealthFunc) error {
	fn, ok := c.KV.(models.KVWithHealth)
	if !ok {
		return nil // No health check available, consider it healthy
	}

	eg := errgroup.Group{}
	eg.Go(func() error {
		return fn.Health(ctx)
	})
	for _, check := range additionalChecks {
		eg.Go(func() error {
			return check(ctx, c)
		})
	}
	return eg.Wait()
}

// HealthCheck is a periodic check to see if the client is healthy.
// It returns a channel that will receive the health status.
func (c Client) HealthCheck(ctx context.Context, ho HealthOptions) <-chan error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ticker := time.NewTicker(ho.Interval)
		if ho.Interval <= 0 {
			ticker = time.NewTicker(1 * time.Minute) // Default to 1 minute if no interval is set
		}
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.Health(ctx, ho.AdditionalChecks); err != nil {
					ch <- err
				} else {
					ch <- nil
				}
			}
		}
	}()
	return ch
}
