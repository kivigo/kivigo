//nolint:funlen
package client_test

import (
	"context"
	"testing"

	"github.com/azrod/kivigo/pkg/client"
	"github.com/azrod/kivigo/pkg/encoder"
	"github.com/azrod/kivigo/pkg/mock"
)

func Test_BatchGet(t *testing.T) {
	type (
		test struct {
			name       string
			keys       []string
			want       map[string]any
			wantErr    bool
			createKeys bool
		}

		testStruct struct {
			Value string
		}
	)

	tests := []test{
		{
			name:       "valid keys",
			keys:       []string{"key1", "key2"},
			want:       map[string]any{"key1": testStruct{"value1"}, "key2": testStruct{"value2"}},
			wantErr:    false,
			createKeys: true,
		},
		{
			name:       "empty keys",
			keys:       []string{},
			want:       map[string]any{},
			wantErr:    true,
			createKeys: false,
		},
		{
			name:       "non-existent keys",
			keys:       []string{"nonExistentKey1", "nonExistentKey2"},
			want:       nil,
			wantErr:    false,
			createKeys: false,
		},
		{
			name:       "partial keys",
			keys:       []string{"key1", "nonExistentKey"},
			want:       map[string]any{"key1": testStruct{"value1"}},
			wantErr:    false,
			createKeys: true,
		},
		{
			name:       "invalid keys",
			keys:       []string{"", "key2"},
			want:       map[string]any{"": nil, "key2": testStruct{"value2"}},
			wantErr:    true,
			createKeys: true,
		},
	}

	mockKV := &mock.MockKV{Data: map[string][]byte{}}

	c, err := client.New(mockKV, client.Option{Encoder: encoder.JSON})
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createKeys {
				for k, v := range tt.want {
					if v != nil {
						if err := c.Set(context.Background(), k, v); err != nil {
							t.Fatalf("failed to set key %s: %v", k, err)
						}
					}
				}
			}

			values := make(map[string]testStruct)

			err := c.BatchGet(context.Background(), tt.keys, values)
			if (err != nil) != tt.wantErr {
				t.Errorf("BatchGet() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
	}
}

func Test_BatchSet(t *testing.T) {
	type (
		test struct {
			name       string
			kv         map[string]any
			wantErr    bool
			createKeys bool
		}

		testStruct struct {
			Value string
		}
	)

	tests := []test{
		{
			name:       "valid batch set",
			kv:         map[string]any{"key1": testStruct{"value1"}, "key2": testStruct{"value2"}},
			wantErr:    false,
			createKeys: false,
		},
		{
			name:       "empty batch set",
			kv:         map[string]any{},
			wantErr:    true,
			createKeys: false,
		},
		{
			name:       "partial batch set with existing keys",
			kv:         map[string]any{"key1": testStruct{"newValue1"}, "key3": testStruct{"value3"}},
			wantErr:    false,
			createKeys: true,
		},
		{
			name:       "batch set with nil values",
			kv:         map[string]any{"key1": nil, "key2": testStruct{"value2"}},
			wantErr:    true,
			createKeys: true,
		},
	}

	mockKV := &mock.MockKV{Data: map[string][]byte{}}

	c, err := client.New(mockKV, client.Option{Encoder: encoder.JSON})
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createKeys {
				for k, v := range tt.kv {
					if v != nil {
						if err := c.Set(context.Background(), k, v); err != nil {
							t.Fatalf("failed to set key %s: %v", k, err)
						}
					}
				}
			}

			err := c.BatchSet(context.Background(), tt.kv)
			if (err != nil) != tt.wantErr {
				t.Errorf("BatchSet() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
	}
}

// Mock client with a backend that does NOT implement KVWithBatch
type dummyKV struct{}

func (d *dummyKV) GetRaw(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
}

func (d *dummyKV) SetRaw(_ context.Context, _ string, _ []byte) error {
	return nil
}

func (d *dummyKV) Delete(_ context.Context, _ string) error {
	return nil
}

func (d *dummyKV) List(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func (d *dummyKV) Close() error {
	return nil
}

func Test_BatchGet_NotSupported(t *testing.T) {
	type testStruct struct {
		Value string
	}

	c, err := client.New(&dummyKV{}, client.Option{Encoder: encoder.JSON})
	if err != nil {
		t.Fatal(err)
	}

	values := make(map[string]testStruct)

	err = c.BatchGet(context.Background(), []string{"key1"}, values)
	if err == nil || err.Error() != "BatchGet not supported by backend" {
		t.Errorf("expected 'BatchGet not supported by backend' error, got %v", err)
	}
}

func Test_BatchSet_NotSupported(t *testing.T) {
	c, err := client.New(&dummyKV{}, client.Option{Encoder: encoder.JSON})
	if err != nil {
		t.Fatal(err)
	}

	err = c.BatchSet(context.Background(), map[string]any{"key1": "value1"})
	if err == nil || err.Error() != "BatchSet not supported by backend" {
		t.Errorf("expected 'BatchSet not supported by backend' error, got %v", err)
	}
}

func Test_BatchGet_DestTypeErrors(t *testing.T) {
	type testStruct struct {
		Value string
	}

	// Backend supports batch
	mockKV := &mock.MockKV{Data: map[string][]byte{}}

	cBatch, err := client.New(mockKV, client.Option{Encoder: encoder.JSON})
	if err != nil {
		t.Fatal(err)
	}

	// Prepare a valid key/value for decoding
	raw, _ := encoder.JSON.Encode(testStruct{Value: "foo"})
	mockKV.Data["key1"] = raw

	tests := []struct {
		name    string
		dest    any
		wantErr string
	}{
		{
			name:    "dest is not a map",
			dest:    &[]testStruct{},
			wantErr: "destination must be a map with string keys",
		},
		{
			name:    "dest map key is not string",
			dest:    &map[int]testStruct{},
			wantErr: "destination must be a map with string keys",
		},
		{
			name:    "dest map value is interface{}",
			dest:    &map[string]interface{}{},
			wantErr: "destination must be a map with string keys, got *map[string]interface {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := []string{"key1"}

			err := cBatch.BatchGet(context.Background(), keys, tt.dest)
			if err == nil || err.Error()[:len(tt.wantErr)] != tt.wantErr {
				t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func Test_BatchGet_DecodeError(t *testing.T) {
	type testStruct struct {
		Value string
	}

	batchKV := &mock.MockKV{Data: map[string][]byte{}}

	c, err := client.New(batchKV, client.Option{Encoder: encoder.JSON})
	if err != nil {
		t.Fatal(err)
	}

	// Insert invalid JSON for key
	batchKV.Data["key1"] = []byte("{invalid json}")

	dest := map[string]testStruct{}

	err = c.BatchGet(context.Background(), []string{"key1"}, &dest)
	if err == nil || err.Error() != "destination must be a map with string keys, got *map[string]client_test.testStruct" {
		t.Errorf("expected decode error, got %v", err)
	}
}
