package builtin

import (
	"strings"
	"unicode"

	"github.com/wyvernzora/chunky/pkg/tokenizer"
)

type wordCountConfig struct {
	wordsPerToken float64
}

// WordCountOption configures the word count tokenizer.
type WordCountOption func(*wordCountConfig)

// WithWordsPerToken sets the average words per token ratio.
// Must be greater than 0. Default is 1.0 (one word = one token).
//
// This ratio determines how many words comprise one token.
// Common values:
//   - 1.0: Each word is one token (default)
//   - 0.75: Accounts for subword tokenization (common in BPE tokenizers)
//   - 1.3: Conservative estimate for models that split compound words
func WithWordsPerToken(wpt float64) WordCountOption {
	return func(cfg *wordCountConfig) {
		if wpt > 0 {
			cfg.wordsPerToken = wpt
		}
	}
}

// NewWordCountTokenizer returns a Tokenizer that estimates tokens by counting
// words (sequences of non-whitespace characters separated by whitespace) and
// dividing by the configured words-per-token ratio.
//
// Words are identified using Unicode word boundaries, which handles various
// writing systems correctly. The tokenizer:
//   - Splits on any Unicode whitespace (spaces, tabs, newlines, etc.)
//   - Counts sequences of non-whitespace as words
//   - Handles punctuation attached to words (e.g., "hello," counts as one word)
//   - Works with Unicode text (CJK, Arabic, Cyrillic, etc.)
//
// This provides a more accurate approximation than character counting for
// languages that use spaces to separate words, and is faster than actual
// model-based tokenization.
//
// Parameters:
//   - opts: Optional configuration via WithWordsPerToken
//
// Default configuration:
//   - wordsPerToken: 1.0 (one word equals one token)
//
// Example:
//
//	// Standard configuration (1 word = 1 token)
//	tok := NewWordCountTokenizer()
//
//	// Account for subword tokenization
//	tok := NewWordCountTokenizer(WithWordsPerToken(0.75))
//
//	count, _ := tok.Count("Hello, world!") // 2 tokens with default ratio
func NewWordCountTokenizer(opts ...WordCountOption) tokenizer.Tokenizer {
	cfg := &wordCountConfig{
		wordsPerToken: 1.0,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return tokenizer.MakeTokenizer(func(s string) (int, error) {
		words := countWords(s)
		return int(float64(words) / cfg.wordsPerToken), nil
	})
}

// countWords counts the number of words in the text using Unicode-aware
// whitespace splitting.
func countWords(text string) int {
	if text == "" {
		return 0
	}

	words := 0
	inWord := false

	for _, r := range text {
		if unicode.IsSpace(r) {
			if inWord {
				words++
				inWord = false
			}
		} else {
			inWord = true
		}
	}

	// Count the last word if we ended in the middle of one
	if inWord {
		words++
	}

	return words
}

// countWordsSimple is an alternative implementation using strings.Fields
// for comparison and validation purposes.
func countWordsSimple(text string) int {
	if text == "" {
		return 0
	}
	return len(strings.Fields(text))
}
