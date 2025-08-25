package memcached

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/azrod/kivigo/pkg/errs"
)

var testMemcached *container

func TestMain(m *testing.M) {
	var err error

	testMemcached, err = start(&testing.T{})
	if err != nil {
		// If we can't start memcached, skip tests
		os.Exit(0)
	}

	// Run all tests
	code := m.Run()

	// Cleanup: stop the container
	_ = testMemcached.Stop(context.Background())

	os.Exit(code)
}

func memcachedAvailable(t *testing.T) bool {
	t.Helper()

	opt := DefaultOptions()
	opt.Servers = []string{testMemcached.addr}

	c, err := New(opt)
	if err != nil {
		return false
	}

	defer c.Close()

	return c.Health(context.Background()) == nil
}

func newTestClient(t *testing.T) *Client {
	t.Helper()

	opt := DefaultOptions()
	opt.Servers = []string{testMemcached.addr}

	c, err := New(opt)
	require.NoError(t, err)

	return c
}

func TestMemcached_BasicOps(t *testing.T) {
	if os.Getenv("CI") == "" && !memcachedAvailable(t) {
		t.Skip("Memcached not available")
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
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestMemcached_BatchOps(t *testing.T) {
	if os.Getenv("CI") == "" && !memcachedAvailable(t) {
		t.Skip("Memcached not available")
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
		require.ErrorIs(t, err, errs.ErrNotFound)
	}
}

// ---------- SetRaw ----------
func TestMemcached_SetRaw_EmptyKey(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	err := c.SetRaw(ctx, "", []byte("value"))
	require.ErrorIs(t, err, errs.ErrEmptyKey)
}

// ---------- GetRaw ----------
func TestMemcached_GetRaw_EmptyKey(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	_, err := c.GetRaw(ctx, "")
	require.ErrorIs(t, err, errs.ErrEmptyKey)
}

func TestMemcached_GetRaw_NotFound(t *testing.T) {
	if os.Getenv("CI") == "" && !memcachedAvailable(t) {
		t.Skip("Memcached not available")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	_, err := c.GetRaw(ctx, "kivigo:test:notfound")
	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestMemcached_GetRaw_WithValue(t *testing.T) {
	if os.Getenv("CI") == "" && !memcachedAvailable(t) {
		t.Skip("Memcached not available")
	}

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
func TestMemcached_List_EmptyPrefix(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	_, err := c.List(ctx, "")
	require.ErrorIs(t, err, errs.ErrEmptyPrefix)
}

// ---------- Delete ----------
func TestMemcached_Delete_EmptyKey(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	err := c.Delete(ctx, "")
	require.ErrorIs(t, err, errs.ErrEmptyKey)
}

func TestMemcached_Delete_NonExistentKey(t *testing.T) {
	if os.Getenv("CI") == "" && !memcachedAvailable(t) {
		t.Skip("Memcached not available")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	// Deleting a non-existent key should not return an error
	require.NoError(t, c.Delete(ctx, "kivigo:test:nonexistent"))
}

// ---------- Health ----------
func TestMemcached_Health(t *testing.T) {
	if os.Getenv("CI") == "" && !memcachedAvailable(t) {
		t.Skip("Memcached not available")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, c.Health(ctx))
}

func TestMemcached_Health_ClientNotInitialized(t *testing.T) {
	// Simulate an uninitialized client
	c := &Client{c: nil}
	ctx := context.Background()

	err := c.Health(ctx)
	require.ErrorIs(t, err, errs.ErrClientNotInitialized)
}

// ---------- BatchGetRaw ----------
func TestMemcached_BatchGetRaw_EmptyKeys(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	_, err := c.BatchGetRaw(ctx, []string{})
	require.ErrorIs(t, err, errs.ErrEmptyBatch)
}

func TestMemcached_BatchGetRaw_MixedFound(t *testing.T) {
	if os.Getenv("CI") == "" && !memcachedAvailable(t) {
		t.Skip("Memcached not available")
	}

	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()

	// Set some keys
	require.NoError(t, c.SetRaw(ctx, "found1", []byte("value1")))
	require.NoError(t, c.SetRaw(ctx, "found2", []byte("value2")))

	// Try to get mix of found and not found keys
	keys := []string{"found1", "notfound", "found2"}
	values, err := c.BatchGetRaw(ctx, keys)
	require.NoError(t, err)

	require.Equal(t, []byte("value1"), values["found1"])
	require.Nil(t, values["notfound"])
	require.Equal(t, []byte("value2"), values["found2"])

	// Cleanup
	require.NoError(t, c.Delete(ctx, "found1"))
	require.NoError(t, c.Delete(ctx, "found2"))
}

// ---------- BatchSetRaw ----------
func TestMemcached_BatchSetRaw_EmptyBatch(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	err := c.BatchSetRaw(ctx, map[string][]byte{})
	require.ErrorIs(t, err, errs.ErrEmptyBatch)
}

// ---------- BatchDelete ----------
func TestMemcached_BatchDelete_EmptyBatch(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	ctx := context.Background()
	err := c.BatchDelete(ctx, []string{})
	require.ErrorIs(t, err, errs.ErrEmptyBatch)
}
