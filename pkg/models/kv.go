package models

import (
	"context"
)

// KV is the main interface for key-value backends.
type (
	KV interface {
		Close() error

		// List returns a slice of all keys stored in the backend, optionally filtered by prefix.
		// Returns an error if the operation fails.
		//
		// Example:
		//   keys, err := client.List(ctx, "user:")
		//   for _, k := range keys {
		//       fmt.Println("Key:", k)
		//   }
		List(ctx context.Context, prefix string) (keys []string, err error)

		// GetRaw retrieves the raw (encoded) value stored under the specified key.
		// Returns the value as a byte slice, or an error if the key does not exist.
		//
		// Example:
		//   raw, err := backend.GetRaw(ctx, "myKey")
		//   if err != nil {
		//       log.Fatal(err)
		//   }
		//   fmt.Println("Raw value:", string(raw))
		GetRaw(ctx context.Context, key string) (value []byte, err error)

		// SetRaw stores the given raw (encoded) value under the specified key.
		// Returns an error if the operation fails.
		//
		// Example:
		//   err := backend.SetRaw(ctx, "myKey", []byte("myValue"))
		//   if err != nil {
		//       log.Fatal(err)
		//   }
		SetRaw(ctx context.Context, key string, value []byte) error

		// Delete removes the value associated with the specified key.
		// Returns an error if the operation fails.
		//
		// Example:
		//   err := client.Delete(ctx, "myKey")
		//   if err != nil {
		//       log.Fatal(err)
		//   }
		Delete(ctx context.Context, key string) error
	}

	KVWithBatch interface {
		// BatchGet retrieves multiple raw values for the given keys.
		// Returns a map of key to raw value, or an error if the operation fails or is not supported.
		//
		// Example:
		//   raws, err := backend.BatchGet(ctx, []string{"foo", "bar"})
		//   if err != nil {
		//       log.Fatal(err)
		//   }
		//   for k, v := range raws {
		//       fmt.Printf("%s: %s\n", k, string(v))
		//   }
		BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error)

		// BatchSet stores multiple key-value pairs atomically if supported by the backend.
		// Returns an error if the operation fails or is not supported.
		//
		// Example:
		//   err := backend.BatchSet(ctx, map[string][]byte{"foo": []byte("1"), "bar": []byte("2")})
		//   if err != nil {
		//       log.Fatal(err)
		//   }
		BatchSetRaw(ctx context.Context, kv map[string][]byte) error

		// BatchDelete removes multiple keys atomically if supported by the backend.
		// Returns an error if the operation fails or is not supported.
		//
		// Example:
		//   err := backend.BatchDelete(ctx, []string{"foo", "bar"})
		//   if err != nil {
		//       log.Fatal(err)
		//   }
		BatchDelete(ctx context.Context, keys []string) error
	}

	KVWithHealth interface {
		// Health checks the health of the backend connection.
		// Returns nil if healthy, or an error otherwise.
		//
		// Example:
		//   err := backend.Health(ctx)
		//   if err != nil {
		//       log.Fatal("Backend not healthy:", err)
		//   }
		Health(ctx context.Context) error
	}
)
