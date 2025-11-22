package builtin

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
	"github.com/wyvernzora/chunky/pkg/tokenizer"
)

type tiktokenConfig struct {
	encodingName string // e.g. "gpt-4o", "cl100k_base", "o200k_base"
}

// TiktokenOption configures the tiktoken tokenizer.
type TiktokenOption func(*tiktokenConfig)

// WithEncoding sets the tiktoken encoding to use.
// Must be a valid encoding name recognized by tiktoken-go.
//
// Common encodings:
//   - "o200k_base": GPT-4o and newer models (default)
//   - "cl100k_base": GPT-4, GPT-3.5-turbo
//   - "p50k_base": Older GPT-3 models
//   - "gpt2": GPT-2 models
//
// See tiktoken documentation for the full list of supported encodings.
func WithEncoding(name string) TiktokenOption {
	return func(cfg *tiktokenConfig) {
		if name != "" {
			cfg.encodingName = name
		}
	}
}

// NewTiktokenTokenizer returns a Tokenizer backed by tiktoken-go, which provides
// accurate token counting for OpenAI models.
//
// This tokenizer uses the actual tokenization algorithm from OpenAI's models,
// providing exact token counts that match what the model will see. This is
// essential for staying within model context limits.
//
// Parameters:
//   - opts: Optional configuration via WithEncoding
//
// Default configuration:
//   - encodingName: "o200k_base" (for GPT-4o and newer)
//
// Returns an error if the specified encoding cannot be loaded.
//
// Example:
//
//	// Default encoding (o200k_base)
//	tok, err := NewTiktokenTokenizer()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// GPT-3.5-turbo encoding
//	tok, err := NewTiktokenTokenizer(
//	    WithEncoding("cl100k_base"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	count, _ := tok.Count("Hello, world!")
func NewTiktokenTokenizer(opts ...TiktokenOption) (tokenizer.Tokenizer, error) {
	cfg := &tiktokenConfig{
		encodingName: "o200k_base",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	enc, err := tiktoken.GetEncoding(cfg.encodingName)
	if err != nil {
		return nil, fmt.Errorf("tiktoken: failed to load encoding %q: %w", cfg.encodingName, err)
	}

	// Counter closure that encodes the text and returns the token count
	counter := func(s string) (int, error) {
		ids := enc.Encode(s, nil, nil)
		return len(ids), nil
	}

	return tokenizer.MakeTokenizer(counter), nil
}
