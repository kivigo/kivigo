package client

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/azrod/kivigo/pkg/models"
)

type (
	Client struct {
		models.KV
		opts Option
	}

	Options func(Option) Option

	Option struct {
		Encoder models.Encoder
	}

	HealthFunc func(ctx context.Context, c Client) error

	HealthOptions struct {
		Interval time.Duration // Default: 1 minute

		// Additional health checks
		// These checks will be executed in addition to the default health check.
		// If any of these checks fail, the health check will return an error.
		// All checks are executed in parallel.
		AdditionalChecks []HealthFunc
	}
)

func New(kv models.KV, opts Option) (Client, error) {
	return Client{KV: kv, opts: opts}, nil
}

func (c Client) Close() error {
	return c.KV.Close()
}

// Health is a instantaneous check to see if the client is healthy.
// If you want to check regularly, HealthCheck function should be used.
func (c Client) Health(ctx context.Context, additionalChecks []HealthFunc) error {
	eg := errgroup.Group{}
	eg.Go(func() error {
		return c.KV.Health(ctx)
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
