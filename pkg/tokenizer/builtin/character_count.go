package builtin

import (
	"github.com/wyvernzora/chunky/pkg/tokenizer"
)

type charCountConfig struct {
	charsPerToken float64
}

// CharacterCountOption configures the character count tokenizer.
type CharacterCountOption func(*charCountConfig)

// WithCharsPerToken sets the average characters per token ratio.
// Must be greater than 0. Default is 4.0.
//
// This ratio determines how many characters are estimated to comprise one token.
// Common values:
//   - 4.0: Standard English text (default)
//   - 3.0: Dense technical content
//   - 5.0: Prose with longer words
func WithCharsPerToken(cpt float64) CharacterCountOption {
	return func(cfg *charCountConfig) {
		if cpt > 0 {
			cfg.charsPerToken = cpt
		}
	}
}

// NewCharCountTokenizer returns a Tokenizer that estimates tokens by dividing
// the Unicode rune count by the configured characters-per-token ratio.
//
// This is a simple, fast approximation suitable for when exact token counts
// aren't critical or when no model-specific tokenizer is available.
//
// The tokenizer counts Unicode runes (not bytes), so it handles multi-byte
// characters correctly.
//
// Parameters:
//   - opts: Optional configuration via WithCharsPerToken
//
// Default configuration:
//   - charsPerToken: 4.0
//
// Example:
//
//	// Standard configuration
//	tok := NewCharCountTokenizer()
//
//	// Custom ratio for dense content
//	tok := NewCharCountTokenizer(WithCharsPerToken(3.0))
//
//	count, _ := tok.Count("Hello, world!") // ~3 tokens with default ratio
func NewCharCountTokenizer(opts ...CharacterCountOption) tokenizer.Tokenizer {
	cfg := &charCountConfig{
		charsPerToken: 4.0,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return tokenizer.MakeTokenizer(func(s string) (int, error) {
		return int(float64(len([]rune(s))) / cfg.charsPerToken), nil
	})
}
