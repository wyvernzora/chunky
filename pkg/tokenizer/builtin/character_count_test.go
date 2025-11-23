package builtin

import (
	"context"
	"testing"

	"github.com/wyvernzora/chunky/pkg/section"
)

func TestNewCharCountTokenizer_Default(t *testing.T) {
	tok := NewCharCountTokenizer()
	if tok == nil {
		t.Fatal("NewCharCountTokenizer returned nil")
	}

	// 12 characters / 4 = 3 tokens
	count, err := tok.Count("hello world!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 tokens, got %d", count)
	}
}

func TestNewCharCountTokenizer_CustomRatio(t *testing.T) {
	tok := NewCharCountTokenizer(WithCharsPerToken(2.0))

	// 10 characters / 2 = 5 tokens
	count, err := tok.Count("hello test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 tokens, got %d", count)
	}
}

func TestNewCharCountTokenizer_MultipleOptions(t *testing.T) {
	// Last option should win
	tok := NewCharCountTokenizer(
		WithCharsPerToken(2.0),
		WithCharsPerToken(5.0),
	)

	// 10 characters / 5 = 2 tokens
	count, err := tok.Count("hello test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 tokens, got %d", count)
	}
}

func TestNewCharCountTokenizer_InvalidRatio(t *testing.T) {
	// Zero and negative ratios should be ignored, falling back to default
	tok := NewCharCountTokenizer(WithCharsPerToken(0))

	// Should use default ratio of 4
	count, err := tok.Count("12345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 tokens with default ratio, got %d", count)
	}
}

func TestCharCountTokenizer_Count_Empty(t *testing.T) {
	tok := NewCharCountTokenizer()

	count, err := tok.Count("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tokens for empty string, got %d", count)
	}
}

func TestCharCountTokenizer_Count_Unicode(t *testing.T) {
	tok := NewCharCountTokenizer(WithCharsPerToken(1.0))

	testCases := []struct {
		name     string
		text     string
		expected int
	}{
		{"ascii", "test", 4},
		{"chinese", "你好世界", 4},
		{"emoji", "🌍🚀🎉", 3},
		{"mixed", "Hi 你好 🌍", 7},
		{"arabic", "مرحبا", 5},
		{"cyrillic", "Привет", 6},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("expected %d runes, got %d", tc.expected, count)
			}
		})
	}
}

func TestCharCountTokenizer_Count_Whitespace(t *testing.T) {
	tok := NewCharCountTokenizer(WithCharsPerToken(1.0))

	testCases := []struct {
		name     string
		text     string
		expected int
	}{
		{"spaces", "   ", 3},
		{"tabs", "\t\t\t", 3},
		{"newlines", "\n\n\n", 3},
		{"mixed", " \t\n ", 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("expected %d chars, got %d", tc.expected, count)
			}
		})
	}
}

func TestCharCountTokenizer_Count_LongText(t *testing.T) {
	tok := NewCharCountTokenizer()

	// 1000 character string
	text := ""
	for i := 0; i < 100; i++ {
		text += "0123456789"
	}

	count, err := tok.Count(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1000 chars / 4 = 250 tokens
	if count != 250 {
		t.Errorf("expected 250 tokens, got %d", count)
	}
}

func TestCharCountTokenizer_Tokenize(t *testing.T) {
	tok := NewCharCountTokenizer(WithCharsPerToken(2.0))

	root := section.NewRoot("root")
	root.SetContent("1234") // 4 chars = 2 tokens

	root.CreateChild("child", 1, "123456") // 6 chars = 3 tokens

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

func TestCharCountTokenizer_DifferentRatios(t *testing.T) {
	testCases := []struct {
		ratio    float64
		text     string
		expected int
	}{
		{1.0, "12345", 5},
		{2.0, "12345", 2},
		{3.0, "123456789", 3},
		{4.0, "12345678", 2},
		{5.0, "1234567890", 2},
		{10.0, "12345678901234567890", 2},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			tok := NewCharCountTokenizer(WithCharsPerToken(tc.ratio))
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("ratio %.1f, text %q: expected %d, got %d",
					tc.ratio, tc.text, tc.expected, count)
			}
		})
	}
}

func TestCharCountTokenizer_Truncation(t *testing.T) {
	tok := NewCharCountTokenizer(WithCharsPerToken(3.0))

	testCases := []struct {
		text     string
		expected int
	}{
		{"1", 0},         // 1/3 = 0.33 -> 0
		{"12", 0},        // 2/3 = 0.66 -> 0
		{"123", 1},       // 3/3 = 1.00 -> 1
		{"1234", 1},      // 4/3 = 1.33 -> 1
		{"12345", 1},     // 5/3 = 1.66 -> 1
		{"123456", 2},    // 6/3 = 2.00 -> 2
		{"1234567", 2},   // 7/3 = 2.33 -> 2
		{"12345678", 2},  // 8/3 = 2.66 -> 2
		{"123456789", 3}, // 9/3 = 3.00 -> 3
	}

	for _, tc := range testCases {
		t.Run(tc.text, func(t *testing.T) {
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

func TestCharCountTokenizer_SpecialCharacters(t *testing.T) {
	tok := NewCharCountTokenizer(WithCharsPerToken(1.0))

	testCases := []struct {
		name  string
		text  string
		runes int
	}{
		{"punctuation", "!@#$%^&*()", 10},
		{"quotes", `"'` + "`", 3},
		{"brackets", "[]{}()<>", 8},
		{"math", "+=×÷≠≈", 6},
		{"currency", "$€£¥", 4},
		{"arrows", "→←↑↓", 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.runes {
				t.Errorf("expected %d runes, got %d", tc.runes, count)
			}
		})
	}
}
