---
sidebar_position: 2
---

# Getting Started

This guide will help you get up and running with KiviGo in just a few minutes.

## Installation

First, install the main KiviGo library:

```bash
go get github.com/kivigo/kivigo
```

Then install the backend you want to use. For example, to use the BadgerDB backend:

```bash
go get github.com/kivigo/kivigo/backend/badger
```

## Quick Start Example

Here's a simple example using the BadgerDB backend with the default JSON encoder:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/kivigo/backend/badger"
)

func main() {
    // Create a BadgerDB backend
    opt := badger.DefaultOptions("./data")
    kvStore, err := badger.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()

    // Create KiviGo client
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Store a simple value
    err = client.Set(ctx, "greeting", "Hello, KiviGo!")
    if err != nil {
        log.Fatal(err)
    }

    // Retrieve the value
    var greeting string
    err = client.Get(ctx, "greeting", &greeting)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Retrieved:", greeting)
    // Output: Retrieved: Hello, KiviGo!
}
```

## Working with Structs

KiviGo automatically marshals and unmarshals Go structs:

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // ... setup client as above ...

    ctx := context.Background()

    // Store a struct
    user := User{
        ID:    1,
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    err := client.Set(ctx, "user:1", user)
    if err != nil {
        log.Fatal(err)
    }

    // Retrieve the struct
    var retrievedUser User
    err = client.Get(ctx, "user:1", &retrievedUser)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("User: %+v\n", retrievedUser)
    // Output: User: {ID:1 Name:John Doe Email:john@example.com}
}
```

## Common Operations

### Setting Values

```go
// Simple values
err := client.Set(ctx, "counter", 42)

// Complex structs
err := client.Set(ctx, "config", ConfigStruct{...})

// Slices and maps
err := client.Set(ctx, "items", []string{"a", "b", "c"})
```

### Getting Values

```go
// Get into a variable of the correct type
var counter int
err := client.Get(ctx, "counter", &counter)

var config ConfigStruct
err := client.Get(ctx, "config", &config)

var items []string
err := client.Get(ctx, "items", &items)
```

### Listing Keys

```go
// List all keys with a prefix
keys, err := client.List(ctx, "user:")
// Returns: ["user:1", "user:2", "user:3", ...]
```

### Deleting Values

```go
err := client.Delete(ctx, "user:1")
```

## Different Backends

KiviGo supports multiple backends. Here are some quick examples:

### Redis Backend

```bash
go get github.com/kivigo/kivigo/backend/redis
```

```go
import "github.com/kivigo/kivigo/backend/redis"

opt := redis.DefaultOptions()
opt.Addr = "localhost:6379"
kvStore, err := redis.New(opt)
```

### Local/BoltDB Backend

```bash
go get github.com/kivigo/kivigo/backend/local
```

```go
import "github.com/kivigo/kivigo/backend/local"

kvStore, err := local.New(local.Option{Path: "./data.db"})
```

### Consul Backend

```bash
go get github.com/kivigo/kivigo/backend/consul
```

```go
import "github.com/kivigo/kivigo/backend/consul"

opt := consul.DefaultOptions()
opt.Address = "localhost:8500"
kvStore, err := consul.New(opt)
```

## Custom Encoders

By default, KiviGo uses JSON encoding. You can specify different encoders:

### YAML Encoder

```go
import "github.com/kivigo/kivigo/pkg/encoder"

client, err := client.New(kvStore, client.Option{
    Encoder: encoder.YAML,
})
```

### JSON Encoder (Default)

```go
import "github.com/kivigo/kivigo/pkg/encoder"

client, err := client.New(kvStore, client.Option{
    Encoder: encoder.JSON, // This is the default
})
```

## Error Handling

KiviGo provides specific error types for common scenarios:

```go
import "github.com/kivigo/kivigo/pkg/errs"

var value string
err := client.Get(ctx, "nonexistent", &value)
if err != nil {
    if errors.Is(err, errs.ErrNotFound) {
        fmt.Println("Key not found")
    } else {
        log.Fatal("Other error:", err)
    }
}
```

## Health Checks

Most backends support health checks:

```go
err := kvStore.Health(ctx)
if err != nil {
    log.Printf("Backend unhealthy: %v", err)
} else {
    log.Println("Backend is healthy")
}
```

## Next Steps

Now that you have the basics:

1. **Explore [Backends](./backends/overview)** - Learn about all available storage backends
2. **Check [Advanced Features](./advanced/health-checks)** - Discover batch operations, custom backends, and more
3. **Read about [Testing](./advanced/mock-testing)** - Learn how to test your code with the mock backend

## Complete Example

Here's a more comprehensive example that demonstrates error handling and multiple operations:

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/kivigo/pkg/errs"
    "github.com/kivigo/kivigo/backend/badger"
)

type Config struct {
    AppName     string `json:"app_name"`
    Port        int    `json:"port"`
    Debug       bool   `json:"debug"`
    DatabaseURL string `json:"database_url"`
}

func main() {
    // Setup
    opt := badger.DefaultOptions("./app_data")
    kvStore, err := badger.New(opt)
    if err != nil {
        log.Fatal("Failed to create backend:", err)
    }
    defer kvStore.Close()

    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal("Failed to create client:", err)
    }

    ctx := context.Background()

    // Check backend health
    if err := kvStore.Health(ctx); err != nil {
        log.Fatal("Backend unhealthy:", err)
    }
    fmt.Println("✅ Backend is healthy")

    // Store configuration
    config := Config{
        AppName:     "MyApp",
        Port:        8080,
        Debug:       true,
        DatabaseURL: "postgres://localhost/myapp",
    }

    err = client.Set(ctx, "app:config", config)
    if err != nil {
        log.Fatal("Failed to store config:", err)
    }
    fmt.Println("✅ Configuration stored")

    // Retrieve configuration
    var retrievedConfig Config
    err = client.Get(ctx, "app:config", &retrievedConfig)
    if err != nil {
        log.Fatal("Failed to retrieve config:", err)
    }
    fmt.Printf("✅ Configuration retrieved: %+v\n", retrievedConfig)

    // Try to get a non-existent key
    var missing string
    err = client.Get(ctx, "nonexistent", &missing)
    if err != nil {
        if errors.Is(err, errs.ErrNotFound) {
            fmt.Println("✅ Correctly handled missing key")
        } else {
            log.Fatal("Unexpected error:", err)
        }
    }

    // List keys
    keys, err := client.List(ctx, "app:")
    if err != nil {
        log.Fatal("Failed to list keys:", err)
    }
    fmt.Printf("✅ Found keys: %v\n", keys)

    // Clean up
    err = client.Delete(ctx, "app:config")
    if err != nil {
        log.Fatal("Failed to delete config:", err)
    }
    fmt.Println("✅ Configuration deleted")
}
```

This example shows a complete workflow including error handling, health checks, and cleanup operations.
