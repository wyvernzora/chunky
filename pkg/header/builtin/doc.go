// Package builtin provides built-in chunk header generators for the chunky library.
//
// Chunk header generators create the header section that appears at the beginning
// of each chunk. This typically contains metadata from the document's frontmatter
// but can be customized for different formats and use cases.
//
// # Available Generators
//
// FrontMatterYamlHeader generates YAML-formatted headers with --- delimiters:
//
//	gen := builtin.FrontMatterYamlHeader()
//	chunker, err := chunker.New(
//	    chunker.WithChunkTokenBudget(1000),
//	    chunker.WithChunkHeaderGenerator(gen),
//	)
//
// Output:
//
//	---
//	title: My Document
//	author: John Doe
//	---
//
// KeyValueHeader generates simple key-value formatted headers with configurable fields:
//
//	gen := builtin.KeyValueHeader(
//	    builtin.RequiredField("title", "Title"),
//	    builtin.OptionalField("author", "Author"),
//	    builtin.OptionalField("tags", "Tags"),
//	)
//
// Output:
//
//	Title: My Document
//	Author: John Doe
//	Tags: [tag1 tag2 tag3]
//
// # Key-Value Header Fields
//
// The KeyValueHeader generator supports two types of fields:
//
//   - Required fields: Must be present and non-empty, or an error is returned
//   - Optional fields: Skipped if missing or empty
//
// Field values must be scalars (string, bool, number) or slices of scalars.
// Maps and nested structures are not supported and will return an error.
//
// # Empty Value Handling
//
// A field is considered empty if:
//   - The value is nil
//   - The value is an empty string or whitespace-only string
//   - The value is an empty slice
//
// # Value Rendering
//
// The KeyValueHeader generator renders values as follows:
//   - Strings: Rendered as-is without quotes
//   - Numbers and bools: Rendered using Go's default string representation
//   - Slices: Rendered as compact arrays, e.g., [item1 item2 item3]
//
// # Examples
//
// Simple document metadata:
//
//	gen := builtin.KeyValueHeader(
//	    builtin.RequiredField("title", "Title"),
//	    builtin.OptionalField("author", "Author"),
//	    builtin.OptionalField("date", "Date"),
//	)
//
// With tags:
//
//	gen := builtin.KeyValueHeader(
//	    builtin.RequiredField("file_path", "Document"),
//	    builtin.OptionalField("tags", "Tags"),
//	)
//
// If a required field is missing:
//
//	// Returns error: "required field missing or empty: title"
package builtin
