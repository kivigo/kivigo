package main

import (
	"context"
	"errors"
	"time"

	"github.com/azrod/kivigo"
	"github.com/azrod/kivigo/pkg/backend"
	"github.com/azrod/kivigo/pkg/backend/local"
	"github.com/azrod/kivigo/pkg/client"
)

func myCustomHealth(ctx context.Context, c client.Client) error {
	// Example: check if a specific key exists
	var value string
	err := c.Get(ctx, "health:ping", &value)
	if err != nil {
		return errors.New("custom health check failed: " + err.Error())
	}
	return nil
}

func main() {
	c, err := kivigo.New(
		backend.Local(local.Option{
			Path: "./",
		}),
	)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// Use HealthCheck with your custom logic
	healthCh := c.HealthCheck(context.Background(), client.HealthOptions{
		Interval:         5 * time.Second,
		AdditionalChecks: []client.HealthFunc{myCustomHealth},
	})

	go func() {
		for err := range healthCh {
			if err != nil {
				println(time.Now().Format(time.RFC3339), "Custom health issue:", err.Error())
			} else {
				println(time.Now().Format(time.RFC3339), "Custom health OK")
			}
		}
	}()

	time.Sleep(7 * time.Second)

	// Simulate setting a health key
	if err := c.Set(context.Background(), "health:ping", "pong"); err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Second)
}
