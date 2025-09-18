---
sidebar_position: 6
---

# Client-Side Hooks

KiviGo provides a powerful client-side hooks system that allows you to register callback functions triggered after successful key-value operations. This enables event-driven architectures, auditing, cache invalidation, and custom business logic without modifying your core application flow.

## Overview

The hooks system supports the following features:

- **Event-driven callbacks**: Automatically triggered after successful operations
- **Flexible filtering**: Event type and key pattern filtering
- **Execution modes**: Synchronous and asynchronous execution
- **Error handling**: Best-effort error delivery without affecting operations
- **Thread safety**: Concurrent hook registration and execution

## Supported Events

KiviGo triggers hooks for these operation types:

| Event Type | Triggered After | Key Present | Value Present |
|------------|----------------|:-----------:|:-------------:|
| `EventSet` | `Set()` operation | ✅ | ✅ |
| `EventSetRaw` | `SetRaw()` operation | ✅ | ✅ |
| `EventDelete` | `Delete()` operation | ✅ | ❌ |
| `EventBatchSet` | `BatchSet()` operation (per key) | ✅ | ✅ |
| `EventBatchDel` | `BatchDelete()` operation (per key) | ✅ | ❌ |

## Basic Hook Registration

### Simple Hook

Register a hook that responds to all events:

```go
package main

import (
    "context"
    "log"
    
    "github.com/kivigo/kivigo"
    "github.com/kivigo/encoders/json"
    "github.com/kivigo/backends/redis"
)

func main() {
    // Create client
    opt := redis.DefaultOptions()
    backend, err := redis.New(opt)
    if err != nil {
        log.Fatal(err)
    }

    client, err := kivigo.New(backend, kivigo.Option{
        Encoder: json.New(),
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Register a simple logging hook
    id, errCh, unregister := client.RegisterHook(
        func(ctx context.Context, evt client.EventType, key string, value []byte) error {
            log.Printf("Operation: %s on key: %s", evt, key)
            if value != nil {
                log.Printf("Value: %s", string(value))
            }
            return nil
        },
        client.HookOptions{}, // No filtering - responds to all events
    )
    defer unregister()
    
    // Monitor hook errors
    go func() {
        for err := range errCh {
            log.Printf("Hook error: %v", err)
        }
    }()
    
    // Your application logic
    client.Set(context.Background(), "user:123", "John Doe")
    client.Delete(context.Background(), "user:123")
}
```

### Hook with Options

Register a hook with specific configuration:

```go
// Register hook for specific events with key filtering
id, errCh, unregister := client.RegisterHook(
    func(ctx context.Context, evt client.EventType, key string, value []byte) error {
        // Custom logic here
        return auditUserOperation(evt, key, value)
    },
    client.HookOptions{
        Events:  []client.EventType{client.EventSet, client.EventDelete}, // Only Set and Delete
        Filter:  client.PrefixFilter("user:"),                            // Only keys starting with "user:"
        Async:   true,                                                    // Execute asynchronously
        Timeout: 5 * time.Second,                                         // Timeout for sync hooks (ignored for async)
    },
)
defer unregister()
```

## Event Filtering

### Filter by Event Type

Register hooks for specific operation types:

```go
// Hook for Set operations only
client.RegisterHook(handleSetOperation, client.HookOptions{
    Events: []client.EventType{client.EventSet, client.EventSetRaw},
})

// Hook for Delete operations only
client.RegisterHook(handleDeleteOperation, client.HookOptions{
    Events: []client.EventType{client.EventDelete, client.EventBatchDel},
})

// Hook for all Batch operations
client.RegisterHook(handleBatchOperation, client.HookOptions{
    Events: []client.EventType{client.EventBatchSet, client.EventBatchDel},
})
```

## Key Filtering

### Built-in Filters

KiviGo provides several built-in filter functions:

```go
// Prefix filter - matches keys starting with prefix
client.RegisterHook(userHook, client.HookOptions{
    Filter: client.PrefixFilter("user:"),
})

// Suffix filter - matches keys ending with suffix
client.RegisterHook(configHook, client.HookOptions{
    Filter: client.SuffixFilter(":config"),
})

// List filter - matches keys in the provided list
client.RegisterHook(importantKeysHook, client.HookOptions{
    Filter: client.ListFilter([]string{"critical-key", "important-data"}),
})

// Regex filter - matches keys against regex pattern
client.RegisterHook(patternHook, client.HookOptions{
    Filter: client.RegexFilter(`^user:\d+$`), // Matches "user:123" format
})
```

### Custom Filters

Create custom filter functions:

```go
// Custom filter for session keys
func sessionFilter(key string) bool {
    return strings.HasPrefix(key, "session:") && len(key) > 8
}

// Custom filter for temporary keys
func temporaryFilter(key string) bool {
    return strings.Contains(key, ":temp:") || strings.HasSuffix(key, ":tmp")
}

client.RegisterHook(sessionHook, client.HookOptions{
    Filter: sessionFilter,
})
```

### Combined Filtering

Combine event and key filtering:

```go
// Monitor Set operations on user configuration keys
client.RegisterHook(userConfigAudit, client.HookOptions{
    Events: []client.EventType{client.EventSet},
    Filter: func(key string) bool {
        return strings.HasPrefix(key, "user:") && strings.HasSuffix(key, ":config")
    },
})
```

## Execution Modes

### Synchronous Execution

Hooks execute in the main goroutine and can block operations:

```go
client.RegisterHook(
    func(ctx context.Context, evt client.EventType, key string, value []byte) error {
        // This executes synchronously and can block the operation
        return validateOperation(evt, key, value)
    },
    client.HookOptions{
        Async:   false,                // Synchronous execution
        Timeout: 2 * time.Second,     // Timeout to prevent blocking
    },
)
```

### Asynchronous Execution

Hooks execute in separate goroutines without blocking:

```go
client.RegisterHook(
    func(ctx context.Context, evt client.EventType, key string, value []byte) error {
        // This executes asynchronously in a separate goroutine
        return sendNotification(evt, key, value)
    },
    client.HookOptions{
        Async: true, // Asynchronous execution
        // Timeout is ignored for async hooks
    },
)
```

## Error Handling

### Error Channels

Hook errors are delivered via buffered channels:

```go
id, errCh, unregister := client.RegisterHook(hookFunc, options)
defer unregister()

// Handle errors asynchronously
go func() {
    for err := range errCh {
        log.Printf("Hook %s error: %v", id, err)
        // Handle error (log, metrics, retry, etc.)
    }
}()
```

### Best-Effort Delivery

- Hook errors never fail the main operation
- Errors are delivered on a best-effort basis
- If the error channel is full, errors are dropped
- Operations complete successfully regardless of hook status

```go
func riskyHook(ctx context.Context, evt client.EventType, key string, value []byte) error {
    // Even if this fails, the main operation (Set/Delete) will succeed
    return someRiskyOperation()
}
```

## Advanced Examples

### Audit System

Implement comprehensive auditing with hooks:

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "time"
    
    "github.com/kivigo/kivigo/pkg/client"
)

type AuditEvent struct {
    Timestamp time.Time       `json:"timestamp"`
    Operation string          `json:"operation"`
    Key       string          `json:"key"`
    Value     string          `json:"value,omitempty"`
    UserID    string          `json:"user_id,omitempty"`
}

func setupAuditHooks(c client.Client) {
    // Audit all user operations
    c.RegisterHook(
        func(ctx context.Context, evt client.EventType, key string, value []byte) error {
            event := AuditEvent{
                Timestamp: time.Now(),
                Operation: string(evt),
                Key:       key,
            }
            
            if value != nil {
                event.Value = string(value)
            }
            
            // Extract user ID from context if available
            if userID, ok := ctx.Value("user_id").(string); ok {
                event.UserID = userID
            }
            
            return logAuditEvent(event)
        },
        client.HookOptions{
            Filter: client.PrefixFilter("user:"),
            Async:  true, // Don't block operations
        },
    )
}

func logAuditEvent(event AuditEvent) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    log.Printf("AUDIT: %s", string(data))
    // Send to audit service, write to file, etc.
    return nil
}
```

### Cache Invalidation

Implement cache invalidation with hooks:

```go
package main

import (
    "context"
    "strings"
    
    "github.com/kivigo/kivigo/pkg/client"
)

type CacheInvalidator struct {
    cacheClient *redis.Client // Your cache client
}

func (ci *CacheInvalidator) setupHooks(kvClient client.Client) {
    // Invalidate cache on data changes
    kvClient.RegisterHook(
        ci.invalidateCache,
        client.HookOptions{
            Events: []client.EventType{
                client.EventSet,
                client.EventDelete,
                client.EventBatchSet,
                client.EventBatchDel,
            },
            Async: true, // Don't block operations
        },
    )
}

func (ci *CacheInvalidator) invalidateCache(ctx context.Context, evt client.EventType, key string, value []byte) error {
    // Generate cache keys to invalidate
    cacheKeys := ci.generateCacheKeys(key)
    
    // Invalidate related cache entries
    for _, cacheKey := range cacheKeys {
        if err := ci.cacheClient.Del(ctx, cacheKey).Err(); err != nil {
            return err
        }
    }
    
    return nil
}

func (ci *CacheInvalidator) generateCacheKeys(key string) []string {
    var cacheKeys []string
    
    // Direct cache key
    cacheKeys = append(cacheKeys, "cache:"+key)
    
    // Related patterns
    if strings.HasPrefix(key, "user:") {
        cacheKeys = append(cacheKeys, "user_list:*")
    }
    
    return cacheKeys
}
```

### Metrics Collection

Collect metrics with hooks:

```go
package main

import (
    "context"
    "time"
    
    "github.com/kivigo/kivigo/pkg/client"
    "github.com/prometheus/client_golang/prometheus"
)

var (
    operationCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "kivigo_operations_total",
            Help: "Total number of KiviGo operations",
        },
        []string{"operation", "key_prefix"},
    )
)

func setupMetricsHooks(c client.Client) {
    c.RegisterHook(
        func(ctx context.Context, evt client.EventType, key string, value []byte) error {
            // Determine key prefix
            prefix := "other"
            if idx := strings.Index(key, ":"); idx > 0 {
                prefix = key[:idx]
            }
            
            // Increment counter
            operationCounter.WithLabelValues(string(evt), prefix).Inc()
            
            return nil
        },
        client.HookOptions{
            Async: true, // Don't block operations
        },
    )
}
```

### Real-time Notifications

Send real-time notifications with hooks:

```go
package main

import (
    "context"
    "encoding/json"
    
    "github.com/kivigo/kivigo/pkg/client"
    "github.com/gorilla/websocket"
)

type NotificationService struct {
    connections map[string]*websocket.Conn
}

func (ns *NotificationService) setupHooks(c client.Client) {
    // Notify on user data changes
    c.RegisterHook(
        ns.notifyUserChange,
        client.HookOptions{
            Events: []client.EventType{client.EventSet, client.EventDelete},
            Filter: client.PrefixFilter("user:"),
            Async:  true,
        },
    )
}

func (ns *NotificationService) notifyUserChange(ctx context.Context, evt client.EventType, key string, value []byte) error {
    notification := map[string]interface{}{
        "type":      "user_change",
        "operation": string(evt),
        "key":       key,
        "timestamp": time.Now().Unix(),
    }
    
    data, err := json.Marshal(notification)
    if err != nil {
        return err
    }
    
    // Send to all connected clients
    for _, conn := range ns.connections {
        if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
            // Handle connection error
            continue
        }
    }
    
    return nil
}
```

## Best Practices

### 1. Use Appropriate Execution Mode

- **Synchronous**: For validation, critical business logic
- **Asynchronous**: For logging, notifications, metrics, cache invalidation

```go
// Validation should be synchronous to potentially affect the operation
client.RegisterHook(validateData, client.HookOptions{
    Async:   false,
    Timeout: 1 * time.Second,
})

// Logging should be asynchronous to not slow down operations
client.RegisterHook(logOperation, client.HookOptions{
    Async: true,
})
```

### 2. Handle Errors Gracefully

Always monitor error channels and handle errors appropriately:

```go
id, errCh, unregister := client.RegisterHook(hookFunc, options)
defer unregister()

go func() {
    for err := range errCh {
        // Log, send to error tracking service, etc.
        log.Printf("Hook %s failed: %v", id, err)
    }
}()
```

### 3. Use Specific Filters

Avoid global hooks when possible. Use specific filters to reduce overhead:

```go
// Good - specific filter
client.RegisterHook(userHook, client.HookOptions{
    Filter: client.PrefixFilter("user:"),
    Events: []client.EventType{client.EventSet},
})

// Avoid - global hook (unless truly needed)
client.RegisterHook(globalHook, client.HookOptions{})
```

### 4. Implement Cleanup

Always unregister hooks when they're no longer needed:

```go
func setupTemporaryHook(client client.Client) func() {
    _, _, unregister := client.RegisterHook(hookFunc, options)
    return unregister // Return cleanup function
}

// Usage
cleanup := setupTemporaryHook(client)
defer cleanup() // Ensure cleanup
```

### 5. Consider Performance

- Use async hooks for non-critical operations
- Implement efficient filters
- Monitor error channel capacity

```go
// Efficient filter using map lookup
var criticalKeys = map[string]bool{
    "config:database": true,
    "config:api":      true,
}

func criticalKeyFilter(key string) bool {
    return criticalKeys[key]
}
```

## Troubleshooting

### Hook Not Triggered

Check these common issues:

1. **Event filtering**: Ensure the hook is registered for the correct event types
2. **Key filtering**: Verify the filter function matches your keys
3. **Hook registration**: Confirm the hook was registered before operations

```go
// Debug hook registration
id, errCh, unregister := client.RegisterHook(
    func(ctx context.Context, evt client.EventType, key string, value []byte) error {
        log.Printf("Hook triggered: evt=%s key=%s", evt, key)
        return nil
    },
    client.HookOptions{}, // No filters for debugging
)
```

### Performance Issues

If hooks are causing performance problems:

1. **Use async execution** for non-critical hooks
2. **Implement efficient filters** to reduce unnecessary executions
3. **Monitor error channels** to prevent channel blocking

```go
// Performance-optimized hook
client.RegisterHook(hookFunc, client.HookOptions{
    Async:  true,              // Don't block operations
    Filter: efficientFilter,   // Specific, fast filter
})
```

### Error Channel Overflow

If error channels are overflowing:

1. **Monitor error channels** actively
2. **Implement error handling** logic
3. **Consider hook complexity** - simpler hooks fail less

```go
// Proper error handling
go func() {
    for err := range errCh {
        // Process errors without blocking
        go handleHookError(err)
    }
}()
```

## Integration Examples

### With Different Backends

The hooks system works identically across all backends:

```go
// Works with any backend
func setupHooksForBackend(backend models.KV) {
    client, _ := client.New(backend, client.Option{Encoder: encoder.JSON})
    
    client.RegisterHook(universalHook, client.HookOptions{
        Async: true,
    })
}

// Usage with different backends
setupHooksForBackend(redisBackend)
setupHooksForBackend(badgerBackend)
setupHooksForBackend(consulBackend)
```

### With Middleware

Hooks can complement middleware patterns:

```go
type ClientWithHooks struct {
    client.Client
    logger *log.Logger
}

func NewClientWithHooks(backend models.KV, logger *log.Logger) *ClientWithHooks {
    c, _ := kivigo.New(backend, kivigo.Option{Encoder: encoder.JSON})
    
    cwh := &ClientWithHooks{
        Client: c,
        logger: logger,
    }
    
    // Setup automatic logging
    c.RegisterHook(cwh.logOperation, client.HookOptions{Async: true})
    
    return cwh
}

func (c *ClientWithHooks) logOperation(ctx context.Context, evt client.EventType, key string, value []byte) error {
    c.logger.Printf("Operation: %s on %s", evt, key)
    return nil
}
```

The hooks system provides a powerful foundation for building event-driven applications with KiviGo while maintaining clean separation of concerns and excellent performance characteristics.
