---
sidebar_position: 2
---

# Installation

This guide covers how to install KiviGo and its backends for your Go projects.

## Core Library

First, install the main KiviGo library:

```bash
go get github.com/azrod/kivigo@v1.5.0
```

## Backend Installation

KiviGo uses a modular architecture where each backend is a separate Go module. This keeps your dependencies minimal - you only install what you need.

### Embedded Storage Backends

For local, embedded storage that doesn't require external services:

#### BadgerDB Backend
```bash
go get github.com/azrod/kivigo/backend/badger@v1.5.0
```

#### Local/BoltDB Backend
```bash
go get github.com/azrod/kivigo/backend/local@v1.5.0
```

### Distributed Storage Backends

For distributed storage that requires external services:

#### Redis Backend
```bash
go get github.com/azrod/kivigo/backend/redis@v1.5.0
```

#### Consul Backend
```bash
go get github.com/azrod/kivigo/backend/consul@v1.5.0
```

#### etcd Backend
```bash
go get github.com/azrod/kivigo/backend/etcd@v1.5.0
```

#### Memcached Backend
```bash
go get github.com/azrod/kivigo/backend/memcached@v1.5.0
```

### Database Backends

For traditional SQL and NoSQL databases:

#### MongoDB Backend
```bash
go get github.com/azrod/kivigo/backend/mongodb@v1.5.0
```

#### MySQL Backend
```bash
go get github.com/azrod/kivigo/backend/mysql@v1.5.0
```

#### PostgreSQL Backend
```bash
go get github.com/azrod/kivigo/backend/postgresql@v1.5.0
```

### Cloud Backends

For cloud-native storage services:

#### AWS DynamoDB Backend
```bash
go get github.com/azrod/kivigo/backend/dynamodb@v1.5.0
```

#### Azure Cosmos DB Backend
```bash
go get github.com/azrod/kivigo/backend/azurecosmos@v1.5.0
```

## Version Management

### Using Specific Versions

To ensure reproducible builds, always specify a version:

```bash
# Install a specific version
go get github.com/azrod/kivigo@v1.5.0
go get github.com/azrod/kivigo/backend/redis@v1.5.0
```

### Using Latest Version

To get the latest release:

```bash
# Get the latest release
go get github.com/azrod/kivigo@latest
go get github.com/azrod/kivigo/backend/redis@latest
```

### Updating Dependencies

To update to a newer version:

```bash
# Update to a specific version
go get -u github.com/azrod/kivigo@v1.5.0

# Update to latest
go get -u github.com/azrod/kivigo@latest
```

## Verification

After installation, verify that KiviGo is working correctly:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/azrod/kivigo"
    "github.com/azrod/kivigo/pkg/mock"
)

func main() {
    // Create a mock backend for testing
    mockKV := &mock.MockKV{Data: map[string][]byte{}}
    
    // Create KiviGo client
    client, err := kivigo.New(mockKV)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Test basic functionality
    err = client.Set(ctx, "test", "Hello, KiviGo!")
    if err != nil {
        log.Fatal(err)
    }

    var value string
    err = client.Get(ctx, "test", &value)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("✅ KiviGo is working! Retrieved: %s\n", value)
}
```

## Troubleshooting

### Common Issues

**Module not found errors:**
- Ensure you're using Go 1.21 or later
- Run `go mod tidy` after installation
- Check your `GOPROXY` settings if behind a corporate proxy

**Version conflicts:**
- Use `go mod graph` to identify conflicting dependencies
- Pin specific versions using `go get package@version`

**Build errors:**
- Some backends require CGO (like BadgerDB) - ensure you have a C compiler
- Check backend-specific documentation for additional requirements

## Next Steps

Now that KiviGo is installed:

1. **Start with [Quick Start](./quick-start)** - Learn the basics with simple examples
2. **Explore [Operations](./operations)** - Understand all available operations
3. **Try [Examples](./examples)** - See practical usage patterns
4. **Choose your [Backend](./backends/overview)** - Pick the right storage for your use case