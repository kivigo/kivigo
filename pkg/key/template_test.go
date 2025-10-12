package key

import (
	"context"
	"testing"
)

func TestTemplateKeyBuilder_FieldDelimiters(t *testing.T) {
	cases := []struct {
		template string
		data     map[string]interface{}
		want     string
	}{
		{"foo:{foo-bar}", map[string]interface{}{"foo-bar": "A"}, "foo:A"},
		{"foo:{foo_bar}", map[string]interface{}{"foo_bar": "B"}, "foo:B"},
		{"foo:{foo/bar}", map[string]interface{}{"foo/bar": "C"}, "foo:C"},
		{"foo:{foo|bar}", map[string]interface{}{"foo|bar": "D"}, "foo:D"},
		{"foo:{foo:bar}", map[string]interface{}{"foo:bar": "E"}, "foo:E"},
	}
	for _, tc := range cases {
		builder, err := Template(tc.template)
		if err != nil {
			t.Fatalf("unexpected error for template %q: %v", tc.template, err)
		}
		key, err := builder.Build(context.Background(), tc.data)
		if err != nil {
			t.Fatalf("unexpected error for Build on template %q: %v", tc.template, err)
		}
		if key != tc.want {
			t.Errorf("template %q: got %q, want %q", tc.template, key, tc.want)
		}
	}
}

func TestTemplateKeyBuilder_Map(t *testing.T) {
	builder, err := Template("user:{userID}:data:{dataID}")
	if err != nil {
		t.Fatalf("unexpected error for template: %v", err)
	}
	key, err := builder.Build(context.Background(), map[string]interface{}{"userID": 42, "dataID": "abc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "user:42:data:abc" {
		t.Errorf("got %q, want %q", key, "user:42:data:abc")
	}
}

func TestTemplateKeyBuilder_Struct(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}
	builder, err := Template("user:{ID}:name:{Name}")
	if err != nil {
		t.Fatalf("unexpected error for template: %v", err)
	}
	key, err := builder.Build(context.Background(), User{ID: 7, Name: "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "user:7:name:alice" {
		t.Errorf("got %q, want %q", key, "user:7:name:alice")
	}
}

type Session struct {
	UserID    string
	SessionID string
}

func (s Session) KeyVars() map[string]interface{} {
	return map[string]interface{}{"userID": s.UserID, "sessionID": s.SessionID}
}

func TestTemplateKeyBuilder_KeyVars(t *testing.T) {
	builder, err := Template("session:{userID}:{sessionID}")
	if err != nil {
		t.Fatalf("unexpected error for template: %v", err)
	}
	key, err := builder.Build(context.Background(), Session{UserID: "bob", SessionID: "xyz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "session:bob:xyz" {
		t.Errorf("got %q, want %q", key, "session:bob:xyz")
	}
}

func TestTemplateKeyBuilder_MissingFields(t *testing.T) {
	builder, err := Template("user:{userID}:data:{dataID}")
	if err != nil {
		t.Fatalf("unexpected error for template: %v", err)
	}
	_, err = builder.Build(context.Background(), map[string]interface{}{"userID": 42})
	if err == nil {
		t.Fatal("expected error for missing field")
	}
	if got, want := err.Error(), "missing template fields: [dataID]"; got != want {
		t.Errorf("got error %q, want %q", got, want)
	}
}

func TestTemplateKeyBuilder_PointerStruct(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}
	builder, err := Template("user:{ID}:name:{Name}")
	if err != nil {
		t.Fatalf("unexpected error for template: %v", err)
	}
	user := &User{ID: 99, Name: "bob"}
	key, err := builder.Build(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "user:99:name:bob" {
		t.Errorf("got %q, want %q", key, "user:99:name:bob")
	}
}

func TestTemplateKeyBuilder_EmptyTemplate(t *testing.T) {
	_, err := Template("")
	if err == nil {
		t.Fatal("expected error for empty template, got nil")
	}
}

func TestTemplateKeyBuilder_ExtraFields(t *testing.T) {
	builder, err := Template("user:{userID}")
	if err != nil {
		t.Fatalf("unexpected error for template: %v", err)
	}
	key, err := builder.Build(context.Background(), map[string]interface{}{"userID": 1, "extra": 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "user:1" {
		t.Errorf("got %q, want %q", key, "user:1")
	}
}

func TestTemplateKeyBuilder_FieldTypes(t *testing.T) {
	builder, err := Template("foo:{int}:{str}:{bool}")
	if err != nil {
		t.Fatalf("unexpected error for template: %v", err)
	}
	key, err := builder.Build(context.Background(), map[string]interface{}{"int": 1, "str": "x", "bool": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "foo:1:x:true" {
		t.Errorf("got %q, want %q", key, "foo:1:x:true")
	}
}

func TestTemplateKeyBuilder_ContextIgnored(t *testing.T) {
	builder, err := Template("static")
	if err != nil {
		t.Fatalf("unexpected error for template: %v", err)
	}
	key, err := builder.Build(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "static" {
		t.Errorf("got %q, want %q", key, "static")
	}
}

func TestTemplateKeyBuilder_ReflectUnexported(t *testing.T) {
	type hidden struct {
		Visible string
		secret  string
	}
	builder, err := Template("{Visible}")
	if err != nil {
		t.Fatalf("unexpected error for template: %v", err)
	}
	key, err := builder.Build(context.Background(), hidden{Visible: "ok", secret: "nope"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "ok" {
		t.Errorf("got %q, want %q", key, "ok")
	}
}
