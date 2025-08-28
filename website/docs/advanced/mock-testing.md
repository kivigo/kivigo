---
sidebar_position: 4
---

# Mock Testing

KiviGo provides a built-in mock backend that makes it easy to write unit tests without requiring external dependencies like Redis or databases. This is essential for fast, reliable tests that can run in any environment.

## Using the Mock Backend

The mock backend (`pkg/mock.MockKV`) provides an in-memory implementation of the KiviGo interface:

```go
package main

import (
    "context"
    "testing"
    
    "github.com/azrod/kivigo"
    "github.com/azrod/kivigo/pkg/mock"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMyService(t *testing.T) {
    // Create mock backend
    mockKV := &mock.MockKV{
        Data: map[string][]byte{},
    }
    
    // Create KiviGo client with mock backend
    client, err := client.New(mockKV, client.Option{})
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Your test logic here
    err = client.Set(ctx, "test:key", "test value")
    require.NoError(t, err)
    
    var value string
    err = client.Get(ctx, "test:key", &value)
    require.NoError(t, err)
    assert.Equal(t, "test value", value)
}
```

## Pre-populating Test Data

You can pre-populate the mock with test data:

```go
func TestWithPreData(t *testing.T) {
    // Pre-populate mock with test data
    mockKV := &mock.MockKV{
        Data: map[string][]byte{
            "user:1": []byte(`{"id":1,"name":"Alice","email":"alice@example.com"}`),
            "user:2": []byte(`{"id":2,"name":"Bob","email":"bob@example.com"}`),
            "config:app": []byte(`{"debug":true,"port":8080}`),
        },
    }
    
    client, err := client.New(mockKV, client.Option{})
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Test that pre-populated data is accessible
    var user struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    
    err = client.Get(ctx, "user:1", &user)
    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
    
    // Test listing keys
    keys, err := client.List(ctx, "user:")
    require.NoError(t, err)
    assert.Contains(t, keys, "user:1")
    assert.Contains(t, keys, "user:2")
}
```

## Testing Service Layer

Here's how to test a service that depends on KiviGo:

```go
// Service under test
type UserService struct {
    kv models.KV
}

func NewUserService(kv models.KV) *UserService {
    return &UserService{kv: kv}
}

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (s *UserService) CreateUser(ctx context.Context, user User) error {
    client, err := client.New(s.kv, client.Option{})
    if err != nil {
        return err
    }
    
    key := fmt.Sprintf("user:%d", user.ID)
    return client.Set(ctx, key, user)
}

func (s *UserService) GetUser(ctx context.Context, id int) (User, error) {
    client, err := client.New(s.kv, client.Option{})
    if err != nil {
        return User{}, err
    }
    
    var user User
    key := fmt.Sprintf("user:%d", id)
    err = client.Get(ctx, key, &user)
    return user, err
}

func (s *UserService) ListUsers(ctx context.Context) ([]User, error) {
    client, err := client.New(s.kv, client.Option{})
    if err != nil {
        return nil, err
    }
    
    keys, err := client.List(ctx, "user:")
    if err != nil {
        return nil, err
    }
    
    var users []User
    for _, key := range keys {
        var user User
        if err := client.Get(ctx, key, &user); err == nil {
            users = append(users, user)
        }
    }
    
    return users, nil
}

// Test the service
func TestUserService(t *testing.T) {
    mockKV := &mock.MockKV{Data: map[string][]byte{}}
    service := NewUserService(mockKV)
    ctx := context.Background()
    
    t.Run("CreateAndGetUser", func(t *testing.T) {
        user := User{
            ID:    1,
            Name:  "John Doe",
            Email: "john@example.com",
        }
        
        // Test creation
        err := service.CreateUser(ctx, user)
        require.NoError(t, err)
        
        // Test retrieval
        retrieved, err := service.GetUser(ctx, 1)
        require.NoError(t, err)
        assert.Equal(t, user, retrieved)
    })
    
    t.Run("GetNonexistentUser", func(t *testing.T) {
        _, err := service.GetUser(ctx, 999)
        assert.Error(t, err)
    })
    
    t.Run("ListUsers", func(t *testing.T) {
        // Create multiple users
        users := []User{
            {ID: 1, Name: "Alice", Email: "alice@example.com"},
            {ID: 2, Name: "Bob", Email: "bob@example.com"},
            {ID: 3, Name: "Charlie", Email: "charlie@example.com"},
        }
        
        for _, user := range users {
            err := service.CreateUser(ctx, user)
            require.NoError(t, err)
        }
        
        // List users
        retrieved, err := service.ListUsers(ctx)
        require.NoError(t, err)
        assert.Len(t, retrieved, 3)
    })
}
```

## Advanced Mock Scenarios

### Simulating Errors

You can create custom mock implementations to simulate specific error conditions:

```go
type FailingMockKV struct {
    *mock.MockKV
    FailOnSet    bool
    FailOnGet    bool
    FailOnList   bool
    FailOnDelete bool
}

func (m *FailingMockKV) SetRaw(ctx context.Context, key string, value []byte) error {
    if m.FailOnSet {
        return errors.New("simulated set failure")
    }
    return m.MockKV.SetRaw(ctx, key, value)
}

func (m *FailingMockKV) GetRaw(ctx context.Context, key string) ([]byte, error) {
    if m.FailOnGet {
        return nil, errors.New("simulated get failure")
    }
    return m.MockKV.GetRaw(ctx, key)
}

func (m *FailingMockKV) List(ctx context.Context, prefix string) ([]string, error) {
    if m.FailOnList {
        return nil, errors.New("simulated list failure")
    }
    return m.MockKV.List(ctx, prefix)
}

func (m *FailingMockKV) Delete(ctx context.Context, key string) error {
    if m.FailOnDelete {
        return errors.New("simulated delete failure")
    }
    return m.MockKV.Delete(ctx, key)
}

// Test error handling
func TestErrorHandling(t *testing.T) {
    failingMock := &FailingMockKV{
        MockKV:    &mock.MockKV{Data: map[string][]byte{}},
        FailOnSet: true,
    }
    
    client, err := client.New(failingMock, client.Option{})
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // This should fail
    err = client.Set(ctx, "test", "value")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "simulated set failure")
}
```

### Testing Concurrent Access

Test your code under concurrent conditions:

```go
func TestConcurrentAccess(t *testing.T) {
    mockKV := &mock.MockKV{Data: map[string][]byte{}}
    client, err := client.New(mockKV, client.Option{})
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Use a WaitGroup to coordinate goroutines
    var wg sync.WaitGroup
    numGoroutines := 10
    numOperations := 100
    
    // Start multiple goroutines performing operations
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(goroutineID int) {
            defer wg.Done()
            
            for j := 0; j < numOperations; j++ {
                key := fmt.Sprintf("key:%d:%d", goroutineID, j)
                value := fmt.Sprintf("value:%d:%d", goroutineID, j)
                
                // Set
                err := client.Set(ctx, key, value)
                assert.NoError(t, err)
                
                // Get
                var retrieved string
                err = client.Get(ctx, key, &retrieved)
                assert.NoError(t, err)
                assert.Equal(t, value, retrieved)
                
                // Delete
                err = client.Delete(ctx, key)
                assert.NoError(t, err)
            }
        }(i)
    }
    
    wg.Wait()
}
```

### Testing with Different Encoders

Test your code with different encoding formats:

```go
func TestWithDifferentEncoders(t *testing.T) {
    testCases := []struct {
        name    string
        encoder models.Encoder
    }{
        {"JSON", encoder.JSON},
        {"YAML", encoder.YAML},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            mockKV := &mock.MockKV{Data: map[string][]byte{}}
            client, err := client.New(mockKV, client.Option{
                Encoder: tc.encoder,
            })
            require.NoError(t, err)
            
            ctx := context.Background()
            
            data := struct {
                Name  string `json:"name" yaml:"name"`
                Count int    `json:"count" yaml:"count"`
            }{
                Name:  "Test",
                Count: 42,
            }
            
            err = client.Set(ctx, "test", data)
            require.NoError(t, err)
            
            var retrieved struct {
                Name  string `json:"name" yaml:"name"`
                Count int    `json:"count" yaml:"count"`
            }
            
            err = client.Get(ctx, "test", &retrieved)
            require.NoError(t, err)
            assert.Equal(t, data, retrieved)
        })
    }
}
```

## Test Helpers

Create reusable test helpers for common patterns:

```go
// Test helper functions
func setupTestClient(t *testing.T) client.Client {
    mockKV := &mock.MockKV{Data: map[string][]byte{}}
    client, err := client.New(mockKV, client.Option{})
    require.NoError(t, err)
    return client
}

func setupTestClientWithData(t *testing.T, data map[string]interface{}) client.Client {
    // Encode test data
    encoded := make(map[string][]byte)
    for key, value := range data {
        jsonData, err := json.Marshal(value)
        require.NoError(t, err)
        encoded[key] = jsonData
    }
    
    mockKV := &mock.MockKV{Data: encoded}
    client, err := client.New(mockKV, client.Option{})
    require.NoError(t, err)
    return client
}

// Usage in tests
func TestWithHelpers(t *testing.T) {
    t.Run("EmptyClient", func(t *testing.T) {
        client := setupTestClient(t)
        ctx := context.Background()
        
        err := client.Set(ctx, "key", "value")
        require.NoError(t, err)
    })
    
    t.Run("PrePopulatedClient", func(t *testing.T) {
        initialData := map[string]interface{}{
            "user:1": map[string]string{"name": "Alice"},
            "config":  map[string]int{"port": 8080},
        }
        
        client := setupTestClientWithData(t, initialData)
        ctx := context.Background()
        
        var user map[string]string
        err := client.Get(ctx, "user:1", &user)
        require.NoError(t, err)
        assert.Equal(t, "Alice", user["name"])
    })
}
```

## Integration with Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name        string
        user        User
        wantErr     bool
        expectedErr string
    }{
        {
            name:    "valid user",
            user:    User{ID: 1, Name: "John", Email: "john@example.com"},
            wantErr: false,
        },
        {
            name:        "empty name",
            user:        User{ID: 1, Name: "", Email: "john@example.com"},
            wantErr:     true,
            expectedErr: "name cannot be empty",
        },
        {
            name:        "invalid email",
            user:        User{ID: 1, Name: "John", Email: "invalid-email"},
            wantErr:     true,
            expectedErr: "invalid email format",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := setupTestClient(t)
            service := NewUserService(client)
            ctx := context.Background()
            
            err := service.CreateUser(ctx, tt.user)
            
            if tt.wantErr {
                assert.Error(t, err)
                if tt.expectedErr != "" {
                    assert.Contains(t, err.Error(), tt.expectedErr)
                }
            } else {
                assert.NoError(t, err)
                
                // Verify user was created
                retrieved, err := service.GetUser(ctx, tt.user.ID)
                assert.NoError(t, err)
                assert.Equal(t, tt.user, retrieved)
            }
        })
    }
}
```

## Best Practices for Mock Testing

1. **Fast Tests**: Mock tests should run quickly without external dependencies
2. **Isolated Tests**: Each test should start with a clean mock state
3. **Test Real Scenarios**: Use mocks to test business logic, not just happy paths
4. **Error Simulation**: Test error handling with custom mock implementations
5. **Concurrent Testing**: Test thread safety when applicable
6. **Helper Functions**: Create reusable test setup functions
7. **Clear Assertions**: Make test expectations explicit and clear

The mock backend enables you to write comprehensive, fast, and reliable tests for any code that uses KiviGo, ensuring your application logic works correctly regardless of the underlying storage backend.
