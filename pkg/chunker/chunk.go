package chunker

import (
	"strings"
)

// Chunk represents a single chunk of markdown content with metadata.
// Each chunk contains the full frontmatter plus a portion of the body content,
// sized to fit within the configured token budget.
type Chunk struct {
	// FilePath is the logical path of the source document.
	FilePath string

	// FileTitle is the human-readable title of the source document.
	FileTitle string

	// ChunkIndex is the 1-indexed position of this chunk within the document.
	// For a document split into N chunks, indices range from 1 to N.
	ChunkIndex int

	// Text is the complete markdown content including frontmatter and body.
	// Format: "---\nfrontmatter\n---\n\nbody content"
	Text string

	// Tokens is the total token count of the Text field.
	Tokens int
}

// chunkBuilder accumulates markdown content into chunks based on token budgets.
// It uses a greedy algorithm to pack content until the budget is exceeded.
type chunkBuilder struct {
	filePath    string
	fileTitle   string
	frontBlock  string // Serialized frontmatter block
	frontTokens int    // Token count of frontBlock
	bodyBudget  int    // Max tokens for body content

	parts  []string // Accumulated body parts for current chunk
	tokens int      // Current token count (body only)
	index  int      // Next chunk index (1-indexed)
}

// newChunkBuilder creates a new builder for chunking a document.
func newChunkBuilder(filePath, fileTitle, frontBlock string, frontTokens, bodyBudget int) *chunkBuilder {
	return &chunkBuilder{
		filePath:    filePath,
		fileTitle:   fileTitle,
		frontBlock:  frontBlock,
		frontTokens: frontTokens,
		bodyBudget:  bodyBudget,
		parts:       make([]string, 0),
		tokens:      0,
		index:       1,
	}
}

// appendUnit adds a content unit to the builder and returns any chunks produced.
// Units are added greedily until they don't fit, at which point a chunk is emitted.
//
// Special case: "jumbo" units that exceed bodyBudget get their own dedicated chunk.
func (b *chunkBuilder) appendUnit(unitText string, unitTokens int) []Chunk {
	if unitTokens <= 0 {
		return nil
	}

	var chunks []Chunk

	// Case 1: JUMBO unit (exceeds body budget entirely)
	if unitTokens > b.bodyBudget {
		// Flush any accumulated content first
		if flushed := b.flush(); flushed != nil {
			chunks = append(chunks, *flushed)
		}

		// Emit jumbo unit as its own chunk
		text := b.frontBlock + unitText
		tokens := b.frontTokens + unitTokens
		chunkIndex := b.index
		b.index++

		jumbo := Chunk{
			FilePath:   b.filePath,
			FileTitle:  b.fileTitle,
			ChunkIndex: chunkIndex,
			Text:       text,
			Tokens:     tokens,
		}
		chunks = append(chunks, jumbo)
		return chunks
	}

	// Case 2: Normal unit
	needTokens := b.tokens + unitTokens

	if needTokens > b.bodyBudget {
		// Won't fit: flush current chunk first
		if flushed := b.flush(); flushed != nil {
			chunks = append(chunks, *flushed)
		}
	}

	// Add unit to current chunk
	b.parts = append(b.parts, unitText)
	b.tokens += unitTokens

	return chunks
}

// flush creates a chunk from accumulated content and resets the builder.
// Returns nil if there's nothing to flush.
func (b *chunkBuilder) flush() *Chunk {
	if len(b.parts) == 0 {
		return nil
	}

	// Build chunk text: frontmatter + accumulated body parts
	body := strings.Join(b.parts, "")
	text := b.frontBlock + body
	tokens := b.frontTokens + b.tokens
	chunkIndex := b.index
	b.index++

	// Reset builder for next chunk
	b.parts = make([]string, 0)
	b.tokens = 0

	chunk := Chunk{
		FilePath:   b.filePath,
		FileTitle:  b.fileTitle,
		ChunkIndex: chunkIndex,
		Text:       text,
		Tokens:     tokens,
	}

	return &chunk
}
