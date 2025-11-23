// Package tokenizer provides token counting and tree tokenization for text content.
//
// It defines interfaces for counting tokens in strings and tokenizing section trees
// to add token count metadata. Multiple tokenizer implementations are available in
// the builtin subpackage.
//
// # Tokenizer Interface
//
// The Tokenizer interface provides two main operations:
//
//	type Tokenizer interface {
//	    // Count returns token count for a string
//	    Count(text string) (int, error)
//
//	    // Tokenize adds token counts to a section tree
//	    Tokenize(ctx context.Context, root *section.Section) (*TokenizedSection, error)
//	}
//
// # TokenizedSection
//
// TokenizedSection represents a section tree annotated with token counts:
//
//   - ContentTokens: Tokens in this section's content only
//   - SubtreeTokens: Total tokens including all descendants
//   - Maintains tree structure with children
//
// Fields are read-only (accessed via getters) to maintain token count invariants.
//
// # Built-in Tokenizers
//
// The builtin subpackage provides three implementations:
//
//  1. TiktokenTokenizer: Uses tiktoken library (OpenAI's tokenizer)
//     - Default encoding: o200k_base
//     - Most accurate for LLM applications
//
//  2. WordCountTokenizer: Approximates tokens by counting words
//     - Configurable words-per-token ratio
//     - Fast, no external dependencies
//
//  3. CharacterCountTokenizer: Approximates tokens by counting characters
//     - Configurable characters-per-token ratio
//     - Fastest option for rough estimates
//
// # Usage Example
//
//	tok, err := builtin.NewTiktokenTokenizer()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	count, err := tok.Count("Hello, world!")
//	tokenized, err := tok.Tokenize(ctx, sectionTree)
//	total := tokenized.GetSubtreeTokens()
package tokenizer
