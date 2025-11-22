package tokenizer

import (
	"strings"

	"github.com/wyvernzora/chunky/pkg/section"
)

// TokenizedSection represents the measured/token-counted form of a Section node.
// It maintains a reference to the original Section and adds token count metadata
// for both the node's immediate content and its entire subtree.
//
// The token counts satisfy the following invariants:
//   - ContentTokens represents tokens in Section.Content() only
//   - SubtreeTokens = ContentTokens + sum of all children's SubtreeTokens
//   - SubtreeTokens >= ContentTokens
type TokenizedSection struct {
	// Section is the original section node this TokenizedSection represents.
	Section *section.Section

	// ContentTokens is the token count for this node's content only,
	// excluding all descendants.
	ContentTokens int

	// SubtreeTokens is the total token count for this node's content
	// plus all descendant contents combined.
	SubtreeTokens int

	// Children contains the tokenized versions of all child sections,
	// maintaining the same tree structure as the original Section tree.
	Children []*TokenizedSection
}

// NewTokenizedSection constructs a TokenizedSection node with the given values.
//
// Caller is responsible for ensuring token count invariants:
//   - SubtreeTokens >= ContentTokens
//   - SubtreeTokens equals ContentTokens + sum of children's SubtreeTokens
//
// Parameters:
//   - sec: The original section node
//   - contentTokens: Token count for this section's content only
//   - subtreeTokens: Total token count including all descendants
//   - children: Tokenized child sections
//
// Example:
//
//	child := NewTokenizedSection(childSec, 10, 10, nil)
//	parent := NewTokenizedSection(parentSec, 5, 15, []*TokenizedSection{child})
func NewTokenizedSection(
	sec *section.Section,
	contentTokens int,
	subtreeTokens int,
	children []*TokenizedSection,
) *TokenizedSection {
	return &TokenizedSection{
		Section:       sec,
		ContentTokens: contentTokens,
		SubtreeTokens: subtreeTokens,
		Children:      children,
	}
}

// Render concatenates this section's content with all descendant contents in
// depth-first pre-order traversal. This reconstructs the complete text representation
// of the entire subtree.
//
// Returns an empty string if the TokenizedSection is nil.
//
// Example:
//
//	root := tokenizer.Tokenize(ctx, section)
//	fullText := root.Render() // Complete document text
func (t *TokenizedSection) Render() string {
	if t == nil {
		return ""
	}

	var b strings.Builder

	// Local content first
	b.WriteString(t.Section.Content())

	// Then recursive children in order
	for _, c := range t.Children {
		if c != nil {
			b.WriteString(c.Render())
		}
	}

	return b.String()
}
