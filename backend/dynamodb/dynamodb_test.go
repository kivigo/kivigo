package dynamodb

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kivigo/kivigo/pkg/errs"
)

var testDynamoDB *container

func TestMain(m *testing.M) {
	var err error

	testDynamoDB, err = start(&testing.T{})
	if err != nil {
		panic(err)
	}

	code := m.Run()

	_ = testDynamoDB.Stop(context.Background())

	os.Exit(code)
}

func newTestClient(t *testing.T) Client {
	t.Helper()

	opt := DefaultOptions()
	opt.Endpoint = testDynamoDB.endpoint
	opt.AccessKey = "test"
	opt.SecretKey = "test"
	opt.TableName = "test_kivigo"

	c, err := New(opt)
	require.NoError(t, err)

	return c
}

type testCase struct {
	name    string
	key     string
	value   []byte
	wantErr bool
}

func TestDynamoDB_SetRaw(t *testing.T) {
	tests := []testCase{
		{"Valid", "foo", []byte("bar"), false},
		{"EmptyKey", "", []byte("bar"), true},
		{"EmptyValue", "foo", []byte(""), false},
		{"LargeValue", "foo", make([]byte, 1024), false},
	}

	c := newTestClient(t)
	defer c.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.SetRaw(context.Background(), tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDynamoDB_GetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test getting non-existent key
	_, err := c.GetRaw(ctx, "nonexistent")
	require.ErrorIs(t, err, errs.ErrNotFound)

	// Test empty key
	_, err = c.GetRaw(ctx, "")
	require.ErrorIs(t, err, errs.ErrEmptyKey)

	// Test setting and getting a key
	key := "test-key"
	value := []byte("test-value")

	err = c.SetRaw(ctx, key, value)
	require.NoError(t, err)

	retrieved, err := c.GetRaw(ctx, key)
	require.NoError(t, err)
	require.Equal(t, value, retrieved)
}

func TestDynamoDB_List(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test empty prefix
	_, err := c.List(ctx, "")
	require.ErrorIs(t, err, errs.ErrEmptyPrefix)

	// Set up test data
	testData := map[string][]byte{
		"user:1":  []byte("alice"),
		"user:2":  []byte("bob"),
		"user:3":  []byte("charlie"),
		"admin:1": []byte("admin1"),
	}

	for k, v := range testData {
		err := c.SetRaw(ctx, k, v)
		require.NoError(t, err)
	}

	// Test listing with prefix
	keys, err := c.List(ctx, "user:")
	require.NoError(t, err)
	require.Len(t, keys, 3)

	// Verify all user keys are present
	expectedKeys := []string{"user:1", "user:2", "user:3"}
	for _, expectedKey := range expectedKeys {
		require.Contains(t, keys, expectedKey)
	}

	// Test listing with different prefix
	keys, err = c.List(ctx, "admin:")
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.Contains(t, keys, "admin:1")
}

func TestDynamoDB_Delete(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test empty key
	err := c.Delete(ctx, "")
	require.ErrorIs(t, err, errs.ErrEmptyKey)

	// Test deleting non-existent key (should not error)
	err = c.Delete(ctx, "nonexistent")
	require.NoError(t, err)

	// Test setting and deleting a key
	key := "delete-test"
	value := []byte("delete-value")

	err = c.SetRaw(ctx, key, value)
	require.NoError(t, err)

	// Verify key exists
	_, err = c.GetRaw(ctx, key)
	require.NoError(t, err)

	// Delete key
	err = c.Delete(ctx, key)
	require.NoError(t, err)

	// Verify key is deleted
	_, err = c.GetRaw(ctx, key)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestDynamoDB_Health(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	err := c.Health(context.Background())
	require.NoError(t, err)
}

func TestDynamoDB_BatchGetRaw(t *testing.T) {
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

	// Verify existing keys
	require.Equal(t, testData["batch1"], results["batch1"])
	require.Equal(t, testData["batch2"], results["batch2"])
	require.Equal(t, testData["batch3"], results["batch3"])
	require.Nil(t, results["nonexistent"])
}

func TestDynamoDB_BatchSetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test empty batch
	err := c.BatchSetRaw(ctx, map[string][]byte{})
	require.ErrorIs(t, err, errs.ErrEmptyBatch)

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
}

func TestDynamoDB_BatchDelete(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test empty batch
	err := c.BatchDelete(ctx, []string{})
	require.ErrorIs(t, err, errs.ErrEmptyBatch)

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

func TestDynamoDB_BasicOps(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test complete workflow
	key := "workflow-test"
	value := []byte("workflow-value")

	// Set
	err := c.SetRaw(ctx, key, value)
	require.NoError(t, err)

	// Get
	retrieved, err := c.GetRaw(ctx, key)
	require.NoError(t, err)
	require.Equal(t, value, retrieved)

	// List (set a few more keys)
	err = c.SetRaw(ctx, "workflow-test-2", []byte("value2"))
	require.NoError(t, err)
	err = c.SetRaw(ctx, "workflow-test-3", []byte("value3"))
	require.NoError(t, err)

	keys, err := c.List(ctx, "workflow-test")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(keys), 3)

	// Delete
	err = c.Delete(ctx, key)
	require.NoError(t, err)

	// Verify deleted
	_, err = c.GetRaw(ctx, key)
	require.ErrorIs(t, err, errs.ErrNotFound)

	// Health check
	err = c.Health(ctx)
	require.NoError(t, err)
}

// Test large batch operations
func TestDynamoDB_LargeBatch(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Test with more than batch size (25 for writes, 100 for reads)
	const numItems = 50
	testData := make(map[string][]byte, numItems)
	keys := make([]string, 0, numItems)

	for i := 0; i < numItems; i++ {
		key := fmt.Sprintf("large-batch-%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		testData[key] = value

		keys = append(keys, key)
	}

	// Batch set
	err := c.BatchSetRaw(ctx, testData)
	require.NoError(t, err)

	// Batch get
	results, err := c.BatchGetRaw(ctx, keys)
	require.NoError(t, err)
	require.Len(t, results, numItems)

	// Verify all values
	for k, expectedV := range testData {
		actualV, ok := results[k]
		require.True(t, ok)
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
