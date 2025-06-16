package json

import (
	"encoding/json"
	"fmt"

	"github.com/azrod/kivigo/pkg/models"
)

var _ models.Encoder = (*Encoder)(nil)

type Encoder struct{}

// Encode encodes the given value into JSON format.
func (f *Encoder) Encode(value any) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to encode value to JSON: %w", err)
	}

	return data, nil
}

// Decode decodes the given JSON data into the provided value.
func (f *Encoder) Decode(data []byte, value any) error {
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	if len(data) == 0 {
		return nil // No data to decode
	}

	err := json.Unmarshal(data, value)
	if err != nil {
		return fmt.Errorf("failed to decode JSON data: %w", err)
	}

	return nil
}
