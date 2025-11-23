package builtin

import (
	"context"
	"strings"
	"testing"

	"github.com/wyvernzora/chunky/pkg/section"
)

func TestNewTiktokenTokenizer_Default(t *testing.T) {
	tok, err := NewTiktokenTokenizer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok == nil {
		t.Fatal("NewTiktokenTokenizer returned nil")
	}

	count, err := tok.Count("hello world")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count == 0 {
		t.Error("expected non-zero token count")
	}
}

func TestNewTiktokenTokenizer_CustomEncoding(t *testing.T) {
	testCases := []struct {
		name     string
		encoding string
	}{
		{"o200k_base", "o200k_base"},
		{"cl100k_base", "cl100k_base"},
		{"p50k_base", "p50k_base"},
		{"r50k_base", "r50k_base"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tok, err := NewTiktokenTokenizer(WithEncoding(tc.encoding))
			if err != nil {
				t.Fatalf("failed to create tokenizer with %s: %v", tc.encoding, err)
			}

			count, err := tok.Count("test")
			if err != nil {
				t.Fatalf("Count failed: %v", err)
			}
			if count == 0 {
				t.Error("expected non-zero token count")
			}
		})
	}
}

func TestNewTiktokenTokenizer_InvalidEncoding(t *testing.T) {
	_, err := NewTiktokenTokenizer(WithEncoding("invalid_encoding_name"))
	if err == nil {
		t.Fatal("expected error for invalid encoding, got nil")
	}

	if !strings.Contains(err.Error(), "invalid_encoding_name") {
		t.Errorf("error should mention encoding name, got: %v", err)
	}
}

func TestNewTiktokenTokenizer_MultipleOptions(t *testing.T) {
	// Last option should win
	tok, err := NewTiktokenTokenizer(
		WithEncoding("p50k_base"),
		WithEncoding("cl100k_base"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should work with cl100k_base encoding
	count, err := tok.Count("test")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count == 0 {
		t.Error("expected non-zero token count")
	}
}

func TestNewTiktokenTokenizer_EmptyEncodingOption(t *testing.T) {
	// Empty encoding should be ignored, falling back to default
	tok, err := NewTiktokenTokenizer(WithEncoding(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count, err := tok.Count("test")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count == 0 {
		t.Error("expected non-zero token count")
	}
}

func TestTiktokenTokenizer_Count_Empty(t *testing.T) {
	tok, err := NewTiktokenTokenizer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count, err := tok.Count("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tokens for empty string, got %d", count)
	}
}

func TestTiktokenTokenizer_Count_SimpleText(t *testing.T) {
	tok, err := NewTiktokenTokenizer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testCases := []struct {
		text string
	}{
		{"hello"},
		{"hello world"},
		{"The quick brown fox jumps over the lazy dog."},
		{"1234567890"},
	}

	for _, tc := range testCases {
		t.Run(tc.text, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count == 0 {
				t.Error("expected non-zero token count")
			}
			// Sanity check: tokens should be reasonable
			if count > len(tc.text) {
				t.Errorf("token count %d exceeds character count %d", count, len(tc.text))
			}
		})
	}
}

func TestTiktokenTokenizer_Count_Unicode(t *testing.T) {
	tok, err := NewTiktokenTokenizer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testCases := []string{
		"你好世界",
		"Hello 世界",
		"🌍🚀🎉",
		"Привет мир",
		"مرحبا بالعالم",
	}

	for _, text := range testCases {
		t.Run(text, func(t *testing.T) {
			count, err := tok.Count(text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count == 0 {
				t.Error("expected non-zero token count")
			}
		})
	}
}

func TestTiktokenTokenizer_Count_Code(t *testing.T) {
	tok, err := NewTiktokenTokenizer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	code := `func main() {
    fmt.Println("Hello, world!")
}`

	count, err := tok.Count(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count == 0 {
		t.Error("expected non-zero token count")
	}
}

func TestTiktokenTokenizer_Count_Markdown(t *testing.T) {
	tok, err := NewTiktokenTokenizer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	markdown := `# Heading

This is a paragraph with **bold** and *italic* text.

- List item 1
- List item 2

[Link](https://example.com)`

	count, err := tok.Count(markdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count == 0 {
		t.Error("expected non-zero token count")
	}
}

func TestTiktokenTokenizer_Count_Whitespace(t *testing.T) {
	tok, err := NewTiktokenTokenizer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testCases := []struct {
		name string
		text string
	}{
		{"spaces", "   "},
		{"tabs", "\t\t\t"},
		{"newlines", "\n\n\n"},
		{"mixed", " \t\n "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Whitespace might tokenize to 0 or more tokens depending on encoding
			_ = count
		})
	}
}

func TestTiktokenTokenizer_Tokenize(t *testing.T) {
	tok, err := NewTiktokenTokenizer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	root := section.NewRoot("root")
	root.SetContent("Hello, world!")

	root.CreateChild("child", 1, "This is a test.")

	ctx := context.Background()
	result, err := tok.Tokenize(ctx, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.GetContentTokens() == 0 {
		t.Error("expected non-zero root ContentTokens")
	}

	if result.GetSubtreeTokens() <= result.GetContentTokens() {
		t.Error("SubtreeTokens should be greater than ContentTokens with children")
	}

	if len(result.GetChildren()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(result.GetChildren()))
	}

	if result.GetChildren()[0].GetContentTokens() == 0 {
		t.Error("expected non-zero child ContentTokens")
	}

	// Verify SubtreeTokens invariant
	expectedSubtree := result.GetContentTokens() + result.GetChildren()[0].GetSubtreeTokens()
	if result.GetSubtreeTokens() != expectedSubtree {
		t.Errorf("SubtreeTokens invariant violated: expected %d, got %d",
			expectedSubtree, result.GetSubtreeTokens())
	}
}

func TestTiktokenTokenizer_Count_LongText(t *testing.T) {
	tok, err := NewTiktokenTokenizer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Create a long text
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString("This is a test sentence. ")
	}
	longText := sb.String()

	count, err := tok.Count(longText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count == 0 {
		t.Error("expected non-zero token count")
	}
}

func TestTiktokenTokenizer_DifferentEncodings_SameText(t *testing.T) {
	encodings := []string{"o200k_base", "cl100k_base"}
	text := "Hello, world! This is a test."

	counts := make(map[string]int)

	for _, enc := range encodings {
		tok, err := NewTiktokenTokenizer(WithEncoding(enc))
		if err != nil {
			t.Fatalf("failed to create tokenizer for %s: %v", enc, err)
		}

		count, err := tok.Count(text)
		if err != nil {
			t.Fatalf("Count failed for %s: %v", enc, err)
		}

		counts[enc] = count
	}

	// Different encodings may produce different counts
	// Just verify they're all non-zero
	for enc, count := range counts {
		if count == 0 {
			t.Errorf("encoding %s produced zero tokens", enc)
		}
	}
}

func TestTiktokenTokenizer_SpecialTokens(t *testing.T) {
	tok, err := NewTiktokenTokenizer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with various special characters
	testCases := []string{
		"<|endoftext|>",
		"[INST] instruction [/INST]",
		"<s>text</s>",
		"<!-- comment -->",
	}

	for _, text := range testCases {
		t.Run(text, func(t *testing.T) {
			count, err := tok.Count(text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Just verify it doesn't crash, counts will vary
			_ = count
		})
	}
}

func TestTiktokenTokenizer_Consistency(t *testing.T) {
	tok, err := NewTiktokenTokenizer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := "This is a consistency test."

	// Count same text multiple times
	counts := make([]int, 5)
	for i := range counts {
		count, err := tok.Count(text)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}
		counts[i] = count
	}

	// All counts should be identical
	first := counts[0]
	for i, count := range counts {
		if count != first {
			t.Errorf("count %d: expected %d, got %d", i, first, count)
		}
	}
}
