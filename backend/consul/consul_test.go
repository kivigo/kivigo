package consul

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
)

var testConsul *container

func TestMain(m *testing.M) {
	var err error

	testConsul, err = start(&testing.T{})
	if err != nil {
		println("Failed to start Consul:", err.Error())
		os.Exit(1)
	}

	code := m.Run()
	_ = testConsul.Stop(context.Background())

	os.Exit(code)
}

func newTestClient(t *testing.T) *Client {
	t.Helper()

	cfg := api.DefaultConfig()
	cfg.Address = testConsul.addr
	client, err := New(cfg)
	require.NoError(t, err)

	return client
}

func TestSetRaw(t *testing.T) {
	type testCase struct {
		name      string
		key       string
		value     []byte
		expectErr bool
	}

	tests := []testCase{
		{"Valid", "foo", []byte("bar"), false},
		{"EmptyKey", "", []byte("bar"), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t)
			ctx := context.Background()

			err := c.SetRaw(ctx, tc.key, tc.value)
			if tc.expectErr && err == nil {
				t.Errorf("expected error for SetRaw, got nil")
			}

			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error for SetRaw: %v", err)
			}
		})
	}
}

func TestGetRaw(t *testing.T) {
	type testCase struct {
		name      string
		key       string
		setup     func(c *Client, ctx context.Context)
		want      []byte
		expectErr bool
	}

	tests := []testCase{
		{
			name: "Valid",
			key:  "foo",
			setup: func(c *Client, ctx context.Context) {
				_ = c.SetRaw(ctx, "foo", []byte("bar"))
			},
			want:      []byte("bar"),
			expectErr: false,
		},
		{
			name:      "NotFound",
			key:       "notfound",
			setup:     func(_ *Client, _ context.Context) {},
			want:      nil,
			expectErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t)
			ctx := context.Background()
			tc.setup(c, ctx)

			got, err := c.GetRaw(ctx, tc.key)
			if tc.expectErr && err == nil {
				t.Errorf("expected error for GetRaw, got nil")
			}

			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error for GetRaw: %v", err)
			}

			if !tc.expectErr && string(got) != string(tc.want) {
				t.Errorf("GetRaw = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type testCase struct {
		name      string
		key       string
		setup     func(c *Client, ctx context.Context)
		expectErr bool
	}

	tests := []testCase{
		{
			name: "Valid",
			key:  "foo",
			setup: func(c *Client, ctx context.Context) {
				_ = c.SetRaw(ctx, "foo", []byte("bar"))
			},
			expectErr: false,
		},
		{
			name:      "NotFound",
			key:       "notfound",
			setup:     func(_ *Client, _ context.Context) {},
			expectErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t)
			ctx := context.Background()
			tc.setup(c, ctx)

			err := c.Delete(ctx, tc.key)
			if tc.expectErr && err == nil {
				t.Errorf("expected error for Delete, got nil")
			}

			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error for Delete: %v", err)
			}
		})
	}
}

func TestList(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	_ = c.SetRaw(ctx, "prefix/a", []byte("A"))
	_ = c.SetRaw(ctx, "prefix/b", []byte("B"))
	keys, err := c.List(ctx, "prefix/")
	require.NoError(t, err)

	if len(keys) < 2 {
		t.Errorf("List = %v, want at least 2 keys", keys)
	}
}

func TestClose(t *testing.T) {
	c := newTestClient(t)
	require.NoError(t, c.Close())
}

func TestBatchSetRaw(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		kv        map[string][]byte
		expectErr bool
	}{
		{"Valid", map[string][]byte{"batch:key1": []byte("v1"), "batch:key2": []byte("v2")}, false},
		{"EmptyBatch", map[string][]byte{}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := c.BatchSetRaw(ctx, tc.kv)
			if tc.expectErr && err == nil {
				t.Errorf("expected error for BatchSetRaw, got nil")
			}

			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error for BatchSetRaw: %v", err)
			}
		})
	}
}

func TestBatchGetRaw(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// Set values for valid test
	_ = c.SetRaw(ctx, "batchget:key1", []byte("v1"))
	_ = c.SetRaw(ctx, "batchget:key2", []byte("v2"))

	tests := []struct {
		name      string
		keys      []string
		want      map[string][]byte
		expectErr bool
	}{
		{
			name:      "Valid",
			keys:      []string{"batchget:key1", "batchget:key2"},
			want:      map[string][]byte{"batchget:key1": []byte("v1"), "batchget:key2": []byte("v2")},
			expectErr: false,
		},
		{
			name:      "EmptyBatch",
			keys:      []string{},
			want:      nil,
			expectErr: false,
		},
		{
			name:      "PartialNotFound",
			keys:      []string{"batchget:key1", "batchget:notfound"},
			want:      map[string][]byte{"batchget:key1": []byte("v1"), "batchget:notfound": nil},
			expectErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.BatchGetRaw(ctx, tc.keys)
			if tc.expectErr && err == nil {
				t.Errorf("expected error for BatchGetRaw, got nil")
			}

			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error for BatchGetRaw: %v", err)
			}

			if tc.expectErr {
				return
			}

			for k, v := range tc.want {
				require.Equal(t, v, got[k])
			}
		})
	}
}

func TestBatchDelete(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// Set values for valid test
	_ = c.SetRaw(ctx, "batchdel:key1", []byte("v1"))
	_ = c.SetRaw(ctx, "batchdel:key2", []byte("v2"))

	tests := []struct {
		name      string
		keys      []string
		expectErr bool
	}{
		{"Valid", []string{"batchdel:key1", "batchdel:key2"}, false},
		{"EmptyBatch", []string{}, false},
		{"PartialNotFound", []string{"batchdel:key1", "batchdel:notfound"}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := c.BatchDelete(ctx, tc.keys)
			if tc.expectErr && err == nil {
				t.Errorf("expected error for BatchDelete, got nil")
			}

			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error for BatchDelete: %v", err)
			}
		})
	}
}

func TestHealth(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.Health(ctx))
}

func TestHealth_Error(t *testing.T) {
	// Simulate an uninitialized client (nil Consul client)
	c := &Client{cli: nil}
	ctx := context.Background()
	err := c.Health(ctx)
	require.Error(t, err)
}
