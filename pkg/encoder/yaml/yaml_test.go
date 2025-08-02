package yaml

import (
	"reflect"
	"testing"
)

type testStruct struct {
	Name string
	Age  int
}

func TestEncoder_Encode(t *testing.T) {
	enc := &Encoder{}

	// Test nil value
	data, err := enc.Encode(nil)
	if err != nil {
		t.Errorf("Encode(nil) returned error: %v", err)
	}

	if data != nil {
		t.Errorf("Encode(nil) = %v, want nil", data)
	}

	// Test struct encoding
	val := testStruct{Name: "Alice", Age: 30}

	data, err = enc.Encode(val)
	if err != nil {
		t.Errorf("Encode(struct) returned error: %v", err)
	}

	if len(data) == 0 {
		t.Errorf("Encode(struct) returned empty data")
	}
}

func TestEncoder_Decode(t *testing.T) {
	enc := &Encoder{}

	// Test nil value
	err := enc.Decode([]byte("name: Alice\nage: 30\n"), nil)
	if err == nil {
		t.Errorf("Decode with nil value did not return error")
	}

	// Test empty data
	var out testStruct

	err = enc.Decode([]byte{}, &out)
	if err != nil {
		t.Errorf("Decode with empty data returned error: %v", err)
	}

	// Test valid decode
	data := []byte("name: Alice\nage: 30\n")

	var result testStruct

	err = enc.Decode(data, &result)
	if err != nil {
		t.Errorf("Decode returned error: %v", err)
	}

	expected := testStruct{Name: "Alice", Age: 30}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Decode = %+v, want %+v", result, expected)
	}
}

func TestEncoder_Decode_UnmarshalError(t *testing.T) {
	enc := &Encoder{}

	var out testStruct
	// Provide invalid YAML
	err := enc.Decode([]byte(":\nbad_yaml"), &out)
	if err == nil {
		t.Errorf("Decode with invalid YAML did not return error")
	}
}
