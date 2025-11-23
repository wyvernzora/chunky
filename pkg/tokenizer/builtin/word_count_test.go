package builtin

import (
	"context"
	"testing"

	"github.com/wyvernzora/chunky/pkg/section"
)

func TestNewWordCountTokenizer_Default(t *testing.T) {
	tok := NewWordCountTokenizer()
	if tok == nil {
		t.Fatal("NewWordCountTokenizer returned nil")
	}

	// 2 words / 1.0 = 2 tokens
	count, err := tok.Count("hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 tokens, got %d", count)
	}
}

func TestNewWordCountTokenizer_CustomRatio(t *testing.T) {
	tok := NewWordCountTokenizer(WithWordsPerToken(0.75))

	// 3 words / 0.75 = 4 tokens
	count, err := tok.Count("hello world test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 4 {
		t.Errorf("expected 4 tokens, got %d", count)
	}
}

func TestNewWordCountTokenizer_MultipleOptions(t *testing.T) {
	// Last option should win
	tok := NewWordCountTokenizer(
		WithWordsPerToken(0.5),
		WithWordsPerToken(2.0),
	)

	// 4 words / 2.0 = 2 tokens
	count, err := tok.Count("one two three four")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 tokens, got %d", count)
	}
}

func TestNewWordCountTokenizer_InvalidRatio(t *testing.T) {
	// Zero and negative ratios should be ignored, falling back to default
	tok := NewWordCountTokenizer(WithWordsPerToken(0))

	// Should use default ratio of 1.0
	count, err := tok.Count("hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 tokens with default ratio, got %d", count)
	}
}

func TestWordCountTokenizer_Count_Empty(t *testing.T) {
	tok := NewWordCountTokenizer()

	count, err := tok.Count("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tokens for empty string, got %d", count)
	}
}

func TestWordCountTokenizer_Count_SimpleText(t *testing.T) {
	tok := NewWordCountTokenizer()

	testCases := []struct {
		name     string
		text     string
		expected int
	}{
		{"single word", "hello", 1},
		{"two words", "hello world", 2},
		{"multiple words", "the quick brown fox", 4},
		{"with punctuation", "hello, world!", 2},
		{"with numbers", "test 123 456", 3},
		{"sentence", "This is a test sentence.", 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, count)
			}
		})
	}
}

func TestWordCountTokenizer_Count_Whitespace(t *testing.T) {
	tok := NewWordCountTokenizer()

	testCases := []struct {
		name     string
		text     string
		expected int
	}{
		{"only spaces", "   ", 0},
		{"only tabs", "\t\t\t", 0},
		{"only newlines", "\n\n\n", 0},
		{"mixed whitespace", " \t\n ", 0},
		{"leading spaces", "  hello", 1},
		{"trailing spaces", "hello  ", 1},
		{"extra spaces between", "hello    world", 2},
		{"newlines between", "hello\n\nworld", 2},
		{"tabs between", "hello\t\tworld", 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, count)
			}
		})
	}
}

func TestWordCountTokenizer_Count_Punctuation(t *testing.T) {
	tok := NewWordCountTokenizer()

	testCases := []struct {
		name     string
		text     string
		expected int
	}{
		{"period", "Hello.", 1},
		{"comma", "Hello,", 1},
		{"exclamation", "Hello!", 1},
		{"question", "Hello?", 1},
		{"quotes", `"Hello"`, 1},
		{"apostrophe", "don't", 1},
		{"hyphen", "well-known", 1},
		{"multiple punct", "Hello, world!", 2},
		{"parentheses", "(Hello)", 1},
		{"brackets", "[Hello]", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, count)
			}
		})
	}
}

func TestWordCountTokenizer_Count_Unicode(t *testing.T) {
	tok := NewWordCountTokenizer()

	testCases := []struct {
		name     string
		text     string
		expected int
	}{
		{"chinese", "你好 世界", 2},
		{"japanese", "こんにちは 世界", 2},
		{"arabic", "مرحبا بالعالم", 2},
		{"cyrillic", "Привет мир", 2},
		{"mixed", "Hello 世界", 2},
		{"emoji with text", "Hello 🌍 world", 3},
		{"only emoji", "🌍🚀🎉", 1}, // Emojis without spaces count as one word
		{"emoji separated", "🌍 🚀 🎉", 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, count)
			}
		})
	}
}

func TestWordCountTokenizer_Count_Multiline(t *testing.T) {
	tok := NewWordCountTokenizer()

	text := `Line one has five words here
Line two has five words too
Line three now`

	count, err := tok.Count(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 6 + 6 + 3 = 15 words
	if count != 15 {
		t.Errorf("expected 15 words, got %d", count)
	}
}

func TestWordCountTokenizer_Count_Code(t *testing.T) {
	tok := NewWordCountTokenizer()

	code := `func main() {
    fmt.Println("Hello")
}`

	count, err := tok.Count(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// func, main(), {, fmt.Println("Hello"), } = 5 words
	if count != 5 {
		t.Errorf("expected 5 words, got %d", count)
	}
}

func TestWordCountTokenizer_Count_LongText(t *testing.T) {
	tok := NewWordCountTokenizer()

	// Create text with 1000 words
	text := ""
	for i := range 1000 {
		if i > 0 {
			text += " "
		}
		text += "word"
	}

	count, err := tok.Count(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 1000 {
		t.Errorf("expected 1000 words, got %d", count)
	}
}

func TestWordCountTokenizer_Tokenize(t *testing.T) {
	tok := NewWordCountTokenizer(WithWordsPerToken(2.0))

	root := section.NewRoot("root")
	root.SetContent("one two three four") // 4 words = 2 tokens

	root.CreateChild("child", 1, "five six seven eight nine ten") // 6 words = 3 tokens

	ctx := context.Background()
	result, err := tok.Tokenize(ctx, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.GetContentTokens() != 2 {
		t.Errorf("root ContentTokens: expected 2, got %d", result.GetContentTokens())
	}

	// Root: 2 + 3 = 5 total
	if result.GetSubtreeTokens() != 5 {
		t.Errorf("root SubtreeTokens: expected 5, got %d", result.GetSubtreeTokens())
	}

	if len(result.GetChildren()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(result.GetChildren()))
	}

	if result.GetChildren()[0].GetContentTokens() != 3 {
		t.Errorf("child ContentTokens: expected 3, got %d", result.GetChildren()[0].GetContentTokens())
	}
}

func TestWordCountTokenizer_DifferentRatios(t *testing.T) {
	testCases := []struct {
		ratio    float64
		text     string
		words    int
		expected int
	}{
		{1.0, "one two three", 3, 3},
		{0.75, "one two three", 3, 4},
		{2.0, "one two three four", 4, 2},
		{1.5, "one two three", 3, 2},
		{0.5, "one two", 2, 4},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			tok := NewWordCountTokenizer(WithWordsPerToken(tc.ratio))
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("ratio %.2f, %d words: expected %d tokens, got %d",
					tc.ratio, tc.words, tc.expected, count)
			}
		})
	}
}

func TestWordCountTokenizer_Truncation(t *testing.T) {
	tok := NewWordCountTokenizer(WithWordsPerToken(3.0))

	testCases := []struct {
		text     string
		words    int
		expected int
	}{
		{"one", 1, 0},                // 1/3 = 0.33 -> 0
		{"one two", 2, 0},            // 2/3 = 0.66 -> 0
		{"one two three", 3, 1},      // 3/3 = 1.00 -> 1
		{"one two three four", 4, 1}, // 4/3 = 1.33 -> 1
		{"a b c d e", 5, 1},          // 5/3 = 1.66 -> 1
		{"a b c d e f", 6, 2},        // 6/3 = 2.00 -> 2
	}

	for _, tc := range testCases {
		t.Run(tc.text, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("%d words: expected %d tokens, got %d", tc.words, tc.expected, count)
			}
		})
	}
}

func TestWordCountTokenizer_SpecialCases(t *testing.T) {
	tok := NewWordCountTokenizer()

	testCases := []struct {
		name     string
		text     string
		expected int
	}{
		{"url", "https://example.com/path", 1},
		{"email", "user@example.com", 1},
		{"version", "v1.2.3", 1},
		{"date", "2024-01-15", 1},
		{"math", "2+2=4", 1},
		{"underscore", "hello_world", 1},
		{"camelCase", "helloWorld", 1},
		{"PascalCase", "HelloWorld", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, count)
			}
		})
	}
}

func TestWordCountTokenizer_ConsistentWithSimple(t *testing.T) {
	// Verify countWords produces same results as strings.Fields for basic cases
	testCases := []string{
		"hello world",
		"  hello  world  ",
		"one two three four five",
		"test\nwith\nnewlines",
		"test\twith\ttabs",
	}

	for _, text := range testCases {
		t.Run(text, func(t *testing.T) {
			got := countWords(text)
			expected := countWordsSimple(text)
			if got != expected {
				t.Errorf("countWords=%d, countWordsSimple=%d", got, expected)
			}
		})
	}
}

func TestWordCountTokenizer_EdgeCases(t *testing.T) {
	tok := NewWordCountTokenizer()

	testCases := []struct {
		name     string
		text     string
		expected int
	}{
		{"single char", "a", 1},
		{"single space", " ", 0},
		{"newline only", "\n", 0},
		{"tab only", "\t", 0},
		{"zero-width space", "\u200b", 1}, // Zero-width space is not whitespace, counts as word
		{"non-breaking space", "hello\u00a0world", 2},
		{"multiple unicode spaces", "hello\u2000world", 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, count)
			}
		})
	}
}

func TestWordCountTokenizer_Markdown(t *testing.T) {
	tok := NewWordCountTokenizer()

	markdown := `# Heading One

This is a paragraph with **bold** and *italic* text.

## Heading Two

- List item one
- List item two
- List item three

[Link text](https://example.com)

` + "```" + `
code block
with multiple lines
` + "```"

	count, err := tok.Count(markdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should count all words including markdown syntax
	if count == 0 {
		t.Error("expected non-zero word count for markdown")
	}
}
