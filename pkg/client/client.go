package client

import (
	"context"
	"time"

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

// New creates and returns a new KiviGo client instance using the provided backend and options.
//
// The kv parameter must implement the models.KV interface and represents the backend storage (e.g. local, Redis).
// The opts parameter allows you to specify options such as the encoder to use for value serialization.
//
// Example usage:
//
//	backend := backend.Local(local.Option{Path: "./"})
//	client, err := client.New(backend, client.Option{Encoder: encoder.JSON})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
// Returns a Client and an error if initialization fails.
func New(kv models.KV, opts Option) (Client, error) {
	return Client{KV: kv, opts: opts}, nil
}

func (c Client) Close() error {
	return c.KV.Close()
}
