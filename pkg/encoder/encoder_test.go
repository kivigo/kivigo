package encoder

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/azrod/kivigo/pkg/encoder/json"
	"github.com/azrod/kivigo/pkg/encoder/yaml"
)

func TestJSONEncoder(t *testing.T) {
	// Test that JSON variable is properly initialized and is a JSON encoder
	require.NotNil(t, JSON)
	require.IsType(t, &json.Encoder{}, JSON)
	
	// Test that it actually works as an encoder
	testData := map[string]interface{}{
		"key":   "value",
		"number": 42,
	}
	
	// Test encoding
	data, err := JSON.Encode(testData)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	
	// Test decoding
	var decoded map[string]interface{}
	err = JSON.Decode(data, &decoded)
	require.NoError(t, err)
	require.Equal(t, "value", decoded["key"])
	require.Equal(t, float64(42), decoded["number"]) // JSON numbers are decoded as float64
}

func TestYAMLEncoder(t *testing.T) {
	// Test that YAML variable is properly initialized and is a YAML encoder
	require.NotNil(t, YAML)
	require.IsType(t, &yaml.Encoder{}, YAML)
	
	// Test that it actually works as an encoder
	testData := map[string]interface{}{
		"key":   "value",
		"number": 42,
	}
	
	// Test encoding
	data, err := YAML.Encode(testData)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	
	// Test decoding
	var decoded map[string]interface{}
	err = YAML.Decode(data, &decoded)
	require.NoError(t, err)
	require.Equal(t, "value", decoded["key"])
	require.Equal(t, 42, decoded["number"]) // YAML preserves int type
}

func TestEncodersAreDistinct(t *testing.T) {
	// Ensure JSON and YAML are different instances
	require.NotEqual(t, JSON, YAML)
	require.IsType(t, &json.Encoder{}, JSON)
	require.IsType(t, &yaml.Encoder{}, YAML)
}

func TestEncodersProduceDifferentOutput(t *testing.T) {
	testData := map[string]string{
		"greeting": "hello",
		"target":   "world",
	}
	
	jsonData, err := JSON.Encode(testData)
	require.NoError(t, err)
	
	yamlData, err := YAML.Encode(testData)
	require.NoError(t, err)
	
	// JSON and YAML should produce different byte representations
	require.NotEqual(t, jsonData, yamlData)
	
	// But both should be able to decode back to the same data
	var jsonDecoded, yamlDecoded map[string]string
	
	err = JSON.Decode(jsonData, &jsonDecoded)
	require.NoError(t, err)
	
	err = YAML.Decode(yamlData, &yamlDecoded)
	require.NoError(t, err)
	
	require.Equal(t, testData, jsonDecoded)
	require.Equal(t, testData, yamlDecoded)
	require.Equal(t, jsonDecoded, yamlDecoded)
}

func TestEncodersCanHandleComplexData(t *testing.T) {
	type TestStruct struct {
		Name   string            `json:"name" yaml:"name"`
		Age    int               `json:"age" yaml:"age"`
		Tags   []string          `json:"tags" yaml:"tags"`
		Meta   map[string]string `json:"meta" yaml:"meta"`
	}
	
	testData := TestStruct{
		Name: "John Doe",
		Age:  30,
		Tags: []string{"developer", "go"},
		Meta: map[string]string{
			"location": "USA",
			"level":    "senior",
		},
	}
	
	// Test JSON encoder
	jsonData, err := JSON.Encode(testData)
	require.NoError(t, err)
	
	var jsonDecoded TestStruct
	err = JSON.Decode(jsonData, &jsonDecoded)
	require.NoError(t, err)
	require.Equal(t, testData, jsonDecoded)
	
	// Test YAML encoder
	yamlData, err := YAML.Encode(testData)
	require.NoError(t, err)
	
	var yamlDecoded TestStruct
	err = YAML.Decode(yamlData, &yamlDecoded)
	require.NoError(t, err)
	require.Equal(t, testData, yamlDecoded)
}

func TestEncodersSingletonBehavior(t *testing.T) {
	// Test that accessing the variables multiple times returns the same instances
	json1 := JSON
	json2 := JSON
	require.Same(t, json1, json2)
	
	yaml1 := YAML
	yaml2 := YAML
	require.Same(t, yaml1, yaml2)
}