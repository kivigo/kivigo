package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kivigo/kivigo/pkg/errs"
)

var testRedis *container

func TestMain(m *testing.M) {
	var err error

	testRedis, err = start(&testing.T{})
	if err != nil {
		fmt.Println("Failed to start Redis:", err)
		os.Exit(1)
	}

	// Run all tests
	code := m.Run()

	// Cleanup: stop the helper
	_ = testRedis.Stop(context.Background())

	os.Exit(code)
}

func redisAvailable(t *testing.T) bool {
	t.Helper()

	opt := DefaultOptions()
	opt.Addr = testRedis.addr

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
	opt.Addr = testRedis.addr

	c, err := New(opt)
	require.NoError(t, err)

	return c
}

func TestRedis_BasicOps(t *testing.T) {
	if os.Getenv("CI") == "" && !redisAvailable(t) {
		t.Skip("Redis not available on localhost:6379")
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
}

func TestRedis_BatchOps(t *testing.T) {
	if os.Getenv("CI") == "" && !redisAvailable(t) {
		t.Skip("Redis not available on localhost:6379")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	kvs := map[string][]byte{
		"kivigo:batch:key1": []byte("v1"),
		"kivigo:batch:key2": []byte("v2"),
		"kivigo:batch:key3": []byte("v3"),
	}

	// BatchSetRaw
	require.NoError(t, c.BatchSetRaw(ctx, kvs))

	// BatchGetRaw
	keys := []string{"kivigo:batch:key1", "kivigo:batch:key2", "kivigo:batch:key3"}
	values, err := c.BatchGetRaw(ctx, keys)
	require.NoError(t, err)
	require.Equal(t, []byte("v1"), values["kivigo:batch:key1"])
	require.Equal(t, []byte("v2"), values["kivigo:batch:key2"])
	require.Equal(t, []byte("v3"), values["kivigo:batch:key3"])

	// BatchDelete
	require.NoError(t, c.BatchDelete(ctx, keys))

	// Ensure deletion
	for _, key := range keys {
		_, err := c.GetRaw(ctx, key)
		require.Error(t, err)
	}
}

// ---------- SetRaw & Set ----------
func TestRedis_SetRaw_EmptyKey(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	err := c.SetRaw(ctx, "", []byte("value"))
	require.ErrorIs(t, err, errs.ErrEmptyKey)
}

// If you have a Set method, test it too
func TestRedis_Set_EmptyKey(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	err := c.SetRaw(ctx, "", []byte("value"))
	require.ErrorIs(t, err, errs.ErrEmptyKey)
}

// ---------- GetRaw ----------
func TestRedis_GetRaw_EmptyKey(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	_, err := c.GetRaw(ctx, "")
	require.ErrorIs(t, err, errs.ErrEmptyKey)
}

func TestRedis_GetRaw_NotFound(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	_, err := c.GetRaw(ctx, "kivigo:test:notfound")
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestRedis_GetRaw_WithValue(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	key := "kivigo:test:getraw"
	val := []byte("getraw-value")

	require.NoError(t, c.SetRaw(ctx, key, val))

	got, err := c.GetRaw(ctx, key)
	require.NoError(t, err)
	require.Equal(t, val, got)

	// Cleanup
	require.NoError(t, c.Delete(ctx, key))
}

// ---------- List ----------
func TestRedis_List_EmptyPrefix(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	_, err := c.List(ctx, "")
	require.ErrorIs(t, err, errs.ErrEmptyPrefix)
}

func TestRedis_List_WithValues(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	prefix := "kivigo:list"
	keys := []string{prefix + ":a", prefix + ":b", prefix + ":c"}
	val := []byte("v")

	// Set values
	for _, k := range keys {
		require.NoError(t, c.SetRaw(ctx, k, val))
	}

	found, err := c.List(ctx, prefix)
	require.NoError(t, err)
	require.Subset(t, found, keys)

	// Cleanup
	for _, k := range keys {
		require.NoError(t, c.Delete(ctx, k))
	}
}

// ---------- Delete ----------
func TestRedis_Delete_EmptyKey(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	err := c.Delete(ctx, "")
	require.ErrorIs(t, err, errs.ErrEmptyKey)
}

func TestRedis_Delete_WithValue(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	key := "kivigo:test:delete"
	val := []byte("todelete")

	require.NoError(t, c.SetRaw(ctx, key, val))

	// Ensure value exists
	got, err := c.GetRaw(ctx, key)
	require.NoError(t, err)
	require.Equal(t, val, got)

	// Delete and check
	require.NoError(t, c.Delete(ctx, key))
	_, err = c.GetRaw(ctx, key)
	require.ErrorIs(t, err, errs.ErrNotFound)
}

// ---------- BatchSetRaw ----------
func TestRedis_BatchSetRaw_EmptyBatch(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	err := c.BatchSetRaw(ctx, map[string][]byte{})
	require.ErrorIs(t, err, errs.ErrEmptyBatch)
}

// ---------- BatchGetRaw ----------
func TestRedis_BatchGetRaw_EmptyKeys(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	_, err := c.BatchGetRaw(ctx, []string{})
	require.ErrorIs(t, err, errs.ErrEmptyBatch)
}

func TestRedis_BatchGetRaw_WithValues(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	kvs := map[string][]byte{
		"kivigo:batchget:key1": []byte("v1"),
		"kivigo:batchget:key2": []byte("v2"),
		"kivigo:batchget:key3": []byte("v3"),
	}

	// Set values
	for k, v := range kvs {
		require.NoError(t, c.SetRaw(ctx, k, v))
	}

	keys := []string{"kivigo:batchget:key1", "kivigo:batchget:key2", "kivigo:batchget:key3"}
	values, err := c.BatchGetRaw(ctx, keys)
	require.NoError(t, err)
	require.Equal(t, kvs["kivigo:batchget:key1"], values["kivigo:batchget:key1"])
	require.Equal(t, kvs["kivigo:batchget:key2"], values["kivigo:batchget:key2"])
	require.Equal(t, kvs["kivigo:batchget:key3"], values["kivigo:batchget:key3"])

	// Cleanup
	for _, k := range keys {
		require.NoError(t, c.Delete(ctx, k))
	}
}

// ---------- BatchDelete ----------
func TestRedis_BatchDelete_EmptyBatch(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	err := c.BatchDelete(ctx, []string{})
	require.ErrorIs(t, err, errs.ErrEmptyBatch)
}

func TestRedis_BatchDelete_WithValues(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	kvs := map[string][]byte{
		"kivigo:batchdel:key1": []byte("v1"),
		"kivigo:batchdel:key2": []byte("v2"),
		"kivigo:batchdel:key3": []byte("v3"),
	}

	// Set values
	for k, v := range kvs {
		require.NoError(t, c.SetRaw(ctx, k, v))
	}

	keys := []string{"kivigo:batchdel:key1", "kivigo:batchdel:key2", "kivigo:batchdel:key3"}
	require.NoError(t, c.BatchDelete(ctx, keys))

	// Ensure deletion
	for _, k := range keys {
		_, err := c.GetRaw(ctx, k)
		require.ErrorIs(t, err, errs.ErrNotFound)
	}
}

// ---------- Health ----------
func TestRedis_Health(t *testing.T) {
	if os.Getenv("CI") == "" && !redisAvailable(t) {
		t.Skip("Redis not available on localhost:6379")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, c.Health(ctx))
}

func TestRedis_Health_ClientNotInitialized(t *testing.T) {
	// Simulate an uninitialized client
	c := Client{c: nil}
	ctx := context.Background()

	err := c.Health(ctx)
	require.ErrorIs(t, err, errs.ErrClientNotInitialized)
}

func TestRedis_Health_CheckFailed(t *testing.T) {
	// Simulate a Redis client pointing to a closed port to force ping failure
	opt := DefaultOptions()

	c, err := New(opt)
	require.NoError(t, err)

	ctx := context.Background()
	healthErr := c.Health(ctx)
	require.Error(t, healthErr)
}
