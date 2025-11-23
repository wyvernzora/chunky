// Package section provides hierarchical document structure representation.
//
// It defines the Section type for representing markdown documents as trees
// of headings and content, along with a transform system for modifying
// section content during processing.
//
// # Section Type
//
// Section represents a node in the document tree:
//
//	type Section struct {
//	    // Title: Heading text
//	    // Level: Heading level (1-6, 0 for root)
//	    // Content: Text content before first child
//	    // Children: Subsections
//	}
//
// Sections form a tree where:
//   - Root has level 0 and document title
//   - Child levels must be greater than parent level
//   - Content includes text up to first subheading
//   - Children maintain document order
//
// # Transform System
//
// Transforms modify section content during processing:
//
//	type Transform func(
//	    ctx context.Context,
//	    frontmatter frontmatter.FrontMatter,
//	    section *Section,
//	) error
//
// Transforms can:
//   - Normalize whitespace and line endings
//   - Add heading annotations
//   - Trim or collapse blank lines
//   - Inject metadata comments
//
// # Built-in Transforms
//
// The builtin subpackage provides:
//
//  1. NormalizeNewlinesTransform: Converts CRLF to LF
//  2. NormalizeHardWrapsTransform: Merges hard-wrapped paragraphs
//  3. PruneLeadingBlankLinesTransform: Removes leading blank lines
//  4. PruneTrailingBlankLinesTransform: Removes trailing blank lines
//  5. CollapseBlankLinesTransform: Collapses 3+ blank lines to 2
//  6. HeadingPrefixTransform: Add Section heading to its content
//  7. HeadingPathCommentTransform: Adds "<!-- heading.path -->" comments
//
// # Tree Operations
//
// Common patterns for working with section trees:
//
//	// Create tree
//	root := section.NewRoot("Document Title")
//	child := root.CreateChild("Section 1", 1, "Content...")
//	grandchild := child.CreateChild("Subsection", 2, "More content...")
//
//	// Traverse tree
//	for _, child := range root.Children() {
//	    fmt.Println(child.Title())
//	}
//
//	// Modify content
//	section.SetContent("New content")
//	section.AppendContent("\nAdditional text")
//
//	// Get hierarchical path
//	path := section.Path() // ["Document Title", "Section 1", "Subsection"]
//
// # Usage with Transforms
//
//	// Apply single transform
//	err := section.ApplyTransform(ctx, fm, root, builtin.NormalizeNewlinesTransform())
//
//	// Apply multiple transforms
//	transforms := []section.Transform{
//	    builtin.NormalizeNewlinesTransform(),
//	    builtin.CollapseBlankLinesTransform(),
//	    builtin.HeadingPrefixTransform(),
//	}
//	for _, transform := range transforms {
//	    if err := section.ApplyTransform(ctx, fm, root, transform); err != nil {
//	        log.Fatal(err)
//	    }
//	}
package section
