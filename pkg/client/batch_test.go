package client_test

import (
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
						if err := c.Set(t.Context(), k, v); err != nil {
							t.Fatalf("failed to set key %s: %v", k, err)
						}
					}
				}
			}

			var values = make(map[string]testStruct)

			err := c.BatchGet(t.Context(), tt.keys, values)
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
						if err := c.Set(t.Context(), k, v); err != nil {
							t.Fatalf("failed to set key %s: %v", k, err)
						}
					}
				}
			}

			err := c.BatchSet(t.Context(), tt.kv)
			if (err != nil) != tt.wantErr {
				t.Errorf("BatchSet() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
	}
}
