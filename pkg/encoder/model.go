package encoder

import (
	"github.com/azrod/kivigo/pkg/encoder/json"
	"github.com/azrod/kivigo/pkg/encoder/yaml"
)

var (
	JSON = &json.Encoder{}
	YAML = &yaml.Encoder{}
)
