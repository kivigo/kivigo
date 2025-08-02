![KiviGo Logo](/docs/kivigo-readme.png)

[![Go Reference](https://pkg.go.dev/badge/github.com/azrod/kivigo.svg)](https://pkg.go.dev/github.com/azrod/kivigo)
[![Go Report Card](https://goreportcard.com/badge/github.com/azrod/kivigo)](https://goreportcard.com/report/github.com/azrod/kivigo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org/dl/)

KiviGo is a lightweight key-value store library for Go. It provides a simple interface for storing and retrieving key-value pairs, supporting multiple backends (such as Redis and BoltDB) and encoders (JSON, YAML, etc.). KiviGo is designed to be easy to use, performant, and flexible.

> **Why "KiviGo"?**  
> The name is a play on words: "Kivi" sounds like "key-value" (the core concept of the library) and "Go" refers to the Go programming language. It also playfully evokes the fruit "kiwi" 🥝 !

## ✨ Features

- Unified interface for different backends ([Redis](pkg/backend/redis/redis.go), [local/BoltDB](pkg/backend/local/local.go), etc.)
- Pluggable encoding/decoding ([JSON](pkg/encoder/json/json.go), [YAML](pkg/encoder/yaml/yaml.go), etc.)
- Health check support (with custom checks)
- List, add, and delete keys
- Easily extensible for new backends or encoders

## 🥝 Motivation

KiviGo was created to simplify and unify key-value storage in Go applications. In many projects, developers face the challenge of switching between different storage backends (like Redis, BoltDB, or in-memory stores) or need to support multiple serialization formats (such as JSON or YAML). Each backend often comes with its own API, error handling, and setup, making code harder to maintain and test.

**What problem does KiviGo solve?**

- Provides a single, consistent API for key-value operations, regardless of the underlying backend.
- Makes it easy to swap storage engines (e.g., from local development with BoltDB to production with Redis) without changing your application logic.
- Supports pluggable encoders, so you can choose or implement the serialization format that fits your needs.
- Simplifies testing by providing a mock backend.
- Enables health checks and batch operations in a backend-agnostic way.

**Use cases:**

- Building microservices that need simple, fast key-value storage with the flexibility to change backends.
- Prototyping applications locally with BoltDB, then moving to Redis for production.
- Writing unit tests for storage logic without requiring a real database.
- Implementing custom backends (e.g., in-memory, cloud-based) while keeping the same client code.
- Supporting advanced features like batch operations, health checks, and custom serialization with minimal effort.

KiviGo helps you focus on your application logic, not on backend-specific details.

## 📊 Comparison: KiviGo vs Other Go Key-Value Libraries

> ⚠️ **Note:** The following comparison is provided for convenience and is based on the state of these libraries at the time of writing. Features and APIs may evolve over time—please refer to each project's documentation for the most up-to-date information.

There are several Go libraries for key-value storage, each with different goals and trade-offs. Here’s how KiviGo compares to some popular alternatives:

| Library         | Unified API | Pluggable Backends | Pluggable Encoders | Health Checks | Batch Ops | Mock/Test Support | Extensible |
|-----------------|:----------:|:------------------:|:------------------:|:-------------:|:--------:|:-----------------:|:----------:|
| **KiviGo**      | ✅         | ✅                 | ✅                 | ✅            | ✅       | ✅                | ✅         |
| [go-redis](https://github.com/redis/go-redis) | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ |
| [bbolt](https://github.com/etcd-io/bbolt)     | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| [badger](https://github.com/dgraph-io/badger) | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| [gokv](https://github.com/philippgille/gokv)  | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ➖ |
| [cache2go](https://github.com/muesli/cache2go) | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

**Key differences:**

- **KiviGo** provides a unified API, supports multiple backends (local, Redis, custom), pluggable encoders (JSON, YAML, custom), health checks, batch operations, and a mock backend for testing.
- **go-redis**, **bbolt**, and **badger** are excellent for their specific storage engines but do not abstract over multiple backends or provide pluggable encoding.
- **gokv** offers a unified API for multiple backends but lacks pluggable encoders, health checks, and mock/test support.
- **cache2go** is focused on in-memory caching and does not provide backend abstraction or encoding options.

KiviGo is designed for projects that need flexibility, testability, and the ability to swap storage or serialization strategies with minimal code changes.

### Backend feature matrix

| Backend      | Default (List/Get/Set/Delete) | Batch (Get/Set/Delete) | Health |
|--------------|:----------------------------:|:----------------------:|:------:|
| Local (Bolt) | ✅                           | ✅                     | ✅     |
| Redis        | ✅                           | ✅                     | ✅     |

## 📦 Installation

```sh
go get github.com/azrod/kivigo
```

## 🚀 Quickstart Example

This example shows how to use KiviGo with the local backend (BoltDB) and the default JSON encoder:

```go
package main

import (
    "context"
    "fmt"

    "github.com/azrod/kivigo"
    "github.com/azrod/kivigo/pkg/backend"
    "github.com/azrod/kivigo/pkg/backend/local"
)

func main() {
    client, err := kivigo.New(
        backend.Local(local.Option{
            Path: "./",
        }),
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
```

## 🧑‍💻 Advanced Example

Using Redis as a backend, YAML as encoder, and periodic health checks:

```go
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
```

Full example is available in the [`examples/advanced/main.go`](examples/advanced/main.go) file.

## 🩺 Custom Health Check Example

You can provide your own custom health check logic by using the `AdditionalChecks` field in `client.HealthOptions` when calling `HealthCheck`.  
This field accepts a slice of `client.HealthFunc`, and each function will be called during the health check.

```go
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
    client, err := kivigo.New(
        backend.Local(local.Option{
            Path: "./",
        }),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // Use HealthCheck with your custom logic
    healthCh := client.HealthCheck(context.Background(), client.HealthOptions{
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
    if err := client.Set(context.Background(), "health:ping", "pong"); err != nil {
        panic(err)
    }

    time.Sleep(10 * time.Second)
}
```

**Output:**

```
$> 2025-06-13T17:04:56+02:00 Custom health issue: custom health check failed: key not found
$> 2025-06-13T17:05:01+02:00 Custom health OK
```

Full example is available in the [`examples/custom_health_check/main.go`](examples/custom_health_check/main.go) file.

## 🧩 Extending KiviGo: Custom In-Memory Backend Example

You can add your own backend by implementing the [`models.KV`](pkg/models/kv.go) interface.  
Here is a minimal example of an in-memory backend:

```go
package memory

import (
 "context"
 "errors"
 "strings"
 "sync"

 "github.com/azrod/kivigo/pkg/models"
)

var _ models.KV = (*BMemory)(nil)

type BMemory struct {
 mu    sync.RWMutex
 store map[string][]byte
}

func New() *BMemory {
 return &BMemory{
  store: make(map[string][]byte),
 }
}

func (b *BMemory) SetRaw(_ context.Context, key string, value []byte) error {
 b.mu.Lock()
 defer b.mu.Unlock()
 b.store[key] = value
 return nil
}

func (b *BMemory) GetRaw(_ context.Context, key string) ([]byte, error) {
 b.mu.RLock()
 defer b.mu.RUnlock()
 val, ok := b.store[key]
 if !ok {
  return nil, errors.New("key not found")
 }
 return val, nil
}

func (b *BMemory) Delete(_ context.Context, key string) error {
 b.mu.Lock()
 defer b.mu.Unlock()
 delete(b.store, key)
 return nil
}

func (b *BMemory) List(_ context.Context, prefix string) (keys []string, err error) {
 b.mu.RLock()
 defer b.mu.RUnlock()

 keys = make([]string, 0, len(b.store))
 for k := range b.store {
  if strings.HasPrefix(k, prefix) {
   keys = append(keys, k)
  }
 }
 return keys, nil
}

func (b *BMemory) Close() error {
 return nil
}

func (b *BMemory) Health(_ context.Context) error {
 // Memory backend is always healthy as it does not depend on external resources
 return nil
}

```

You can then use your backend with KiviGo:

```go
import (
    "github.com/azrod/kivigo"
    "github.com/azrod/kivigo/pkg/backend"
    "yourmodule/memory"
)

[...]

client, err := kivigo.New(
    backend.CustomBackend(memory.New()),
)

[...]
```

Full example is available in the [`examples/custom_backend/main.go`](examples/custom_backend/main.go) file.

## 🧪 Using the Mock Backend for Testing

KiviGo provides a mock backend (`pkg/mock.MockKV`) to make it easy to write unit tests without relying on a real database.

### Example Usage

```go
import (
    "context"
    "testing"

    "github.com/azrod/kivigo/pkg/client"
    "github.com/azrod/kivigo/pkg/encoder"
    "github.com/azrod/kivigo/pkg/mock"
)

func TestWithMockKV(t *testing.T) {
    mockKV := &mock.MockKV{Data: map[string][]byte{"foo": []byte("bar")}}
    c, err := client.New(mockKV, client.Option{Encoder: encoder.JSON})
    if err != nil {
        t.Fatal(err)
    }

    var val string
    if err := c.Get(context.Background(), "foo", &val); err != nil {
        t.Fatal(err)
    }
    if val != "bar" {
        t.Errorf("expected bar, got %s", val)
    }
}
```

You can inject `MockKV` into your tests to simulate all key-value backend behaviors, including batch operations and health checks.

## 📚 Documentation

See [pkg.go.dev/github.com/azrod/kivigo](https://pkg.go.dev/github.com/azrod/kivigo) for full documentation, and explore the [`pkg/backend`](pkg/backend/) and [`pkg/encoder`](pkg/encoder/) folders for more details on implementations.

---

**License**: MIT

---

## 🤝 Open Source & Contributions

KiviGo is an open source project. **Contributions, issues, and feature requests are welcome!**  
Feel free to open a pull request or an issue to help improve the project.
