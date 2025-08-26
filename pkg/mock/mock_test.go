package mock

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/azrod/kivigo/pkg/errs"
)

func newTestMockKV() *MockKV {
	return &MockKV{
		Data: make(map[string][]byte),
	}
}

func TestMockKV_SetRaw(t *testing.T) {
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
			m := newTestMockKV()
			ctx := context.Background()

			err := m.SetRaw(ctx, tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, errs.ErrEmptyKey, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.value, m.Data[tt.key])
			}
		})
	}
}

func getGetRawTestCases() []struct {
	name      string
	key       string
	setup     func(m *MockKV)
	want      []byte
	expectErr bool
} {
	return []struct {
		name      string
		key       string
		setup     func(m *MockKV)
		want      []byte
		expectErr bool
	}{
		{
			name: "Valid",
			key:  "test-key",
			setup: func(m *MockKV) {
				m.Data["test-key"] = []byte("test-value")
			},
			want:      []byte("test-value"),
			expectErr: false,
		},
		{
			name:      "EmptyKey",
			key:       "",
			setup:     func(m *MockKV) {},
			want:      nil,
			expectErr: true,
		},
		{
			name:      "NotFound",
			key:       "missing-key",
			setup:     func(m *MockKV) {},
			want:      nil,
			expectErr: true,
		},
		{
			name: "EmptyValue",
			key:  "empty-key",
			setup: func(m *MockKV) {
				m.Data["empty-key"] = []byte("")
			},
			want:      []byte(""),
			expectErr: false,
		},
		{
			name: "NilValue",
			key:  "nil-key",
			setup: func(m *MockKV) {
				m.Data["nil-key"] = nil
			},
			want:      nil,
			expectErr: false,
		},
	}
}

func runGetRawSubtest(t *testing.T, tt struct {
	name      string
	key       string
	setup     func(m *MockKV)
	want      []byte
	expectErr bool
},
) {
	t.Helper()

	m := newTestMockKV()
	tt.setup(m)

	ctx := context.Background()

	got, err := m.GetRaw(ctx, tt.key)
	if tt.expectErr {
		require.Error(t, err)

		if tt.key == "" {
			require.Equal(t, errs.ErrEmptyKey, err)
		} else {
			require.Equal(t, errs.ErrKeyNotFound, err)
		}
	} else {
		require.NoError(t, err)
		require.Equal(t, tt.want, got)
	}
}

func TestMockKV_GetRaw(t *testing.T) {
	tests := getGetRawTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runGetRawSubtest(t, tt)
		})
	}
}

func TestMockKV_Delete(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		setup   func(m *MockKV)
		wantErr bool
	}{
		{
			name: "Valid",
			key:  "test-key",
			setup: func(m *MockKV) {
				m.Data["test-key"] = []byte("test-value")
			},
			wantErr: false,
		},
		{
			name:    "EmptyKey",
			key:     "",
			setup:   func(m *MockKV) {},
			wantErr: true,
		},
		{
			name:    "NonExistentKey",
			key:     "missing-key",
			setup:   func(m *MockKV) {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestMockKV()
			tt.setup(m)

			ctx := context.Background()

			err := m.Delete(ctx, tt.key)
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, errs.ErrEmptyKey, err)
			} else {
				require.NoError(t, err)
				// Verify key was removed
				_, exists := m.Data[tt.key]
				require.False(t, exists)
			}
		})
	}
}

func TestMockKV_List(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		setup    func(m *MockKV)
		wantKeys []string
	}{
		{
			name:   "ValidPrefix",
			prefix: "user:",
			setup: func(m *MockKV) {
				m.Data["user:1"] = []byte("user1")
				m.Data["user:2"] = []byte("user2")
				m.Data["admin:1"] = []byte("admin1")
			},
			wantKeys: []string{"user:1", "user:2"},
		},
		{
			name:   "EmptyPrefix",
			prefix: "",
			setup: func(m *MockKV) {
				m.Data["key1"] = []byte("value1")
				m.Data["key2"] = []byte("value2")
			},
			wantKeys: []string{"key1", "key2"},
		},
		{
			name:     "NoMatches",
			prefix:   "nonexistent:",
			setup:    func(m *MockKV) {},
			wantKeys: []string{},
		},
		{
			name:   "PartialPrefix",
			prefix: "test",
			setup: func(m *MockKV) {
				m.Data["test-key"] = []byte("value1")
				m.Data["testing"] = []byte("value2")
				m.Data["other"] = []byte("value3")
			},
			wantKeys: []string{"test-key", "testing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestMockKV()
			tt.setup(m)

			ctx := context.Background()

			keys, err := m.List(ctx, tt.prefix)
			require.NoError(t, err)
			require.ElementsMatch(t, tt.wantKeys, keys)
		})
	}
}

func TestMockKV_Close(t *testing.T) {
	m := newTestMockKV()
	err := m.Close()
	require.NoError(t, err)
}

func TestMockKV_Health(t *testing.T) {
	m := newTestMockKV()
	ctx := context.Background()

	err := m.Health(ctx)
	require.NoError(t, err)
}

func TestMockKV_BatchSetRaw(t *testing.T) {
	tests := []struct {
		name string
		kv   map[string][]byte
	}{
		{
			name: "Valid",
			kv: map[string][]byte{
				"key1": []byte("value1"),
				"key2": []byte("value2"),
			},
		},
		{
			name: "EmptyBatch",
			kv:   map[string][]byte{},
		},
		{
			name: "NilBatch",
			kv:   nil,
		},
		{
			name: "WithNilValues",
			kv: map[string][]byte{
				"key1": []byte("value1"),
				"key2": nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestMockKV()
			ctx := context.Background()

			err := m.BatchSetRaw(ctx, tt.kv)
			require.NoError(t, err)

			// Verify all keys were set
			for k, v := range tt.kv {
				require.Equal(t, v, m.Data[k])
			}
		})
	}
}

func getBatchGetRawBasicTestCases() []struct {
	name      string
	keys      []string
	setup     func(m *MockKV)
	want      map[string][]byte
	expectErr bool
} {
	return []struct {
		name      string
		keys      []string
		setup     func(m *MockKV)
		want      map[string][]byte
		expectErr bool
	}{
		{
			name: "Valid",
			keys: []string{"key1", "key2"},
			setup: func(m *MockKV) {
				m.Data["key1"] = []byte("value1")
				m.Data["key2"] = []byte("value2")
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
			setup: func(m *MockKV) {
				m.Data["key1"] = []byte("value1")
			},
			want: map[string][]byte{
				"key1": []byte("value1"),
			},
			expectErr: false,
		},
	}
}

func getBatchGetRawEdgeTestCases() []struct {
	name      string
	keys      []string
	setup     func(m *MockKV)
	want      map[string][]byte
	expectErr bool
} {
	return []struct {
		name      string
		keys      []string
		setup     func(m *MockKV)
		want      map[string][]byte
		expectErr bool
	}{
		{
			name:      "EmptyKeys",
			keys:      []string{},
			setup:     func(m *MockKV) {},
			want:      map[string][]byte{},
			expectErr: false,
		},
		{
			name:      "NilKeys",
			keys:      nil,
			setup:     func(m *MockKV) {},
			want:      map[string][]byte{},
			expectErr: false,
		},
		{
			name: "AllMissing",
			keys: []string{"missing1", "missing2"},
			setup: func(m *MockKV) {
				m.Data["other"] = []byte("value")
			},
			want:      map[string][]byte{},
			expectErr: false,
		},
	}
}

func getBatchGetRawTestCases() []struct {
	name      string
	keys      []string
	setup     func(m *MockKV)
	want      map[string][]byte
	expectErr bool
} {
	basic := getBatchGetRawBasicTestCases()
	edge := getBatchGetRawEdgeTestCases()

	return append(basic, edge...)
}

func runBatchGetRawSubtest(t *testing.T, tt struct {
	name      string
	keys      []string
	setup     func(m *MockKV)
	want      map[string][]byte
	expectErr bool
},
) {
	t.Helper()

	m := newTestMockKV()
	tt.setup(m)

	ctx := context.Background()

	got, err := m.BatchGetRaw(ctx, tt.keys)
	if tt.expectErr {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
		require.Equal(t, tt.want, got)
	}
}

func TestMockKV_BatchGetRaw(t *testing.T) {
	tests := getBatchGetRawTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBatchGetRawSubtest(t, tt)
		})
	}
}

func TestMockKV_BatchDelete(t *testing.T) {
	tests := []struct {
		name  string
		keys  []string
		setup func(m *MockKV)
	}{
		{
			name: "Valid",
			keys: []string{"key1", "key2"},
			setup: func(m *MockKV) {
				m.Data["key1"] = []byte("value1")
				m.Data["key2"] = []byte("value2")
				m.Data["key3"] = []byte("value3")
			},
		},
		{
			name: "PartialExists",
			keys: []string{"key1", "missing"},
			setup: func(m *MockKV) {
				m.Data["key1"] = []byte("value1")
				m.Data["key2"] = []byte("value2")
			},
		},
		{
			name:  "EmptyKeys",
			keys:  []string{},
			setup: func(m *MockKV) {},
		},
		{
			name:  "NilKeys",
			keys:  nil,
			setup: func(m *MockKV) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestMockKV()
			tt.setup(m)

			ctx := context.Background()

			err := m.BatchDelete(ctx, tt.keys)
			require.NoError(t, err)

			// Verify keys were deleted
			for _, k := range tt.keys {
				_, exists := m.Data[k]
				require.False(t, exists)
			}
		})
	}
}

func TestMockKV_ConcurrentAccess(t *testing.T) {
	m := newTestMockKV()
	ctx := context.Background()

	// Test concurrent read/write to verify thread safety
	done := make(chan bool, 3)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = m.SetRaw(ctx, "key", []byte("value"))
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = m.GetRaw(ctx, "key")
		}
		done <- true
	}()

	// List goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = m.List(ctx, "")
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Test should complete without race conditions
	require.True(t, true)
}
