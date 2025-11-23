package chunker

import (
	"context"
	"fmt"
	"log/slog"

	cctx "github.com/wyvernzora/chunky/pkg/context"
	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	fmbuiltin "github.com/wyvernzora/chunky/pkg/frontmatter/builtin"
	pbuiltin "github.com/wyvernzora/chunky/pkg/parser/builtin"
	"github.com/wyvernzora/chunky/pkg/section"
	sbuiltin "github.com/wyvernzora/chunky/pkg/section/builtin"
	tbuiltin "github.com/wyvernzora/chunky/pkg/tokenizer/builtin"
)

// Chunker processes markdown documents and splits them into token-sized chunks.
// It maintains state across multiple Push calls and accumulates chunks.
type Chunker interface {
	// Push processes a document and adds its chunks to the internal collection.
	// Documents with "do_not_embed: true" in frontmatter are skipped.
	Push(ctx context.Context, input Input) error

	// Chunks returns all accumulated chunks from previous Push calls.
	// The returned slice should not be modified by the caller.
	Chunks() []Chunk

	// Reset clears all accumulated chunks, preparing for a new batch.
	Reset()

	// EffectiveBudget returns the actual token budget available for body content
	// after accounting for reserved overhead.
	EffectiveBudget() int
}

// Input represents a document to be chunked.
type Input struct {
	// Path is the logical path of the document (e.g., "docs/guide.md").
	Path string

	// Title is the human-readable title for the document.
	Title string

	// Markdown is the raw markdown content including optional frontmatter.
	Markdown string
}

// New creates a new Chunker with the given options.
//
// Required options:
//   - WithChunkTokenBudget: Maximum tokens per chunk (must be > 0)
//
// Optional configuration:
//   - WithReservedOverheadRatio: Fraction reserved for overhead (default: 0.1)
//   - WithTokenizer: Custom tokenizer (default: TiktokenTokenizer with o200k_base)
//   - WithParser: Custom parser (default: DefaultParser from parser/builtin)
//   - WithChunkHeaderGenerator: Custom header generator (default: YAML frontmatter)
//   - WithFrontMatterTransform: Add frontmatter transforms (appends to defaults)
//   - WithSectionTransform: Add section transforms (appends to defaults)
//
// Default frontmatter transforms (applied in order):
//   - InjectFilePath: Adds file_path from context
//
// Default section transforms (applied in order):
//   - NormalizeNewlines: Convert CRLF to LF
//   - NormalizeHardWraps: Merge hard-wrapped paragraphs
//   - PruneLeadingBlankLines: Remove leading blank lines
//   - PruneTrailingBlankLines: Remove trailing blank lines
//   - CollapseBlankLines: Collapse 3+ blank lines to 2
//   - HeadingPrefix: Add "## " prefix to headings
//   - HeadingPathComment: Add "<!-- heading.path -->" comments
//
// The effective token budget for body content is calculated as:
//
//	effectiveBudget = chunkTokenBudget * (1 - reservedOverheadRatio)
//
// Returns an error if:
//   - WithChunkTokenBudget was not provided or is <= 0
//   - WithReservedOverheadRatio is < 0 or >= 1
//   - Default tokenizer initialization fails
//
// Example:
//
//	chunker, err := New(
//	    WithChunkTokenBudget(1000),
//	    WithReservedOverheadRatio(0.1),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Example with custom tokenizer:
//
//	tok, _ := builtin.NewWordCountTokenizer()
//	chunker, err := New(
//	    WithChunkTokenBudget(1000),
//	    WithTokenizer(tok),
//	)
func New(opts ...Option) (Chunker, error) {
	// Initialize config with defaults
	cfg := &options{
		chunkTokenBudget:      0,
		reservedOverheadRatio: 0.1,
		tokenizer:             nil,
		parser:                nil,
		headerGenerator:       nil,
		fmTransforms: []fm.Transform{
			// Default: inject file path into frontmatter
			fmbuiltin.InjectFilePath("file_path"),
		},
		sectionTransforms: []section.Transform{
			// Default section transforms applied in order
			sbuiltin.NormalizeNewlinesTransform(),
			sbuiltin.NormalizeHardWrapsTransform(),
			sbuiltin.PruneLeadingBlankLinesTransform(0),  // Remove all leading blanks
			sbuiltin.PruneTrailingBlankLinesTransform(0), // Remove all trailing blanks
			sbuiltin.CollapseBlankLinesTransform(),
			sbuiltin.HeadingPrefixTransform(),
			sbuiltin.HeadingPathCommentTransform(),
		},
	}

	// Apply options (these append to or override defaults)
	for _, opt := range opts {
		opt(cfg)
	}

	// Validate configuration
	if cfg.chunkTokenBudget <= 0 {
		return nil, fmt.Errorf("WithChunkTokenBudget is required and must be > 0, got %d", cfg.chunkTokenBudget)
	}

	if cfg.reservedOverheadRatio < 0.0 || cfg.reservedOverheadRatio >= 1.0 {
		return nil, fmt.Errorf("WithReservedOverheadRatio must be >= 0 and < 1, got %f", cfg.reservedOverheadRatio)
	}

	// Set defaults
	if cfg.tokenizer == nil {
		tok, err := tbuiltin.NewTiktokenTokenizer()
		if err != nil {
			return nil, fmt.Errorf("failed to create default tokenizer: %w", err)
		}
		cfg.tokenizer = tok
	}

	if cfg.parser == nil {
		cfg.parser = pbuiltin.DefaultParser
	}

	if cfg.headerGenerator == nil {
		cfg.headerGenerator = defaultHeaderGenerator
	}

	effectiveBudget := int(float64(cfg.chunkTokenBudget) * (1.0 - cfg.reservedOverheadRatio))

	return &defaultChunker{
		config:          cfg,
		effectiveBudget: effectiveBudget,
		chunks:          nil,
	}, nil
}

// defaultChunker is the standard implementation of the Chunker interface.
type defaultChunker struct {
	config          *options
	effectiveBudget int
	chunks          []Chunk
}

// Push implements Chunker.Push.
func (c *defaultChunker) Push(ctx context.Context, input Input) error {
	// Validate input
	if input.Path == "" {
		return fmt.Errorf("Input.Path cannot be empty")
	}
	if input.Title == "" {
		return fmt.Errorf("Input.Title cannot be empty")
	}
	if input.Markdown == "" {
		return fmt.Errorf("Input.Markdown cannot be empty")
	}

	// Add file info to context
	ctx = cctx.WithFileInfo(ctx, cctx.FileInfo{
		Path:  input.Path,
		Title: input.Title,
	})
	logger := cctx.Logger(ctx)

	logger.Debug("chunker: parsing document",
		slog.String("path", input.Path),
		slog.String("title", input.Title))

	// Check for cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled before parsing %s: %w", input.Path, err)
	}

	// Parse markdown
	root, frontmatter, err := c.config.parser(ctx, []byte(input.Markdown))
	if err != nil {
		logger.Error("chunker: parse failed", slog.Any("error", err))
		return fmt.Errorf("parse failed for %s: %w", input.Path, err)
	}

	// Apply frontmatter transforms
	for i, transform := range c.config.fmTransforms {
		// Check for cancellation
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled during frontmatter transform for %s: %w", input.Path, err)
		}

		if err := transform(ctx, frontmatter); err != nil {
			logger.Error("chunker: frontmatter transform failed",
				slog.Int("transform_index", i),
				slog.Any("error", err))
			return fmt.Errorf("frontmatter transform %d failed for %s: %w", i, input.Path, err)
		}
	}

	// Check for do_not_embed flag
	if doNotEmbed, ok := frontmatter["do_not_embed"].(bool); ok && doNotEmbed {
		logger.Debug("chunker: skipping document with do_not_embed=true",
			slog.String("path", input.Path))
		return nil
	}

	// Apply section transforms
	for i, transform := range c.config.sectionTransforms {
		// Check for cancellation
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled during section transform for %s: %w", input.Path, err)
		}

		if err := section.ApplyTransform(ctx, frontmatter, root, transform); err != nil {
			logger.Error("chunker: section transform failed",
				slog.Int("transform_index", i),
				slog.Any("error", err))
			return fmt.Errorf("section transform %d failed for %s: %w", i, input.Path, err)
		}
	}

	// Generate chunk header
	frontBlock, err := c.config.headerGenerator(ctx, frontmatter.View())
	if err != nil {
		logger.Error("chunker: header generation failed", slog.Any("error", err))
		return fmt.Errorf("header generation failed for %s: %w", input.Path, err)
	}

	// Count header tokens
	frontTokens, err := c.config.tokenizer.Count(frontBlock)
	if err != nil {
		logger.Error("chunker: frontmatter token counting failed", slog.Any("error", err))
		return fmt.Errorf("frontmatter token counting failed for %s: %w", input.Path, err)
	}

	logger.Debug("chunker: frontmatter counted",
		slog.Int("tokens", frontTokens),
		slog.Int("bytes", len(frontBlock)))

	// Calculate body budget
	bodyBudget := c.effectiveBudget - frontTokens
	if bodyBudget <= 0 {
		logger.Warn("chunker: no budget remaining for body content",
			slog.Int("frontmatter_tokens", frontTokens),
			slog.Int("effective_budget", c.effectiveBudget))
		return fmt.Errorf("frontmatter (%d tokens) exceeds effective budget (%d tokens) for %s",
			frontTokens, c.effectiveBudget, input.Path)
	}

	logger.Debug("chunker: body budget calculated",
		slog.Int("body_budget", bodyBudget))

	// Tokenize section tree
	tokenizedRoot, err := c.config.tokenizer.Tokenize(ctx, root)
	if err != nil {
		logger.Error("chunker: tokenization failed", slog.Any("error", err))
		return fmt.Errorf("tokenization failed for %s: %w", input.Path, err)
	}

	logger.Debug("chunker: section tree tokenized",
		slog.Int("subtree_tokens", tokenizedRoot.GetSubtreeTokens()))

	// Generate chunks
	chunks := chunkDocument(chunkDocumentParams{
		filePath:    input.Path,
		fileTitle:   input.Title,
		frontBlock:  frontBlock,
		frontTokens: frontTokens,
		bodyBudget:  bodyBudget,
		root:        tokenizedRoot,
	})

	logger.Debug("chunker: document chunked",
		slog.Int("chunk_count", len(chunks)),
		slog.String("path", input.Path))

	// Accumulate chunks
	c.chunks = append(c.chunks, chunks...)

	return nil
}

// Chunks implements Chunker.Chunks.
func (c *defaultChunker) Chunks() []Chunk {
	return c.chunks
}

// Reset implements Chunker.Reset.
func (c *defaultChunker) Reset() {
	c.chunks = nil
}

// EffectiveBudget implements Chunker.EffectiveBudget.
func (c *defaultChunker) EffectiveBudget() int {
	return c.effectiveBudget
}
