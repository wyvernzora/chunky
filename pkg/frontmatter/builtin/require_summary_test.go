package builtin

import (
	"context"
	"strings"
	"testing"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
)

func TestRequireSummary_Success(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"title":   "Test Document",
		"summary": "This is a valid summary.",
	}

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRequireSummary_LongSummary(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"summary": "This is a much longer summary that spans multiple sentences. " +
			"It provides detailed information about the document. " +
			"This should still be valid.",
	}

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("expected no error for long summary, got %v", err)
	}
}

func TestRequireSummary_SingleWord(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"summary": "Valid",
	}

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("expected no error for single word, got %v", err)
	}
}

func TestRequireSummary_WithLeadingTrailingWhitespace(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"summary": "  Valid summary with whitespace  ",
	}

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRequireSummary_NilFrontMatter(t *testing.T) {
	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil frontmatter, got nil")
	}

	expectedMsg := "frontmatter cannot be nil"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestRequireSummary_MissingKey(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"title":  "Test Document",
		"author": "John Doe",
	}

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err == nil {
		t.Fatal("expected error for missing summary key, got nil")
	}

	expectedMsg := "missing required 'summary' field"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestRequireSummary_EmptyString(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"summary": "",
	}

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err == nil {
		t.Fatal("expected error for empty summary, got nil")
	}

	expectedMsg := "'summary' field cannot be empty"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestRequireSummary_WhitespaceOnly(t *testing.T) {
	testCases := []struct {
		name    string
		summary string
	}{
		{"spaces", "   "},
		{"tabs", "\t\t\t"},
		{"newlines", "\n\n\n"},
		{"mixed", " \t \n \t "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			frontmatter := fm.FrontMatter{
				"summary": tc.summary,
			}

			ctx := context.Background()
			transform := RequireSummary()
			err := transform(ctx, frontmatter)
			if err == nil {
				t.Fatalf("expected error for whitespace-only summary %q, got nil", tc.summary)
			}

			expectedMsg := "'summary' field cannot be empty"
			if !strings.Contains(err.Error(), expectedMsg) {
				t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
			}
		})
	}
}

func TestRequireSummary_WrongType_Int(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"summary": 123,
	}

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err == nil {
		t.Fatal("expected error for int summary, got nil")
	}

	expectedMsg := "'summary' field must be a string"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestRequireSummary_WrongType_Bool(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"summary": true,
	}

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err == nil {
		t.Fatal("expected error for bool summary, got nil")
	}

	expectedMsg := "'summary' field must be a string"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestRequireSummary_WrongType_Array(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"summary": []string{"line1", "line2"},
	}

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err == nil {
		t.Fatal("expected error for array summary, got nil")
	}

	expectedMsg := "'summary' field must be a string"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestRequireSummary_WrongType_Map(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"summary": map[string]string{"text": "summary"},
	}

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err == nil {
		t.Fatal("expected error for map summary, got nil")
	}

	expectedMsg := "'summary' field must be a string"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestRequireSummary_EmptyFrontMatter(t *testing.T) {
	frontmatter := fm.EmptyFrontMatter()

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err == nil {
		t.Fatal("expected error for empty frontmatter, got nil")
	}

	expectedMsg := "missing required 'summary' field"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestRequireSummary_DoesNotModifyFrontMatter(t *testing.T) {
	originalSummary := "This is the original summary."
	frontmatter := fm.FrontMatter{
		"title":   "Test",
		"summary": originalSummary,
		"author":  "John Doe",
	}

	ctx := context.Background()
	transform := RequireSummary()
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify no modifications
	if len(frontmatter) != 3 {
		t.Errorf("expected 3 keys, got %d", len(frontmatter))
	}

	if frontmatter["summary"] != originalSummary {
		t.Errorf("summary was modified: expected %q, got %q", originalSummary, frontmatter["summary"])
	}

	if frontmatter["title"] != "Test" {
		t.Error("title was modified")
	}

	if frontmatter["author"] != "John Doe" {
		t.Error("author was modified")
	}
}

func TestRequireSummary_Idempotent(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"summary": "Valid summary",
	}

	ctx := context.Background()
	transform := RequireSummary()

	// First application
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("first application failed: %v", err)
	}

	// Second application
	err = transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("second application failed: %v", err)
	}

	// Verify summary unchanged
	if frontmatter["summary"] != "Valid summary" {
		t.Error("summary was modified across multiple applications")
	}
}

func TestRequireSummary_WithSpecialCharacters(t *testing.T) {
	testCases := []struct {
		name    string
		summary string
	}{
		{"with quotes", `Summary with "quotes" included`},
		{"with apostrophes", "Summary with 'apostrophes' included"},
		{"with unicode", "Summary with unicode: 你好世界 🌍"},
		{"with punctuation", "Summary with punctuation: comma, period. exclamation! question?"},
		{"with newline", "Summary with\nnewline character"},
		{"with emoji", "Summary with emoji 🚀 🎉 ✨"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			frontmatter := fm.FrontMatter{
				"summary": tc.summary,
			}

			ctx := context.Background()
			transform := RequireSummary()
			err := transform(ctx, frontmatter)
			if err != nil {
				t.Fatalf("expected no error for summary %q, got %v", tc.summary, err)
			}
		})
	}
}
