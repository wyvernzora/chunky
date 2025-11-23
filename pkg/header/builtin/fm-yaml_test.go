package builtin

import (
	"context"
	"strings"
	"testing"

	"github.com/wyvernzora/chunky/pkg/frontmatter"
)

func TestFrontMatterYamlHeader_Basic(t *testing.T) {
	gen := FrontMatterYamlHeader()

	fm := frontmatter.FrontMatter{
		"title":  "Test Document",
		"author": "John Doe",
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should start with ---
	if !strings.HasPrefix(result, "---\n") {
		t.Errorf("expected result to start with '---\\n', got: %q", result)
	}

	// Should end with ---\n
	if !strings.HasSuffix(result, "---\n") {
		t.Errorf("expected result to end with '---\\n', got: %q", result)
	}

	// Should contain the fields
	if !strings.Contains(result, "title:") {
		t.Errorf("result missing 'title:': %q", result)
	}
	if !strings.Contains(result, "author:") {
		t.Errorf("result missing 'author:': %q", result)
	}
}

func TestFrontMatterYamlHeader_Empty(t *testing.T) {
	gen := FrontMatterYamlHeader()

	fm := frontmatter.FrontMatter{}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty frontmatter should return empty string
	if result != "" {
		t.Errorf("expected empty string for empty frontmatter, got: %q", result)
	}
}

func TestFrontMatterYamlHeader_SingleField(t *testing.T) {
	gen := FrontMatterYamlHeader()

	fm := frontmatter.FrontMatter{
		"title": "Test Document",
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "---\ntitle: Test Document\n---\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFrontMatterYamlHeader_MultipleFields(t *testing.T) {
	gen := FrontMatterYamlHeader()

	fm := frontmatter.FrontMatter{
		"title":  "Test Document",
		"author": "John Doe",
		"date":   "2024-01-01",
		"draft":  false,
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain all fields
	fields := []string{"title:", "author:", "date:", "draft:"}
	for _, field := range fields {
		if !strings.Contains(result, field) {
			t.Errorf("result missing field %q: %q", field, result)
		}
	}
}

func TestFrontMatterYamlHeader_ComplexTypes(t *testing.T) {
	gen := FrontMatterYamlHeader()

	fm := frontmatter.FrontMatter{
		"title": "Test Document",
		"tags":  []any{"go", "testing", "yaml"},
		"metadata": map[string]any{
			"version": "1.0",
			"status":  "draft",
		},
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be valid YAML format
	if !strings.HasPrefix(result, "---\n") {
		t.Errorf("expected result to start with '---\\n', got: %q", result)
	}

	// Should contain title
	if !strings.Contains(result, "title:") {
		t.Errorf("result missing 'title:': %q", result)
	}

	// Should contain tags
	if !strings.Contains(result, "tags:") {
		t.Errorf("result missing 'tags:': %q", result)
	}

	// Should contain metadata
	if !strings.Contains(result, "metadata:") {
		t.Errorf("result missing 'metadata:': %q", result)
	}
}

func TestFrontMatterYamlHeader_NumericValues(t *testing.T) {
	gen := FrontMatterYamlHeader()

	fm := frontmatter.FrontMatter{
		"count":   42,
		"price":   19.99,
		"enabled": true,
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain all fields
	if !strings.Contains(result, "count:") {
		t.Errorf("result missing 'count:': %q", result)
	}
	if !strings.Contains(result, "price:") {
		t.Errorf("result missing 'price:': %q", result)
	}
	if !strings.Contains(result, "enabled:") {
		t.Errorf("result missing 'enabled:': %q", result)
	}
}

func TestFrontMatterYamlHeader_SpecialCharacters(t *testing.T) {
	gen := FrontMatterYamlHeader()

	fm := frontmatter.FrontMatter{
		"title":       "Test: Document with special chars",
		"description": "A document with\nmultiple lines",
	}

	result, err := gen(context.Background(), fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be valid YAML
	if !strings.HasPrefix(result, "---\n") {
		t.Errorf("expected result to start with '---\\n', got: %q", result)
	}
	if !strings.HasSuffix(result, "---\n") {
		t.Errorf("expected result to end with '---\\n', got: %q", result)
	}
}

func TestFrontMatterYamlHeader_ContextPropagation(t *testing.T) {
	gen := FrontMatterYamlHeader()

	// Create a context with a value
	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	fm := frontmatter.FrontMatter{
		"title": "Test Document",
	}

	result, err := gen(ctx, fm.View())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still work with custom context
	if !strings.Contains(result, "title:") {
		t.Errorf("result missing 'title:': %q", result)
	}
}
