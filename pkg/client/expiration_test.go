package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/azrod/kivigo/pkg/encoder"
	"github.com/azrod/kivigo/pkg/mock"
	"github.com/azrod/kivigo/pkg/models"
)

// mockKVWithTTL extends MockKV to support TTL operations.
type mockKVWithTTL struct {
	*mock.MockKV
	expirations map[string]time.Time
}

func newMockKVWithTTL() *mockKVWithTTL {
	return &mockKVWithTTL{
		MockKV:      &mock.MockKV{Data: make(map[string][]byte)},
		expirations: make(map[string]time.Time),
	}
}

func (m *mockKVWithTTL) SupportsExpiration() bool {
	return true
}

func (m *mockKVWithTTL) Expire(ctx context.Context, key string, ttl time.Duration) error {
	// Check if key exists
	if _, err := m.GetRaw(ctx, key); err != nil {
		return err // Key not found or other error
	}

	// Set expiration time
	m.expirations[key] = time.Now().Add(ttl)
	return nil
}

// Verify interface compliance
var _ models.KVWithTTL = (*mockKVWithTTL)(nil)

func TestClient_SupportsExpiration(t *testing.T) {
	tests := []struct {
		name     string
		backend  models.KV
		expected bool
	}{
		{
			name:     "Backend with TTL support",
			backend:  newMockKVWithTTL(),
			expected: true,
		},
		{
			name:     "Backend without TTL support",
			backend:  &mock.MockKV{Data: make(map[string][]byte)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.backend, Option{Encoder: encoder.JSON})
			require.NoError(t, err)

			result := client.SupportsExpiration()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_Expire(t *testing.T) {
	t.Run("Success with TTL backend", func(t *testing.T) {
		backend := newMockKVWithTTL()
		client, err := New(backend, Option{Encoder: encoder.JSON})
		require.NoError(t, err)

		ctx := context.Background()
		key := "test-key"
		value := []byte("test-value")
		ttl := 30 * time.Second

		// Set a key first
		err = backend.SetRaw(ctx, key, value)
		require.NoError(t, err)

		// Test expiration
		err = client.Expire(ctx, key, ttl)
		assert.NoError(t, err)

		// Verify expiration was set
		expectedExpiration := time.Now().Add(ttl)
		actualExpiration, exists := backend.expirations[key]
		assert.True(t, exists)
		assert.WithinDuration(t, expectedExpiration, actualExpiration, time.Second)
	})

	t.Run("Error with non-TTL backend", func(t *testing.T) {
		backend := &mock.MockKV{Data: make(map[string][]byte)}
		client, err := New(backend, Option{Encoder: encoder.JSON})
		require.NoError(t, err)

		ctx := context.Background()
		key := "test-key"
		ttl := 30 * time.Second

		err = client.Expire(ctx, key, ttl)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expiration not supported by backend")
	})

	t.Run("Error when key does not exist", func(t *testing.T) {
		backend := newMockKVWithTTL()
		client, err := New(backend, Option{Encoder: encoder.JSON})
		require.NoError(t, err)

		ctx := context.Background()
		key := "non-existent-key"
		ttl := 30 * time.Second

		err = client.Expire(ctx, key, ttl)
		assert.Error(t, err)
	})
}

func TestClient_Expire_Integration(t *testing.T) {
	t.Run("Complete workflow with TTL", func(t *testing.T) {
		backend := newMockKVWithTTL()
		client, err := New(backend, Option{Encoder: encoder.JSON})
		require.NoError(t, err)

		ctx := context.Background()

		// Check if expiration is supported
		require.True(t, client.SupportsExpiration(), "Expected backend to support expiration")

		// Set a value
		err = client.Set(ctx, "user:123", map[string]string{"name": "John"})
		require.NoError(t, err)

		// Set expiration
		err = client.Expire(ctx, "user:123", 1*time.Hour)
		assert.NoError(t, err)

		// Verify we can still get the value
		var result map[string]string
		err = client.Get(ctx, "user:123", &result)
		require.NoError(t, err)
		assert.Equal(t, "John", result["name"])
	})
}
