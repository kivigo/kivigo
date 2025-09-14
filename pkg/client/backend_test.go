//nolint:funlen,cyclop
package client_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/kivigo/encoders/json"

	"github.com/kivigo/kivigo/pkg/client"
	"github.com/kivigo/kivigo/pkg/mock"
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
			name:    "invalid key",
			key:     "",
			wantErr: true,
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

	c, err := client.New(mockKV, client.Option{Encoder: json.New()})
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

	c, err := client.New(mockKV, client.Option{Encoder: json.New()})
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

func Test_Delete(t *testing.T) {
	type test struct {
		name    string
		key     string
		wantErr bool
	}

	tests := []test{
		{
			name:    "existing key",
			key:     "existingKey",
			wantErr: false,
		},
		{
			name:    "invalid key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "non-existing key",
			key:     "nonExistingKey",
			wantErr: true,
		},
	}

	mockKV := &mock.MockKV{Data: map[string][]byte{
		"existingKey": []byte(`"existingValue"`),
	}}

	c, err := client.New(mockKV, client.Option{Encoder: json.New()})
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Delete(context.Background(), tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
	}
}

func Test_MatchKeys(t *testing.T) {
	mockKV := &mock.MockKV{Data: map[string][]byte{
		"prefix-1": []byte(`"val1"`),
		"prefix-2": []byte(`"val2"`),
		"other":    []byte(`"val3"`),
	}}

	c, err := client.New(mockKV, client.Option{Encoder: json.New()})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		prefix  string
		f       client.MatchKeysFunc
		want    bool
		wantErr bool
	}{
		{
			name:    "nil func",
			prefix:  "prefix",
			f:       nil,
			want:    false,
			wantErr: true,
		},
		{
			name:   "custom match - at least one key",
			prefix: "prefix",
			f: func(keys []string) (bool, error) {
				return len(keys) > 0, nil
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "custom match - no keys",
			prefix: "notfound",
			f: func(keys []string) (bool, error) {
				return len(keys) == 0, nil
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "custom match - error in func",
			prefix: "prefix",
			f: func(keys []string) (bool, error) {
				return false, fmt.Errorf("custom error")
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.MatchKeys(context.Background(), tt.prefix, tt.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("MatchKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_HasKey(t *testing.T) {
	mockKV := &mock.MockKV{Data: map[string][]byte{
		"foo": []byte(`"bar"`),
	}}
	c, err := client.New(mockKV, client.Option{Encoder: json.New()})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		key     string
		want    bool
		wantErr bool
	}{
		{"existing key", "foo", true, false},
		{"non-existing key", "baz", false, false},
		{"empty key", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.HasKey(context.Background(), tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("HasKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_HasKeys(t *testing.T) {
	mockKV := &mock.MockKV{Data: map[string][]byte{
		"a": []byte(`"1"`),
		"b": []byte(`"2"`),
	}}
	c, err := client.New(mockKV, client.Option{Encoder: json.New()})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		keys    []string
		want    bool
		wantErr bool
	}{
		{"all existing", []string{"a", "b"}, true, false},
		{"one missing", []string{"a", "c"}, false, false},
		{"none existing", []string{"x", "y"}, false, false},
		{"empty keys", []string{}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.HasKeys(context.Background(), tt.keys)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("HasKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
