//go:build !unit

package cassandra

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/azrod/kivigo/pkg/errs"
)

var testCassandra *container

func TestMain(m *testing.M) {
	// Skip testcontainer tests if explicitly disabled
	if os.Getenv("SKIP_TESTCONTAINERS") == "true" {
		fmt.Println("⏭️ Skipping testcontainer tests (SKIP_TESTCONTAINERS=true)")
		os.Exit(0)
	}

	var err error

	testCassandra, err = start(&testing.T{})
	if err != nil {
		fmt.Printf("Failed to start Cassandra: %v\n", err)
		fmt.Println("💡 Tip: Set SKIP_TESTCONTAINERS=true to skip these tests")
		os.Exit(1)
	}

	// Run all tests
	code := m.Run()

	// Cleanup: stop the container
	if testCassandra != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = testCassandra.Stop(ctx)
	}

	os.Exit(code)
}

func cassandraAvailable(t *testing.T) bool {
	t.Helper()

	if testCassandra == nil {
		return false
	}

	opt := DefaultOptions()
	opt.Hosts = testCassandra.hosts

	c, err := New(opt)
	if err != nil {
		t.Logf("Cassandra connection failed: %v", err)
		return false
	}

	defer c.Close()

	// Test basic connectivity with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Health(ctx); err != nil {
		t.Logf("Cassandra health check failed: %v", err)
		return false
	}

	return true
}

func newTestClient(t *testing.T) Client {
	t.Helper()

	opt := DefaultOptions()
	opt.Hosts = testCassandra.hosts

	c, err := New(opt)
	require.NoError(t, err)

	return c
}

func TestCassandra_BasicOps(t *testing.T) {
	if os.Getenv("SKIP_TESTCONTAINERS") == "true" {
		t.Skip("Skipping testcontainer tests (SKIP_TESTCONTAINERS=true)")
	}
	
	if os.Getenv("CI") == "" && !cassandraAvailable(t) {
		t.Skip("Cassandra not available")
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

func TestCassandra_BatchOps(t *testing.T) {
	if os.Getenv("SKIP_TESTCONTAINERS") == "true" {
		t.Skip("Skipping testcontainer tests (SKIP_TESTCONTAINERS=true)")
	}
	
	if os.Getenv("CI") == "" && !cassandraAvailable(t) {
		t.Skip("Cassandra not available")
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
	require.Len(t, got, 3)
	require.Nil(t, got["batch:key1"])
	require.Nil(t, got["batch:key2"])
	require.Nil(t, got["batch:key3"])
}

func TestCassandra_Health(t *testing.T) {
	if os.Getenv("SKIP_TESTCONTAINERS") == "true" {
		t.Skip("Skipping testcontainer tests (SKIP_TESTCONTAINERS=true)")
	}
	
	if os.Getenv("CI") == "" && !cassandraAvailable(t) {
		t.Skip("Cassandra not available")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	require.NoError(t, c.Health(ctx))
}

func TestCassandra_ErrorCases(t *testing.T) {
	if os.Getenv("SKIP_TESTCONTAINERS") == "true" {
		t.Skip("Skipping testcontainer tests (SKIP_TESTCONTAINERS=true)")
	}
	
	if os.Getenv("CI") == "" && !cassandraAvailable(t) {
		t.Skip("Cassandra not available")
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

func TestCassandra_ListWithPrefix(t *testing.T) {
	if os.Getenv("SKIP_TESTCONTAINERS") == "true" {
		t.Skip("Skipping testcontainer tests (SKIP_TESTCONTAINERS=true)")
	}
	
	if os.Getenv("CI") == "" && !cassandraAvailable(t) {
		t.Skip("Cassandra not available")
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

	// Clean up
	for key := range testData {
		require.NoError(t, c.Delete(ctx, key))
	}
}

func TestCassandra_BatchGetPartialResults(t *testing.T) {
	if os.Getenv("SKIP_TESTCONTAINERS") == "true" {
		t.Skip("Skipping testcontainer tests (SKIP_TESTCONTAINERS=true)")
	}
	
	if os.Getenv("CI") == "" && !cassandraAvailable(t) {
		t.Skip("Cassandra not available")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Set only some keys
	require.NoError(t, c.SetRaw(ctx, "exists1", []byte("value1")))
	require.NoError(t, c.SetRaw(ctx, "exists2", []byte("value2")))

	// BatchGet with mix of existing and non-existing keys
	got, err := c.BatchGetRaw(ctx, []string{"exists1", "notexists", "exists2"})
	require.NoError(t, err)
	require.Equal(t, []byte("value1"), got["exists1"])
	require.Equal(t, []byte("value2"), got["exists2"])
	require.Nil(t, got["notexists"])

	// Clean up
	require.NoError(t, c.Delete(ctx, "exists1"))
	require.NoError(t, c.Delete(ctx, "exists2"))
}
