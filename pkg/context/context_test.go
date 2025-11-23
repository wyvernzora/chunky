package context

import (
	"context"
	"log/slog"
	"testing"
)

func TestWithFileInfo(t *testing.T) {
	ctx := context.Background()
	info := FileInfo{
		Path:    "test.md",
		Title:   "Test Document",
		Content: []byte("test content"),
	}

	ctx = WithFileInfo(ctx, info)

	retrieved, ok := FileInfoFrom(ctx)
	if !ok {
		t.Fatal("expected FileInfo in context")
	}

	if retrieved.Path != info.Path {
		t.Errorf("Path = %q, want %q", retrieved.Path, info.Path)
	}
	if retrieved.Title != info.Title {
		t.Errorf("Title = %q, want %q", retrieved.Title, info.Title)
	}
	if string(retrieved.Content) != string(info.Content) {
		t.Errorf("Content = %q, want %q", retrieved.Content, info.Content)
	}
}

func TestFileInfoFrom_Missing(t *testing.T) {
	ctx := context.Background()

	_, ok := FileInfoFrom(ctx)
	if ok {
		t.Error("expected no FileInfo in empty context")
	}
}

func TestFileInfoFrom_WrongType(t *testing.T) {
	ctx := context.Background()
	// Manually insert wrong type to test type assertion
	ctx = context.WithValue(ctx, fiKey, "wrong type")

	_, ok := FileInfoFrom(ctx)
	if ok {
		t.Error("expected FileInfoFrom to return false for wrong type")
	}
}

func TestMustFileInfo_Present(t *testing.T) {
	ctx := context.Background()
	info := FileInfo{
		Path:  "test.md",
		Title: "Test",
	}

	ctx = WithFileInfo(ctx, info)
	retrieved := MustFileInfo(ctx)

	if retrieved.Path != info.Path {
		t.Errorf("Path = %q, want %q", retrieved.Path, info.Path)
	}
	if retrieved.Title != info.Title {
		t.Errorf("Title = %q, want %q", retrieved.Title, info.Title)
	}
}

func TestMustFileInfo_Missing(t *testing.T) {
	ctx := context.Background()
	retrieved := MustFileInfo(ctx)

	// Should return zero value
	if retrieved.Path != "" {
		t.Errorf("expected empty Path, got %q", retrieved.Path)
	}
	if retrieved.Title != "" {
		t.Errorf("expected empty Title, got %q", retrieved.Title)
	}
	if retrieved.Content != nil {
		t.Error("expected nil Content")
	}
}

func TestFileInfo_EmptyValues(t *testing.T) {
	ctx := context.Background()
	info := FileInfo{
		Path:    "",
		Title:   "",
		Content: nil,
	}

	ctx = WithFileInfo(ctx, info)
	retrieved, ok := FileInfoFrom(ctx)

	if !ok {
		t.Fatal("expected FileInfo in context")
	}

	if retrieved.Path != "" {
		t.Errorf("expected empty Path")
	}
	if retrieved.Title != "" {
		t.Errorf("expected empty Title")
	}
	if retrieved.Content != nil {
		t.Error("expected nil Content")
	}
}

func TestFileInfo_LargeContent(t *testing.T) {
	ctx := context.Background()
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	info := FileInfo{
		Path:    "large.md",
		Title:   "Large Document",
		Content: largeContent,
	}

	ctx = WithFileInfo(ctx, info)
	retrieved, ok := FileInfoFrom(ctx)

	if !ok {
		t.Fatal("expected FileInfo in context")
	}

	if len(retrieved.Content) != len(largeContent) {
		t.Errorf("Content length = %d, want %d", len(retrieved.Content), len(largeContent))
	}
}

func TestLogger_Default(t *testing.T) {
	ctx := context.Background()
	logger := Logger(ctx)

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	// Should not panic when using logger
	logger.Info("test message")
}

func TestLogger_WithLogger(t *testing.T) {
	ctx := context.Background()

	// Create a custom logger
	customLogger := slog.Default()
	ctx = WithLogger(ctx, customLogger)

	logger := Logger(ctx)
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	// Should not panic when using logger
	logger.Info("test message with custom logger")
}

func TestLogger_Chained(t *testing.T) {
	ctx := context.Background()

	logger1 := slog.Default()
	ctx = WithLogger(ctx, logger1)

	// Add FileInfo after logger
	info := FileInfo{Path: "test.md", Title: "Test"}
	ctx = WithFileInfo(ctx, info)

	// Logger should still be accessible
	logger := Logger(ctx)
	if logger == nil {
		t.Fatal("expected non-nil logger after adding FileInfo")
	}

	// FileInfo should still be accessible
	retrieved, ok := FileInfoFrom(ctx)
	if !ok {
		t.Fatal("expected FileInfo after adding logger")
	}
	if retrieved.Path != info.Path {
		t.Error("FileInfo not preserved after logger")
	}
}

func TestWithFileInfo_Overwrite(t *testing.T) {
	ctx := context.Background()

	info1 := FileInfo{Path: "first.md", Title: "First"}
	ctx = WithFileInfo(ctx, info1)

	info2 := FileInfo{Path: "second.md", Title: "Second"}
	ctx = WithFileInfo(ctx, info2)

	retrieved, ok := FileInfoFrom(ctx)
	if !ok {
		t.Fatal("expected FileInfo in context")
	}

	// Should have the second info
	if retrieved.Path != "second.md" {
		t.Errorf("Path = %q, want \"second.md\"", retrieved.Path)
	}
	if retrieved.Title != "Second" {
		t.Errorf("Title = %q, want \"Second\"", retrieved.Title)
	}
}

func TestFileInfo_SpecialCharacters(t *testing.T) {
	ctx := context.Background()
	info := FileInfo{
		Path:    "path/with spaces/and-特殊字符.md",
		Title:   "Title with émojis 🎉 and symbols!@#$",
		Content: []byte("Content with\nnewlines\tand\ttabs"),
	}

	ctx = WithFileInfo(ctx, info)
	retrieved, ok := FileInfoFrom(ctx)

	if !ok {
		t.Fatal("expected FileInfo in context")
	}

	if retrieved.Path != info.Path {
		t.Errorf("Path not preserved with special characters")
	}
	if retrieved.Title != info.Title {
		t.Errorf("Title not preserved with special characters")
	}
	if string(retrieved.Content) != string(info.Content) {
		t.Errorf("Content not preserved with special characters")
	}
}
