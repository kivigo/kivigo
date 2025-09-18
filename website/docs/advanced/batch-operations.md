---
sidebar_position: 5
---

# Batch Operations

Batch operations allow you to perform multiple KiviGo operations efficiently in a single call. This is particularly useful for improving performance when dealing with large datasets or when you need to maintain consistency across multiple operations.

## Overview

KiviGo provides three types of batch operations:

- **BatchSet**: Set multiple key-value pairs at once
- **BatchGet**: Retrieve multiple values by their keys
- **BatchDelete**: Delete multiple keys at once

## Backend Support

Not all backends support batch operations. Check your backend's documentation:

| Backend | Batch Support | Implementation |
|---------|:-------------:|----------------|
| BadgerDB | ✅ | Native transactions |
| BoltDB | ✅ | Native transactions |
| Redis | ✅ | Pipeline operations |
| Consul | ✅ | Transaction API |
| etcd | ✅ | Transaction API |
| MongoDB | ✅ | Bulk operations |
| MySQL | ✅ | Prepared statements |
| PostgreSQL | ✅ | Prepared statements |
| Memcached | ✅ | Multi-operations |

## Basic Batch Operations

### Batch Set

Store multiple key-value pairs at once:

```go
package main

import (
    "context"
    "log"
    
    "github.com/kivigo/kivigo"
    "github.com/kivigo/backends/redis"
)

func main() {
    // Setup client
    kvStore, err := redis.New(redis.DefaultOptions())
    if err != nil {
        log.Fatal(err)
    }
    defer kvStore.Close()
    
    client, err := kivigo.New(kvStore)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Prepare batch data
    batchData := map[string]interface{}{
        "user:1": map[string]string{
            "name":  "Alice",
            "email": "alice@example.com",
            "role":  "admin",
        },
        "user:2": map[string]string{
            "name":  "Bob", 
            "email": "bob@example.com",
            "role":  "user",
        },
        "user:3": map[string]string{
            "name":  "Charlie",
            "email": "charlie@example.com", 
            "role":  "user",
        },
        "config:database": map[string]interface{}{
            "host":    "localhost",
            "port":    5432,
            "ssl":     true,
            "timeout": 30,
        },
    }
    
    // Perform batch set
    err = client.BatchSet(ctx, batchData)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Batch set completed successfully")
}
```

### Batch Get

Retrieve multiple values by their keys:

```go
func batchGetExample() {
    // ... setup client ...
    
    ctx := context.Background()
    
    // Keys to retrieve
    keys := []string{
        "user:1",
        "user:2", 
        "user:3",
        "config:database",
    }
    
    // Perform batch get
    results, err := client.BatchGet(ctx, keys)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process results
    for key, value := range results {
        fmt.Printf("Key: %s, Value: %v\n", key, value)
    }
    
    // Note: Missing keys are not included in results
    // Check if a specific key was found
    if userData, found := results["user:1"]; found {
        var user map[string]string
        err := client.Get(ctx, "user:1", &user) // Individual get for type safety
        if err != nil {
            log.Printf("Error decoding user:1: %v", err)
        } else {
            fmt.Printf("User 1: %+v\n", user)
        }
    } else {
        fmt.Println("user:1 not found")
    }
}
```

### Batch Delete

Delete multiple keys at once:

```go
func batchDeleteExample() {
    // ... setup client ...
    
    ctx := context.Background()
    
    // Keys to delete
    keysToDelete := []string{
        "user:1",
        "user:2",
        "user:3",
        "temp:session:abc123",
        "temp:session:def456",
    }
    
    // Perform batch delete
    err := client.BatchDelete(ctx, keysToDelete)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Deleted %d keys", len(keysToDelete))
}
```

## Advanced Batch Patterns

### Batch Update Pattern

Update multiple records efficiently:

```go
type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    LastSeen string `json:"last_seen"`
}

func updateUserLastSeen(client client.Client, userIDs []int) error {
    ctx := context.Background()
    currentTime := time.Now().Format(time.RFC3339)
    
    // Step 1: Batch get existing users
    keys := make([]string, len(userIDs))
    for i, id := range userIDs {
        keys[i] = fmt.Sprintf("user:%d", id)
    }
    
    results, err := client.BatchGet(ctx, keys)
    if err != nil {
        return fmt.Errorf("failed to batch get users: %w", err)
    }
    
    // Step 2: Update last seen time
    updates := make(map[string]interface{})
    for key, _ := range results {
        // Get the user individually for proper type handling
        var user User
        err := client.Get(ctx, key, &user)
        if err != nil {
            log.Printf("Failed to decode user %s: %v", key, err)
            continue
        }
        
        // Update last seen
        user.LastSeen = currentTime
        updates[key] = user
    }
    
    // Step 3: Batch set updated users
    if len(updates) > 0 {
        err = client.BatchSet(ctx, updates)
        if err != nil {
            return fmt.Errorf("failed to batch update users: %w", err)
        }
    }
    
    return nil
}
```

### Batch Migration Pattern

Migrate data between different key formats:

```go
func migrateUserKeys(client client.Client) error {
    ctx := context.Background()
    
    // Step 1: List old format keys
    oldKeys, err := client.List(ctx, "user_")
    if err != nil {
        return err
    }
    
    if len(oldKeys) == 0 {
        return nil // No migration needed
    }
    
    // Step 2: Batch get old data
    oldData, err := client.BatchGet(ctx, oldKeys)
    if err != nil {
        return fmt.Errorf("failed to get old data: %w", err)
    }
    
    // Step 3: Prepare new format data
    newData := make(map[string]interface{})
    for oldKey, _ := range oldData {
        // Extract ID from old key format (user_123 -> 123)
        idStr := strings.TrimPrefix(oldKey, "user_")
        
        // Create new key format (user:123)
        newKey := fmt.Sprintf("user:%s", idStr)
        
        // Get the actual data for type-safe handling
        var userData interface{}
        err := client.Get(ctx, oldKey, &userData)
        if err != nil {
            log.Printf("Failed to decode %s: %v", oldKey, err)
            continue
        }
        
        newData[newKey] = userData
    }
    
    // Step 4: Batch set new format data
    if len(newData) > 0 {
        err = client.BatchSet(ctx, newData)
        if err != nil {
            return fmt.Errorf("failed to set new data: %w", err)
        }
    }
    
    // Step 5: Batch delete old format data
    err = client.BatchDelete(ctx, oldKeys)
    if err != nil {
        return fmt.Errorf("failed to delete old data: %w", err)
    }
    
    log.Printf("Migrated %d users from old to new key format", len(oldData))
    return nil
}
```

### Batch Validation Pattern

Validate and process multiple items:

```go
func processUserBatch(client client.Client, users []User) error {
    ctx := context.Background()
    
    // Step 1: Validate all users
    validUsers := make(map[string]interface{})
    var errors []error
    
    for _, user := range users {
        if err := validateUser(user); err != nil {
            errors = append(errors, fmt.Errorf("user %d: %w", user.ID, err))
            continue
        }
        
        key := fmt.Sprintf("user:%d", user.ID)
        validUsers[key] = user
    }
    
    // Return early if any validation failed
    if len(errors) > 0 {
        return fmt.Errorf("validation errors: %v", errors)
    }
    
    // Step 2: Check for existing users (prevent overwrites)
    keys := make([]string, 0, len(validUsers))
    for key := range validUsers {
        keys = append(keys, key)
    }
    
    existing, err := client.BatchGet(ctx, keys)
    if err != nil {
        return fmt.Errorf("failed to check existing users: %w", err)
    }
    
    if len(existing) > 0 {
        existingKeys := make([]string, 0, len(existing))
        for key := range existing {
            existingKeys = append(existingKeys, key)
        }
        return fmt.Errorf("users already exist: %v", existingKeys)
    }
    
    // Step 3: Batch create users
    err = client.BatchSet(ctx, validUsers)
    if err != nil {
        return fmt.Errorf("failed to create users: %w", err)
    }
    
    log.Printf("Successfully created %d users", len(validUsers))
    return nil
}

func validateUser(user User) error {
    if user.ID <= 0 {
        return errors.New("invalid ID")
    }
    if user.Name == "" {
        return errors.New("name is required")
    }
    if user.Email == "" {
        return errors.New("email is required")
    }
    return nil
}
```

## Performance Considerations

### Batch Size Optimization

Different backends have different optimal batch sizes:

```go
const (
    RedisOptimalBatchSize      = 1000
    BadgerDBOptimalBatchSize   = 500
    ConsulOptimalBatchSize     = 100
    DatabaseOptimalBatchSize   = 100
)

func processBatchesOptimally(client client.Client, data map[string]interface{}, batchSize int) error {
    ctx := context.Background()
    
    // Split data into optimal batches
    keys := make([]string, 0, len(data))
    for key := range data {
        keys = append(keys, key)
    }
    
    for i := 0; i < len(keys); i += batchSize {
        end := i + batchSize
        if end > len(keys) {
            end = len(keys)
        }
        
        batch := make(map[string]interface{})
        for _, key := range keys[i:end] {
            batch[key] = data[key]
        }
        
        err := client.BatchSet(ctx, batch)
        if err != nil {
            return fmt.Errorf("batch %d-%d failed: %w", i, end-1, err)
        }
        
        log.Printf("Processed batch %d-%d", i, end-1)
    }
    
    return nil
}
```

### Monitoring Batch Performance

Track batch operation performance:

```go
func monitoredBatchSet(client client.Client, data map[string]interface{}) error {
    ctx := context.Background()
    
    start := time.Now()
    size := len(data)
    
    err := client.BatchSet(ctx, data)
    
    duration := time.Since(start)
    rate := float64(size) / duration.Seconds()
    
    log.Printf("Batch set: %d items in %v (%.2f items/sec)", 
               size, duration, rate)
    
    if err != nil {
        log.Printf("Batch set failed: %v", err)
        return err
    }
    
    // Alert on slow batches
    if duration > 5*time.Second {
        log.Printf("WARNING: Slow batch operation detected: %v", duration)
    }
    
    return nil
}
```

## Error Handling in Batch Operations

### Partial Failure Handling

Handle partial failures gracefully:

```go
func resilientBatchSet(client client.Client, data map[string]interface{}) error {
    ctx := context.Background()
    
    // Try batch operation first
    err := client.BatchSet(ctx, data)
    if err == nil {
        return nil // Success
    }
    
    log.Printf("Batch operation failed, falling back to individual operations: %v", err)
    
    // Fall back to individual operations
    var individualErrors []error
    successCount := 0
    
    for key, value := range data {
        err := client.Set(ctx, key, value)
        if err != nil {
            individualErrors = append(individualErrors, 
                fmt.Errorf("key %s: %w", key, err))
        } else {
            successCount++
        }
    }
    
    log.Printf("Individual operations: %d success, %d failed", 
               successCount, len(individualErrors))
    
    if len(individualErrors) > 0 {
        return fmt.Errorf("partial failure: %v", individualErrors)
    }
    
    return nil
}
```

### Retry Logic for Batches

Implement retry logic for failed batches:

```go
func retryableBatchSet(client client.Client, data map[string]interface{}, maxRetries int) error {
    ctx := context.Background()
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        err := client.BatchSet(ctx, data)
        if err == nil {
            return nil
        }
        
        if attempt < maxRetries {
            backoff := time.Duration(attempt+1) * time.Second
            log.Printf("Batch attempt %d failed, retrying in %v: %v", 
                       attempt+1, backoff, err)
            time.Sleep(backoff)
        } else {
            return fmt.Errorf("batch failed after %d attempts: %w", maxRetries+1, err)
        }
    }
    
    return nil
}
```

## Best Practices

1. **Use Appropriate Batch Sizes**: Test different batch sizes to find the optimal performance
2. **Handle Partial Failures**: Implement fallback strategies for partial failures
3. **Monitor Performance**: Track batch operation metrics
4. **Validate Before Batch**: Validate all data before performing batch operations
5. **Use Transactions When Available**: Some backends provide transaction guarantees for batches
6. **Consider Memory Usage**: Large batches consume more memory
7. **Implement Timeouts**: Use context timeouts for batch operations

Batch operations are powerful tools for improving performance and maintaining consistency when working with multiple key-value pairs in KiviGo.
