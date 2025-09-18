---
sidebar_position: 8
title: Memcached
---

import BackendTemplate from '@site/src/components/BackendTemplate';

<BackendTemplate
  name="Memcached"
  description="Memcached is a high-performance, distributed memory object caching system. It's designed to speed up dynamic web applications by alleviating database load through caching data and objects in RAM."
  category="In-Memory Cache"
  packageName="backend/memcached"
  importPath="github.com/kivigo/backends/memcached"
  features={[
    { name: "Basic Operations", supported: true, description: "Set, Get, Delete, List operations" },
    { name: "Batch Operations", supported: true, description: "BatchSet, BatchGet, BatchDelete for bulk operations" },
    { name: "Health Checks", supported: true, description: "Built-in health monitoring" },
    { name: "TTL/Expiration", supported: true, description: "Automatic key expiration" },
    { name: "Distributed Caching", supported: true, description: "Multiple server support with consistent hashing" },
    { name: "High Performance", supported: true, description: "Sub-millisecond response times" },
    { name: "Memory Efficiency", supported: true, description: "Optimized memory usage" },
    { name: "Simple Protocol", supported: true, description: "Text-based protocol for easy debugging" }
  ]}
  dependencies={[
    "Memcached server(s) running",
    "Network connectivity to Memcached servers",
    "Sufficient memory allocation on servers"
  ]}
  installationNotes="Requires one or more Memcached servers. You can run Memcached locally with Docker: `docker run -d -p 11211:11211 memcached`"
  configurationExample={`package main

import (
    "time"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/backends/memcached"
)

func main() {
    // Single server configuration
    opt := memcached.NewOptions()
    opt.Servers = []string{"localhost:11211"}
    opt.Timeout = 1 * time.Second
    opt.MaxIdleConns = 2

    // Multiple servers with consistent hashing
    multiOpt := memcached.NewOptions()
    multiOpt.Servers = []string{
        "memcached1:11211",
        "memcached2:11211", 
        "memcached3:11211",
    }
    multiOpt.Timeout = 500 * time.Millisecond
    multiOpt.MaxIdleConns = 10
    multiOpt.KeepAlive = 30 * time.Second
    
    // Optional: Configure default TTL
    multiOpt.DefaultTTL = 3600 // 1 hour
    
    // Create backend
    kvStore, err := memcached.New(multiOpt)
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
    "github.com/kivigo/backends/memcached"
)

type CacheItem struct {
    Data      interface{} \`json:"data"\`
    CachedAt  time.Time   \`json:"cached_at"\`
    ExpiresAt time.Time   \`json:"expires_at"\`
}

func main() {
    // Setup
    opt := memcached.NewOptions()
    opt.Servers = []string{"localhost:11211"}
    opt.Timeout = 1 * time.Second
    opt.DefaultTTL = 300 // 5 minutes

    kvStore, err := memcached.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Cache expensive computation result
    cacheKey := "computation:result:user123"
    cachedItem := CacheItem{
        Data: map[string]interface{}{
            "result": 42,
            "computation_time": "2.5s",
            "parameters": []string{"param1", "param2"},
        },
        CachedAt:  time.Now(),
        ExpiresAt: time.Now().Add(5 * time.Minute),
    }
    
    err = client.Set(ctx, cacheKey, cachedItem)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve cached result
    var retrievedItem CacheItem
    err = client.Get(ctx, cacheKey, &retrievedItem)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Cached Item: %+v\\n", retrievedItem)
    
    // Cache multiple session data with TTL
    sessionKeys := []string{
        "session:abc123",
        "session:def456", 
        "session:ghi789",
    }
    
    for _, key := range sessionKeys {
        sessionData := map[string]interface{}{
            "user_id": key[8:], // Extract from key
            "active":  true,
            "last_activity": time.Now().Format(time.RFC3339),
        }
        
        err = client.Set(ctx, key, sessionData)
        if err != nil {
            log.Printf("Failed to cache session %s: %v", key, err)
        }
    }
    
    // List keys with prefix (note: Memcached doesn't natively support this)
    keys, err := client.List(ctx, "session:")
    if err != nil {
        log.Printf("List operation may not be fully supported: %v", err)
    } else {
        fmt.Printf("Session Keys: %v\\n", keys)
    }
}`}
  healthCheckExample={`package main

import (
    "context"
    "log"
    "time"

    "github.com/kivigo/backends/memcached"
)

func main() {
    opt := memcached.NewOptions()
    opt.Servers = []string{"localhost:11211"}
    opt.Timeout = 1 * time.Second

    kvStore, err := memcached.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    ctx := context.Background()
    
    // Simple health check
    err = kvStore.Health(ctx)
    if err != nil {
        log.Printf("Memcached unhealthy: %v", err)
        return
    }
    
    log.Println("Memcached is healthy")
    
    // Monitor cache hit rate and performance
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    testKey := "health:check"
    testValue := map[string]interface{}{
        "timestamp": time.Now().Unix(),
        "check_id":  "health-monitor",
    }
    
    for {
        select {
        case <-ticker.C:
            start := time.Now()
            
            // Perform a write/read test
            if err := kvStore.SetRaw(ctx, testKey, []byte("health-check")); err != nil {
                log.Printf("Health check write failed: %v", err)
                continue
            }
            
            _, err := kvStore.GetRaw(ctx, testKey)
            if err != nil {
                log.Printf("Health check read failed: %v", err)
                continue
            }
            
            latency := time.Since(start)
            log.Printf("Health check passed (latency: %v)", latency)
            
            // Alert on high latency
            if latency > 10*time.Millisecond {
                log.Printf("WARNING: High cache latency: %v", latency)
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
    "github.com/kivigo/backends/memcached"
)

func main() {
    opt := memcached.NewOptions()
    opt.Servers = []string{"localhost:11211"}
    opt.Timeout = 1 * time.Second
    opt.DefaultTTL = 1800 // 30 minutes

    kvStore, err := memcached.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Batch cache frequently accessed user data
    userCache := map[string]interface{}{
        "user:profile:1": map[string]interface{}{
            "name": "Alice", "email": "alice@example.com", "role": "admin",
        },
        "user:profile:2": map[string]interface{}{
            "name": "Bob", "email": "bob@example.com", "role": "user",
        },
        "user:profile:3": map[string]interface{}{
            "name": "Charlie", "email": "charlie@example.com", "role": "moderator",
        },
        "user:settings:1": map[string]interface{}{
            "theme": "dark", "notifications": true, "language": "en",
        },
        "user:settings:2": map[string]interface{}{
            "theme": "light", "notifications": false, "language": "es",
        },
    }
    
    err = client.BatchSet(ctx, userCache)
    if err != nil {
        log.Fatal(err)
    }
    
    // Batch retrieve user data
    keys := []string{
        "user:profile:1", "user:profile:2", "user:profile:3",
        "user:settings:1", "user:settings:2",
    }
    
    results, err := client.BatchGet(ctx, keys)
    if err != nil {
        log.Fatal(err)
    }
    
    for key, value := range results {
        fmt.Printf("%s: %v\\n", key, value)
    }
    
    // Batch invalidate cache on user logout
    logoutKeys := []string{"user:profile:2", "user:settings:2"}
    err = client.BatchDelete(ctx, logoutKeys)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Batch cache operations completed")
}`}
  notes={[
    "Memcached is purely in-memory - data is lost when the server restarts",
    "Maximum value size is 1MB by default",
    "Keys have a maximum length of 250 characters",
    "Use consistent hashing for multi-server setups to minimize rehashing",
    "Monitor memory usage to prevent evictions of important data",
    "Consider using connection pooling for high-traffic applications",
    "Memcached doesn't natively support key enumeration (List operation may be limited)",
    "For persistent caching, consider Redis as an alternative"
  ]}
  links={[
    { text: "Memcached Official Website", url: "https://memcached.org/" },
    { text: "Memcached GitHub Repository", url: "https://github.com/memcached/memcached" },
    { text: "Memcached Wiki", url: "https://github.com/memcached/memcached/wiki" },
    { text: "Go Memcache Client", url: "https://github.com/bradfitz/gomemcache" }
  ]}
/>
