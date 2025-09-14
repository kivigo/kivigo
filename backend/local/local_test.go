package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kivigo/kivigo/pkg/errs"
)

func newTestClient(t *testing.T) Client {
	t.Helper()

	tmpDir := t.TempDir()
	opt := DefaultOptions()
	opt.Path = tmpDir
	opt.FileName = "test.db"

	client, err := New(opt)
	require.NoError(t, err)

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		opt     Option
		wantErr bool
	}{
		{
			name:    "DefaultOptions",
			opt:     DefaultOptions(),
			wantErr: false,
		},
		{
			name: "CustomPath",
			opt: Option{
				Path:     tmpDir,
				FileName: "custom.db",
			},
			wantErr: false,
		},
		{
			name: "PathWithoutTrailingSlash",
			opt: Option{
				Path:     tmpDir,
				FileName: "test.db",
			},
			wantErr: false,
		},
		{
			name: "InvalidPath",
			opt: Option{
				Path:     "/nonexistent/invalid/path/that/does/not/exist",
				FileName: "test.db",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opt.Path != "./" && !tt.wantErr {
				// Use temp dir for custom paths
				tt.opt.Path = tmpDir
			}

			client, err := New(tt.opt)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NoError(t, client.Close())

			// Clean up the created file
			if tt.opt.Path != "./" {
				dbPath := filepath.Join(tt.opt.Path, tt.opt.FileName)
				os.Remove(dbPath)
			}
		})
	}
}

func TestNewOptions(t *testing.T) {
	opt := NewOptions()
	require.Equal(t, Option{}, opt)
}

func TestDefaultOptions(t *testing.T) {
	opt := DefaultOptions()
	require.Equal(t, "./", opt.Path)
	require.Equal(t, "kivigo.db", opt.FileName)
}

func TestSetRaw(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   []byte
		wantErr bool
	}{
		{
			name:    "Valid",
			key:     "test-key",
			value:   []byte("test-value"),
			wantErr: false,
		},
		{
			name:    "EmptyKey",
			key:     "",
			value:   []byte("test-value"),
			wantErr: true,
		},
		{
			name:    "EmptyValue",
			key:     "test-key",
			value:   []byte(""),
			wantErr: false,
		},
		{
			name:    "NilValue",
			key:     "test-key",
			value:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t)
			ctx := context.Background()

			err := client.SetRaw(ctx, tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				if tt.key == "" {
					require.Equal(t, errs.ErrEmptyKey, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetRaw(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		setup     func(client Client, ctx context.Context)
		want      []byte
		expectErr bool
	}{
		{
			name: "Valid",
			key:  "test-key",
			setup: func(client Client, ctx context.Context) {
				_ = client.SetRaw(ctx, "test-key", []byte("test-value"))
			},
			want:      []byte("test-value"),
			expectErr: false,
		},
		{
			name:      "NotFound",
			key:       "missing-key",
			setup:     func(client Client, ctx context.Context) {},
			want:      nil,
			expectErr: true,
		},
		{
			name: "EmptyValue",
			key:  "empty-key",
			setup: func(client Client, ctx context.Context) {
				_ = client.SetRaw(ctx, "empty-key", []byte(""))
			},
			want:      []byte(""),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t)
			ctx := context.Background()
			tt.setup(client, ctx)

			got, err := client.GetRaw(ctx, tt.key)
			if tt.expectErr {
				require.Error(t, err)
				require.Equal(t, errs.ErrNotFound, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetRawBugCheck(t *testing.T) {
	// Specific test to verify the GetRaw implementation works correctly
	// This tests for the potential bug where value assignment inside closure
	// might not be returned properly
	client := newTestClient(t)
	ctx := context.Background()

	// Test with different value sizes to ensure the implementation is correct
	testValues := [][]byte{
		[]byte("small"),
		[]byte("this is a longer value that should test the implementation"),
		[]byte(""), // empty value
		[]byte("single"),
	}

	for i, val := range testValues {
		key := fmt.Sprintf("test-%d", i)

		// Set the value
		err := client.SetRaw(ctx, key, val)
		require.NoError(t, err)

		// Get the value back
		retrieved, err := client.GetRaw(ctx, key)
		require.NoError(t, err)
		require.Equal(t, val, retrieved, "Value mismatch for key %s", key)
		require.Equal(t, len(val), len(retrieved), "Length mismatch for key %s", key)
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		setup   func(client Client, ctx context.Context)
		wantErr bool
	}{
		{
			name: "Valid",
			key:  "test-key",
			setup: func(client Client, ctx context.Context) {
				_ = client.SetRaw(ctx, "test-key", []byte("test-value"))
			},
			wantErr: false,
		},
		{
			name:    "EmptyKey",
			key:     "",
			setup:   func(client Client, ctx context.Context) {},
			wantErr: true,
		},
		{
			name:    "NonExistentKey",
			key:     "missing-key",
			setup:   func(client Client, ctx context.Context) {},
			wantErr: false, // BoltDB doesn't error on deleting non-existent keys
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t)
			ctx := context.Background()
			tt.setup(client, ctx)

			err := client.Delete(ctx, tt.key)
			if tt.wantErr {
				require.Error(t, err)
				if tt.key == "" {
					require.Equal(t, errs.ErrEmptyKey, err)
				}
			} else {
				require.NoError(t, err)

				// Verify key was deleted
				if tt.key != "" {
					_, err := client.GetRaw(ctx, tt.key)
					require.Equal(t, errs.ErrNotFound, err)
				}
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		setup     func(client Client, ctx context.Context)
		wantKeys  []string
		expectErr bool
	}{
		{
			name:   "ValidPrefix",
			prefix: "user:",
			setup: func(client Client, ctx context.Context) {
				_ = client.SetRaw(ctx, "user:1", []byte("user1"))
				_ = client.SetRaw(ctx, "user:2", []byte("user2"))
				_ = client.SetRaw(ctx, "admin:1", []byte("admin1"))
			},
			wantKeys:  []string{"user:1", "user:2"},
			expectErr: false,
		},
		{
			name:   "EmptyPrefix",
			prefix: "",
			setup: func(client Client, ctx context.Context) {
				_ = client.SetRaw(ctx, "key1", []byte("value1"))
				_ = client.SetRaw(ctx, "key2", []byte("value2"))
			},
			wantKeys:  []string{"key1", "key2"},
			expectErr: false,
		},
		{
			name:      "NoMatches",
			prefix:    "nonexistent:",
			setup:     func(client Client, ctx context.Context) {},
			wantKeys:  []string{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t)
			ctx := context.Background()
			tt.setup(client, ctx)

			keys, err := client.List(ctx, tt.prefix)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, tt.wantKeys, keys)
			}
		})
	}
}

func TestHealth(t *testing.T) {
	t.Run("Healthy", func(t *testing.T) {
		client := newTestClient(t)
		ctx := context.Background()

		err := client.Health(ctx)
		require.NoError(t, err)
	})

	t.Run("Uninitialized", func(t *testing.T) {
		client := Client{c: nil}
		ctx := context.Background()

		err := client.Health(ctx)
		require.Error(t, err)
		require.Equal(t, errs.ErrClientNotInitialized, err)
	})
}

func TestBatchSetRaw(t *testing.T) {
	tests := []struct {
		name    string
		kv      map[string][]byte
		wantErr bool
	}{
		{
			name: "Valid",
			kv: map[string][]byte{
				"key1": []byte("value1"),
				"key2": []byte("value2"),
			},
			wantErr: false,
		},
		{
			name:    "EmptyBatch",
			kv:      map[string][]byte{},
			wantErr: false,
		},
		{
			name:    "NilBatch",
			kv:      nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t)
			ctx := context.Background()

			err := client.BatchSetRaw(ctx, tt.kv)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify all keys were set
				for k, v := range tt.kv {
					got, err := client.GetRaw(ctx, k)
					require.NoError(t, err)
					require.Equal(t, v, got)
				}
			}
		})
	}
}

func TestBatchGetRaw(t *testing.T) {
	tests := []struct {
		name      string
		keys      []string
		setup     func(client Client, ctx context.Context)
		want      map[string][]byte
		expectErr bool
	}{
		{
			name: "Valid",
			keys: []string{"key1", "key2"},
			setup: func(client Client, ctx context.Context) {
				_ = client.SetRaw(ctx, "key1", []byte("value1"))
				_ = client.SetRaw(ctx, "key2", []byte("value2"))
			},
			want: map[string][]byte{
				"key1": []byte("value1"),
				"key2": []byte("value2"),
			},
			expectErr: false,
		},
		{
			name: "PartialFound",
			keys: []string{"key1", "missing"},
			setup: func(client Client, ctx context.Context) {
				_ = client.SetRaw(ctx, "key1", []byte("value1"))
			},
			want: map[string][]byte{
				"key1":    []byte("value1"),
				"missing": nil,
			},
			expectErr: false,
		},
		{
			name:      "EmptyKeys",
			keys:      []string{},
			setup:     func(client Client, ctx context.Context) {},
			want:      map[string][]byte{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t)
			ctx := context.Background()
			tt.setup(client, ctx)

			got, err := client.BatchGetRaw(ctx, tt.keys)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBatchDelete(t *testing.T) {
	tests := []struct {
		name    string
		keys    []string
		setup   func(client Client, ctx context.Context)
		wantErr bool
	}{
		{
			name: "Valid",
			keys: []string{"key1", "key2"},
			setup: func(client Client, ctx context.Context) {
				_ = client.SetRaw(ctx, "key1", []byte("value1"))
				_ = client.SetRaw(ctx, "key2", []byte("value2"))
			},
			wantErr: false,
		},
		{
			name: "PartialExists",
			keys: []string{"key1", "missing"},
			setup: func(client Client, ctx context.Context) {
				_ = client.SetRaw(ctx, "key1", []byte("value1"))
			},
			wantErr: false,
		},
		{
			name:    "EmptyKeys",
			keys:    []string{},
			setup:   func(client Client, ctx context.Context) {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t)
			ctx := context.Background()
			tt.setup(client, ctx)

			err := client.BatchDelete(ctx, tt.keys)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify keys were deleted
				for _, k := range tt.keys {
					_, err := client.GetRaw(ctx, k)
					require.Equal(t, errs.ErrNotFound, err)
				}
			}
		})
	}
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	opt := DefaultOptions()
	opt.Path = tmpDir
	opt.FileName = "test.db"

	client, err := New(opt)
	require.NoError(t, err)

	err = client.Close()
	require.NoError(t, err)

	// Verify database is closed by trying to use it
	ctx := context.Background()
	_, err = client.GetRaw(ctx, "test")
	require.Error(t, err) // Should fail because DB is closed
}
