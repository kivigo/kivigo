package json

import (
	"reflect"
	"testing"
)

func TestEncoder_Encode(t *testing.T) {
	enc := &Encoder{}

	type sample struct {
		Name string
		Age  int
	}

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{"nil value", nil, true},
		{"simple struct", sample{"Alice", 30}, false},
		{"map value", map[string]int{"a": 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := enc.Encode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(got) == 0 {
				t.Errorf("Encode() got empty result for valid input")
			}
		})
	}
}

func TestEncoder_Decode(t *testing.T) {
	enc := &Encoder{}

	type sample struct {
		Name string
		Age  int
	}

	tests := []struct {
		name    string
		data    []byte
		target  any
		want    any
		wantErr bool
	}{
		{
			name:    "nil value",
			data:    []byte(`{"Name":"Bob","Age":25}`),
			target:  nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte{},
			target:  &sample{},
			want:    &sample{},
			wantErr: false,
		},
		{
			name:    "valid decode",
			data:    []byte(`{"Name":"Bob","Age":25}`),
			target:  &sample{},
			want:    &sample{Name: "Bob", Age: 25},
			wantErr: false,
		},
		{
			name:    "invalid json",
			data:    []byte(`{invalid json}`),
			target:  &sample{},
			want:    &sample{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := enc.Decode(tt.data, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.target != nil && !reflect.DeepEqual(tt.target, tt.want) {
				t.Errorf("Decode() got = %v, want %v", tt.target, tt.want)
			}
		})
	}
}

func TestEncoder_ImplementsEncoderInterface(t *testing.T) {
	var _ interface {
		Encode(any) ([]byte, error)
		Decode([]byte, any) error
	} = &Encoder{}
}
