package memcached

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kivigo/kivigo/pkg/errs"
)

var testMemcached *container

func TestMain(m *testing.M) {
	var err error

	testMemcached, err = start(&testing.T{})
	if err != nil {
		panic(err)
	}

	code := m.Run()

	if testMemcached != nil {
		_ = testMemcached.Stop(context.Background())
	}

	os.Exit(code)
}

func newTestClient(t *testing.T) Client {
	t.Helper()

	opt := DefaultOptions()
	opt.Servers = []string{testMemcached.addr}
	opt.Timeout = 5 * time.Second

	c, err := New(opt)
	require.NoError(t, err)

	return c
}

func TestNew(t *testing.T) {
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
			name: "CustomOptions",
			opt: Option{
				Servers: []string{"localhost:11211", "localhost:11212"},
				Timeout: 200 * time.Millisecond,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.opt)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NoError(t, client.Close())
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	opt := DefaultOptions()
	require.Equal(t, []string{"localhost:11211"}, opt.Servers)
	require.Equal(t, 100*time.Millisecond, opt.Timeout)
}

func TestNewOptions(t *testing.T) {
	opt := NewOptions()
	require.Empty(t, opt.Servers)
	require.Zero(t, opt.Timeout)
}

func TestSetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	tests := []struct {
		name    string
		key     string
		value   []byte
		wantErr bool
	}{
		{
			name:    "Valid",
			key:     "foo",
			value:   []byte("bar"),
			wantErr: false,
		},
		{
			name:    "EmptyKey",
			key:     "",
			value:   []byte("bar"),
			wantErr: true,
		},
		{
			name:    "EmptyValue",
			key:     "empty",
			value:   []byte(""),
			wantErr: false,
		},
		{
			name:    "NilValue",
			key:     "nil",
			value:   nil,
			wantErr: false,
		},
	}

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

func TestGetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	// Set up test data
	testKey := "test-get"
	testValue := []byte("test-value")
	err := c.SetRaw(context.Background(), testKey, testValue)
	require.NoError(t, err)

	tests := []struct {
		name      string
		key       string
		wantValue []byte
		wantErr   bool
	}{
		{
			name:      "ExistingKey",
			key:       testKey,
			wantValue: testValue,
			wantErr:   false,
		},
		{
			name:      "NonExistentKey",
			key:       "non-existent",
			wantValue: nil,
			wantErr:   true,
		},
		{
			name:      "EmptyKey",
			key:       "",
			wantValue: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := c.GetRaw(context.Background(), tt.key)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantValue, value)
			}
		})
	}
}

func TestList(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	// Memcached doesn't support listing keys
	keys, err := c.List(context.Background(), "prefix")
	require.Error(t, err)
	require.Equal(t, errs.ErrOperationNotSupported, err)
	require.Nil(t, keys)
}

func TestDelete(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	// Set up test data
	testKey := "test-delete"
	testValue := []byte("test-value")
	err := c.SetRaw(context.Background(), testKey, testValue)
	require.NoError(t, err)

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "ExistingKey",
			key:     testKey,
			wantErr: false,
		},
		{
			name:    "NonExistentKey",
			key:     "non-existent",
			wantErr: false, // Memcached treats missing key as success
		},
		{
			name:    "EmptyKey",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Delete(context.Background(), tt.key)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	// Verify the key was actually deleted
	_, err = c.GetRaw(context.Background(), testKey)
	require.Error(t, err)
	require.Equal(t, errs.ErrNotFound, err)
}

func TestHealth(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	err := c.Health(context.Background())
	require.NoError(t, err)
}

func TestBatchSetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	tests := []struct {
		name    string
		kv      map[string][]byte
		wantErr bool
	}{
		{
			name: "ValidBatch",
			kv: map[string][]byte{
				"batch1": []byte("value1"),
				"batch2": []byte("value2"),
				"batch3": []byte("value3"),
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
			err := c.BatchSetRaw(context.Background(), tt.kv)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify all keys were set
				for key, expectedValue := range tt.kv {
					value, err := c.GetRaw(context.Background(), key)
					require.NoError(t, err)
					require.Equal(t, expectedValue, value)
				}
			}
		})
	}
}

func TestBatchGetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	// Set up test data
	testData := map[string][]byte{
		"bget1": []byte("value1"),
		"bget2": []byte("value2"),
		"bget3": []byte("value3"),
	}

	for key, value := range testData {
		err := c.SetRaw(context.Background(), key, value)
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		keys      []string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "ExistingKeys",
			keys:      []string{"bget1", "bget2", "bget3"},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "MixedKeys",
			keys:      []string{"bget1", "non-existent", "bget2"},
			wantCount: 3, // Should return all keys with nil for non-existent
			wantErr:   false,
		},
		{
			name:      "EmptyKeys",
			keys:      []string{},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := c.BatchGetRaw(context.Background(), tt.keys)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, results, tt.wantCount)

				// Verify existing keys have correct values
				for _, key := range tt.keys {
					result, exists := results[key]
					require.True(t, exists)

					if expectedValue, hasExpected := testData[key]; hasExpected {
						require.Equal(t, expectedValue, result)
					} else {
						require.Nil(t, result) // Non-existent key should have nil value
					}
				}
			}
		})
	}
}

func TestBatchDelete(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	// Set up test data
	testKeys := []string{"bdel1", "bdel2", "bdel3"}
	for _, key := range testKeys {
		err := c.SetRaw(context.Background(), key, []byte("value"))
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		keys    []string
		wantErr bool
	}{
		{
			name:    "ExistingKeys",
			keys:    testKeys,
			wantErr: false,
		},
		{
			name:    "NonExistentKeys",
			keys:    []string{"non-existent1", "non-existent2"},
			wantErr: false, // Should not error on non-existent keys
		},
		{
			name:    "EmptyKeys",
			keys:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.BatchDelete(context.Background(), tt.keys)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	// Verify keys were actually deleted
	for _, key := range testKeys {
		_, err := c.GetRaw(context.Background(), key)
		require.Error(t, err)
		require.Equal(t, errs.ErrNotFound, err)
	}
}
