![KiviGo Logo](/docs/kivigo-white.png)

KiviGo is a lightweight key-value store library for Go. It provides a simple interface for storing and retrieving key-value pairs, supporting multiple backends (such as Redis and BoltDB) and encoders (JSON, YAML, etc.). KiviGo is designed to be easy to use, performant, and flexible.

> **Why "KiviGo"?**  
> The name is a play on words: "Kivi" sounds like "key-value" (the core concept of the library) and "Go" refers to the Go programming language. It also playfully evokes the fruit "kiwi"!

[![Go Reference](https://pkg.go.dev/badge/github.com/azrod/kivigo.svg)](https://pkg.go.dev/github.com/azrod/kivigo)

## ✨ Features

- Unified interface for different backends ([Redis](pkg/backend/redis/redis.go), [local/BoltDB](pkg/backend/local/local.go), etc.)
- Pluggable encoding/decoding ([JSON](pkg/encoder/json/json.go), [YAML](pkg/encoder/yaml/yaml.go), etc.)
- Health check support (with custom checks)
- List, add, and delete keys
- Easily extensible for new backends or encoders

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

## 📚 Documentation

See [pkg.go.dev/github.com/azrod/kivigo](https://pkg.go.dev/github.com/azrod/kivigo) for full documentation, and explore the [`pkg/backend`](pkg/backend/) and [`pkg/encoder`](pkg/encoder/) folders for more details on implementations.

---

**License**: MIT

---

## 🤝 Open Source & Contributions

KiviGo is an open source project. **Contributions, issues, and feature requests are welcome!**  
Feel free to open a pull request or an issue to help improve the project.
