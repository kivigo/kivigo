/*
Package kivigo is a lightweight, extensible key-value store library for Go.

KiviGo provides a unified interface for storing and retrieving key-value pairs,
with support for multiple backends (such as Redis and BoltDB) and pluggable encoders (JSON, YAML, etc.).

Features:
  - Simple API for Get, Set, Delete, List operations
  - Multiple backends: Redis, BoltDB (local), and more
  - Pluggable encoding (JSON, YAML, ...)
  - Health checks (with custom checks)
  - Extensible: add your own backend or encoder
  - Batch operations and transactions (if supported by backend)
  - Key expiration (TTL) and hooks (events) [if supported]

Example usage:

	client, err := kivigo.New(
	    backend.Local(local.Option{Path: "./"}),
	)
	if err != nil {
	    panic(err)
	}
	defer client.Close()

	if err := client.Set(context.Background(), "myKey", "myValue"); err != nil {
	    panic(err)
	}

	var value string
	if err := client.Get(context.Background(), "myKey", &value); err != nil {
	    panic(err)
	}
	fmt.Println("Value:", value)

See README.md for more advanced examples and backend-specific features.
*/
package kivigo
