package main

import (
	"context"
	"fmt"

	"github.com/azrod/kivigo"
	"github.com/azrod/kivigo/examples/custom_backend/internal/backend/memory"
)

func main() {
	client, err := kivigo.New(
		memory.New(),
	)
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
