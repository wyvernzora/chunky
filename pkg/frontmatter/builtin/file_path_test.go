package builtin

import (
	"context"
	"testing"

	cctx "github.com/wyvernzora/chunky/pkg/context"
	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
)

func TestInjectFilePath_Success(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"title": "Test Document",
	}

	ctx := cctx.WithFileInfo(context.Background(), cctx.FileInfo{
		Path:  "/path/to/document.md",
		Title: "Test Document",
	})

	transform := InjectFilePath("")
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if path, ok := frontmatter["file_path"].(string); !ok || path != "/path/to/document.md" {
		t.Errorf("expected file_path '/path/to/document.md', got %v", frontmatter["file_path"])
	}
}

func TestInjectFilePath_CustomKey(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"title": "Test Document",
	}

	ctx := cctx.WithFileInfo(context.Background(), cctx.FileInfo{
		Path:  "/custom/path.md",
		Title: "Test Document",
	})

	transform := InjectFilePath("source_file")
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if path, ok := frontmatter["source_file"].(string); !ok || path != "/custom/path.md" {
		t.Errorf("expected source_file '/custom/path.md', got %v", frontmatter["source_file"])
	}

	// Ensure default key wasn't set
	if _, exists := frontmatter["file_path"]; exists {
		t.Error("default file_path key should not be set when using custom key")
	}
}

func TestInjectFilePath_AlreadyExists(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"file_path": "/existing/path.md",
	}

	ctx := cctx.WithFileInfo(context.Background(), cctx.FileInfo{
		Path:  "/new/path.md",
		Title: "Test",
	})

	transform := InjectFilePath("")
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should preserve existing value
	if path := frontmatter["file_path"].(string); path != "/existing/path.md" {
		t.Errorf("expected existing path to be preserved, got %v", path)
	}
}

func TestInjectFilePath_NilFrontMatter(t *testing.T) {
	ctx := cctx.WithFileInfo(context.Background(), cctx.FileInfo{
		Path:  "/path/to/document.md",
		Title: "Test",
	})

	transform := InjectFilePath("")
	err := transform(ctx, nil)
	if err == nil {
		t.Fatal("expected error for nil frontmatter, got nil")
	}

	expectedMsg := "frontmatter cannot be nil"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestInjectFilePath_NoFileInfo(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"title": "Test Document",
	}

	ctx := context.Background() // No FileInfo in context

	transform := InjectFilePath("")
	err := transform(ctx, frontmatter)
	if err == nil {
		t.Fatal("expected error for missing FileInfo, got nil")
	}

	expectedMsg := "file path not found in context"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestInjectFilePath_EmptyPath(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"title": "Test Document",
	}

	ctx := cctx.WithFileInfo(context.Background(), cctx.FileInfo{
		Path:  "", // Empty path
		Title: "Test",
	})

	transform := InjectFilePath("")
	err := transform(ctx, frontmatter)
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}

	expectedMsg := "file path not found in context"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

func TestInjectFilePath_Idempotent(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"title": "Test Document",
	}

	ctx := cctx.WithFileInfo(context.Background(), cctx.FileInfo{
		Path:  "/path/to/document.md",
		Title: "Test",
	})

	transform := InjectFilePath("")

	// First application
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("first application failed: %v", err)
	}

	firstPath := frontmatter["file_path"].(string)

	// Second application (should be no-op)
	err = transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("second application failed: %v", err)
	}

	secondPath := frontmatter["file_path"].(string)

	if firstPath != secondPath {
		t.Errorf("transform not idempotent: first=%q, second=%q", firstPath, secondPath)
	}
}

func TestInjectFilePath_PreservesOtherKeys(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"title":  "Test Document",
		"author": "John Doe",
		"tags":   []string{"test", "example"},
	}

	ctx := cctx.WithFileInfo(context.Background(), cctx.FileInfo{
		Path:  "/path/to/document.md",
		Title: "Test",
	})

	transform := InjectFilePath("")
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check original keys preserved
	if frontmatter["title"] != "Test Document" {
		t.Errorf("title was modified")
	}
	if frontmatter["author"] != "John Doe" {
		t.Errorf("author was modified")
	}
	if len(frontmatter) != 4 { // original 3 + file_path
		t.Errorf("unexpected number of keys: %d", len(frontmatter))
	}
}

func TestInjectFilePath_EmptyFrontMatter(t *testing.T) {
	frontmatter := fm.EmptyFrontMatter()

	ctx := cctx.WithFileInfo(context.Background(), cctx.FileInfo{
		Path:  "/path/to/document.md",
		Title: "Test",
	})

	transform := InjectFilePath("")
	err := transform(ctx, frontmatter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(frontmatter) != 1 {
		t.Errorf("expected 1 key in frontmatter, got %d", len(frontmatter))
	}

	if path := frontmatter["file_path"].(string); path != "/path/to/document.md" {
		t.Errorf("expected path '/path/to/document.md', got %q", path)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOfSubstring(s, substr) >= 0))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
