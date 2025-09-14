// Package postgresql implements a PostgreSQL backend for KiviGo.
//
// This backend uses PostgreSQL as the storage layer for key-value operations.
// It provides a SQL-based implementation with support for transactions,
// batch operations, and health checks.
//
// Example usage:
//
//	import "github.com/kivigo/kivigo/backend/postgresql"
//
//	// Create a new PostgreSQL client
//	client, err := postgresql.New(postgresql.DefaultOptions())
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Use with KiviGo
//	kv, err := kivigo.New(client)
//	if err != nil {
//		log.Fatal(err)
//	}
package postgresql
