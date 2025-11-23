// Package context provides context utilities for the chunky library.
//
// It extends the standard context package with typed values for passing
// metadata throughout the chunking pipeline, particularly file information
// and logging.
//
// # FileInfo
//
// FileInfo holds document metadata:
//
//	type FileInfo struct {
//	    Path  string // Logical file path (e.g., "docs/guide.md")
//	    Title string // Human-readable title
//	}
//
// Store and retrieve file info from context:
//
//	ctx = context.WithFileInfo(ctx, context.FileInfo{
//	    Path:  "docs/guide.md",
//	    Title: "User Guide",
//	})
//
//	info, ok := context.GetFileInfo(ctx)
//	if ok {
//	    fmt.Println(info.Path)
//	}
//
// # Logging
//
// The package provides access to structured logging via slog:
//
//	logger := context.Logger(ctx)
//	logger.Info("processing document",
//	    slog.String("path", path),
//	    slog.Int("chunks", count))
//
// If no logger is configured in the context, a default logger is returned.
//
// # Usage in Pipeline
//
// Context flows through the entire chunking pipeline:
//
//  1. Chunker.Push adds FileInfo to context
//  2. Parser receives context for logging
//  3. Transforms access FileInfo and Logger
//  4. Tokenizer uses context for cancellation
//  5. Header generator reads FileInfo for metadata
//
// This allows transforms to access document metadata without
// explicit parameter passing:
//
//	func MyTransform(ctx context.Context, fm frontmatter.FrontMatter, s *section.Section) error {
//	    info, _ := context.GetFileInfo(ctx)
//	    logger := context.Logger(ctx)
//	    logger.Info("transforming", slog.String("path", info.Path))
//	    // ... transformation logic
//	}
package context
