package models

import (
	"context"
	"time"
)

// KVWithTTL is an optional interface for backends that support key expiration (TTL).
// Backends implementing this interface can set expiration times on keys.
type KVWithTTL interface {
	// SupportsExpiration returns true if the backend supports key expiration.
	// This method allows clients to check TTL capability at runtime.
	//
	// Example:
	//   if backend.SupportsExpiration() {
	//       // Safe to use Expire method
	//   }
	SupportsExpiration() bool

	// Expire sets a time-to-live (TTL) for the specified key.
	// The key will be automatically deleted after the TTL duration.
	// Returns an error if the operation fails or if the key does not exist.
	//
	// Example:
	//   err := backend.Expire(ctx, "myKey", 30*time.Second)
	//   if err != nil {
	//       log.Fatal("Failed to set expiration:", err)
	//   }
	Expire(ctx context.Context, key string, ttl time.Duration) error
}
