package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/azrod/kivigo/pkg/errs"
)

// HasKey checks if the specified key exists in the store.
// Returns an error if the key is empty.
func (c Client) HasKey(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, errs.ErrEmptyKey
	}

	_, err := c.GetRaw(ctx, key)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return true, nil
}

// HasKeys checks if the specified keys exists in the store.
// Returns an error if the keys is empty.
func (c Client) HasKeys(ctx context.Context, keys []string) (bool, error) {
	if len(keys) == 0 {
		return false, errs.ErrEmptyKey
	}

	for _, key := range keys {
		exists, err := c.HasKey(ctx, key)
		if err != nil {
			return false, fmt.Errorf("failed to check key existence: %w", err)
		}
		if !exists {
			return false, nil
		}
	}

	return true, nil
}

type MatchKeysFunc func(keys []string) (bool, error)

// MatchKeys lists all keys with the given prefix and applies the provided MatchKeysFunc to determine if they match a custom condition.
// Returns false and an error if the function is nil or listing keys fails.
//
// Example:
//
//	ok, err := client.MatchKeys(ctx, "prefix", func(keys []string) (bool, error) {
//	    // Custom matching logic
//	    return len(keys) > 0, nil
//	})
func (c Client) MatchKeys(ctx context.Context, prefix string, f MatchKeysFunc) (bool, error) {
	if f == nil {
		return false, errs.ErrEmptyFunc
	}

	// List all keys in the store
	keys, err := c.List(ctx, prefix)
	if err != nil {
		return false, fmt.Errorf("failed to list keys: %w", err)
	}

	// Check if any key matches the criteria
	return f(keys)
}

// Get retrieves the value stored under the specified key and decodes it into dest.
// Returns an error if the key does not exist or decoding fails.
//
// Example:
//
//	var value string
//	err := client.Get(ctx, "myKey", &value)
func (c Client) Get(ctx context.Context, key string, value any) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	vV, err := c.GetRaw(ctx, key)
	if err != nil {
		return err
	}

	return c.opts.Encoder.Decode(vV, value)
}

// Set stores the given value under the specified key.
// Returns an error if the operation fails.
//
// Example:
//
//	err := client.Set(ctx, "myKey", "myValue")
func (c Client) Set(ctx context.Context, key string, value any) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	vV, err := c.opts.Encoder.Encode(value)
	if err != nil {
		return err
	}

	err = c.SetRaw(ctx, key, vV)
	if err != nil {
		return err
	}

	// Trigger hooks after successful operation
	if c.hooks != nil {
		c.hooks.Run(ctx, EventSet, key, vV)
	}

	return nil
}

// Delete removes the value associated with the specified key.
// Returns an error if the operation fails.
//
// Example:
//
//	err := client.Delete(ctx, "myKey")
func (c Client) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	err := c.KV.Delete(ctx, key)
	if err != nil {
		return err
	}

	// Trigger hooks after successful operation
	if c.hooks != nil {
		c.hooks.Run(ctx, EventDelete, key, nil)
	}

	return nil
}
