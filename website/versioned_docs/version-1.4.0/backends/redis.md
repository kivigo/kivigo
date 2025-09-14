---
sidebar_position: 3
---

import BackendTemplate from '@site/src/components/BackendTemplate';

<BackendTemplate
  name="Redis"
  description="Redis is an in-memory data structure store, used as a database, cache, and message broker. It supports data structures such as strings, hashes, lists, sets, and more."
  category="Distributed Cache"
  packageName="backend/redis"
  importPath="github.com/kivigo/kivigo/backend/redis"
  features={[
    { name: "Basic Operations", supported: true, description: "Set, Get, Delete, List operations" },
    { name: "Batch Operations", supported: true, description: "Pipeline-based batch operations" },
    { name: "Health Checks", supported: true, description: "PING-based health monitoring" },
    { name: "Persistence", supported: true, description: "RDB and AOF persistence options" },
    { name: "TTL/Expiration", supported: true, description: "Automatic key expiration" },
    { name: "Clustering", supported: true, description: "Redis Cluster support" },
    { name: "Pub/Sub", supported: false, description: "Not exposed through KiviGo interface" },
    { name: "Transactions", supported: true, description: "MULTI/EXEC transactions" }
  ]}
  dependencies={[
    "Redis server (version 6.0+)",
    "Network connectivity to Redis instance",
    "Optional: Redis Cluster for high availability"
  ]}
  installationNotes="Requires a running Redis server. You can run Redis locally, use a managed service, or deploy it in containers."
  configurationExample={`package main

import (
    "github.com/kivigo/kivigo"
    "github.com/kivigo/kivigo/backend/redis"
)

func main() {
    // Basic configuration
    opt := redis.DefaultOptions()
    opt.Addr = "localhost:6379"

    // Custom configuration
    customOpt := &redis.Options{
        Addr:         "localhost:6379",
        Password:     "mypassword",
        DB:           0,
        PoolSize:     20,
        MinIdleConns: 5,
        MaxRetries:   3,
    }
    
    // TLS configuration
    tlsOpt := redis.DefaultOptions()
    tlsOpt.Addr = "redis.example.com:6380"
    tlsOpt.TLSConfig = &tls.Config{
        ServerName: "redis.example.com",
    }
    
    // Redis Cluster configuration
    clusterOpt := &redis.ClusterOptions{
        Addrs: []string{
            "redis-node-1:6379",
            "redis-node-2:6379", 
            "redis-node-3:6379",
        },
        Password: "cluster-password",
    }
    
    // Create backend
    kvStore, err := redis.New(customOpt)
    if err != nil {
        panic(err)
    }
    defer kvStore.Close()
    
    // Create client
    client, err := kivigo.New(kvStore)
    if err != nil {
        panic(err)
    }
}`}
  usageExample={`package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/kivigo/backend/redis"
)

type Session struct {
    UserID    string    \`json:"user_id"\`
    LoginTime time.Time \`json:"login_time"\`
    IPAddress string    \`json:"ip_address"\`
}

func main() {
    // Setup Redis connection
    opt := redis.DefaultOptions()
    opt.Addr = "localhost:6379"

    kvStore, err := redis.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Store session data
    session := Session{
        UserID:    "user123",
        LoginTime: time.Now(),
        IPAddress: "192.168.1.100",
    }
    
    err = client.Set(ctx, "session:abc123", session)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve session
    var retrievedSession Session
    err = client.Get(ctx, "session:abc123", &retrievedSession)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Session: %+v\\n", retrievedSession)
    
    // List all sessions
    sessionKeys, err := client.List(ctx, "session:")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Active sessions: %v\\n", sessionKeys)
    
    // Delete session
    err = client.Delete(ctx, "session:abc123")
    if err != nil {
        log.Fatal(err)
    }
}`}
  healthCheckExample={`package main

import (
    "context"
    "log"
    "time"

    "github.com/kivigo/kivigo/backend/redis"
)

func main() {
    opt := redis.DefaultOptions()
    opt.Addr = "localhost:6379"

    kvStore, err := redis.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    ctx := context.Background()
    
    // Simple health check
    err = kvStore.Health(ctx)
    if err != nil {
        log.Printf("Redis unhealthy: %v", err)
        return
    }
    
    log.Println("Redis is healthy")
    
    // Advanced health monitoring with timeout
    healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    err = kvStore.Health(healthCtx)
    if err != nil {
        log.Printf("Health check timed out or failed: %v", err)
    } else {
        log.Println("Redis responded within timeout")
    }
    
    // Periodic health monitoring
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := kvStore.Health(ctx); err != nil {
                log.Printf("Redis health check failed: %v", err)
                // Could implement reconnection logic here
            } else {
                log.Println("Redis health check passed")
            }
        case <-ctx.Done():
            return
        }
    }
}`}
  batchExample={`package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/kivigo/backend/redis"
)

func main() {
    opt := redis.DefaultOptions()
    opt.Addr = "localhost:6379"

    kvStore, err := redis.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Batch set multiple cache entries
    cacheData := map[string]interface{}{
        "cache:user:1":    map[string]string{"name": "Alice", "role": "admin"},
        "cache:user:2":    map[string]string{"name": "Bob", "role": "user"},
        "cache:user:3":    map[string]string{"name": "Charlie", "role": "user"},
        "cache:config:db": map[string]string{"host": "localhost", "port": "5432"},
        "cache:config:api": map[string]string{"timeout": "30s", "retries": "3"},
    }
    
    // Batch operations use Redis pipelines for efficiency
    err = client.BatchSet(ctx, cacheData)
    if err != nil {
        log.Fatal(err)
    }
    
    // Batch get multiple values
    keys := []string{"cache:user:1", "cache:user:2", "cache:config:db"}
    results, err := client.BatchGet(ctx, keys)
    if err != nil {
        log.Fatal(err)
    }
    
    for key, value := range results {
        fmt.Printf("%s: %v\\n", key, value)
    }
    
    // Batch delete cache entries
    deleteKeys := []string{"cache:user:1", "cache:user:2"}
    err = client.BatchDelete(ctx, deleteKeys)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Cache batch operations completed")
    
    // Verify deletion
    remainingKeys, err := client.List(ctx, "cache:")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Remaining cache keys: %v\\n", remainingKeys)
}`}
  notes={[
    "Redis is primarily an in-memory store - ensure adequate RAM for your dataset",
    "Configure persistence (RDB snapshots or AOF) based on your durability requirements",
    "Use connection pooling for high-throughput applications",
    "Redis Cluster provides automatic sharding and high availability",
    "Consider using Redis Sentinel for high availability in non-cluster setups",
    "Monitor memory usage and configure appropriate eviction policies",
    "Network latency affects performance - deploy Redis close to your application",
    "Use Redis AUTH and TLS for security in production environments"
  ]}
  links={[
    { text: "Redis Official Documentation", url: "https://redis.io/documentation" },
    { text: "Redis Commands Reference", url: "https://redis.io/commands/" },
    { text: "Redis Persistence Guide", url: "https://redis.io/topics/persistence" },
    { text: "Redis Cluster Tutorial", url: "https://redis.io/topics/cluster-tutorial" },
    { text: "go-redis Client Documentation", url: "https://redis.uptrace.dev/" }
  ]}
/>
