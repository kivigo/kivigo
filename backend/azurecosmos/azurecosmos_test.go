package azurecosmos

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kivigo/kivigo/pkg/errs"
)

var testAzureCosmos *container

func TestMain(m *testing.M) {
	var err error

	testAzureCosmos, err = start(&testing.T{})
	if err != nil {
		panic(err)
	}

	code := m.Run()

	_ = testAzureCosmos.Stop(context.Background())

	os.Exit(code)
}

func newTestClient(t *testing.T) Client {
	t.Helper()

	opt := DefaultOptions()
	opt.Endpoint = testAzureCosmos.endpoint
	opt.Database = "test_kivigo"
	// Use unique container for each test to avoid conflicts
	opt.Container = fmt.Sprintf("test_%s_%d", strings.ReplaceAll(t.Name(), "/", "_"), time.Now().UnixNano())

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

func TestAzureCosmos_SetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	tests := []testCase{
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
			key:     "foo",
			value:   []byte(""),
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

func TestAzureCosmos_GetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	// First set a value
	key := "test_get"
	value := []byte("test_value")
	err := c.SetRaw(context.Background(), key, value)
	require.NoError(t, err)

	tests := []struct {
		name     string
		key      string
		expected []byte
		wantErr  bool
	}{
		{
			name:     "ExistingKey",
			key:      key,
			expected: value,
			wantErr:  false,
		},
		{
			name:     "NonExistentKey",
			key:      "nonexistent",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "EmptyKey",
			key:      "",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.GetRaw(context.Background(), tt.key)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestAzureCosmos_Delete(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	// First set a value
	key := "test_delete"
	value := []byte("test_value")
	err := c.SetRaw(context.Background(), key, value)
	require.NoError(t, err)

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "ExistingKey",
			key:     key,
			wantErr: false,
		},
		{
			name:    "NonExistentKey",
			key:     "nonexistent",
			wantErr: false, // Should not error for non-existent key
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
	_, err = c.GetRaw(context.Background(), key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestAzureCosmos_List(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	// Set up test data
	testData := map[string][]byte{
		"prefix_1": []byte("value1"),
		"prefix_2": []byte("value2"),
		"other_1":  []byte("value3"),
		"prefix_3": []byte("value4"),
	}

	for k, v := range testData {
		err := c.SetRaw(context.Background(), k, v)
		require.NoError(t, err)
	}

	// Add a small delay to ensure items are available for querying (eventual consistency)
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name     string
		prefix   string
		expected []string
	}{
		{
			name:     "AllKeys",
			prefix:   "",
			expected: []string{"prefix_1", "prefix_2", "other_1", "prefix_3"},
		},
		{
			name:     "PrefixFilter",
			prefix:   "prefix_",
			expected: []string{"prefix_1", "prefix_2", "prefix_3"},
		},
		{
			name:     "NoMatches",
			prefix:   "nonexistent_",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.List(context.Background(), tt.prefix)
			require.NoError(t, err)

			if tt.name == "AllKeys" {
				require.Len(t, got, len(tt.expected))
				for _, expected := range tt.expected {
					require.Contains(t, got, expected)
				}
			} else {
				require.ElementsMatch(t, tt.expected, got)
			}
		})
	}
}

func TestAzureCosmos_Health(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	err := c.Health(context.Background())
	require.NoError(t, err)
}

func TestAzureCosmos_BatchGetRaw(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	// Set up test data
	testData := map[string][]byte{
		"batch_key1": []byte("value1"),
		"batch_key2": []byte("value2"),
		"batch_key3": []byte("value3"),
	}

	for k, v := range testData {
		err := c.SetRaw(context.Background(), k, v)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		keys    []string
		wantErr bool
	}{
		{
			name:    "ValidKeys",
			keys:    []string{"batch_key1", "batch_key2", "batch_key3"},
			wantErr: false,
		},
		{
			name:    "MixedKeys",
			keys:    []string{"batch_key1", "nonexistent", "batch_key2"},
			wantErr: false,
		},
		{
			name:    "EmptyBatch",
			keys:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.BatchGetRaw(context.Background(), tt.keys)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tt.keys), len(got))

			for _, key := range tt.keys {
				if expectedValue, exists := testData[key]; exists {
					require.Equal(t, expectedValue, got[key])
				} else {
					require.Nil(t, got[key]) // Non-existent keys should have nil values
				}
			}
		})
	}
}

func TestAzureCosmos_BatchSetRaw(t *testing.T) {
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
				"batch_set1": []byte("value1"),
				"batch_set2": []byte("value2"),
				"batch_set3": []byte("value3"),
			},
			wantErr: false,
		},
		{
			name:    "EmptyBatch",
			kv:      map[string][]byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.BatchSetRaw(context.Background(), tt.kv)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify all keys were set
			for key, expectedValue := range tt.kv {
				got, err := c.GetRaw(context.Background(), key)
				require.NoError(t, err)
				require.Equal(t, expectedValue, got)
			}
		})
	}
}

func TestAzureCosmos_BatchDelete(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	// Set up test data
	testData := map[string][]byte{
		"batch_del1": []byte("value1"),
		"batch_del2": []byte("value2"),
		"batch_del3": []byte("value3"),
	}

	for k, v := range testData {
		err := c.SetRaw(context.Background(), k, v)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		keys    []string
		wantErr bool
	}{
		{
			name:    "ValidKeys",
			keys:    []string{"batch_del1", "batch_del2"},
			wantErr: false,
		},
		{
			name:    "MixedKeys",
			keys:    []string{"batch_del3", "nonexistent"},
			wantErr: false,
		},
		{
			name:    "EmptyBatch",
			keys:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.BatchDelete(context.Background(), tt.keys)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify all keys were deleted
			for _, key := range tt.keys {
				if _, exists := testData[key]; exists {
					_, err := c.GetRaw(context.Background(), key)
					require.Error(t, err)
					require.True(t,
						errors.Is(err, errs.ErrNotFound) ||
							strings.Contains(fmt.Sprintf("%v", err), "not found"),
						"Expected key %s to be deleted", key)
				}
			}
		})
	}
}
