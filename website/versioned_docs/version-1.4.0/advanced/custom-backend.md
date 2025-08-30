---
sidebar_position: 3
---

# Custom Backend Development

KiviGo's architecture makes it easy to create custom backends for any storage system. This guide shows you how to implement your own backend that integrates seamlessly with the KiviGo ecosystem.

## Understanding the Interface

To create a custom backend, you need to implement the `KV` interface and optionally other interfaces for additional features:

```go
// Required interface - all backends must implement this
type KV interface {
    SetRaw(ctx context.Context, key string, value []byte) error
    GetRaw(ctx context.Context, key string) ([]byte, error)
    Delete(ctx context.Context, key string) error
    List(ctx context.Context, prefix string) ([]string, error)
    Close() error
}

// Optional interfaces for additional functionality
type KVWithHealth interface {
    KV
    Health(ctx context.Context) error
}

type KVWithBatch interface {
    KV
    BatchSetRaw(ctx context.Context, data map[string][]byte) error
    BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error)
    BatchDelete(ctx context.Context, keys []string) error
}
```

## Creating a Simple In-Memory Backend

Let's create a simple in-memory backend as an example:

```go
package memory

import (
    "context"
    "fmt"
    "strings"
    "sync"
    
    "github.com/azrod/kivigo/pkg/errs"
    "github.com/azrod/kivigo/pkg/models"
)

// Ensure our backend implements the required interfaces
var (
    _ models.KV           = (*Backend)(nil)
    _ models.KVWithHealth = (*Backend)(nil)
    _ models.KVWithBatch  = (*Backend)(nil)
)

type Backend struct {
    data  map[string][]byte
    mutex sync.RWMutex
}

type Options struct {
    InitialCapacity int
}

func DefaultOptions() Options {
    return Options{
        InitialCapacity: 100,
    }
}

func New(opts Options) *Backend {
    return &Backend{
        data: make(map[string][]byte, opts.InitialCapacity),
    }
}

// Implement KV interface
func (b *Backend) SetRaw(ctx context.Context, key string, value []byte) error {
    if key == "" {
        return errs.ErrEmptyKey
    }
    
    // Check context cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    b.mutex.Lock()
    defer b.mutex.Unlock()
    
    // Make a copy of the value to avoid external modifications
    valueCopy := make([]byte, len(value))
    copy(valueCopy, value)
    
    b.data[key] = valueCopy
    return nil
}

func (b *Backend) GetRaw(ctx context.Context, key string) ([]byte, error) {
    if key == "" {
        return nil, errs.ErrEmptyKey
    }
    
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    b.mutex.RLock()
    defer b.mutex.RUnlock()
    
    value, exists := b.data[key]
    if !exists {
        return nil, errs.ErrNotFound
    }
    
    // Return a copy to prevent external modifications
    result := make([]byte, len(value))
    copy(result, value)
    
    return result, nil
}

func (b *Backend) Delete(ctx context.Context, key string) error {
    if key == "" {
        return errs.ErrEmptyKey
    }
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    b.mutex.Lock()
    defer b.mutex.Unlock()
    
    delete(b.data, key)
    return nil
}

func (b *Backend) List(ctx context.Context, prefix string) ([]string, error) {
    if prefix == "" {
        return nil, errs.ErrEmptyPrefix
    }
    
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    b.mutex.RLock()
    defer b.mutex.RUnlock()
    
    var keys []string
    for key := range b.data {
        if strings.HasPrefix(key, prefix) {
            keys = append(keys, key)
        }
    }
    
    return keys, nil
}

func (b *Backend) Close() error {
    b.mutex.Lock()
    defer b.mutex.Unlock()
    
    // Clear the data map
    b.data = make(map[string][]byte)
    return nil
}

// Implement KVWithHealth interface
func (b *Backend) Health(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // For in-memory backend, we're always healthy if we can acquire the lock
    b.mutex.RLock()
    defer b.mutex.RUnlock()
    
    return nil
}

// Implement KVWithBatch interface
func (b *Backend) BatchSetRaw(ctx context.Context, data map[string][]byte) error {
    if len(data) == 0 {
        return nil // No-op for empty batch
    }
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    b.mutex.Lock()
    defer b.mutex.Unlock()
    
    // Set all key-value pairs
    for key, value := range data {
        if key == "" {
            return errs.ErrEmptyKey
        }
        
        // Make a copy of the value
        valueCopy := make([]byte, len(value))
        copy(valueCopy, value)
        
        b.data[key] = valueCopy
    }
    
    return nil
}

func (b *Backend) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
    if len(keys) == 0 {
        return make(map[string][]byte), nil
    }
    
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    b.mutex.RLock()
    defer b.mutex.RUnlock()
    
    result := make(map[string][]byte)
    
    for _, key := range keys {
        if key == "" {
            return nil, errs.ErrEmptyKey
        }
        
        if value, exists := b.data[key]; exists {
            // Return a copy
            valueCopy := make([]byte, len(value))
            copy(valueCopy, value)
            result[key] = valueCopy
        }
        // Note: We don't return an error for missing keys in batch operations
        // Individual missing keys are simply omitted from the result
    }
    
    return result, nil
}

func (b *Backend) BatchDelete(ctx context.Context, keys []string) error {
    if len(keys) == 0 {
        return nil
    }
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    b.mutex.Lock()
    defer b.mutex.Unlock()
    
    for _, key := range keys {
        if key == "" {
            return errs.ErrEmptyKey
        }
        delete(b.data, key)
    }
    
    return nil
}
```

## Creating a File-Based Backend

Here's a more complex example that persists data to files:

```go
package filestore

import (
    "context"
    "encoding/json"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "sync"
    
    "github.com/azrod/kivigo/pkg/errs"
    "github.com/azrod/kivigo/pkg/models"
)

var (
    _ models.KV           = (*Backend)(nil)
    _ models.KVWithHealth = (*Backend)(nil)
)

type Backend struct {
    baseDir string
    mutex   sync.RWMutex
}

type Options struct {
    BaseDirectory string
    CreateDir     bool
}

func DefaultOptions() Options {
    return Options{
        BaseDirectory: "./filestore",
        CreateDir:     true,
    }
}

func New(opts Options) (*Backend, error) {
    if opts.CreateDir {
        err := os.MkdirAll(opts.BaseDirectory, 0755)
        if err != nil {
            return nil, fmt.Errorf("failed to create directory: %w", err)
        }
    }
    
    return &Backend{
        baseDir: opts.BaseDirectory,
    }, nil
}

func (b *Backend) keyToPath(key string) string {
    // Convert key to safe filename
    safeKey := strings.ReplaceAll(key, "/", "_")
    safeKey = strings.ReplaceAll(safeKey, ":", "_")
    return filepath.Join(b.baseDir, safeKey+".json")
}

func (b *Backend) SetRaw(ctx context.Context, key string, value []byte) error {
    if key == "" {
        return errs.ErrEmptyKey
    }
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    b.mutex.Lock()
    defer b.mutex.Unlock()
    
    path := b.keyToPath(key)
    
    // Ensure directory exists
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }
    
    // Write file atomically
    tempPath := path + ".tmp"
    err := os.WriteFile(tempPath, value, 0644)
    if err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }
    
    err = os.Rename(tempPath, path)
    if err != nil {
        os.Remove(tempPath) // Clean up temp file
        return fmt.Errorf("failed to rename temp file: %w", err)
    }
    
    return nil
}

func (b *Backend) GetRaw(ctx context.Context, key string) ([]byte, error) {
    if key == "" {
        return nil, errs.ErrEmptyKey
    }
    
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    b.mutex.RLock()
    defer b.mutex.RUnlock()
    
    path := b.keyToPath(key)
    
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, errs.ErrNotFound
        }
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    return data, nil
}

func (b *Backend) Delete(ctx context.Context, key string) error {
    if key == "" {
        return errs.ErrEmptyKey
    }
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    b.mutex.Lock()
    defer b.mutex.Unlock()
    
    path := b.keyToPath(key)
    
    err := os.Remove(path)
    if err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to delete file: %w", err)
    }
    
    return nil
}

func (b *Backend) List(ctx context.Context, prefix string) ([]string, error) {
    if prefix == "" {
        return nil, errs.ErrEmptyPrefix
    }
    
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    b.mutex.RLock()
    defer b.mutex.RUnlock()
    
    var keys []string
    
    err := filepath.WalkDir(b.baseDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        
        if d.IsDir() {
            return nil
        }
        
        if !strings.HasSuffix(path, ".json") {
            return nil
        }
        
        // Extract key from filename
        rel, err := filepath.Rel(b.baseDir, path)
        if err != nil {
            return err
        }
        
        key := strings.TrimSuffix(rel, ".json")
        key = strings.ReplaceAll(key, "_", ":")
        
        if strings.HasPrefix(key, prefix) {
            keys = append(keys, key)
        }
        
        return nil
    })
    
    return keys, err
}

func (b *Backend) Close() error {
    // No cleanup needed for file-based backend
    return nil
}

func (b *Backend) Health(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // Check if base directory is accessible
    _, err := os.Stat(b.baseDir)
    if err != nil {
        return fmt.Errorf("base directory not accessible: %w", err)
    }
    
    // Try to create a temp file to test write permissions
    tempFile := filepath.Join(b.baseDir, ".health_check")
    err = os.WriteFile(tempFile, []byte("health"), 0644)
    if err != nil {
        return fmt.Errorf("cannot write to directory: %w", err)
    }
    
    // Clean up temp file
    os.Remove(tempFile)
    
    return nil
}
```

## Backend Testing

Create comprehensive tests for your custom backend:

```go
package memory

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "github.com/azrod/kivigo/pkg/errs"
)

func TestMemoryBackend(t *testing.T) {
    backend := New(DefaultOptions())
    defer backend.Close()
    
    ctx := context.Background()
    
    t.Run("SetAndGet", func(t *testing.T) {
        key := "test:key"
        value := []byte("test value")
        
        err := backend.SetRaw(ctx, key, value)
        require.NoError(t, err)
        
        retrieved, err := backend.GetRaw(ctx, key)
        require.NoError(t, err)
        assert.Equal(t, value, retrieved)
    })
    
    t.Run("GetNonExistent", func(t *testing.T) {
        _, err := backend.GetRaw(ctx, "nonexistent")
        assert.ErrorIs(t, err, errs.ErrNotFound)
    })
    
    t.Run("Delete", func(t *testing.T) {
        key := "delete:me"
        value := []byte("data")
        
        err := backend.SetRaw(ctx, key, value)
        require.NoError(t, err)
        
        err = backend.Delete(ctx, key)
        require.NoError(t, err)
        
        _, err = backend.GetRaw(ctx, key)
        assert.ErrorIs(t, err, errs.ErrNotFound)
    })
    
    t.Run("List", func(t *testing.T) {
        // Set up test data
        testData := map[string][]byte{
            "prefix:one":   []byte("1"),
            "prefix:two":   []byte("2"),
            "prefix:three": []byte("3"),
            "other:key":    []byte("4"),
        }
        
        for key, value := range testData {
            err := backend.SetRaw(ctx, key, value)
            require.NoError(t, err)
        }
        
        keys, err := backend.List(ctx, "prefix:")
        require.NoError(t, err)
        
        expected := []string{"prefix:one", "prefix:two", "prefix:three"}
        assert.ElementsMatch(t, expected, keys)
    })
    
    t.Run("Health", func(t *testing.T) {
        err := backend.Health(ctx)
        assert.NoError(t, err)
    })
    
    t.Run("BatchOperations", func(t *testing.T) {
        batchData := map[string][]byte{
            "batch:1": []byte("one"),
            "batch:2": []byte("two"),
            "batch:3": []byte("three"),
        }
        
        // Batch set
        err := backend.BatchSetRaw(ctx, batchData)
        require.NoError(t, err)
        
        // Batch get
        keys := []string{"batch:1", "batch:2", "batch:3"}
        results, err := backend.BatchGetRaw(ctx, keys)
        require.NoError(t, err)
        
        for key, expectedValue := range batchData {
            actualValue, exists := results[key]
            assert.True(t, exists)
            assert.Equal(t, expectedValue, actualValue)
        }
        
        // Batch delete
        err = backend.BatchDelete(ctx, keys)
        require.NoError(t, err)
        
        // Verify deletion
        for _, key := range keys {
            _, err := backend.GetRaw(ctx, key)
            assert.ErrorIs(t, err, errs.ErrNotFound)
        }
    })
}
```

## Integration with KiviGo Client

Once your backend is implemented, you can use it with the KiviGo client:

```go
package main

import (
    "context"
    "log"
    
    "github.com/azrod/kivigo"
    "your-module/memory" // Your custom backend
)

func main() {
    // Create your custom backend
    backend := memory.New(memory.DefaultOptions())
    defer backend.Close()
    
    // Create KiviGo client with your backend
    kvClient, err := client.New(backend, client.Option{})
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Use the standard KiviGo API
    err = kvClient.Set(ctx, "user:123", map[string]string{
        "name":  "John Doe",
        "email": "john@example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    var user map[string]string
    err = kvClient.Get(ctx, "user:123", &user)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("User: %+v", user)
}
```

## Best Practices for Custom Backends

1. **Thread Safety**: Always implement proper locking for concurrent access
2. **Context Handling**: Respect context cancellation and timeouts
3. **Error Handling**: Use appropriate KiviGo error types
4. **Data Copying**: Make copies of data to prevent external modifications
5. **Resource Cleanup**: Implement proper cleanup in the `Close()` method
6. **Comprehensive Testing**: Test all interface methods thoroughly
7. **Documentation**: Document your backend's behavior and limitations

Creating custom backends allows you to integrate KiviGo with any storage system while maintaining the same familiar API for your applications.
