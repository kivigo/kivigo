---
sidebar_position: 3
---

# Quick Start

Get up and running with KiviGo in just a few minutes. This guide shows you the basics of storing and retrieving data.

## Basic Example

Here's a simple example using the BadgerDB backend with the default JSON encoder:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/azrod/kivigo"
    "github.com/azrod/kivigo/backend/badger"
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

## Working with Different Data Types

KiviGo automatically handles different Go data types:

### Strings and Numbers

```go
// Strings
err := client.Set(ctx, "name", "John Doe")

// Numbers
err = client.Set(ctx, "age", 30)
err = client.Set(ctx, "score", 95.5)

// Booleans
err = client.Set(ctx, "active", true)
```

### Slices and Maps

```go
// Slices
items := []string{"apple", "banana", "cherry"}
err := client.Set(ctx, "fruits", items)

// Maps
config := map[string]interface{}{
    "timeout": 30,
    "debug":   true,
    "host":    "localhost",
}
err = client.Set(ctx, "config", config)
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

## Using Different Backends

KiviGo's unified interface means you can easily switch between backends:

### Redis Backend

```go
import "github.com/azrod/kivigo/backend/redis"

opt := redis.DefaultOptions()
opt.Addr = "localhost:6379"
kvStore, err := redis.New(opt)
```

### Local/BoltDB Backend

```go
import "github.com/azrod/kivigo/backend/local"

kvStore, err := local.New(local.Option{Path: "./data.db"})
```

### Consul Backend

```go
import "github.com/azrod/kivigo/backend/consul"

opt := consul.DefaultOptions()
opt.Address = "localhost:8500"
kvStore, err := consul.New(opt)
```

The client usage remains exactly the same regardless of which backend you choose!

## Custom Encoders

By default, KiviGo uses JSON encoding. You can specify different encoders:

### YAML Encoder

```go
import "github.com/azrod/kivigo/pkg/encoder"

client, err := kivigo.New(kvStore, kivigo.Option{
    Encoder: encoder.YAML,
})
```

### JSON Encoder (Default)

```go
import "github.com/azrod/kivigo/pkg/encoder"

client, err := kivigo.New(kvStore, kivigo.Option{
    Encoder: encoder.JSON, // This is the default
})
```

## Error Handling

KiviGo provides specific error types for common scenarios:

```go
import "github.com/azrod/kivigo/pkg/errs"

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

Now that you understand the basics:

1. **Learn about [Operations](./operations)** - Discover all available operations like List, Delete, and batch operations
2. **See more [Examples](./examples)** - Explore practical patterns and advanced usage
3. **Choose your [Backend](../backends/overview)** - Find the perfect storage backend for your use case
4. **Explore [Advanced Features](../advanced/health-checks)** - Dive into custom backends, batch operations, and more
