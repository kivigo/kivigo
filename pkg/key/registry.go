package key

import (
	"fmt"
	"sync"
)

// globalRegistry is a package-level registry for shared templates.
var globalRegistry = NewRegistry()

// GlobalRegistry returns the package-level registry instance.
func GlobalRegistry() *Registry {
	return globalRegistry
}

// Registry stores named key templates for reuse.
type Registry struct {
	mu        sync.RWMutex
	templates map[string]*TemplateKeyBuilder
	funcs     map[string]TransformFunc // registry-level functions
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		templates: make(map[string]*TemplateKeyBuilder),
		funcs:     make(map[string]TransformFunc),
	}
}

// Register adds a named template and injects all registry-level functions into it.
func (r *Registry) Register(name string, tmpl *TemplateKeyBuilder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.templates[name]; exists {
		return fmt.Errorf("template with name '%s' already exists", name)
	}
	// Inject all registry-level funcs into the template
	for fname, fn := range r.funcs {
		tmpl.RegisterFunc(fname, fn)
	}
	r.templates[name] = tmpl
	return nil
}

// RegisterFunc registers a transformation function globally for all templates in the registry.
func (r *Registry) RegisterFunc(name string, fn TransformFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.funcs[name] = fn
	// Inject into all existing templates
	for _, tmpl := range r.templates {
		tmpl.RegisterFunc(name, fn)
	}
}

// Get retrieves a named template builder.
func (r *Registry) Get(name string) (*TemplateKeyBuilder, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.templates[name]
	return t, ok
}

// MustGet retrieves a named template builder or panics if not found.
func (r *Registry) MustGet(name string) *TemplateKeyBuilder {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.templates[name]
	if !ok {
		panic(fmt.Sprintf("template with name '%s' not found", name))
	}
	return t
}

// Delete removes a named template.
func (r *Registry) Delete(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.templates, name)
}
