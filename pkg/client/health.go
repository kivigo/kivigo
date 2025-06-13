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

// HealthCheck periodically checks the health of the backend and additional checks if provided.
// Returns a channel that receives errors if a check fails, or nil if healthy.
//
// Example (basic):
//
//	healthCh := client.HealthCheck(ctx, client.HealthOptions{Interval: 10 * time.Second})
//	go func() {
//	    for err := range healthCh {
//	        if err != nil {
//	            fmt.Println("Health issue:", err)
//	        } else {
//	            fmt.Println("Backend healthy")
//	        }
//	    }
//	}()
//
// Example with AdditionalChecks:
//
//	customCheck := func(ctx context.Context, c client.Client) error {
//	    var value string
//	    if err := c.Get(ctx, "health:ping", &value); err != nil {
//	        return fmt.Errorf("custom health check failed: %w", err)
//	    }
//	    return nil
//	}
//	healthCh := client.HealthCheck(ctx, client.HealthOptions{
//	    Interval:         10 * time.Second,
//	    AdditionalChecks: []client.HealthFunc{customCheck},
//	})
//	go func() {
//	    for err := range healthCh {
//	        if err != nil {
//	            fmt.Println("Health issue:", err)
//	        } else {
//	            fmt.Println("Backend healthy")
//	        }
//	    }
//	}()
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
