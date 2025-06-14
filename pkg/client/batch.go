package client

import (
	"context"
	"fmt"
	"reflect"

	"github.com/azrod/kivigo/pkg/errs"
	"github.com/azrod/kivigo/pkg/models"
)

// BatchGet retrieves multiple values from the key-value store in a single batch operation.
// It takes a context, a slice of keys to retrieve, and a destination map where the decoded values will be stored.
// Returns an error if the backend does not support batch operations or if decoding fails.
// The destination map should have string keys corresponding to the keys being retrieved, and values of the type
// that the encoder can decode into. If a key does not exist, it will not be set in the destination map.
// Example usage:
//
//	var values map[string]string
//	err := client.BatchGet(ctx, []string{"key1", "key2"}, values)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Retrieved values:", values)
func (c Client) BatchGet(ctx context.Context, keys []string, dest any) error {
	batch, ok := c.KV.(models.KVWithBatch)
	if !ok {
		return fmt.Errorf("BatchGet not supported by backend")
	}

	if len(keys) == 0 {
		return errs.ErrEmptyKey
	}

	for _, key := range keys {
		if key == "" {
			return errs.ErrEmptyKey
		}
	}

	raws, err := batch.BatchGetRaw(ctx, keys)
	if err != nil {
		return err
	}

	destType := reflect.TypeOf(dest)
	if destType.Kind() != reflect.Map || destType.Key().Kind() != reflect.String {
		return fmt.Errorf("destination must be a map with string keys, got %T", dest)
	}

	// For each key in the raws, decode the value into the corresponding entry in dest
	// To determine the type of the destination map's values, we can use reflection
	if destType.Elem().Kind() == reflect.Interface {
		// If the destination map's values are of type interface{}, we can decode directly
		return fmt.Errorf("destination map's values cannot be of type interface{}; please specify a concrete type")
	}

	// If the destination map's values are of a specific type, we need to ensure that
	// the encoder can decode into that type
	destValueType := destType.Elem()

	for k, raw := range raws {
		if raw == nil {
			// If the raw value is nil, we skip setting it in the destination map
			continue
		}
		// Create a zero value of the destination type to decode into
		destValue := reflect.New(destValueType).Interface()
		if err := c.opts.Encoder.Decode(raw, destValue); err != nil {
			return fmt.Errorf("failed to decode value for key %s: %w", k, err)
		}
		// Set the decoded value in the destination map
		reflect.ValueOf(dest).SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(destValue).Elem())
	}

	return nil
}

// BatchSet sets multiple values in the key-value store in a single batch operation.
// It takes a context, a map of keys to values, and returns an error if the backend does not support batch operations
// or if encoding fails. The keys must be strings, and the values must be of a type that the encoder can encode.
// Example usage:
//
//	err := client.BatchSet(ctx, map[string]string{"key1": "value1", "key2": "value2"})
func (c Client) BatchSet(ctx context.Context, kv map[string]any) error {
	batch, ok := c.KV.(models.KVWithBatch)
	if !ok {
		return fmt.Errorf("BatchSet not supported by backend")
	}

	if len(kv) == 0 {
		return errs.ErrEmptyBatch
	}

	for k, v := range kv {
		if k == "" {
			return errs.ErrEmptyKey
		}

		if v == nil {
			return fmt.Errorf("value for key %s cannot be nil", k)
		}
	}

	raws := make(map[string][]byte, len(kv))

	for k, v := range kv {
		raw, err := c.opts.Encoder.Encode(v)
		if err != nil {
			return fmt.Errorf("failed to encode value for key %s: %w", k, err)
		}

		raws[k] = raw
	}

	return batch.BatchSetRaw(ctx, raws)
}
