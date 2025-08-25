/*
Package memcached provides a Memcached backend implementation for KiviGo.

This package implements the models.KV, models.KVWithBatch, and models.KVWithHealth
interfaces using Memcached as the storage backend.

# Features

  - Full KV interface support (Get, Set, Delete, List)
  - Batch operations (not atomic due to Memcached limitations)
  - Health checking via ping operations
  - Configurable server list for distributed Memcached clusters

# Usage

Basic usage:

	import "github.com/azrod/kivigo/backend/memcached"

	// Create client with default options (localhost:11211)
	client, err := memcached.New(memcached.DefaultOptions())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Use with KiviGo
	kv, err := kivigo.New(client)
	if err != nil {
		log.Fatal(err)
	}

Custom configuration:

	opt := memcached.NewOptions()
	opt.Servers = []string{"server1:11211", "server2:11211"}

	client, err := memcached.New(opt)
	if err != nil {
		log.Fatal(err)
	}

# Limitations

  - Memcached doesn't natively support key listing, so this implementation
    maintains a local key index. This means the List() operation only returns
    keys that were set through this client instance.
  - Batch operations are not atomic - each operation in a batch is performed
    independently.
  - The key index is not persistent across client restarts.

# Configuration

The Option struct supports the following fields:

  - Servers: []string - List of Memcached server addresses (default: ["localhost:11211"])
*/
package memcached
