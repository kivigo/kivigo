//nolint:funlen,cyclop
package client_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/azrod/kivigo/pkg/client"
	"github.com/azrod/kivigo/pkg/encoder"
	"github.com/azrod/kivigo/pkg/mock"
)

func Test_Get(t *testing.T) {
	type test struct {
		name      string
		key       string
		want      any
		wantErr   bool
		createKey bool
	}

	type teststruct struct {
		Value string
	}

	tests := []test{
		{
			name:      "existing key",
			key:       "existingKey",
			want:      teststruct{"existingValue"},
			wantErr:   false,
			createKey: true, // This will create the key before testing
		},
		{
			name:      "existing-invalid-key",
			key:       "existing-invalid-key",
			want:      "value",
			wantErr:   true,
			createKey: true, // This will create the key before testing
		},
		{
			name:    "non-existing key",
			key:     "nonExistingKey",
			want:    "",
			wantErr: true,
		},
	}

	mockKV := &mock.MockKV{Data: map[string][]byte{}}

	c, err := client.New(mockKV, client.Option{Encoder: encoder.JSON})
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createKey {
				switch v := tt.want.(type) {
				// struct
				case teststruct:
					t.Logf("Creating key %s with value %v", tt.key, v)

					err := c.Set(context.Background(), tt.key, v)
					if err != nil {
						t.Fatalf("Set() error = %v", err)
					}
				case string:
					t.Logf("Creating key %s with value %s", tt.key, v)

					err := c.SetRaw(context.Background(), tt.key, []byte(v))
					if err != nil {
						t.Fatalf("SetRaw() error = %v", err)
					}
				}
			}

			var got teststruct

			err = c.Get(context.Background(), tt.key, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Set(t *testing.T) {
	type teststruct struct {
		Value string
	}

	type test struct {
		name    string
		key     string
		value   any
		wantErr bool
	}

	tests := []test{
		{
			name:    "valid key-value pair",
			key:     "validKey",
			value:   teststruct{"validValue"},
			wantErr: false,
		},
		{
			name:    "invalid key",
			key:     "",
			value:   teststruct{"invalidValue"},
			wantErr: true,
		},
		{
			name:    "invalid value type",
			key:     "invalidValueTypeKey",
			value:   nil,
			wantErr: true,
		},
	}

	mockKV := &mock.MockKV{Data: map[string][]byte{}}

	c, err := client.New(mockKV, client.Option{Encoder: encoder.JSON})
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Set(context.Background(), tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr {
				var got teststruct

				err := c.Get(context.Background(), tt.key, &got)
				if err != nil {
					t.Errorf("Get() error = %v", err)

					return
				}

				if !reflect.DeepEqual(got, tt.value) {
					t.Errorf("Get() = %v, want %v", got, tt.value)
				}
			}
		})
	}
}
