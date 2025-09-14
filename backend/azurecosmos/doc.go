// Package azurecosmos provides a Azure Cosmos DB backend for KiviGo.
//
// This package implements the KiviGo interfaces using Azure Cosmos DB as the storage backend.
// It supports all standard operations including batch operations and health checks.
//
// Usage:
//
//	import "github.com/kivigo/kivigo/backend/azurecosmos"
//
//	// Create client with default options (emulator)
//	client, err := azurecosmos.New(azurecosmos.DefaultOptions())
//
//	// Create client with custom options
//	opts := azurecosmos.NewOptions()
//	opts.Endpoint = "https://myaccount.documents.azure.com:443/"
//	opts.Key = "your-cosmos-db-key"
//	opts.Database = "mydb"
//	opts.Container = "mycontainer"
//	client, err := azurecosmos.New(opts)
//
// The implementation uses Azure Cosmos DB's SQL API with JSON documents.
// Each key-value pair is stored as a document with the key as both the ID and partition key.
package azurecosmos
