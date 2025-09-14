/*
Package mysql provides a MySQL backend for KiviGo.

This backend stores key-value pairs in a MySQL database table,
implementing all KiviGo interfaces including batch operations and health checks.

Usage:

	import "github.com/kivigo/kivigo/backend/mysql"

	// Create client with default options
	opts := mysql.DefaultOptions()
	client, err := mysql.New(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Use with KiviGo
	kv, err := kivigo.New(client)
	if err != nil {
		log.Fatal(err)
	}

The MySQL backend requires a running MySQL server and creates a table
named 'kv_store' to store the key-value pairs.
*/
package mysql
