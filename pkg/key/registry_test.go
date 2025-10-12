package key

import (
	"testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	tmpl, err := Template("foo:{bar}")
	if err != nil {
		t.Fatalf("unexpected error on Template: %v", err)
	}
	err = r.Register("foo", tmpl)
	if err != nil {
		t.Fatalf("unexpected error on Register: %v", err)
	}
	b, ok := r.Get("foo")
	if !ok {
		t.Fatal("expected template to be found")
	}
	key, err := b.Build(nil, map[string]interface{}{"bar": 123})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "foo:123" {
		t.Errorf("got %q, want %q", key, "foo:123")
	}
}

func TestRegistry_Delete(t *testing.T) {
	r := NewRegistry()
	tmpl, err := Template("foo:{bar}")
	if err != nil {
		t.Fatalf("unexpected error on Template: %v", err)
	}
	err = r.Register("foo", tmpl)
	if err != nil {
		t.Fatalf("unexpected error on Register: %v", err)
	}
	r.Delete("foo")
	_, ok := r.Get("foo")
	if ok {
		t.Fatal("expected template to be deleted")
	}
}

func TestGlobalRegistry(t *testing.T) {
	tmpl, err := Template("user:{id}")
	if err != nil {
		t.Fatalf("unexpected error on Template: %v", err)
	}
	err = GlobalRegistry().Register("user", tmpl)
	if err != nil {
		t.Fatalf("unexpected error on Register: %v", err)
	}
	b, ok := GlobalRegistry().Get("user")
	if !ok {
		t.Fatal("expected global template to be found")
	}
	key, err := b.Build(nil, map[string]interface{}{"id": 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "user:42" {
		t.Errorf("got %q, want %q", key, "user:42")
	}
	GlobalRegistry().Delete("user")

	_, ok = GlobalRegistry().Get("user")
	if ok {
		t.Fatal("expected global template to be deleted")
	}
}
