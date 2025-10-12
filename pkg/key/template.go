package key

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// TemplateKeyBuilder is a KeyBuilder implementation that builds a key from a template string.
//
// Example:
//
//	builder := &TemplateKeyBuilder{Template: "user:{userID}:data:{dataID}"}
//	key, _ := builder.Build(ctx, map[string]any{"userID": 42, "dataID": "abc"}) // key == "user:42:data:abc"
//
// Useful for readable, structured keys.
// KeyVars is an optional interface for custom structs to provide template variables.
type KeyVars interface { //nolint:revive
	KeyVars() map[string]interface{}
}

// TemplateKeyBuilder builds keys from a template string.
type TemplateKeyBuilder struct {
	Template string
}

func Template(tmpl string) (*TemplateKeyBuilder, error) {
	// Basic validation: must not be empty, braces must be balanced, fields must be non-empty, and only allowed characters
	if len(tmpl) == 0 {
		return nil, fmt.Errorf("template string cannot be empty")
	}
	allowedRegex := regexp.MustCompile(`^[a-zA-Z0-9/_|:\-{}]+$`)
	if !allowedRegex.MatchString(tmpl) {
		return nil, fmt.Errorf("template contains invalid characters. Allowed: a-z, A-Z, 0-9, /, |, -, _, :, {}")
	}
	fieldRegex := regexp.MustCompile(`^[a-zA-Z0-9/_|:\-]+$`)
	for i := 0; i < len(tmpl); i++ {
		if tmpl[i] == '{' {
			j := strings.IndexByte(tmpl[i:], '}')
			if j == -1 {
				return nil, fmt.Errorf("template contains unclosed '{'")
			}
			j += i
			if j == i+1 {
				return nil, fmt.Errorf("template contains empty field name: {}")
			}
			field := tmpl[i+1 : j]
			if !fieldRegex.MatchString(field) {
				return nil, fmt.Errorf("template field contains invalid characters in {%s}. Allowed: a-z, A-Z, 0-9", field)
			}
			i = j
		}
	}
	return &TemplateKeyBuilder{Template: tmpl}, nil
}

// Build replaces {field} in the template using data from map, struct, or KeyVars interface.
//
// Data source priority (highest to lowest):
//  1. If data implements KeyVars, uses KeyVars() map.
//  2. If data is map[string]interface{}, uses map values.
//  3. If data is a struct, uses exported struct fields.
//
// Returns an error if any template fields are missing.
func (t *TemplateKeyBuilder) Build(ctx context.Context, data any) (string, error) { //nolint:cyclop
	vars := make(map[string]interface{})
	switch v := data.(type) {
	case KeyVars:
		vars = v.KeyVars()
	case map[string]interface{}:
		vars = v
	default:
		rv := reflect.ValueOf(data)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		if rv.Kind() == reflect.Struct {
			rt := rv.Type()
			for i := 0; i < rt.NumField(); i++ {
				f := rt.Field(i)
				if f.PkgPath == "" { // exported
					vars[f.Name] = rv.Field(i).Interface()
				}
			}
		}
	}

	key := t.Template
	missing := []string{}
	result := ""
	start := 0
	for {
		i := strings.Index(key[start:], "{")
		if i == -1 {
			result += key[start:]
			break
		}
		i += start
		result += key[start:i]
		j := strings.Index(key[i:], "}")
		if j == -1 {
			// Unclosed brace, treat as literal
			result += key[i:]
			break
		}
		j += i
		field := key[i+1 : j]
		val, ok := vars[field]
		if ok {
			result += fmt.Sprintf("%v", val)
		} else {
			missing = append(missing, field)
			result += "{" + field + "}"
		}
		start = j + 1
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("missing template fields: %v", missing)
	}
	return result, nil
}
