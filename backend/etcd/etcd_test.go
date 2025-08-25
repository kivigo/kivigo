package etcd

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var testEtcd *container

func TestMain(m *testing.M) {
	var err error

	testEtcd, err = start(&testing.T{})
	if err != nil {
		println("Failed to start etcd:", err.Error())
		os.Exit(1)
	}

	code := m.Run()
	_ = testEtcd.Stop(context.Background())

	os.Exit(code)
}

func newTestClient(t *testing.T) *Client {
	t.Helper()

	opt := DefaultOptions(testEtcd.endpoints...)
	client, err := New(opt)
	require.NoError(t, err)

	return client
}

func TestSetRaw(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.SetRaw(ctx, "foo", []byte("bar")))
}

func TestGetRaw(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	_ = c.SetRaw(ctx, "foo", []byte("bar"))
	val, err := c.GetRaw(ctx, "foo")
	require.NoError(t, err)
	require.Equal(t, []byte("bar"), val)
}

func TestDelete(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	_ = c.SetRaw(ctx, "foo", []byte("bar"))
	require.NoError(t, c.Delete(ctx, "foo"))
	_, err := c.GetRaw(ctx, "foo")
	require.Error(t, err)
}

func TestList(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	_ = c.SetRaw(ctx, "prefix/a", []byte("A"))
	_ = c.SetRaw(ctx, "prefix/b", []byte("B"))
	keys, err := c.List(ctx, "prefix/")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(keys), 2)
}

func TestBatchSetRaw(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.BatchSetRaw(ctx, map[string][]byte{
		"batch:key1": []byte("v1"),
		"batch:key2": []byte("v2"),
	}))
}

func TestBatchGetRaw(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	_ = c.BatchSetRaw(ctx, map[string][]byte{
		"batchget:key1": []byte("v1"),
		"batchget:key2": []byte("v2"),
	})
	got, err := c.BatchGetRaw(ctx, []string{"batchget:key1", "batchget:key2"})
	require.NoError(t, err)
	require.Equal(t, []byte("v1"), got["batchget:key1"])
	require.Equal(t, []byte("v2"), got["batchget:key2"])
}

func TestBatchDelete(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	_ = c.BatchSetRaw(ctx, map[string][]byte{
		"batchdel:key1": []byte("v1"),
		"batchdel:key2": []byte("v2"),
	})

	require.NoError(t, c.BatchDelete(ctx, []string{"batchdel:key1", "batchdel:key2"}))
}

func TestHealth(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.Health(ctx))
}

func TestClose(t *testing.T) {
	c := newTestClient(t)
	require.NoError(t, c.Close())
}
