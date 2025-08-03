package main

import (
	"context"
	"errors"
	"time"

	"github.com/azrod/kivigo"
	"github.com/azrod/kivigo/backend/local"
	"github.com/azrod/kivigo/pkg/client"
)

func myCustomHealth(ctx context.Context, c client.Client) error {
	// Example: check if a specific key exists
	var value string

	err := c.Get(ctx, "health:ping", &value)
	if err != nil || value != "pong" {
		return errors.New("custom health check failed: " + err.Error())
	}

	return nil
}

func main() {
	kvStore, err := local.New(local.Option{Path: "./"})
	if err != nil {
		panic(err)
	}

	c, err := kivigo.New(kvStore)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// Use HealthCheck with your custom logic
	healthCh := c.HealthCheck(context.Background(), client.HealthOptions{
		Interval:         500 * time.Millisecond,
		AdditionalChecks: []client.HealthFunc{myCustomHealth},
	})

	defer func() {
		if err := c.Delete(context.Background(), "health:ping"); err != nil {
			panic(err)
		}
	}()

	go func() {
		for err := range healthCh {
			if err != nil {
				println(time.Now().Format(time.RFC3339), "Custom health issue:", err.Error())
			} else {
				println(time.Now().Format(time.RFC3339), "Custom health OK")
			}
		}
	}()

	time.Sleep(1500 * time.Millisecond)

	// Simulate setting a health key
	if err := c.Set(context.Background(), "health:ping", "pong"); err != nil {
		panic(err)
	}

	time.Sleep(3 * time.Second)
}
