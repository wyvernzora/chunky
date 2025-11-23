// Package frontmatter provides frontmatter handling and transformation.
//
// It defines types for working with YAML frontmatter metadata and provides
// a transform system for modifying frontmatter during document processing.
//
// # FrontMatter Type
//
// FrontMatter is a map holding metadata key-value pairs:
//
//	type FrontMatter map[string]any
//
// Common fields include: title, author, date, tags, description, etc.
//
// # Transform System
//
// Transforms modify frontmatter during processing:
//
//	type Transform func(ctx context.Context, fm FrontMatter) error
//
// Transforms can:
//   - Add metadata from context (file paths, timestamps)
//   - Validate required fields
//   - Merge default values
//   - Convert field types
//
// # Built-in Transforms
//
// The builtin subpackage provides:
//
//  1. InjectFilePath: Adds file path from context to frontmatter
//     - Reads from context.FileInfo
//     - Configurable field name
//
//  2. MergeFrontMatter: Merges additional metadata without overwriting
//     - Only adds missing keys
//     - Preserves existing values
//     - Useful for defaults
//
//  3. RequireSummary: Validates that summary field exists
//     - Returns error if missing
//     - Useful for enforcing metadata requirements
//
// # FrontMatterView
//
// FrontMatterView provides read-only access to frontmatter:
//   - Prevents external modification
//   - Used in header generation
//   - Deep copies data for safety
//
// # Usage Example
//
//	// Create transform
//	injectPath := builtin.InjectFilePath("file_path")
//	addDefaults := builtin.MergeFrontMatter(frontmatter.FrontMatter{
//	    "author": "Anonymous",
//	    "version": "1.0",
//	})
//
//	// Apply transforms
//	fm := frontmatter.FrontMatter{"title": "Doc"}
//	injectPath(ctx, fm)
//	addDefaults(ctx, fm)
//
//	// Serialize to YAML
//	yaml, err := frontmatter.Serialize(fm)
package frontmatter
