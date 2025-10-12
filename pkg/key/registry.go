package key

import "fmt"

// globalRegistry is a package-level registry for shared templates.
var globalRegistry = NewRegistry()

// GlobalRegistry returns the package-level registry instance.
func GlobalRegistry() *Registry {
	return globalRegistry
}

// Registry stores named key templates for reuse.
type Registry struct {
	templates map[string]*TemplateKeyBuilder
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{templates: make(map[string]*TemplateKeyBuilder)}
}

// Register adds a named template if it does not already exist.
// Returns an error if the name is already registered.
func (r *Registry) Register(name string, tmpl *TemplateKeyBuilder) error {
	if _, exists := r.templates[name]; exists {
		return fmt.Errorf("template with name '%s' already exists", name)
	}
	r.templates[name] = tmpl
	return nil
}

// Get retrieves a named template builder.
func (r *Registry) Get(name string) (*TemplateKeyBuilder, bool) {
	t, ok := r.templates[name]
	return t, ok
}

// Delete removes a named template.
func (r *Registry) Delete(name string) {
	delete(r.templates, name)
}
