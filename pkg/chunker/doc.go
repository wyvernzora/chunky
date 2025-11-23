// Package chunker provides intelligent markdown document chunking for embedding pipelines.
//
// It splits large documents into token-sized chunks while preserving semantic structure
// through a configurable pipeline of transforms. The chunker maintains document hierarchy
// and allows customization of tokenization, parsing, and header generation.
//
// # Basic Usage
//
// Create a chunker with a token budget and process documents:
//
//	chunker, err := chunker.New(
//	    chunker.WithChunkTokenBudget(1000),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	err = chunker.Push(ctx, chunker.Input{
//	    Path:     "docs/guide.md",
//	    Title:    "User Guide",
//	    Markdown: content,
//	})
//
//	chunks := chunker.Chunks()
//
// # Architecture
//
// The chunker operates through a pipeline:
//
//  1. Parse markdown into section tree with frontmatter
//  2. Apply frontmatter transforms (inject metadata, validate, etc.)
//  3. Apply section transforms (normalize text, add annotations)
//  4. Tokenize the section tree
//  5. Generate chunks using greedy algorithm to fit token budget
//
// # Transforms
//
// Default frontmatter transforms:
//   - InjectFilePath: Adds file_path from context
//
// Default section transforms:
//   - NormalizeNewlines: Convert CRLF to LF
//   - NormalizeHardWraps: Merge hard-wrapped paragraphs
//   - PruneLeadingBlankLines: Remove leading blank lines
//   - PruneTrailingBlankLines: Remove trailing blank lines
//   - CollapseBlankLines: Collapse 3+ blank lines to 2
//   - HeadingPrefix: Add Section heading to its content
//   - HeadingPathComment: Add "<!-- heading.path -->" comments
//
// # Customization
//
// All components can be customized via options:
//
//	chunker, err := chunker.New(
//	    chunker.WithChunkTokenBudget(2000),
//	    chunker.WithReservedOverheadRatio(0.15),
//	    chunker.WithTokenizer(myTokenizer),
//	    chunker.WithParser(myParser),
//	    chunker.WithChunkHeaderGenerator(myGenerator),
//	    chunker.WithFrontMatterTransform(myTransform),
//	    chunker.WithSectionTransform(myTransform),
//	)
//
// # Token Budget
//
// The token budget accounts for both chunk headers and body content:
//
//	effectiveBudget = chunkTokenBudget * (1 - reservedOverheadRatio)
//
// The reserved overhead ratio (default 0.1 or 10%) reserves budget for
// downstream processing in the embedding pipeline, ensuring the total
// doesn't exceed limits. Chunk header tokens are separately accounted for
// in each chunk's token count.
package chunker
