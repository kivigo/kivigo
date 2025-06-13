package main

import (
	"context"
	"fmt"

	"github.com/azrod/kivigo"
	"github.com/azrod/kivigo/pkg/backend"
	"github.com/azrod/kivigo/pkg/backend/local"
)

func main() {
	// Example usage of the backend
	client, err := kivigo.New(
		backend.Local(local.Option{
			Path: "./",
		}),
	)
	if err != nil {
		panic(err)
	}

	defer client.Close()

	// Use the client to interact with the local database
	// For example, you can set a key-value pair
	if err = client.Set(context.Background(), "exampleKey", "exampleValue"); err != nil {
		panic(err)
	}

	// Retrieve the value
	var value string
	if err = client.Get(context.Background(), "exampleKey", &value); err != nil {
		panic(err)
	}

	if value != "exampleValue" {
		panic("expected value to be 'exampleValue', got: " + value)
	}

	// Output: Successfully set and retrieved value from local database.
	// Note: This is a simple example. In a real application, you would handle errors and context properly.
	// You might also want to implement more complex operations like transactions, batch operations, etc.
	// Make sure to import the necessary packages and handle any errors appropriately.
	// This example assumes that the kivigo package and its dependencies are correctly set up in your Go environment.
	// You can also implement more complex operations like transactions, batch operations, etc.

	fmt.Println("Successfully set and retrieved value from local database.")

}
