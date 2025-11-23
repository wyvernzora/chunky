// Package parser provides markdown parsing functionality.
//
// It defines the Parser type for converting markdown text into section trees
// with extracted frontmatter. The builtin subpackage provides a default
// implementation using the goldmark library.
//
// # Parser Type
//
// Parser is a function type that parses markdown:
//
//	type Parser func(
//	    ctx context.Context,
//	    markdown []byte,
//	) (*section.Section, frontmatter.FrontMatter, error)
//
// # Output Structure
//
// The parser returns:
//   - Section tree: Hierarchical document structure with headings and content
//   - FrontMatter: YAML frontmatter extracted from document header
//   - Error: Parsing errors if any
//
// # Default Parser
//
// The builtin.DefaultParser provides:
//   - YAML frontmatter extraction
//   - Hierarchical section tree based on heading levels
//   - Content preservation with proper nesting
//   - Support for documents without frontmatter
//
// # Usage Example
//
//	root, fm, err := builtin.DefaultParser(ctx, []byte(markdown))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Access section tree
//	fmt.Println(root.Title())
//	for _, child := range root.Children() {
//	    fmt.Println(child.Title())
//	}
//
//	// Access frontmatter
//	if author, ok := fm["author"].(string); ok {
//	    fmt.Println("Author:", author)
//	}
package parser
