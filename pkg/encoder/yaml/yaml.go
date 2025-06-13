package yaml

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/azrod/kivigo/pkg/models"
)

var _ models.Encoder = (*Encoder)(nil)

type Encoder struct{}

// Encode encodes the given value into YAML format.
func (f *Encoder) Encode(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	data, err := yaml.Marshal(value)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Decode decodes the given YAML data into the provided value.
func (f *Encoder) Decode(data []byte, value any) error {
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	if len(data) == 0 {
		return nil // No data to decode
	}

	err := yaml.Unmarshal(data, value)
	if err != nil {
		return err
	}
	return nil
}
