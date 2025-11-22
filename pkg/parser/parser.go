package parser

import (
	"context"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
)

// Parser transforms a Markdown document into a hierarchical Section tree
// with structured frontmatter metadata.
//
// Parameters:
//   - ctx: Context for cancellation and logger propagation (see github.com/wyvernzora/chunky/pkg/context)
//   - markdown: Raw Markdown content (may include YAML frontmatter)
//
// Returns:
//   - *section.Section: Root section representing the entire document. The root section's
//     title should be derived from context (e.g., file path). All heading-based sections
//     in the document appear as descendants of this root.
//   - fm.FrontMatter: Parsed frontmatter metadata. If no frontmatter exists,
//     an empty FrontMatter is returned (never nil).
//   - error: Any parsing errors encountered
//
// Implementations are free to use any parsing strategy as long as they produce
// a consistent Section hierarchy from the Markdown input.
type Parser func(ctx context.Context, markdown []byte) (*section.Section, fm.FrontMatter, error)
