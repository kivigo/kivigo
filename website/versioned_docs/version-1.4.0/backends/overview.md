---
sidebar_position: 1
---

# Backends Overview

KiviGo supports a wide variety of backends, each designed for different use cases and environments. This page provides an overview of all available backends and helps you choose the right one for your project.

## Backend Categories

### 🗄️ Embedded Stores

These backends store data locally in files, perfect for single-node applications or development environments.

- **[BadgerDB](./badger)** - Fast, embedded key-value store with built-in caching
- **[BoltDB (Local)](./local)** - Simple, reliable embedded database

### 🌐 Distributed Stores

These backends are designed for distributed systems and service discovery.

- **[Consul](./consul)** - Service discovery and configuration management
- **[etcd](./etcd)** - Distributed key-value store for critical data

### ⚡ Distributed Caches

High-performance caching solutions for fast data access.

- **[Redis](./redis)** - In-memory data structure store
- **[Memcached](./memcached)** - Distributed memory caching system

### ☁️ Cloud Services

Managed cloud backends that scale automatically.

- **[Azure Cosmos DB](./azurecosmos)** - Globally distributed, multi-model database
- **[DynamoDB](./dynamodb)** - Amazon's managed NoSQL database

### 🗃️ SQL Databases

Traditional relational databases with strong consistency guarantees.

- **[MySQL](./mysql)** - Popular open-source relational database
- **[PostgreSQL](./postgresql)** - Advanced open-source relational database

### 🍃 NoSQL Databases

Document and collection-based databases for flexible schemas.

- **[MongoDB](./mongodb)** - Document-oriented NoSQL database

## Feature Comparison

| Backend | Default Ops | Batch Ops | Health Check | Local/Remote | Persistence |
|---------|:-----------:|:---------:|:------------:|:------------:|:-----------:|
| BadgerDB | ✅ | ✅ | ✅ | Local | Disk |
| BoltDB | ✅ | ✅ | ✅ | Local | Disk |
| Redis | ✅ | ✅ | ✅ | Remote | Memory/Disk |
| Consul | ✅ | ✅ | ✅ | Remote | Disk |
| etcd | ✅ | ✅ | ✅ | Remote | Disk |
| Memcached | ✅ | ✅ | ✅ | Remote | Memory |
| Azure Cosmos | ✅ | ✅ | ✅ | Cloud | Disk |
| DynamoDB | ✅ | ✅ | ✅ | Cloud | Disk |
| MySQL | ✅ | ✅ | ✅ | Remote | Disk |
| PostgreSQL | ✅ | ✅ | ✅ | Remote | Disk |
| MongoDB | ✅ | ✅ | ✅ | Remote | Disk |

## Choosing the Right Backend

### 🏠 For Local Development

- **BadgerDB**: Fast development with persistence
- **BoltDB**: Simple, reliable file-based storage

### 📊 For Caching

- **Redis**: Feature-rich caching with data structures
- **Memcached**: Simple, high-performance memory caching

### 🔧 For Configuration Management

- **Consul**: Service discovery + configuration
- **etcd**: Kubernetes-style distributed configuration

### 🌍 For Production Applications

- **Redis**: High-performance applications
- **PostgreSQL**: ACID compliance requirements
- **MongoDB**: Flexible schema applications

### ☁️ For Cloud-Native Applications

- **Azure Cosmos DB**: Multi-region Azure applications
- **DynamoDB**: AWS-native applications

## Installation Patterns

All backends follow a consistent installation pattern:

```bash
# Main library
go get github.com/kivigo/kivigo

# Specific backend
go get github.com/kivigo/kivigo/backend/{backend-name}
```

## Usage Patterns

All backends implement the same interface, so switching is easy:

```go
// BadgerDB
import "github.com/kivigo/kivigo/backend/badger"
kvStore, err := badger.New(badger.DefaultOptions("./data"))

// Redis  
import "github.com/kivigo/kivigo/backend/redis"
kvStore, err := redis.New(redis.DefaultOptions())

// The client usage is identical regardless of backend
client, err := kivigo.New(kvStore)
```

## Performance Characteristics

### Read Performance (approximate)

1. **Memcached** - Fastest (in-memory)
2. **Redis** - Very fast (in-memory with persistence)
3. **BadgerDB** - Fast (local disk with caching)
4. **BoltDB** - Good (local disk)
5. **Consul/etcd** - Good (distributed)
6. **Cloud services** - Variable (network dependent)

### Write Performance (approximate)

1. **Memcached** - Fastest (memory only)
2. **Redis** - Very fast (async persistence)
3. **BadgerDB** - Fast (optimized writes)
4. **BoltDB** - Good (single writer)
5. **Consul/etcd** - Good (distributed consensus)
6. **Cloud services** - Variable (network dependent)

## Deployment Considerations

### Local/Embedded Backends

- **Pros**: No external dependencies, fast, simple deployment
- **Cons**: Single node only, backup complexity
- **Best for**: Development, single-node applications

### Remote/Distributed Backends

- **Pros**: Shared across services, scalable, managed backups
- **Cons**: Network dependency, operational complexity
- **Best for**: Production applications, microservices

### Cloud Backends

- **Pros**: Fully managed, auto-scaling, global distribution
- **Cons**: Vendor lock-in, cost, network latency
- **Best for**: Cloud-native applications, global scale

## Migration Between Backends

Thanks to KiviGo's unified interface, migrating between backends requires minimal code changes:

```go
// Before: Local development with BadgerDB
kvStore, err := badger.New(badger.DefaultOptions("./data"))

// After: Production with Redis
kvStore, err := redis.New(redis.DefaultOptions())

// Client code remains the same!
client, err := kivigo.New(kvStore)
```

For data migration, you can use KiviGo's List functionality to transfer data:

```go
// List all keys from old backend
keys, err := oldClient.List(ctx, "")

// Transfer each key-value pair
for _, key := range keys {
    var value interface{}
    err := oldClient.Get(ctx, key, &value)
    if err != nil {
        continue
    }
    
    err = newClient.Set(ctx, key, value)
    if err != nil {
        log.Printf("Failed to migrate key %s: %v", key, err)
    }
}
```

## Getting Help

- Check the individual backend documentation pages for detailed setup instructions
- Review the [Advanced Features](../advanced/health-checks) for optimization tips
- Use the [Mock Backend](../advanced/mock-testing) for testing your integration
