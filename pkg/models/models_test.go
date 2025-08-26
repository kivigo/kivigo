package models

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestInterfaces ensures the interfaces are properly defined and can be implemented
func TestEncoderInterface(t *testing.T) {
	// Test that Encoder interface has the expected methods
	var encoder Encoder

	require.Nil(t, encoder) // Interface should be nil when not implemented

	// Test with a mock implementation
	mockEncoder := &mockEncoder{}
	encoder = mockEncoder
	require.NotNil(t, encoder)

	// Test interface methods
	data, err := encoder.Encode("test")
	require.NoError(t, err)
	require.Equal(t, []byte("encoded:test"), data)

	var decoded string
	err = encoder.Decode([]byte("test"), &decoded)
	require.NoError(t, err)
	require.Equal(t, "decoded:test", decoded)
}

func TestKVInterface(t *testing.T) {
	// Test that KV interface has the expected methods
	var kv KV

	require.Nil(t, kv)

	// Test with a mock implementation
	mockKV := &mockKV{}
	kv = mockKV
	require.NotNil(t, kv)

	ctx := context.Background()

	// Test interface methods
	err := kv.SetRaw(ctx, "key", []byte("value"))
	require.NoError(t, err)

	value, err := kv.GetRaw(ctx, "key")
	require.NoError(t, err)
	require.Equal(t, []byte("value"), value)

	keys, err := kv.List(ctx, "")
	require.NoError(t, err)
	require.Equal(t, []string{"key"}, keys)

	err = kv.Delete(ctx, "key")
	require.NoError(t, err)

	err = kv.Close()
	require.NoError(t, err)
}

func TestKVWithHealthInterface(t *testing.T) {
	var kvHealth KVWithHealth

	require.Nil(t, kvHealth)

	mockKVHealth := &mockKVWithHealth{}
	kvHealth = mockKVHealth
	require.NotNil(t, kvHealth)

	ctx := context.Background()
	err := kvHealth.Health(ctx)
	require.NoError(t, err)
}

func TestKVWithBatchInterface(t *testing.T) {
	var kvBatch KVWithBatch

	require.Nil(t, kvBatch)

	mockKVBatch := &mockKVWithBatch{}
	kvBatch = mockKVBatch
	require.NotNil(t, kvBatch)

	ctx := context.Background()

	// Test batch operations
	kv := map[string][]byte{"key1": []byte("value1"), "key2": []byte("value2")}
	err := kvBatch.BatchSetRaw(ctx, kv)
	require.NoError(t, err)

	result, err := kvBatch.BatchGetRaw(ctx, []string{"key1", "key2"})
	require.NoError(t, err)
	require.Equal(t, kv, result)

	err = kvBatch.BatchDelete(ctx, []string{"key1", "key2"})
	require.NoError(t, err)
}

func TestInterfaceComposition(t *testing.T) {
	// Test that a struct can implement multiple interfaces
	composite := &compositeKV{}

	// Should implement all interfaces
	var kv KV = composite

	var kvHealth KVWithHealth = composite

	var kvBatch KVWithBatch = composite

	require.NotNil(t, kv)
	require.NotNil(t, kvHealth)
	require.NotNil(t, kvBatch)

	ctx := context.Background()

	// Test that all methods work
	err := composite.SetRaw(ctx, "test", []byte("value"))
	require.NoError(t, err)

	err = composite.Health(ctx)
	require.NoError(t, err)

	err = composite.BatchSetRaw(ctx, map[string][]byte{"batch": []byte("test")})
	require.NoError(t, err)
}

// Mock implementations for testing

type mockEncoder struct{}

func (m *mockEncoder) Encode(v any) ([]byte, error) {
	return []byte("encoded:" + v.(string)), nil
}

func (m *mockEncoder) Decode(data []byte, v any) error {
	*(v.(*string)) = "decoded:" + string(data)

	return nil
}

type mockKV struct {
	data map[string][]byte
}

func (m *mockKV) Close() error { return nil }

func (m *mockKV) List(_ context.Context, _ string) ([]string, error) {
	if m.data == nil {
		return []string{"key"}, nil
	}

	keys := make([]string, 0, len(m.data))

	for k := range m.data {
		keys = append(keys, k)
	}

	return keys, nil
}

func (m *mockKV) GetRaw(_ context.Context, key string) ([]byte, error) {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}

	if val, ok := m.data[key]; ok {
		return val, nil
	}

	return []byte("value"), nil
}

func (m *mockKV) SetRaw(_ context.Context, key string, value []byte) error {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}

	m.data[key] = value

	return nil
}

func (m *mockKV) Delete(_ context.Context, key string) error {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}

	delete(m.data, key)

	return nil
}

type mockKVWithHealth struct{}

func (m *mockKVWithHealth) Health(_ context.Context) error {
	return nil
}

type mockKVWithBatch struct {
	data map[string][]byte
}

func (m *mockKVWithBatch) BatchGetRaw(_ context.Context, keys []string) (map[string][]byte, error) {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}

	result := make(map[string][]byte)

	for _, key := range keys {
		if val, ok := m.data[key]; ok {
			result[key] = val
		}
	}

	return result, nil
}

func (m *mockKVWithBatch) BatchSetRaw(_ context.Context, kv map[string][]byte) error {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}

	for k, v := range kv {
		m.data[k] = v
	}

	return nil
}

func (m *mockKVWithBatch) BatchDelete(_ context.Context, keys []string) error {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}

	for _, key := range keys {
		delete(m.data, key)
	}

	return nil
}

type compositeKV struct {
	data map[string][]byte
}

func (c *compositeKV) Close() error { return nil }

func (c *compositeKV) List(_ context.Context, _ string) ([]string, error) {
	keys := make([]string, 0, len(c.data))

	for k := range c.data {
		keys = append(keys, k)
	}

	return keys, nil
}

func (c *compositeKV) GetRaw(_ context.Context, key string) ([]byte, error) {
	return c.data[key], nil
}

func (c *compositeKV) SetRaw(_ context.Context, key string, value []byte) error {
	if c.data == nil {
		c.data = make(map[string][]byte)
	}

	c.data[key] = value

	return nil
}

func (c *compositeKV) Delete(_ context.Context, key string) error {
	delete(c.data, key)

	return nil
}

func (c *compositeKV) Health(_ context.Context) error {
	return nil
}

func (c *compositeKV) BatchGetRaw(_ context.Context, keys []string) (map[string][]byte, error) {
	result := make(map[string][]byte)

	for _, key := range keys {
		if val, ok := c.data[key]; ok {
			result[key] = val
		}
	}

	return result, nil
}

func (c *compositeKV) BatchSetRaw(_ context.Context, kv map[string][]byte) error {
	if c.data == nil {
		c.data = make(map[string][]byte)
	}

	for k, v := range kv {
		c.data[k] = v
	}

	return nil
}

func (c *compositeKV) BatchDelete(_ context.Context, keys []string) error {
	for _, key := range keys {
		delete(c.data, key)
	}

	return nil
}
