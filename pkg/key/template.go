package key

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// KeyVars is an optional interface for custom structs to provide template variables.
type KeyVars interface { //nolint:revive
	KeyVars() map[string]interface{}
}

// TransformFunc is a function that transforms a string value (with optional args).
type TransformFunc func(val string, args ...string) (string, error)

// TemplateKeyBuilder builds keys from a template string with dynamic variables, transformations, and conditionals.
type TemplateKeyBuilder struct {
	Template string
	funcs    map[string]TransformFunc
}

// Template creates a new TemplateKeyBuilder with built-in functions.
func Template(tmpl string) (*TemplateKeyBuilder, error) {
	if len(tmpl) == 0 {
		return nil, fmt.Errorf("template string cannot be empty")
	}
	allowedRegex := regexp.MustCompile(`^[a-zA-Z0-9/_|:\-{}(),\"' +]+$`)
	if !allowedRegex.MatchString(tmpl) {
		return nil, fmt.Errorf("template contains invalid characters. Allowed: a-z, A-Z, 0-9, /, |, -, _, :, {}, (), ',', \" and space")
	}
	tb := &TemplateKeyBuilder{
		Template: tmpl,
		funcs:    make(map[string]TransformFunc),
	}
	for k, v := range builtinFuncs {
		tb.funcs[k] = v
	}
	return tb, nil
}

// RegisterFunc registers a custom transformation function for this template.
func (t *TemplateKeyBuilder) RegisterFunc(name string, fn TransformFunc) {
	t.funcs[name] = fn
}

// Build (Render) replaces {field|func1|func2(args)} in the template using data and applies transformations, defaults, and conditionals.
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
				if f.PkgPath == "" {
					vars[f.Name] = rv.Field(i).Interface()
				}
			}
		}
	}

	tmpl := t.Template
	result := ""
	start := 0
	for {
		op := strings.Index(tmpl[start:], "{")
		if op == -1 {
			result += tmpl[start:]
			break
		}
		op += start
		result += tmpl[start:op]
		cl := strings.Index(tmpl[op:], "}")
		if cl == -1 {
			result += tmpl[op:]
			break
		}
		cl += op
		tok := tmpl[op+1 : cl]
		val, err := t.evalToken(tok, vars)
		if err != nil {
			return "", err
		}
		result += val
		start = cl + 1
	}
	return result, nil
}

// evalToken parses and evaluates a token like "field|upper|default('x')".
func (t *TemplateKeyBuilder) evalToken(token string, vars map[string]interface{}) (string, error) {
	parts := strings.Split(token, "|")
	if len(parts) == 0 {
		return "", nil
	}
	name := parts[0]
	val, ok := vars[name]
	var sval string
	if ok {
		sval = fmt.Sprintf("%v", val)
	} else {
		sval = ""
	}
	for _, fncall := range parts[1:] {
		fn, args := parseFuncCall(fncall)
		f, ok := t.funcs[fn]
		if !ok {
			return "", fmt.Errorf("unknown transform function: %s", fn)
		}
		var err error
		sval, err = f(sval, args...)
		if err != nil {
			return "", err
		}
	}
	return sval, nil
}

// parseFuncCall parses "func(arg1,arg2)" or "func".
func parseFuncCall(s string) (string, []string) {
	op := strings.Index(s, "(")
	if op == -1 {
		return s, nil
	}
	cl := strings.LastIndex(s, ")")
	if cl == -1 || cl < op {
		return s, nil
	}
	fn := s[:op]
	args := parseArgs(s[op+1 : cl])
	return fn, args
}

// parseArgs splits a comma-separated argument string, handling quotes and escapes.
func parseArgs(s string) []string { //nolint:cyclop
	var args []string
	var arg strings.Builder
	inSingle := false
	inDouble := false
	escape := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if escape {
			arg.WriteByte(c)
			escape = false
			continue
		}
		switch c {
		case '\\':
			escape = true
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				continue
			}
			arg.WriteByte(c)
		case '"':
			if !inSingle {
				inDouble = !inDouble
				continue
			}
			arg.WriteByte(c)
		case ',':
			if !inSingle && !inDouble {
				args = append(args, strings.TrimSpace(arg.String()))
				arg.Reset()
				continue
			}
			arg.WriteByte(c)
		default:
			arg.WriteByte(c)
		}
	}
	if arg.Len() > 0 {
		args = append(args, strings.TrimSpace(arg.String()))
	}
	// Remove surrounding quotes from each argument, and unescape quotes
	for i := range args {
		a := args[i]
		if len(a) >= 2 && ((a[0] == '\'' && a[len(a)-1] == '\'') || (a[0] == '"' && a[len(a)-1] == '"')) {
			a = a[1 : len(a)-1]
		}
		a = strings.ReplaceAll(a, `\'`, `'`)
		a = strings.ReplaceAll(a, `\"`, `"`)
		args[i] = a
	}
	return args
}

// Built-in transformation functions
var builtinFuncs = map[string]TransformFunc{
	"upper": func(val string, _ ...string) (string, error) {
		return strings.ToUpper(val), nil
	},
	"lower": func(val string, _ ...string) (string, error) {
		return strings.ToLower(val), nil
	},
	"trim": func(val string, _ ...string) (string, error) {
		return strings.TrimSpace(val), nil
	},
	"default": func(val string, args ...string) (string, error) {
		if val == "" && len(args) > 0 {
			return args[0], nil
		}
		return val, nil
	},
	"slugify": func(val string, _ ...string) (string, error) {
		return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(val), " ", "-")), nil
	},
	"if": func(val string, args ...string) (string, error) {
		if len(args) < 2 {
			return val, nil
		}
		if val != "" && val != "0" && val != "false" {
			return args[0], nil
		}
		return args[1], nil
	},
	"intadd": func(val string, args ...string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("intadd: missing argument for addition")
		}
		v, err := strconv.Atoi(val)
		if err != nil {
			return "", fmt.Errorf("intadd: cannot parse value %q as int: %w", val, err)
		}
		delta, err := strconv.Atoi(args[0])
		if err != nil {
			return "", fmt.Errorf("intadd: cannot parse argument %q as int: %w", args[0], err)
		}
		return strconv.Itoa(v + delta), nil
	},
}
