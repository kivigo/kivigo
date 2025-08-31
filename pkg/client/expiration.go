package client

import (
	"context"
	"fmt"
	"time"

	"github.com/azrod/kivigo/pkg/models"
)

// SupportsExpiration returns true if the underlying backend supports key expiration (TTL).
// This allows clients to check TTL capability at runtime before attempting to set expiration.
//
// Example usage:
//
//	if client.SupportsExpiration() {
//	    err := client.Expire(ctx, "mykey", 30*time.Second)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	} else {
//	    log.Println("Backend does not support expiration")
//	}
func (c Client) SupportsExpiration() bool {
	if backend, ok := c.KV.(models.KVWithTTL); ok {
		return backend.SupportsExpiration()
	}
	return false
}

// Expire sets a time-to-live (TTL) for the specified key.
// The key will be automatically deleted after the TTL duration.
// Returns an error if the operation fails, if the key does not exist,
// or if the backend does not support expiration.
//
// Example usage:
//
//	if c.SupportsExpiration() {
//	    if err := c.Expire(ctx, "mykey", 30*time.Second); err != nil {
//	        log.Fatal("Failed to set expiration:", err)
//	    }
//	} else {
//	    // Handle fallback behavior for backends without TTL support
//	}
func (c Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	backend, ok := c.KV.(models.KVWithTTL)
	if !ok {
		return fmt.Errorf("expiration not supported by backend")
	}
	return backend.Expire(ctx, key, ttl)
}