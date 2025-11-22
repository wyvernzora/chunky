package parser

import "context"

// Parser transforms a Markdown document into a hierarchical Section tree
// with structured frontmatter metadata.
//
// Parameters:
//   - ctx: Context for cancellation and logger propagation (see github.com/wyvernzora/chunky/pkg/log)
//   - title: The title to assign to the root Section
//   - markdown: Raw Markdown content (may include YAML frontmatter)
//
// Returns:
//   - *Section: Root section representing the entire document. All heading-based
//     sections in the document appear as descendants of this root.
//   - FrontMatter: Parsed frontmatter metadata as a map. If no frontmatter exists,
//     an empty FrontMatter map is returned (never nil).
//   - error: Any parsing errors encountered
//
// Implementations are free to use any parsing strategy as long as they produce
// a consistent Section hierarchy from the Markdown input.
type Parser func(ctx context.Context, title string, markdown []byte) (*Section, FrontMatter, error)
