package builtin

import (
	"context"
	"strings"
	"testing"

	"github.com/wyvernzora/chunky/pkg/frontmatter"
)

func TestKeyValueHeader_Basic(t *testing.T) {
	gen := KeyValueHeader(
		RequiredField("title", "Title"),
		OptionalField("author", "Author"),
	)

	fm := frontmatter.FrontMatter{
		"title":  "Test Document",
		"author": "John Doe",
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Title: Test Document\nAuthor: John Doe\n\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestKeyValueHeader_MissingRequired(t *testing.T) {
	gen := KeyValueHeader(
		RequiredField("title", "Title"),
		OptionalField("author", "Author"),
	)

	fm := frontmatter.FrontMatter{
		"author": "John Doe",
	}

	_, err := gen(context.Background(), fm.View())
	if err == nil {
		t.Fatal("expected error for missing required field")
	}
	if !strings.Contains(err.Error(), "required field missing or empty: title") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestKeyValueHeader_EmptyRequired(t *testing.T) {
	gen := KeyValueHeader(
		RequiredField("title", "Title"),
	)

	fm := frontmatter.FrontMatter{
		"title": "   ", // whitespace only
	}

	_, err := gen(context.Background(), fm.View())
	if err == nil {
		t.Fatal("expected error for empty required field")
	}
	if !strings.Contains(err.Error(), "required field missing or empty: title") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestKeyValueHeader_MissingOptional(t *testing.T) {
	gen := KeyValueHeader(
		RequiredField("title", "Title"),
		OptionalField("author", "Author"),
		OptionalField("date", "Date"),
	)

	fm := frontmatter.FrontMatter{
		"title": "Test Document",
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Title: Test Document\n\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestKeyValueHeader_EmptyOptional(t *testing.T) {
	gen := KeyValueHeader(
		RequiredField("title", "Title"),
		OptionalField("author", "Author"),
	)

	fm := frontmatter.FrontMatter{
		"title":  "Test Document",
		"author": "",
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty optional field should be skipped
	expected := "Title: Test Document\n\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestKeyValueHeader_SliceValues(t *testing.T) {
	gen := KeyValueHeader(
		RequiredField("title", "Title"),
		OptionalField("tags", "Tags"),
	)

	fm := frontmatter.FrontMatter{
		"title": "Test Document",
		"tags":  []any{"go", "testing", "documentation"},
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain the title
	if !strings.Contains(result, "Title: Test Document") {
		t.Errorf("result missing title: %q", result)
	}

	// Should contain tags array
	if !strings.Contains(result, "Tags:") {
		t.Errorf("result missing tags: %q", result)
	}
}

func TestKeyValueHeader_EmptySliceSkipped(t *testing.T) {
	gen := KeyValueHeader(
		RequiredField("title", "Title"),
		OptionalField("tags", "Tags"),
	)

	fm := frontmatter.FrontMatter{
		"title": "Test Document",
		"tags":  []any{},
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty slice should be skipped
	expected := "Title: Test Document\n\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestKeyValueHeader_InvalidMapValue(t *testing.T) {
	gen := KeyValueHeader(
		RequiredField("title", "Title"),
		OptionalField("metadata", "Metadata"),
	)

	fm := frontmatter.FrontMatter{
		"title":    "Test Document",
		"metadata": map[string]any{"key": "value"},
	}

	_, err := gen(context.Background(), fm.View())
	if err == nil {
		t.Fatal("expected error for map value")
	}
	if !strings.Contains(err.Error(), "value must be a scalar or slice of scalars") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestKeyValueHeader_InvalidSliceElement(t *testing.T) {
	gen := KeyValueHeader(
		RequiredField("title", "Title"),
		OptionalField("tags", "Tags"),
	)

	fm := frontmatter.FrontMatter{
		"title": "Test Document",
		"tags":  []any{"valid", map[string]any{"nested": "invalid"}},
	}

	_, err := gen(context.Background(), fm.View())
	if err == nil {
		t.Fatal("expected error for slice with non-scalar element")
	}
	if !strings.Contains(err.Error(), "slice element at index") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestKeyValueHeader_LabelTrimming(t *testing.T) {
	gen := KeyValueHeader(
		RequiredField("title", "  Title  "),
	)

	fm := frontmatter.FrontMatter{
		"title": "Test Document",
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Label should be trimmed
	expected := "Title: Test Document\n\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"whitespace string", "   ", true},
		{"non-empty string", "hello", false},
		{"empty slice", []any{}, true},
		{"non-empty slice", []any{"item"}, false},
		{"zero int", 0, false},
		{"non-zero int", 42, false},
		{"false bool", false, false},
		{"true bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEmpty(tt.value)
			if got != tt.want {
				t.Errorf("isEmpty(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValidateValue(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"nil", nil, false},
		{"string", "hello", false},
		{"int", 42, false},
		{"bool", true, false},
		{"float", 3.14, false},
		{"slice of strings", []any{"a", "b", "c"}, false},
		{"slice of ints", []any{1, 2, 3}, false},
		{"slice of mixed scalars", []any{"a", 1, true, 3.14}, false},
		{"empty slice", []any{}, false},
		{"slice with nil", []any{"valid", nil}, false},
		{"map", map[string]any{"key": "value"}, true},
		{"slice with map", []any{"valid", map[string]any{"nested": "invalid"}}, true},
		{"slice with slice", []any{"valid", []any{"nested"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateValue(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsScalar(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, true},
		{"string", "hello", true},
		{"int", 42, true},
		{"int8", int8(42), true},
		{"int16", int16(42), true},
		{"int32", int32(42), true},
		{"int64", int64(42), true},
		{"uint", uint(42), true},
		{"uint8", uint8(42), true},
		{"uint16", uint16(42), true},
		{"uint32", uint32(42), true},
		{"uint64", uint64(42), true},
		{"float32", float32(3.14), true},
		{"float64", float64(3.14), true},
		{"bool true", true, true},
		{"bool false", false, true},
		{"slice", []any{"item"}, false},
		{"map", map[string]any{"key": "value"}, false},
		{"struct", struct{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isScalar(tt.value)
			if got != tt.want {
				t.Errorf("isScalar(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
