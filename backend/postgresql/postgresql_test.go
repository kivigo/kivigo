package postgresql

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kivigo/kivigo/pkg/errs"
)

var testPostgreSQL *container

func TestMain(m *testing.M) {
	var err error

	testPostgreSQL, err = start(&testing.T{})
	if err != nil {
		fmt.Println("Failed to start PostgreSQL:", err)
		os.Exit(1)
	}

	// Run all tests
	code := m.Run()

	// Cleanup: stop the container
	_ = testPostgreSQL.Stop(context.Background())

	os.Exit(code)
}

func newTestClient(t *testing.T) Client {
	t.Helper()

	opt := DefaultOptions()
	opt.DSN = testPostgreSQL.dsn

	c, err := New(opt)
	require.NoError(t, err)

	return c
}

func TestPostgreSQL_BasicOps(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test SetRaw and GetRaw
	key := "test:key"
	value := []byte("test value")

	err := c.SetRaw(ctx, key, value)
	require.NoError(t, err)

	retrievedValue, err := c.GetRaw(ctx, key)
	require.NoError(t, err)
	require.Equal(t, value, retrievedValue)

	// Test Delete
	err = c.Delete(ctx, key)
	require.NoError(t, err)

	// Verify deletion
	_, err = c.GetRaw(ctx, key)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestPostgreSQL_ErrorCases(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test empty key
	_, err := c.GetRaw(ctx, "")
	require.ErrorIs(t, err, errs.ErrEmptyKey)

	err = c.SetRaw(ctx, "", []byte("value"))
	require.ErrorIs(t, err, errs.ErrEmptyKey)

	err = c.Delete(ctx, "")
	require.ErrorIs(t, err, errs.ErrEmptyKey)

	// Test not found
	_, err = c.GetRaw(ctx, "nonexistent")
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestPostgreSQL_List(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Set up test data
	testData := map[string][]byte{
		"user:1":   []byte("user1data"),
		"user:2":   []byte("user2data"),
		"admin:1":  []byte("admin1data"),
		"config:1": []byte("config1data"),
	}

	for k, v := range testData {
		err := c.SetRaw(ctx, k, v)
		require.NoError(t, err)
	}

	// Test listing with prefix
	keys, err := c.List(ctx, "user:")
	require.NoError(t, err)
	require.Len(t, keys, 2)
	for _, expectedKey := range []string{"user:1", "user:2"} {
		require.Contains(t, keys, expectedKey)
	}

	// Test listing with different prefix
	keys, err = c.List(ctx, "admin:")
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.Contains(t, keys, "admin:1")

	// Test listing all keys (empty prefix)
	keys, err = c.List(ctx, "")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(keys), 4) // At least our test data

	// Clean up
	for key := range testData {
		_ = c.Delete(ctx, key)
	}
}

func TestPostgreSQL_BatchGetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Set up test data
	testData := map[string][]byte{
		"batch1": []byte("value1"),
		"batch2": []byte("value2"),
		"batch3": []byte("value3"),
	}

	for k, v := range testData {
		err := c.SetRaw(ctx, k, v)
		require.NoError(t, err)
	}

	// Test batch get
	keys := []string{"batch1", "batch2", "batch3", "nonexistent"}
	results, err := c.BatchGetRaw(ctx, keys)
	require.NoError(t, err)
	require.Len(t, results, 4)

	require.Equal(t, testData["batch1"], results["batch1"])
	require.Equal(t, testData["batch2"], results["batch2"])
	require.Equal(t, testData["batch3"], results["batch3"])
	require.Nil(t, results["nonexistent"])

	// Clean up
	for key := range testData {
		_ = c.Delete(ctx, key)
	}
}

func TestPostgreSQL_BatchSetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test batch set
	testData := map[string][]byte{
		"batchset1": []byte("value1"),
		"batchset2": []byte("value2"),
		"batchset3": []byte("value3"),
	}

	err := c.BatchSetRaw(ctx, testData)
	require.NoError(t, err)

	// Verify all values were set
	for expectedK, expectedV := range testData {
		actualV, err := c.GetRaw(ctx, expectedK)
		require.NoError(t, err)
		require.Equal(t, expectedV, actualV)
	}

	// Clean up
	for key := range testData {
		_ = c.Delete(ctx, key)
	}
}

func TestPostgreSQL_BatchDelete(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Set up test data
	testData := map[string][]byte{
		"batchdel1": []byte("value1"),
		"batchdel2": []byte("value2"),
		"batchdel3": []byte("value3"),
	}

	for k, v := range testData {
		err := c.SetRaw(ctx, k, v)
		require.NoError(t, err)
	}

	// Test batch delete
	keys := []string{"batchdel1", "batchdel2", "batchdel3"}
	err := c.BatchDelete(ctx, keys)
	require.NoError(t, err)

	// Verify all values were deleted
	for _, key := range keys {
		_, err := c.GetRaw(ctx, key)
		require.ErrorIs(t, err, errs.ErrNotFound)
	}
}

// Test large batch operations
func TestPostgreSQL_LargeBatch(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Create 100 key-value pairs
	batchData := make(map[string][]byte)
	keys := make([]string, 100)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("large:batch:%d", i)
		value := []byte(fmt.Sprintf("value%d", i))
		batchData[key] = value
		keys[i] = key
	}

	// Test batch set
	err := c.BatchSetRaw(ctx, batchData)
	require.NoError(t, err)

	// Test batch get
	results, err := c.BatchGetRaw(ctx, keys)
	require.NoError(t, err)
	require.Len(t, results, 100)

	for key, expectedValue := range batchData {
		actualValue := results[key]
		require.Equal(t, expectedValue, actualValue)
	}

	// Test batch delete
	err = c.BatchDelete(ctx, keys)
	require.NoError(t, err)

	// Verify all deleted
	for _, key := range keys {
		_, err := c.GetRaw(ctx, key)
		require.ErrorIs(t, err, errs.ErrNotFound)
	}
}

func TestPostgreSQL_Health(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test health check
	err := c.Health(ctx)
	require.NoError(t, err)
}
