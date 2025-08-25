package main

import (
	"context"
	"fmt"
	"log"

	"github.com/azrod/kivigo"
	"github.com/azrod/kivigo/backend/memcached"
)

func main() {
	// Create memcached backend with default options (localhost:11211)
	backend, err := memcached.New(memcached.DefaultOptions())
	if err != nil {
		log.Fatal("Failed to create memcached backend:", err)
	}

	// Create KiviGo client
	client, err := kivigo.New(backend)
	if err != nil {
		log.Fatal("Failed to create KiviGo client:", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Store some values
	fmt.Println("Storing values...")
	if err := client.Set(ctx, "user:1", "Alice"); err != nil {
		log.Fatal("Failed to set value:", err)
	}

	if err := client.Set(ctx, "user:2", "Bob"); err != nil {
		log.Fatal("Failed to set value:", err)
	}

	// Retrieve a value
	fmt.Println("Retrieving values...")
	var name string
	if err := client.Get(ctx, "user:1", &name); err != nil {
		log.Fatal("Failed to get value:", err)
	}
	fmt.Printf("user:1 = %s\n", name)

	// List all user keys
	fmt.Println("Listing keys...")
	keys, err := client.List(ctx, "user:")
	if err != nil {
		log.Fatal("Failed to list keys:", err)
	}
	fmt.Printf("Found keys: %v\n", keys)

	// Check backend health
	fmt.Println("Checking health...")
	if err := client.Health(ctx, nil); err != nil {
		log.Fatal("Backend health check failed:", err)
	}
	fmt.Println("Backend is healthy!")

	// Clean up
	fmt.Println("Cleaning up...")
	if err := client.Delete(ctx, "user:1"); err != nil {
		log.Fatal("Failed to delete key:", err)
	}
	if err := client.Delete(ctx, "user:2"); err != nil {
		log.Fatal("Failed to delete key:", err)
	}

	fmt.Println("Example completed successfully!")
}