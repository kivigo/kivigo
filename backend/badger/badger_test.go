package badger

import (
	"context"
	"testing"
)

func newTestClient(t *testing.T) Client {
	t.Helper()

	client, err := New(DefaultOptions("badger").WithDir(t.TempDir() + "/key").WithValueDir(t.TempDir() + "/value/"))
	if err != nil {
		t.Fatalf("failed to create badger client: %v", err)
	}

	t.Cleanup(func() {
		// Remove the database file after tests
		if err := client.Close(); err != nil {
			t.Errorf("failed to close badger client: %v", err)
		}
	})

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
		setup     func(c Client, ctx context.Context)
		want      []byte
		expectErr bool
	}

	tests := []testCase{
		{
			name: "Valid",
			key:  "foo",
			setup: func(c Client, ctx context.Context) {
				_ = c.SetRaw(ctx, "foo", []byte("bar"))
			},
			want:      []byte("bar"),
			expectErr: false,
		},
		{
			name:      "EmptyKey",
			key:       "",
			setup:     func(_ Client, _ context.Context) {},
			want:      nil,
			expectErr: true,
		},
		{
			name:      "NotFound",
			key:       "notfound",
			setup:     func(_ Client, _ context.Context) {},
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
		setup     func(c Client, ctx context.Context)
		expectErr bool
	}

	tests := []testCase{
		{
			name: "Valid",
			key:  "foo",
			setup: func(c Client, ctx context.Context) {
				_ = c.SetRaw(ctx, "foo", []byte("bar"))
			},
			expectErr: false,
		},
		{
			name:      "EmptyKey",
			key:       "",
			setup:     func(_ Client, _ context.Context) {},
			expectErr: true,
		},
		{
			name:      "NotFound",
			key:       "notfound",
			setup:     func(_ Client, _ context.Context) {},
			expectErr: false, // Deleting a non-existent key should not error
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

func TestList(t *testing.T) { //nolint:cyclop
	type testCase struct {
		name      string
		prefix    string
		keys      []string
		expectErr bool
	}

	tests := []testCase{
		{"Valid", "pfx:", []string{"a", "b", "c"}, false},
		{"EmptyPrefix", "", []string{"a"}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t)
			ctx := context.Background()

			for _, k := range tc.keys {
				if tc.prefix != "" {
					if err := c.SetRaw(ctx, tc.prefix+k, []byte(k)); err != nil {
						t.Fatalf("SetRaw failed: %v", err)
					}
				}
			}

			listed, err := c.List(ctx, tc.prefix)
			if tc.expectErr && err == nil {
				t.Errorf("expected error for List, got nil")
			}

			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error for List: %v", err)
			}

			if tc.expectErr {
				return
			}

			if len(listed) != len(tc.keys) {
				t.Errorf("List = %v, want %v", listed, tc.keys)
			}
		})
	}
}

func TestBatchSetRaw(t *testing.T) {
	type testCase struct {
		name      string
		kv        map[string][]byte
		expectErr bool
	}

	tests := []testCase{
		{"Valid", map[string][]byte{"b1": []byte("v1"), "b2": []byte("v2")}, false},
		{"EmptyBatch", map[string][]byte{}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t)
			ctx := context.Background()

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

func TestBatchGetRaw(t *testing.T) { //nolint:funlen
	type testCase struct {
		name      string
		setup     func(c Client, ctx context.Context)
		getKeys   []string
		want      map[string][]byte
		expectErr bool
	}

	tests := []testCase{
		{
			name: "Valid",
			setup: func(c Client, ctx context.Context) {
				_ = c.BatchSetRaw(ctx, map[string][]byte{"b1": []byte("v1"), "b2": []byte("v2")})
			},
			getKeys:   []string{"b1", "b2"},
			want:      map[string][]byte{"b1": []byte("v1"), "b2": []byte("v2")},
			expectErr: false,
		},
		{
			name:      "EmptyBatch",
			setup:     func(_ Client, _ context.Context) {},
			getKeys:   []string{},
			want:      map[string][]byte{},
			expectErr: true,
		},
		{
			name: "PartialNotFound",
			setup: func(c Client, ctx context.Context) {
				_ = c.SetRaw(ctx, "b1", []byte("v1"))
			},
			getKeys:   []string{"b1", "b2"},
			want:      map[string][]byte{"b1": []byte("v1"), "b2": nil},
			expectErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t)
			ctx := context.Background()
			tc.setup(c, ctx)

			got, err := c.BatchGetRaw(ctx, tc.getKeys)
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
				if string(got[k]) != string(v) {
					t.Errorf("BatchGetRaw[%s] = %s, want %s", k, got[k], v)
				}
			}
		})
	}
}

func TestBatchDelete(t *testing.T) {
	type testCase struct {
		name      string
		setup     func(c Client, ctx context.Context)
		delKeys   []string
		expectErr bool
	}

	tests := []testCase{
		{
			name: "Valid",
			setup: func(c Client, ctx context.Context) {
				_ = c.BatchSetRaw(ctx, map[string][]byte{"b1": []byte("v1"), "b2": []byte("v2")})
			},
			delKeys:   []string{"b1", "b2"},
			expectErr: false,
		},
		{
			name:      "EmptyBatch",
			setup:     func(_ Client, _ context.Context) {},
			delKeys:   []string{},
			expectErr: true,
		},
		{
			name: "PartialNotFound",
			setup: func(c Client, ctx context.Context) {
				_ = c.SetRaw(ctx, "b1", []byte("v1"))
			},
			delKeys:   []string{"b1", "b2"},
			expectErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t)
			ctx := context.Background()
			tc.setup(c, ctx)

			err := c.BatchDelete(ctx, tc.delKeys)
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
	if err := c.Health(context.Background()); err != nil {
		t.Errorf("Health failed: %v", err)
	}
}
