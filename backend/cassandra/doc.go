/*
Package cassandra provides a Cassandra backend for KiviGo.

This backend uses Apache Cassandra as the underlying storage engine,
implementing the KV, KVWithBatch, and KVWithHealth interfaces.

Example usage:

	import "github.com/azrod/kivigo/backend/cassandra"

	// Create a new Cassandra client
	client, err := cassandra.New(cassandra.DefaultOptions())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Use with KiviGo
	kvStore, err := kivigo.New(client)
	if err != nil {
		log.Fatal(err)
	}
	defer kvStore.Close()
*/
package cassandra