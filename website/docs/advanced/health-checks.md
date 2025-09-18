---
sidebar_position: 1
---

# Health Checks

KiviGo provides built-in health check capabilities for monitoring the status of your storage backends. This is essential for building resilient applications that can detect and respond to backend failures.

## Basic Health Checks

All KiviGo backends that implement the `KVWithHealth` interface support health checking:

```go
package main

import (
    "context"
    "log"
    
    "github.com/kivigo/backends/redis"
)

func main() {
    opt := redis.DefaultOptions()
    kvStore, err := redis.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()

    [...]
    
    ctx := context.Background()
    
    // Perform health check
    err = kvStore.Health(ctx)
    if err != nil {
        log.Printf("Backend is unhealthy: %v", err)
    } else {
        log.Println("Backend is healthy")
    }
}
```

## Supported Backends

The following backends support health checks:

| Backend | Health Check Method | Description |
|---------|-------------------|-------------|
| BadgerDB | Database transaction test | Attempts a read-only transaction |
| BoltDB | Database access test | Checks database file accessibility |
| Redis | PING command | Sends Redis PING command |
| Consul | Status API | Checks Consul cluster status |
| etcd | Cluster health | Verifies etcd cluster health |
| MongoDB | Ping command | MongoDB ping operation |
| MySQL | Connection ping | Database connection ping |
| PostgreSQL | Connection ping | Database connection ping |

## Health Check with Timeout

Use context timeouts to prevent health checks from hanging:

```go
import (
    "context"
    "time"
    "log"
)

func healthCheckWithTimeout(kvStore models.KVWithHealth) error {
    // Create context with 5-second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err := kvStore.Health(ctx)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("health check timed out after 5 seconds")
        }
        return fmt.Errorf("health check failed: %w", err)
    }
    
    return nil
}
```

## Periodic Health Monitoring

Implement continuous health monitoring with a ticker:

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/kivigo/backends/redis"
    "github.com/kivigo/kivigo/pkg/models"
)

func startHealthMonitoring(kvStore models.KVWithHealth, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    ctx := context.Background()
    
    for {
        select {
        case <-ticker.C:
            err := kvStore.Health(ctx)
            if err != nil {
                log.Printf("❌ Health check failed: %v", err)
                // Implement alerting or recovery logic here
            } else {
                log.Println("✅ Health check passed")
            }
        case <-ctx.Done():
            log.Println("Health monitoring stopped")
            return
        }
    }
}

func main() {
    opt := redis.DefaultOptions()
    kvStore, err := redis.New(opt)
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()

    [...]
    
    // Start health monitoring every 30 seconds
    go startHealthMonitoring(kvStore, 30*time.Second)
    
    // Your application logic here
    select {} // Block forever
}
```

## Health Check in HTTP Handlers

Integrate health checks into HTTP health endpoints:

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/kivigo/kivigo/pkg/models"
)

type HealthResponse struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
    Backend   string    `json:"backend"`
    Error     string    `json:"error,omitempty"`
}

func healthHandler(kvStore models.KVWithHealth, backendName string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
        defer cancel()
        
        response := HealthResponse{
            Timestamp: time.Now(),
            Backend:   backendName,
        }
        
        err := kvStore.Health(ctx)
        if err != nil {
            response.Status = "unhealthy"
            response.Error = err.Error()
            w.WriteHeader(http.StatusServiceUnavailable)
        } else {
            response.Status = "healthy"
            w.WriteHeader(http.StatusOK)
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }
}

func main() {
    // Setup your backend
    kvStore, err := redis.New(redis.DefaultOptions())
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()

    [...]
    
    // Register health endpoint
    http.HandleFunc("/health", healthHandler(kvStore, "redis"))
    
    log.Println("Health endpoint available at http://localhost:8080/health")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Advanced Health Monitoring

### Circuit Breaker Pattern

Implement a circuit breaker to prevent cascading failures:

```go
type CircuitBreaker struct {
    kvStore     models.KVWithHealth
    failures    int
    lastFailure time.Time
    maxFailures int
    timeout     time.Duration
    state       string // "closed", "open", "half-open"
}

func (cb *CircuitBreaker) IsHealthy(ctx context.Context) bool {
    now := time.Now()
    
    switch cb.state {
    case "open":
        if now.Sub(cb.lastFailure) > cb.timeout {
            cb.state = "half-open"
            return cb.tryHealth(ctx)
        }
        return false
        
    case "half-open":
        return cb.tryHealth(ctx)
        
    default: // closed
        return cb.tryHealth(ctx)
    }
}

func (cb *CircuitBreaker) tryHealth(ctx context.Context) bool {
    err := cb.kvStore.Health(ctx)
    
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }
        return false
    }
    
    // Success - reset circuit breaker
    cb.failures = 0
    cb.state = "closed"
    return true
}
```

### Health Check Aggregation

When using multiple backends, aggregate their health status:

```go
type MultiBackendHealth struct {
    backends map[string]models.KVWithHealth
}

func (m *MultiBackendHealth) CheckAll(ctx context.Context) map[string]error {
    results := make(map[string]error)
    
    for name, backend := range m.backends {
        results[name] = backend.Health(ctx)
    }
    
    return results
}

func (m *MultiBackendHealth) IsAllHealthy(ctx context.Context) bool {
    for _, backend := range m.backends {
        if err := backend.Health(ctx); err != nil {
            return false
        }
    }
    return true
}
```

## Best Practices

### 1. Set Appropriate Timeouts

Always use context timeouts to prevent health checks from hanging:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### 2. Implement Graceful Degradation

When backends are unhealthy, implement fallback mechanisms:

```go
func getData(primary, fallback models.KV, key string) (interface{}, error) {
    // Try primary backend
    if primary.Health(ctx) == nil {
        return primary.Get(ctx, key, &result)
    }
    
    // Fallback to secondary backend
    log.Println("Using fallback backend")
    return fallback.Get(ctx, key, &result)
}
```

### 3. Monitor Health Check Performance

Track health check latency to detect performance degradation:

```go
func timedHealthCheck(kvStore models.KVWithHealth) (time.Duration, error) {
    start := time.Now()
    err := kvStore.Health(context.Background())
    duration := time.Since(start)
    
    if duration > 1*time.Second {
        log.Printf("Slow health check: %v", duration)
    }
    
    return duration, err
}
```

### 4. Avoid Health Check Storms

Coordinate health checks across multiple instances to avoid overwhelming backends:

```go
import "math/rand"

func jitteredHealthCheck(interval time.Duration) time.Duration {
    // Add ±25% jitter to prevent synchronized health checks
    jitter := time.Duration(rand.Float64() * float64(interval) * 0.5)
    return interval + jitter - time.Duration(float64(interval)*0.25)
}
```

## Troubleshooting

### Common Health Check Failures

1. **Network Timeouts**: Check network connectivity and firewall rules
2. **Authentication Errors**: Verify credentials and permissions
3. **Resource Exhaustion**: Monitor CPU, memory, and disk usage
4. **Service Overload**: Implement rate limiting and connection pooling

### Debugging Health Checks

Enable debug logging to troubleshoot health check issues:

```go
import "log"

func debugHealthCheck(kvStore models.KVWithHealth) {
    ctx := context.Background()
    
    log.Println("Starting health check...")
    start := time.Now()
    
    err := kvStore.Health(ctx)
    duration := time.Since(start)
    
    if err != nil {
        log.Printf("Health check failed after %v: %v", duration, err)
    } else {
        log.Printf("Health check passed in %v", duration)
    }
}
```

Health checks are essential for building robust applications with KiviGo. Use them to implement monitoring, alerting, and automatic recovery mechanisms in your applications.
