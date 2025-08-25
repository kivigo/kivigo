package memcached

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/azrod/kivigo"
)

// TestMemcached_Integration tests the memcached backend with the main KiviGo client
func TestMemcached_Integration(t *testing.T) {
	if os.Getenv("CI") == "" && !memcachedAvailable(t) {
		t.Skip("Memcached not available")
	}

	// Create memcached backend
	backend, err := New(Option{
		Servers: []string{testMemcached.addr},
	})
	require.NoError(t, err)

	// Create KiviGo client
	client, err := kivigo.New(backend)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	// Test basic operations through KiviGo client
	key := "integration:test"
	value := "test-value"

	// Set value
	err = client.Set(ctx, key, value)
	require.NoError(t, err)

	// Get value
	var retrievedValue string
	err = client.Get(ctx, key, &retrievedValue)
	require.NoError(t, err)
	require.Equal(t, value, retrievedValue)

	// List keys
	keys, err := client.List(ctx, "integration:")
	require.NoError(t, err)
	require.Contains(t, keys, key)

	// Health check
	err = client.Health(ctx, nil)
	require.NoError(t, err)

	// Delete key
	err = client.Delete(ctx, key)
	require.NoError(t, err)

	// Verify deletion
	err = client.Get(ctx, key, &retrievedValue)
	require.Error(t, err)
}