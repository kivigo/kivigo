package main

import (
	"context"
	"fmt"

	"github.com/azrod/kivigo"
	"github.com/azrod/kivigo/backend/local"
)

func main() {
	kvStore, err := local.New(local.DefaultOptions())
	if err != nil {
		panic(err)
	}

	client, err := kivigo.New(kvStore)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Store a value
	if err := client.Set(context.Background(), "myKey", "myValue"); err != nil {
		panic(err)
	}

	// Retrieve the value
	var value string
	if err := client.Get(context.Background(), "myKey", &value); err != nil {
		panic(err)
	}

	fmt.Println("Retrieved value:", value)
}
