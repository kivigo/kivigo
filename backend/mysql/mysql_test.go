package mysql

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kivigo/kivigo/pkg/errs"
)

var testMySQL *container

func TestMain(m *testing.M) {
	var err error

	testMySQL, err = start(&testing.T{})
	if err != nil {
		fmt.Println("Failed to start MySQL:", err)
		os.Exit(1)
	}

	// Run all tests
	code := m.Run()

	// Cleanup: stop the container
	_ = testMySQL.Stop(context.Background())

	os.Exit(code)
}

func newTestClient(t *testing.T) Client {
	t.Helper()

	opt := DefaultOptions()
	opt.DSN = testMySQL.dsn

	c, err := New(opt)
	require.NoError(t, err)

	return c
}

func TestMySQL_BasicOps(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test SetRaw
	key := "test-key"
	value := []byte("test-value")
	err := c.SetRaw(ctx, key, value)
	require.NoError(t, err)

	// Test GetRaw
	retrieved, err := c.GetRaw(ctx, key)
	require.NoError(t, err)
	require.Equal(t, value, retrieved)

	// Test List
	keys, err := c.List(ctx, "test")
	require.NoError(t, err)
	require.Contains(t, keys, key)

	// Test Delete
	err = c.Delete(ctx, key)
	require.NoError(t, err)

	// Verify key is deleted
	_, err = c.GetRaw(ctx, key)
	require.ErrorIs(t, err, errs.ErrNotFound)

	// Health check
	err = c.Health(ctx)
	require.NoError(t, err)
}

func TestMySQL_ErrorCases(t *testing.T) {
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

func TestMySQL_List(t *testing.T) {
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

func TestMySQL_BatchGetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test empty batch
	_, err := c.BatchGetRaw(ctx, []string{})
	require.ErrorIs(t, err, errs.ErrEmptyBatch)

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

	// Check results
	require.Equal(t, testData["batch1"], results["batch1"])
	require.Equal(t, testData["batch2"], results["batch2"])
	require.Equal(t, testData["batch3"], results["batch3"])
	require.Nil(t, results["nonexistent"])

	// Clean up
	for key := range testData {
		_ = c.Delete(ctx, key)
	}
}

func TestMySQL_BatchSetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test empty batch
	err := c.BatchSetRaw(ctx, map[string][]byte{})
	require.ErrorIs(t, err, errs.ErrEmptyBatch)

	// Test batch with empty key
	err = c.BatchSetRaw(ctx, map[string][]byte{
		"valid_key": []byte("value"),
		"":          []byte("empty_key_value"),
	})
	require.ErrorIs(t, err, errs.ErrEmptyKey)

	// Test batch set
	testData := map[string][]byte{
		"batchset1": []byte("setvalue1"),
		"batchset2": []byte("setvalue2"),
		"batchset3": []byte("setvalue3"),
	}

	err = c.BatchSetRaw(ctx, testData)
	require.NoError(t, err)

	// Verify all keys were set
	for k, expectedV := range testData {
		actualV, err := c.GetRaw(ctx, k)
		require.NoError(t, err)
		require.Equal(t, expectedV, actualV)
	}

	// Clean up
	for key := range testData {
		_ = c.Delete(ctx, key)
	}
}

func TestMySQL_BatchDelete(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test empty batch
	err := c.BatchDelete(ctx, []string{})
	require.ErrorIs(t, err, errs.ErrEmptyBatch)

	// Test batch with empty key
	err = c.BatchDelete(ctx, []string{"valid_key", ""})
	require.ErrorIs(t, err, errs.ErrEmptyKey)

	// Set up test data
	testData := map[string][]byte{
		"batchdel1": []byte("delvalue1"),
		"batchdel2": []byte("delvalue2"),
		"batchdel3": []byte("delvalue3"),
	}

	for k, v := range testData {
		err := c.SetRaw(ctx, k, v)
		require.NoError(t, err)
	}

	// Verify keys exist
	for k := range testData {
		_, err := c.GetRaw(ctx, k)
		require.NoError(t, err)
	}

	// Batch delete
	keys := []string{"batchdel1", "batchdel2", "batchdel3"}
	err = c.BatchDelete(ctx, keys)
	require.NoError(t, err)

	// Verify keys are deleted
	for _, k := range keys {
		_, err := c.GetRaw(ctx, k)
		require.ErrorIs(t, err, errs.ErrNotFound)
	}
}

// Test large batch operations
func TestMySQL_LargeBatch(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Create a large batch (100 items)
	batchSize := 100
	testData := make(map[string][]byte, batchSize)
	keys := make([]string, 0, batchSize)

	for i := 0; i < batchSize; i++ {
		key := fmt.Sprintf("large_batch_key_%d", i)
		value := []byte(fmt.Sprintf("large_batch_value_%d", i))
		testData[key] = value
		keys = append(keys, key)
	}

	// Batch set
	err := c.BatchSetRaw(ctx, testData)
	require.NoError(t, err)

	// Batch get
	results, err := c.BatchGetRaw(ctx, keys)
	require.NoError(t, err)
	require.Len(t, results, batchSize)

	// Verify all values
	for key, expectedV := range testData {
		actualV := results[key]
		require.Equal(t, expectedV, actualV)
	}

	// Batch delete
	err = c.BatchDelete(ctx, keys)
	require.NoError(t, err)

	// Verify all deleted
	for _, key := range keys {
		_, err := c.GetRaw(ctx, key)
		require.ErrorIs(t, err, errs.ErrNotFound)
	}
}

func TestMySQL_Health(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test healthy connection
	err := c.Health(ctx)
	require.NoError(t, err)

	// Test uninitialized client
	uninitializedClient := Client{db: nil}
	err = uninitializedClient.Health(ctx)
	require.ErrorIs(t, err, errs.ErrClientNotInitialized)
}
