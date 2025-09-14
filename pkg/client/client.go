package client

import (
	"context"
	"time"

	mencoder "github.com/kivigo/encoders/model"

	"github.com/kivigo/kivigo/pkg/models"
)

type (
	Client struct {
		models.KV
		opts  Option
		hooks *HooksRegistry
	}

	Options func(Option) Option

	Option struct {
		Encoder mencoder.Encoder
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
	return Client{
		KV:    kv,
		opts:  opts,
		hooks: NewHooksRegistry(),
	}, nil
}

func (c Client) Close() error {
	return c.KV.Close()
}

// RegisterHook registers a new hook with the client.
// Returns a unique hook ID, an error channel for receiving hook errors,
// and an unregister function to remove the hook.
//
// Example usage:
//
//	id, errCh, unregister := client.RegisterHook(func(ctx context.Context, evt EventType, key string, value []byte) error {
//	    log.Printf("Hook triggered: %s %s", evt, key)
//	    return nil
//	}, HookOptions{Events: []EventType{EventSet}})
//	defer unregister()
func (c Client) RegisterHook(cb HookFunc, opts HookOptions) (string, <-chan error, func()) {
	return c.hooks.RegisterHook(cb, opts)
}

// UnregisterHook removes a hook by its ID.
//
// Example usage:
//
//	client.UnregisterHook(hookID)
func (c Client) UnregisterHook(id string) {
	c.hooks.UnregisterHook(id)
}
