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
//
// All fields are read-only after construction to maintain token count invariants.
type TokenizedSection struct {
	// section is the original section node this TokenizedSection represents.
	section *section.Section

	// contentTokens is the token count for this node's content only,
	// excluding all descendants.
	contentTokens int

	// subtreeTokens is the total token count for this node's content
	// plus all descendant contents combined.
	subtreeTokens int

	// children contains the tokenized versions of all child sections,
	// maintaining the same tree structure as the original Section tree.
	children []*TokenizedSection
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
		section:       sec,
		contentTokens: contentTokens,
		subtreeTokens: subtreeTokens,
		children:      children,
	}
}

// GetSection returns the original section node this TokenizedSection represents.
func (t *TokenizedSection) GetSection() *section.Section {
	if t == nil {
		return nil
	}
	return t.section
}

// GetContentTokens returns the token count for this node's content only,
// excluding all descendants.
func (t *TokenizedSection) GetContentTokens() int {
	if t == nil {
		return 0
	}
	return t.contentTokens
}

// GetSubtreeTokens returns the total token count for this node's content
// plus all descendant contents combined.
func (t *TokenizedSection) GetSubtreeTokens() int {
	if t == nil {
		return 0
	}
	return t.subtreeTokens
}

// GetChildren returns a copy of the tokenized child sections to prevent
// external modification of the tree structure.
func (t *TokenizedSection) GetChildren() []*TokenizedSection {
	if t == nil {
		return nil
	}
	// Return a copy to prevent modification
	result := make([]*TokenizedSection, len(t.children))
	copy(result, t.children)
	return result
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
	b.WriteString(t.section.Content())

	// Then recursive children in order
	for _, c := range t.children {
		if c != nil {
			b.WriteString(c.Render())
		}
	}

	return b.String()
}
