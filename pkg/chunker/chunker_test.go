package chunker

import (
	"context"
	"errors"
	"strings"
	"testing"

	cctx "github.com/wyvernzora/chunky/pkg/context"
	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
	"github.com/wyvernzora/chunky/pkg/tokenizer"
	tbuiltin "github.com/wyvernzora/chunky/pkg/tokenizer/builtin"
)

// TestNew_Success tests successful chunker creation with valid configuration
func TestNew_Success(t *testing.T) {
	c, err := New(WithChunkTokenBudget(1000))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil chunker")
	}
}

// TestNew_MissingBudget tests that New returns error when budget is not set
func TestNew_MissingBudget(t *testing.T) {
	_, err := New()
	if err == nil {
		t.Fatal("expected error for missing budget")
	}
	if !strings.Contains(err.Error(), "WithChunkTokenBudget is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestNew_InvalidBudget tests that New returns error for invalid budget values
func TestNew_InvalidBudget(t *testing.T) {
	tests := []struct {
		name   string
		budget int
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(WithChunkTokenBudget(tt.budget))
			if err == nil {
				t.Fatalf("expected error for budget %d", tt.budget)
			}
			if !strings.Contains(err.Error(), "WithChunkTokenBudget") {
				t.Errorf("unexpected error message: %v", err)
			}
		})
	}
}

// TestNew_InvalidOverheadRatio tests that New returns error for invalid ratios
func TestNew_InvalidOverheadRatio(t *testing.T) {
	tests := []struct {
		name  string
		ratio float64
	}{
		{"negative", -0.1},
		{"exactly one", 1.0},
		{"greater than one", 1.5},
		{"large negative", -10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(
				WithChunkTokenBudget(1000),
				WithReservedOverheadRatio(tt.ratio),
			)
			if err == nil {
				t.Fatalf("expected error for ratio %f", tt.ratio)
			}
			if !strings.Contains(err.Error(), "WithReservedOverheadRatio") {
				t.Errorf("unexpected error message: %v", err)
			}
		})
	}
}

// TestNew_ValidOverheadRatio tests that New accepts valid ratios
func TestNew_ValidOverheadRatio(t *testing.T) {
	tests := []struct {
		name  string
		ratio float64
	}{
		{"zero", 0.0},
		{"small", 0.05},
		{"default", 0.1},
		{"large", 0.5},
		{"almost one", 0.99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(
				WithChunkTokenBudget(1000),
				WithReservedOverheadRatio(tt.ratio),
			)
			if err != nil {
				t.Fatalf("unexpected error for ratio %f: %v", tt.ratio, err)
			}
			if c == nil {
				t.Fatal("expected non-nil chunker")
			}
		})
	}
}

// TestNew_CustomTokenizer tests that custom tokenizer is used
func TestNew_CustomTokenizer(t *testing.T) {
	tok := tbuiltin.NewWordCountTokenizer()

	c, err := New(
		WithChunkTokenBudget(1000),
		WithTokenizer(tok),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil chunker")
	}
}

// TestNew_CustomParser tests that custom parser is used
func TestNew_CustomParser(t *testing.T) {
	customParser := func(ctx context.Context, markdown []byte) (*section.Section, fm.FrontMatter, error) {
		root := section.NewRoot("Custom")
		root.SetContent(string(markdown))
		return root, fm.FrontMatter{}, nil
	}

	c, err := New(
		WithChunkTokenBudget(1000),
		WithParser(customParser),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "# Test",
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}
}

// TestNew_CustomHeaderGenerator tests that custom header generator is used
func TestNew_CustomHeaderGenerator(t *testing.T) {
	customGen := func(ctx context.Context, frontmatter fm.FrontMatterView) (string, error) {
		return "CUSTOM HEADER\n", nil
	}

	c, err := New(
		WithChunkTokenBudget(1000),
		WithChunkHeader(customGen),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "# Test\n\nContent",
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	chunks := c.Chunks()
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	if !strings.Contains(chunks[0].Text, "CUSTOM HEADER") {
		t.Errorf("expected custom header in chunk text, got: %s", chunks[0].Text)
	}
}

// TestPush_EmptyPath tests that Push validates empty path
func TestPush_EmptyPath(t *testing.T) {
	c, err := New(WithChunkTokenBudget(1000))
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "",
		Title:    "Test",
		Markdown: "# Test",
	})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !strings.Contains(err.Error(), "Path cannot be empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestPush_EmptyTitle tests that Push validates empty title
func TestPush_EmptyTitle(t *testing.T) {
	c, err := New(WithChunkTokenBudget(1000))
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "",
		Markdown: "# Test",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	if !strings.Contains(err.Error(), "Title cannot be empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestPush_EmptyMarkdown tests that Push validates empty markdown
func TestPush_EmptyMarkdown(t *testing.T) {
	c, err := New(WithChunkTokenBudget(1000))
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "",
	})
	if err == nil {
		t.Fatal("expected error for empty markdown")
	}
	if !strings.Contains(err.Error(), "Markdown cannot be empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestPush_Success tests successful document processing
func TestPush_Success(t *testing.T) {
	c, err := New(WithChunkTokenBudget(1000))
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "# Heading\n\nSome content here.",
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	chunks := c.Chunks()
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
}

// TestPush_DoNotEmbed tests that documents with do_not_embed flag are skipped
func TestPush_DoNotEmbed(t *testing.T) {
	c, err := New(WithChunkTokenBudget(1000))
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:  "test.md",
		Title: "Test",
		Markdown: `---
do_not_embed: true
---
# Heading

Content`,
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	chunks := c.Chunks()
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for do_not_embed, got %d", len(chunks))
	}
}

// TestPush_MultipleDocs tests pushing multiple documents
func TestPush_MultipleDocs(t *testing.T) {
	c, err := New(WithChunkTokenBudget(1000))
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	// Push first doc
	err = c.Push(context.Background(), Input{
		Path:     "doc1.md",
		Title:    "Doc 1",
		Markdown: "# Doc 1\n\nContent 1",
	})
	if err != nil {
		t.Fatalf("Push 1 failed: %v", err)
	}

	count1 := len(c.Chunks())
	if count1 == 0 {
		t.Fatal("expected chunks after first push")
	}

	// Push second doc
	err = c.Push(context.Background(), Input{
		Path:     "doc2.md",
		Title:    "Doc 2",
		Markdown: "# Doc 2\n\nContent 2",
	})
	if err != nil {
		t.Fatalf("Push 2 failed: %v", err)
	}

	count2 := len(c.Chunks())
	if count2 <= count1 {
		t.Errorf("expected more chunks after second push: %d <= %d", count2, count1)
	}
}

// TestPush_ContextCancellation tests that Push respects context cancellation
func TestPush_ContextCancellation(t *testing.T) {
	c, err := New(WithChunkTokenBudget(1000))
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = c.Push(ctx, Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "# Test\n\nContent",
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !strings.Contains(err.Error(), "context cancel") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestPush_ParserError tests that parser errors are propagated
func TestPush_ParserError(t *testing.T) {
	expectedErr := errors.New("parser error")
	customParser := func(ctx context.Context, markdown []byte) (*section.Section, fm.FrontMatter, error) {
		return nil, nil, expectedErr
	}

	c, err := New(
		WithChunkTokenBudget(1000),
		WithParser(customParser),
	)
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "# Test",
	})
	if err == nil {
		t.Fatal("expected error from parser")
	}
	if !strings.Contains(err.Error(), "parse failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestPush_TransformError tests that transform errors are propagated
func TestPush_TransformError(t *testing.T) {
	expectedErr := errors.New("transform error")
	errorTransform := func(ctx context.Context, frontmatter fm.FrontMatter) error {
		return expectedErr
	}

	c, err := New(
		WithChunkTokenBudget(1000),
		WithFrontMatterTransform(errorTransform),
	)
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "# Test",
	})
	if err == nil {
		t.Fatal("expected error from transform")
	}
	if !strings.Contains(err.Error(), "frontmatter transform") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestPush_FrontmatterExceedsBudget tests error when frontmatter is too large
func TestPush_FrontmatterExceedsBudget(t *testing.T) {
	// Create a chunker with very small budget
	c, err := New(WithChunkTokenBudget(10))
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	// Push document with large frontmatter
	err = c.Push(context.Background(), Input{
		Path:  "test.md",
		Title: "Test",
		Markdown: `---
long_field: This is a very long field that will exceed the token budget when serialized to YAML format
another_field: More content to make it even larger
yet_another: Even more text
---
# Test`,
	})
	if err == nil {
		t.Fatal("expected error when frontmatter exceeds budget")
	}
	if !strings.Contains(err.Error(), "exceeds effective budget") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestReset tests that Reset clears chunks
func TestReset(t *testing.T) {
	c, err := New(WithChunkTokenBudget(1000))
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	// Push a document
	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "# Test\n\nContent",
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	if len(c.Chunks()) == 0 {
		t.Fatal("expected chunks before reset")
	}

	// Reset
	c.Reset()

	if len(c.Chunks()) != 0 {
		t.Errorf("expected 0 chunks after reset, got %d", len(c.Chunks()))
	}
}

// TestChunks tests that Chunks returns accumulated chunks
func TestChunks(t *testing.T) {
	c, err := New(WithChunkTokenBudget(1000))
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	// Initially empty
	if len(c.Chunks()) != 0 {
		t.Errorf("expected 0 chunks initially, got %d", len(c.Chunks()))
	}

	// Push document
	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "# Test\n\nContent",
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	chunks := c.Chunks()
	if len(chunks) == 0 {
		t.Fatal("expected chunks after push")
	}

	// Verify chunk fields
	chunk := chunks[0]
	if chunk.FilePath != "test.md" {
		t.Errorf("expected FilePath 'test.md', got %q", chunk.FilePath)
	}
	if chunk.FileTitle != "Test" {
		t.Errorf("expected FileTitle 'Test', got %q", chunk.FileTitle)
	}
	if chunk.ChunkIndex != 1 {
		t.Errorf("expected ChunkIndex 1, got %d", chunk.ChunkIndex)
	}
	if chunk.Tokens == 0 {
		t.Error("expected non-zero tokens")
	}
	if chunk.Text == "" {
		t.Error("expected non-empty text")
	}
}

// TestEffectiveBudget tests that EffectiveBudget returns correct value
func TestEffectiveBudget(t *testing.T) {
	tests := []struct {
		name           string
		budget         int
		overheadRatio  float64
		expectedResult int
	}{
		{"default ratio", 1000, 0.1, 900},
		{"zero ratio", 1000, 0.0, 1000},
		{"high ratio", 1000, 0.5, 500},
		{"small budget", 100, 0.1, 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(
				WithChunkTokenBudget(tt.budget),
				WithReservedOverheadRatio(tt.overheadRatio),
			)
			if err != nil {
				t.Fatalf("failed to create chunker: %v", err)
			}

			result := c.EffectiveBudget()
			if result != tt.expectedResult {
				t.Errorf("expected %d, got %d", tt.expectedResult, result)
			}
		})
	}
}

// TestWithFrontMatterTransforms tests batch transform addition
func TestWithFrontMatterTransforms(t *testing.T) {
	callCount := 0
	transform1 := func(ctx context.Context, fm fm.FrontMatter) error {
		callCount++
		fm["key1"] = "value1"
		return nil
	}
	transform2 := func(ctx context.Context, fm fm.FrontMatter) error {
		callCount++
		fm["key2"] = "value2"
		return nil
	}

	c, err := New(
		WithChunkTokenBudget(1000),
		WithFrontMatterTransforms(transform1, transform2),
	)
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "# Test",
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	if callCount < 2 {
		t.Errorf("expected at least 2 transform calls, got %d", callCount)
	}
}

// TestWithSectionTransforms tests batch section transform addition
func TestWithSectionTransforms(t *testing.T) {
	callCount := 0
	transform1 := section.Transform(func(ctx context.Context, fmView fm.FrontMatterView, s *section.Section) error {
		callCount++
		return nil
	})
	transform2 := section.Transform(func(ctx context.Context, fmView fm.FrontMatterView, s *section.Section) error {
		callCount++
		return nil
	})

	c, err := New(
		WithChunkTokenBudget(1000),
		WithSectionTransforms(transform1, transform2),
	)
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "# Test",
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	if callCount < 2 {
		t.Errorf("expected at least 2 transform calls, got %d", callCount)
	}
}

// TestChunkMetadata tests that chunk metadata is correctly set
func TestChunkMetadata(t *testing.T) {
	c, err := New(WithChunkTokenBudget(1000))
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "docs/test.md",
		Title:    "Test Document",
		Markdown: "# Test\n\nContent",
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	chunks := c.Chunks()
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}

	chunk := chunks[0]
	if chunk.FilePath != "docs/test.md" {
		t.Errorf("expected FilePath 'docs/test.md', got %q", chunk.FilePath)
	}
	if chunk.FileTitle != "Test Document" {
		t.Errorf("expected FileTitle 'Test Document', got %q", chunk.FileTitle)
	}
	if !strings.Contains(chunk.Text, "file_path: docs/test.md") {
		t.Error("expected file_path in chunk text")
	}
}

// TestContextPropagation tests that context values are accessible in transforms
func TestContextPropagation(t *testing.T) {
	var receivedInfo cctx.FileInfo
	transform := func(ctx context.Context, fm fm.FrontMatter) error {
		info, ok := cctx.FileInfoFrom(ctx)
		if !ok {
			return errors.New("FileInfo not in context")
		}
		receivedInfo = info
		return nil
	}

	c, err := New(
		WithChunkTokenBudget(1000),
		WithFrontMatterTransform(transform),
	)
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test Title",
		Markdown: "# Test",
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	if receivedInfo.Path != "test.md" {
		t.Errorf("expected Path 'test.md', got %q", receivedInfo.Path)
	}
	if receivedInfo.Title != "Test Title" {
		t.Errorf("expected Title 'Test Title', got %q", receivedInfo.Title)
	}
}

// TestTokenizerError tests that tokenizer errors are propagated
func TestTokenizerError(t *testing.T) {
	expectedErr := errors.New("tokenizer error")

	// Create a mock tokenizer that returns an error
	mockTokenizer := &mockTokenizerError{err: expectedErr}

	c, err := New(
		WithChunkTokenBudget(1000),
		WithTokenizer(mockTokenizer),
	)
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	err = c.Push(context.Background(), Input{
		Path:     "test.md",
		Title:    "Test",
		Markdown: "# Test\n\nContent",
	})
	if err == nil {
		t.Fatal("expected error from tokenizer")
	}
}

// mockTokenizerError is a mock tokenizer that returns an error
type mockTokenizerError struct {
	err error
}

func (m *mockTokenizerError) Count(text string) (int, error) {
	return 0, m.err
}

func (m *mockTokenizerError) Tokenize(ctx context.Context, root *section.Section) (*tokenizer.TokenizedSection, error) {
	return nil, m.err
}
