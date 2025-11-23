package frontmatter

import (
	"reflect"
	"testing"
)

func TestClone(t *testing.T) {
	original := FrontMatter{
		"title":  "Test",
		"count":  42,
		"tags":   []string{"a", "b"},
		"nested": map[string]any{"key": "value"},
	}

	cloned := original.Clone()

	// Verify basic structure (note: deepCopyJSON via JSON converts int to float64)
	if cloned["title"] != "Test" {
		t.Errorf("title not cloned correctly")
	}
	if cloned["count"] != float64(42) { // JSON converts int to float64
		t.Errorf("count not cloned correctly, got %v (%T)", cloned["count"], cloned["count"])
	}

	// Verify it's a deep copy (modifying clone doesn't affect original)
	cloned["title"] = "Modified"
	if original["title"] == "Modified" {
		t.Error("modifying clone affected original - not a deep copy")
	}
}

func TestClone_Empty(t *testing.T) {
	original := FrontMatter{}
	cloned := original.Clone()

	if len(cloned) != 0 {
		t.Errorf("expected empty clone, got %d keys", len(cloned))
	}
}

func TestClone_Nil(t *testing.T) {
	var original FrontMatter
	cloned := original.Clone()

	if cloned == nil {
		t.Error("expected non-nil clone")
	}
	if len(cloned) != 0 {
		t.Errorf("expected empty clone, got %d keys", len(cloned))
	}
}

func TestGet(t *testing.T) {
	fm := FrontMatter{
		"string": "value",
		"number": 42,
		"bool":   true,
	}

	view := fm.View()

	tests := []struct {
		name    string
		key     string
		wantVal any
		wantOk  bool
	}{
		{"existing string", "string", "value", true},
		{"existing number", "number", float64(42), true}, // JSON converts int to float64
		{"existing bool", "bool", true, true},
		{"missing key", "missing", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := view.Get(tt.key)
			if ok != tt.wantOk {
				t.Errorf("Get(%q) ok = %v, want %v", tt.key, ok, tt.wantOk)
			}
			if !reflect.DeepEqual(val, tt.wantVal) {
				t.Errorf("Get(%q) val = %v, want %v", tt.key, val, tt.wantVal)
			}
		})
	}
}

func TestKeys(t *testing.T) {
	fm := FrontMatter{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	view := fm.View()
	keys := view.Keys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}

	// Verify all keys are present (order doesn't matter)
	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}

	for _, expected := range []string{"a", "b", "c"} {
		if !keySet[expected] {
			t.Errorf("missing expected key %q", expected)
		}
	}
}

func TestKeys_Empty(t *testing.T) {
	fm := FrontMatter{}
	view := fm.View()
	keys := view.Keys()

	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

func TestAsMap(t *testing.T) {
	fm := FrontMatter{
		"key1": "value1",
		"key2": 42,
	}

	view := fm.View()
	m := view.AsMap()

	// Verify basic structure (JSON serialization converts int to float64)
	if m["key1"] != "value1" {
		t.Error("AsMap did not preserve string value")
	}
	if m["key2"] != float64(42) { // JSON converts int to float64
		t.Errorf("AsMap did not preserve number value, got %v (%T)", m["key2"], m["key2"])
	}

	// Verify it's a copy (modifying returned map doesn't affect original)
	m["key1"] = "modified"
	if fm["key1"] == "modified" {
		t.Error("modifying AsMap result affected original - not a copy")
	}
}

func TestView(t *testing.T) {
	fm := FrontMatter{
		"title": "Test",
		"count": 42,
	}

	view := fm.View()

	// Test Get
	val, ok := view.Get("title")
	if !ok || val != "Test" {
		t.Errorf("View.Get(\"title\") = %v, %v; want \"Test\", true", val, ok)
	}

	// Test Keys
	keys := view.Keys()
	if len(keys) != 2 {
		t.Errorf("View.Keys() returned %d keys, want 2", len(keys))
	}

	// Test AsMap
	m := view.AsMap()
	if len(m) != 2 {
		t.Errorf("View.AsMap() returned %d entries, want 2", len(m))
	}

	// Verify modifying the view's map doesn't affect original
	m["title"] = "Modified"
	if fm["title"] == "Modified" {
		t.Error("modifying view map affected original frontmatter")
	}
}

func TestView_Empty(t *testing.T) {
	fm := FrontMatter{}
	view := fm.View()

	if len(view.Keys()) != 0 {
		t.Error("expected empty view")
	}
}

func TestSerialize(t *testing.T) {
	tests := []struct {
		name    string
		fm      FrontMatter
		wantErr bool
	}{
		{
			name: "simple types",
			fm: FrontMatter{
				"title":  "Test Document",
				"author": "John Doe",
				"year":   2024,
			},
			wantErr: false,
		},
		{
			name: "with arrays",
			fm: FrontMatter{
				"tags": []string{"tag1", "tag2", "tag3"},
			},
			wantErr: false,
		},
		{
			name: "nested maps",
			fm: FrontMatter{
				"metadata": map[string]any{
					"version": "1.0",
					"status":  "draft",
				},
			},
			wantErr: false,
		},
		{
			name:    "empty",
			fm:      FrontMatter{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml, err := Serialize(tt.fm)
			if (err != nil) != tt.wantErr {
				t.Errorf("Serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if yaml == "" && len(tt.fm) > 0 {
					t.Error("expected non-empty YAML for non-empty frontmatter")
				}
				// Verify it's valid YAML by checking for starting ---
				if len(tt.fm) > 0 && len(yaml) > 0 {
					if yaml[:3] != "---" {
						t.Errorf("expected YAML to start with ---, got: %s", yaml[:10])
					}
				}
			}
		})
	}
}

func TestSerialize_EmptyProducesEmptyString(t *testing.T) {
	fm := FrontMatter{}
	yaml, err := Serialize(fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if yaml != "" {
		t.Errorf("expected empty string for empty frontmatter, got: %q", yaml)
	}
}

func TestSerialize_WithSpecialCharacters(t *testing.T) {
	fm := FrontMatter{
		"description": "This has: colons and \"quotes\"",
		"path":        "some/path/to/file.md",
	}

	yaml, err := Serialize(fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if yaml == "" {
		t.Error("expected non-empty YAML")
	}
}

func TestDeepCopyJSON(t *testing.T) {
	original := map[string]any{
		"string": "value",
		"number": 42,
		"nested": map[string]any{
			"key": "value",
		},
		"array": []any{1, 2, 3},
	}

	copied := deepCopyJSON(original)

	// Verify it's not nil
	if copied == nil {
		t.Fatal("expected non-nil copy")
	}

	// Type assert to map
	copiedMap, ok := copied.(map[string]any)
	if !ok {
		t.Fatalf("expected copied to be map[string]any, got %T", copied)
	}

	// Verify it's a deep copy
	copiedMap["string"] = "modified"
	if original["string"] == "modified" {
		t.Error("modifying copy affected original")
	}

	// Verify nested structures are also copied
	if nested, ok := copiedMap["nested"].(map[string]any); ok {
		nested["key"] = "modified"
		if origNested, ok := original["nested"].(map[string]any); ok {
			if origNested["key"] == "modified" {
				t.Error("modifying nested copy affected original")
			}
		}
	}
}

func TestDeepCopyJSON_Nil(t *testing.T) {
	copied := deepCopyJSON(nil)
	if copied != nil {
		t.Error("expected nil copy of nil input")
	}
}

func TestDeepCopyJSON_Empty(t *testing.T) {
	original := map[string]any{}
	copied := deepCopyJSON(original)

	// deepCopyJSON may return nil for empty maps after JSON round-trip
	if copied != nil {
		if copiedMap, ok := copied.(map[string]any); ok && len(copiedMap) != 0 {
			t.Errorf("expected empty or nil copy, got %d keys", len(copiedMap))
		}
	}
}
