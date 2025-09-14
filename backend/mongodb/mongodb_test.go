package mongodb

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kivigo/kivigo/pkg/errs"
)

var testMongoDB *container

func TestMain(m *testing.M) {
	var err error

	testMongoDB, err = start(&testing.T{})
	if err != nil {
		fmt.Println("Failed to start MongoDB:", err)
		os.Exit(1)
	}

	// Run all tests
	code := m.Run()

	// Cleanup: stop the container
	_ = testMongoDB.Stop(context.Background())

	os.Exit(code)
}

func mongoAvailable(t *testing.T) bool {
	t.Helper()

	opt := DefaultOptions()
	opt.URI = testMongoDB.uri

	c, err := New(opt)
	if err != nil {
		return false
	}

	defer c.Close()

	return true
}

func newTestClient(t *testing.T) Client {
	t.Helper()

	opt := DefaultOptions()
	opt.URI = testMongoDB.uri

	c, err := New(opt)
	require.NoError(t, err)

	return c
}

func TestMongoDB_BasicOps(t *testing.T) {
	if os.Getenv("CI") == "" && !mongoAvailable(t) {
		t.Skip("MongoDB not available")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	key := "kivigo:test:key"
	val := []byte("value")

	// SetRaw
	require.NoError(t, c.SetRaw(ctx, key, val))

	// GetRaw
	got, err := c.GetRaw(ctx, key)
	require.NoError(t, err)
	require.Equal(t, val, got)

	// List
	keys, err := c.List(ctx, "kivigo:test")
	require.NoError(t, err)
	require.Contains(t, keys, key)

	// Delete
	require.NoError(t, c.Delete(ctx, key))
	_, err = c.GetRaw(ctx, key)
	require.Error(t, err)
	require.Equal(t, errs.ErrNotFound, err)
}

func TestMongoDB_BatchOps(t *testing.T) {
	if os.Getenv("CI") == "" && !mongoAvailable(t) {
		t.Skip("MongoDB not available")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	kv := map[string][]byte{
		"batch:key1": []byte("value1"),
		"batch:key2": []byte("value2"),
		"batch:key3": []byte("value3"),
	}

	// BatchSetRaw
	require.NoError(t, c.BatchSetRaw(ctx, kv))

	// BatchGetRaw
	got, err := c.BatchGetRaw(ctx, []string{"batch:key1", "batch:key2", "batch:key3"})
	require.NoError(t, err)
	require.Equal(t, kv, got)

	// BatchDelete
	require.NoError(t, c.BatchDelete(ctx, []string{"batch:key1", "batch:key2", "batch:key3"}))

	// Verify deletion
	got, err = c.BatchGetRaw(ctx, []string{"batch:key1", "batch:key2", "batch:key3"})
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestMongoDB_Health(t *testing.T) {
	if os.Getenv("CI") == "" && !mongoAvailable(t) {
		t.Skip("MongoDB not available")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	require.NoError(t, c.Health(ctx))
}

func TestMongoDB_ErrorCases(t *testing.T) {
	if os.Getenv("CI") == "" && !mongoAvailable(t) {
		t.Skip("MongoDB not available")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test empty key
	err := c.SetRaw(ctx, "", []byte("value"))
	require.Equal(t, errs.ErrEmptyKey, err)

	// Test getting non-existent key
	_, err = c.GetRaw(ctx, "non-existent-key")
	require.Equal(t, errs.ErrNotFound, err)

	// Test empty key in delete
	err = c.Delete(ctx, "")
	require.Equal(t, errs.ErrEmptyKey, err)

	// Test empty key in batch operations
	err = c.BatchSetRaw(ctx, map[string][]byte{"": []byte("value")})
	require.Equal(t, errs.ErrEmptyKey, err)

	err = c.BatchDelete(ctx, []string{""})
	require.Equal(t, errs.ErrEmptyKey, err)

	// Test empty batch operations
	_, err = c.BatchGetRaw(ctx, []string{})
	require.Error(t, err)

	err = c.BatchSetRaw(ctx, map[string][]byte{})
	require.Error(t, err)

	err = c.BatchDelete(ctx, []string{})
	require.Error(t, err)
}

func TestMongoDB_ListWithPrefix(t *testing.T) {
	if os.Getenv("CI") == "" && !mongoAvailable(t) {
		t.Skip("MongoDB not available")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Set test data
	testData := map[string][]byte{
		"prefix:key1":    []byte("value1"),
		"prefix:key2":    []byte("value2"),
		"different:key3": []byte("value3"),
	}

	for key, value := range testData {
		require.NoError(t, c.SetRaw(ctx, key, value))
	}

	// List with prefix
	keys, err := c.List(ctx, "prefix:")
	require.NoError(t, err)
	require.Len(t, keys, 2)
	require.Contains(t, keys, "prefix:key1")
	require.Contains(t, keys, "prefix:key2")
	require.NotContains(t, keys, "different:key3")

	// List all
	allKeys, err := c.List(ctx, "")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(allKeys), 3)

	// Clean up
	for key := range testData {
		require.NoError(t, c.Delete(ctx, key))
	}
}
