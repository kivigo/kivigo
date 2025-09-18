---
sidebar_position: 5
---

# Examples

Practical examples showing how to use KiviGo in real-world scenarios.

## Configuration Management

Store and manage application configuration:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/backends/badger"
)

type AppConfig struct {
    Database struct {
        Host     string `json:"host"`
        Port     int    `json:"port"`
        Username string `json:"username"`
        Name     string `json:"name"`
    } `json:"database"`
    Server struct {
        Host string `json:"host"`
        Port int    `json:"port"`
    } `json:"server"`
    Features struct {
        EnableLogging bool `json:"enable_logging"`
        Debug         bool `json:"debug"`
    } `json:"features"`
}

func main() {
    // Setup
    opt := badger.DefaultOptions("./config_data")
    kvStore, err := badger.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()

    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Store configuration
    config := AppConfig{}
    config.Database.Host = "localhost"
    config.Database.Port = 5432
    config.Database.Username = "admin"
    config.Database.Name = "myapp"
    config.Server.Host = "0.0.0.0"
    config.Server.Port = 8080
    config.Features.EnableLogging = true
    config.Features.Debug = false

    err = client.Set(ctx, "config:app", config)
    if err != nil {
        log.Fatal(err)
    }

    // Retrieve and use configuration
    var retrievedConfig AppConfig
    err = client.Get(ctx, "config:app", &retrievedConfig)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Database URL: %s:%d/%s\n", 
        retrievedConfig.Database.Host,
        retrievedConfig.Database.Port,
        retrievedConfig.Database.Name)
    fmt.Printf("Server will run on: %s:%d\n",
        retrievedConfig.Server.Host,
        retrievedConfig.Server.Port)
}
```

## User Session Management

Manage user sessions with TTL and cleanup:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/backends/redis"
)

type UserSession struct {
    UserID    int       `json:"user_id"`
    Username  string    `json:"username"`
    CreatedAt time.Time `json:"created_at"`
    LastSeen  time.Time `json:"last_seen"`
    IPAddress string    `json:"ip_address"`
}

func main() {
    // Setup Redis backend (good for sessions with TTL)
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

    // Create a new session
    sessionID := "sess_abc123"
    session := UserSession{
        UserID:    12345,
        Username:  "john_doe",
        CreatedAt: time.Now(),
        LastSeen:  time.Now(),
        IPAddress: "192.168.1.100",
    }

    err = client.Set(ctx, fmt.Sprintf("session:%s", sessionID), session)
    if err != nil {
        log.Fatal(err)
    }

    // Retrieve session
    var retrievedSession UserSession
    err = client.Get(ctx, fmt.Sprintf("session:%s", sessionID), &retrievedSession)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Session for user %s (ID: %d)\n", 
        retrievedSession.Username, retrievedSession.UserID)

    // Update last seen
    retrievedSession.LastSeen = time.Now()
    err = client.Set(ctx, fmt.Sprintf("session:%s", sessionID), retrievedSession)
    if err != nil {
        log.Fatal(err)
    }

    // List all active sessions
    sessionKeys, err := client.List(ctx, "session:")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Active sessions: %d\n", len(sessionKeys))
    for _, key := range sessionKeys {
        var sess UserSession
        err := client.Get(ctx, key, &sess)
        if err != nil {
            continue
        }
        fmt.Printf("- %s: %s (last seen: %s)\n", 
            key, sess.Username, sess.LastSeen.Format(time.RFC3339))
    }

    // Cleanup old sessions (older than 1 hour)
    cutoff := time.Now().Add(-1 * time.Hour)
    for _, key := range sessionKeys {
        var sess UserSession
        err := client.Get(ctx, key, &sess)
        if err != nil {
            continue
        }
        
        if sess.LastSeen.Before(cutoff) {
            err = client.Delete(ctx, key)
            if err != nil {
                log.Printf("Failed to delete expired session %s: %v", key, err)
            } else {
                fmt.Printf("Deleted expired session: %s\n", key)
            }
        }
    }
}
```

## Caching Layer

Implement a caching layer for expensive operations:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/kivigo/backends/redis"
    "github.com/kivigo/kivigo"
    "github.com/kivigo/kivigo/pkg/errs"
)

type CachedData struct {
    Value     interface{} `json:"value"`
    Timestamp time.Time   `json:"timestamp"`
    TTL       int         `json:"ttl_seconds"`
}

type Cache struct {
    client kivigo.Client
}

func NewCache(client kivigo.Client) *Cache {
    return &Cache{client: client}
}

// Set stores data in cache with TTL
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttlSeconds int) error {
    cached := CachedData{
        Value:     value,
        Timestamp: time.Now(),
        TTL:       ttlSeconds,
    }
    
    return c.client.Set(ctx, fmt.Sprintf("cache:%s", key), cached)
}

// Get retrieves data from cache, checking expiration
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
    var cached CachedData
    err := c.client.Get(ctx, fmt.Sprintf("cache:%s", key), &cached)
    if err != nil {
        if errors.Is(err, errs.ErrNotFound) {
            return false, nil // Cache miss
        }
        return false, err
    }
    
    // Check if expired
    if time.Since(cached.Timestamp) > time.Duration(cached.TTL)*time.Second {
        // Delete expired data
        c.client.Delete(ctx, fmt.Sprintf("cache:%s", key))
        return false, nil // Cache miss (expired)
    }
    
    // Use type assertion or JSON marshaling to copy value
    switch v := dest.(type) {
    case *string:
        if str, ok := cached.Value.(string); ok {
            *v = str
        }
    case *int:
        if num, ok := cached.Value.(float64); ok { // JSON numbers are float64
            *v = int(num)
        }
    default:
        // For complex types, re-marshal through JSON
        // This is a simplified example - in production you might want
        // a more sophisticated approach
        return false, fmt.Errorf("unsupported type for cache get")
    }
    
    return true, nil // Cache hit
}

// Example expensive operation
func expensiveOperation(ctx context.Context, input string) (string, error) {
    // Simulate expensive work
    time.Sleep(2 * time.Second)
    return fmt.Sprintf("processed: %s", input), nil
}

func main() {
    // Setup
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

    cache := NewCache(client)
    ctx := context.Background()

    // Function that uses cache
    processWithCache := func(input string) (string, error) {
        cacheKey := fmt.Sprintf("operation:%s", input)
        
        // Try cache first
        var result string
        hit, err := cache.Get(ctx, cacheKey, &result)
        if err != nil {
            return "", err
        }
        
        if hit {
            fmt.Printf("Cache HIT for %s\n", input)
            return result, nil
        }
        
        fmt.Printf("Cache MISS for %s, computing...\n", input)
        
        // Compute expensive operation
        result, err = expensiveOperation(ctx, input)
        if err != nil {
            return "", err
        }
        
        // Store in cache for 5 minutes
        err = cache.Set(ctx, cacheKey, result, 300)
        if err != nil {
            log.Printf("Failed to cache result: %v", err)
        }
        
        return result, nil
    }

    // Test the caching
    inputs := []string{"hello", "world", "hello"} // "hello" will be cached
    
    for _, input := range inputs {
        start := time.Now()
        result, err := processWithCache(input)
        if err != nil {
            log.Printf("Error processing %s: %v", input, err)
            continue
        }
        
        duration := time.Since(start)
        fmt.Printf("Result: %s (took %v)\n", result, duration)
        fmt.Println()
    }
}
```

## Complete Application Example

A comprehensive example showing error handling, health checks, and multiple operations:

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/kivigo/pkg/errs"
    "github.com/kivigo/backends/badger"
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

    // Store multiple related items
    users := []struct {
        ID   int
        Name string
        Role string
    }{
        {1, "John Doe", "admin"},
        {2, "Jane Smith", "user"},
        {3, "Bob Johnson", "user"},
    }

    for _, user := range users {
        key := fmt.Sprintf("user:%d", user.ID)
        err := client.Set(ctx, key, user)
        if err != nil {
            log.Printf("Failed to store user %d: %v", user.ID, err)
        } else {
            fmt.Printf("✅ Stored user: %s\n", key)
        }
    }

    // List and display all users
    userKeys, err := client.List(ctx, "user:")
    if err != nil {
        log.Fatal("Failed to list users:", err)
    }

    fmt.Printf("✅ Found %d users:\n", len(userKeys))
    for _, key := range userKeys {
        var user struct {
            ID   int    `json:"ID"`
            Name string `json:"Name"`
            Role string `json:"Role"`
        }
        
        err := client.Get(ctx, key, &user)
        if err != nil {
            log.Printf("Failed to get user %s: %v", key, err)
            continue
        }
        
        fmt.Printf("  - %s: %s (%s)\n", key, user.Name, user.Role)
    }

    // Clean up users (but keep config)
    for _, key := range userKeys {
        err := client.Delete(ctx, key)
        if err != nil {
            log.Printf("Failed to delete %s: %v", key, err)
        }
    }
    fmt.Println("✅ Users cleaned up")

    // Verify cleanup
    remainingKeys, err := client.List(ctx, "")
    if err != nil {
        log.Fatal("Failed to list remaining keys:", err)
    }
    fmt.Printf("✅ Remaining keys: %v\n", remainingKeys)

    // Final cleanup
    err = client.Delete(ctx, "app:config")
    if err != nil {
        log.Fatal("Failed to delete config:", err)
    }
    fmt.Println("✅ Configuration deleted")
}
```

## Testing with Mock Backend

Use the mock backend for unit testing:

```go
package main

import (
    "context"
    "testing"

    "github.com/kivigo/kivigo"
    "github.com/kivigo/kivigo/pkg/mock"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserService(t *testing.T) {
    // Create mock backend
    mockKV := &mock.MockKV{Data: map[string][]byte{}}
    client, err := kivigo.New(mockKV)
    require.NoError(t, err)

    ctx := context.Background()

    // Test data
    type User struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }

    user := User{ID: 1, Name: "Test User"}

    // Test storing user
    err = client.Set(ctx, "user:1", user)
    assert.NoError(t, err)

    // Test retrieving user
    var retrieved User
    err = client.Get(ctx, "user:1", &retrieved)
    assert.NoError(t, err)
    assert.Equal(t, user, retrieved)

    // Test listing users
    keys, err := client.List(ctx, "user:")
    assert.NoError(t, err)
    assert.Contains(t, keys, "user:1")

    // Test deleting user
    err = client.Delete(ctx, "user:1")
    assert.NoError(t, err)

    // Verify deletion
    err = client.Get(ctx, "user:1", &retrieved)
    assert.Error(t, err)
}
```

## Next Steps

These examples demonstrate real-world usage patterns. To continue learning:

1. **Explore [Backends](../backends/overview)** - Choose the right backend for your specific use case
2. **Learn [Advanced Features](../advanced/health-checks)** - Implement robust error handling and monitoring
3. **Try [Batch Operations](../advanced/batch-operations)** - Optimize performance for bulk operations
4. **Build [Custom Backends](../advanced/custom-backend)** - Create your own storage implementations
