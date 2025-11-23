package chunker

import (
	"github.com/wyvernzora/chunky/pkg/tokenizer"
)

// chunkDocumentParams holds parameters for document chunking.
type chunkDocumentParams struct {
	filePath    string
	fileTitle   string
	frontBlock  string
	frontTokens int
	bodyBudget  int
	root        *tokenizer.TokenizedSection
}

// chunkDocument splits a tokenized document into chunks based on token budgets.
//
// Algorithm:
//  1. Traverse the tokenized tree in pre-order (parent before children)
//  2. Accumulate content units greedily into chunks
//  3. When a unit doesn't fit, emit current chunk and start a new one
//  4. Units exceeding bodyBudget get their own dedicated "jumbo" chunk
//  5. Flush any remaining content as the final chunk
//
// Returns a slice of chunks, each containing frontmatter + portion of body.
func chunkDocument(params chunkDocumentParams) []Chunk {
	builder := newChunkBuilder(
		params.filePath,
		params.fileTitle,
		params.frontBlock,
		params.frontTokens,
		params.bodyBudget,
	)

	var chunks []Chunk

	// Traverse and accumulate units
	units := traverseUnits(params.root)
	for _, u := range units {
		produced := builder.appendUnit(u.text, u.tokens)
		chunks = append(chunks, produced...)
	}

	// Flush any remaining content
	if final := builder.flush(); final != nil {
		chunks = append(chunks, *final)
	}

	return chunks
}

// unit represents a single content unit from a tokenized section tree.
type unit struct {
	text   string
	tokens int
}

// traverseUnits performs a pre-order traversal of the tokenized section tree,
// yielding each node's self content as a unit.
//
// The traversal uses an explicit stack to avoid recursion and processes nodes
// in document order (parent before children, children in left-to-right order).
func traverseUnits(root *tokenizer.TokenizedSection) []unit {
	if root == nil {
		return nil
	}

	var units []unit
	stack := []*tokenizer.TokenizedSection{root}

	for len(stack) > 0 {
		// Pop from stack
		n := len(stack)
		node := stack[n-1]
		stack = stack[:n-1]

		// Yield this node's self content if it has any tokens
		if node.GetContentTokens() > 0 {
			units = append(units, unit{
				text:   node.GetSection().Content(),
				tokens: node.GetContentTokens(),
			})
		}

		// Push children in reverse order to maintain document order when popping
		children := node.GetChildren()
		for i := len(children) - 1; i >= 0; i-- {
			stack = append(stack, children[i])
		}
	}

	return units
}
