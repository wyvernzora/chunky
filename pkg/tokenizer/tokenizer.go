package tokenizer

import (
	"context"

	"github.com/wyvernzora/chunky/pkg/section"
)

// Tokenizer takes a transformed Section tree and returns a measured/tokenized tree.
// It provides token counting capabilities for both individual strings and entire
// section hierarchies.
type Tokenizer interface {
	// Count returns the number of tokens in the given string.
	// Returns an error if token counting fails.
	Count(s string) (int, error)

	// Tokenize walks the section tree and produces a TokenizedSection tree with
	// token counts for each node's content and subtree. The context can be used
	// for cancellation or passing metadata.
	//
	// The resulting TokenizedSection maintains the same tree structure as the input,
	// with each node annotated with:
	//   - ContentTokens: tokens in this node's content only
	//   - SubtreeTokens: tokens in this node's content plus all descendants
	//
	// Returns an error if token counting fails for any node.
	Tokenize(ctx context.Context, root *section.Section) (*TokenizedSection, error)
}

// TokenCounter is a function that counts tokens in a given text string.
// It returns the token count and any error encountered during counting.
type TokenCounter func(text string) (int, error)

// tokenizer is the internal implementation of the Tokenizer interface.
type tokenizer struct {
	tokenCounter TokenCounter
}

// Count implements Tokenizer.Count by delegating to the configured TokenCounter.
func (t *tokenizer) Count(s string) (int, error) {
	return t.tokenCounter(s)
}

// Tokenize implements Tokenizer.Tokenize by recursively walking the section tree
// and computing token counts for each node.
func (t *tokenizer) Tokenize(ctx context.Context, root *section.Section) (*TokenizedSection, error) {
	var visit func(*section.Section) (*TokenizedSection, error)

	visit = func(node *section.Section) (*TokenizedSection, error) {
		// Count tokens in *this node's* content
		content := node.Content()
		contentTokens, err := t.tokenCounter(content)
		if err != nil {
			return nil, err
		}

		// Measure children
		kids := node.Children()
		tkids := make([]*TokenizedSection, 0, len(kids))
		subtreeTokens := contentTokens

		for _, child := range kids {
			tchild, err := visit(child)
			if err != nil {
				return nil, err
			}
			tkids = append(tkids, tchild)
			subtreeTokens += tchild.SubtreeTokens
		}

		return &TokenizedSection{
			Section:       node,
			ContentTokens: contentTokens,
			SubtreeTokens: subtreeTokens,
			Children:      tkids,
		}, nil
	}

	return visit(root)
}

// MakeTokenizer creates a new Tokenizer using the provided TokenCounter function.
// The TokenCounter will be used for all token counting operations.
//
// Example:
//
//	counter := func(text string) (int, error) {
//	    return len(strings.Fields(text)), nil // Simple word count
//	}
//	tok := tokenizer.MakeTokenizer(counter)
func MakeTokenizer(counter TokenCounter) Tokenizer {
	return &tokenizer{tokenCounter: counter}
}
