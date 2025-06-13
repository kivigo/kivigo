package main

import (
	"context"
	"fmt"
	"time"

	"github.com/azrod/kivigo"
	"github.com/azrod/kivigo/pkg/backend"
	"github.com/azrod/kivigo/pkg/backend/redis"
	"github.com/azrod/kivigo/pkg/client"
	"github.com/azrod/kivigo/pkg/encoder"
)

func main() {
	// Configure client with Redis backend and YAML encoder
	c, err := kivigo.New(
		backend.Redis(redis.Option{
			Addr: "localhost:6379",
		}),
		func(opt client.Option) client.Option {
			opt.Encoder = encoder.YAML
			return opt
		},
	)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	type User struct {
		Name string
		Age  int
	}

	// Store a struct
	user := User{Name: "Alice", Age: 30}
	if err := c.Set(context.Background(), "user:1", user); err != nil {
		panic(err)
	}

	// Retrieve the struct
	var u User
	if err := c.Get(context.Background(), "user:1", &u); err != nil {
		panic(err)
	}
	fmt.Printf("Retrieved user: %+v\n", u)

	// Periodic health check
	healthCh := c.HealthCheck(context.Background(), client.HealthOptions{
		Interval: 10 * time.Second,
	})
	go func() {
		for err := range healthCh {
			if err != nil {
				fmt.Println("Health issue:", err)
			} else {
				fmt.Println("Backend healthy")
			}
		}
	}()

	time.Sleep(12 * time.Second) // Let the health check run at least once
}
