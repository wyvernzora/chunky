package chunker

import (
	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/header"
	"github.com/wyvernzora/chunky/pkg/parser"
	"github.com/wyvernzora/chunky/pkg/section"
	"github.com/wyvernzora/chunky/pkg/tokenizer"
)

// Option configures a Chunker instance.
type Option func(*options)

// options holds the internal configuration for a chunker.
type options struct {
	chunkTokenBudget      int
	reservedOverheadRatio float64
	tokenizer             tokenizer.Tokenizer
	parser                parser.Parser
	headerGenerator       header.ChunkHeader
	fmTransforms          []fm.Transform
	sectionTransforms     []section.Transform
}

// WithChunkTokenBudget sets the maximum total tokens per chunk (frontmatter + body).
// This is a required option and must be > 0.
//
// The effective body budget is calculated as:
//
//	effectiveBudget = chunkTokenBudget * (1 - reservedOverheadRatio)
//
// Example:
//
//	chunker := New(WithChunkTokenBudget(1000))
func WithChunkTokenBudget(budget int) Option {
	return func(opts *options) {
		opts.chunkTokenBudget = budget
	}
}

// WithReservedOverheadRatio sets the fraction of budget reserved for overhead.
// Must be in range [0.0, 1.0). Default: 0.1 (10%).
//
// The effective budget for body content is:
//
//	effectiveBudget = chunkTokenBudget * (1 - reservedOverheadRatio)
//
// New() will return an error if the ratio is outside the valid range.
//
// Example:
//
//	// Reserve 15% for overhead
//	chunker, err := New(
//	    WithChunkTokenBudget(1000),
//	    WithReservedOverheadRatio(0.15),
//	)
func WithReservedOverheadRatio(ratio float64) Option {
	return func(opts *options) {
		opts.reservedOverheadRatio = ratio
	}
}

// WithTokenizer sets a custom tokenizer for counting tokens.
// If not provided, defaults to TiktokenTokenizer with o200k_base encoding.
//
// Example:
//
//	tok, _ := builtin.NewWordCountTokenizer(builtin.WithWordsPerToken(0.75))
//	chunker := New(
//	    WithChunkTokenBudget(1000),
//	    WithTokenizer(tok),
//	)
func WithTokenizer(tok tokenizer.Tokenizer) Option {
	return func(opts *options) {
		opts.tokenizer = tok
	}
}

// WithParser sets a custom parser for parsing markdown into section trees.
// If not provided, defaults to the builtin DefaultParser.
//
// Example:
//
//	customParser := &MyParser{}
//	chunker := New(
//	    WithChunkTokenBudget(1000),
//	    WithParser(customParser),
//	)
func WithParser(p parser.Parser) Option {
	return func(opts *options) {
		opts.parser = p
	}
}

// WithChunkHeader sets a custom generator for chunk headers.
// If not provided, defaults to YAML frontmatter serialization.
//
// The header generator creates the text prepended to each chunk's body content,
// typically containing metadata in frontmatter format.
//
// Example:
//
//	// Custom JSON header generator
//	jsonGenerator := func(ctx context.Context, fm fm.FrontMatterView) (string, error) {
//	    data, err := json.Marshal(fm)
//	    if err != nil {
//	        return "", err
//	    }
//	    return fmt.Sprintf("```json\n%s\n```\n\n", string(data)), nil
//	}
//
//	chunker := New(
//	    WithChunkTokenBudget(1000),
//	    WithChunkHeader(jsonGenerator),
//	)
func WithChunkHeader(generator header.ChunkHeader) Option {
	return func(opts *options) {
		opts.headerGenerator = generator
	}
}

// WithFrontMatterTransform adds a frontmatter transform to apply after parsing.
// Transforms run in the order they are added and can add, modify, or remove
// frontmatter fields.
//
// Can be called multiple times to add multiple transforms.
//
// Example:
//
//	chunker := New(
//	    WithChunkTokenBudget(1000),
//	    WithFrontMatterTransform(builtin.RequireSummary()),
//	)
func WithFrontMatterTransform(transform fm.Transform) Option {
	return func(opts *options) {
		opts.fmTransforms = append(opts.fmTransforms, transform)
	}
}

// WithFrontMatterTransforms adds multiple frontmatter transforms at once.
// This is a convenience function equivalent to calling WithFrontMatterTransform
// multiple times.
//
// Example:
//
//	chunker := New(
//	    WithChunkTokenBudget(1000),
//	    WithFrontMatterTransforms(
//	        builtin.InjectFilePath("file_path"),
//	        builtin.RequireSummary(),
//	    ),
//	)
func WithFrontMatterTransforms(transforms ...fm.Transform) Option {
	return func(opts *options) {
		opts.fmTransforms = append(opts.fmTransforms, transforms...)
	}
}

// WithSectionTransform adds a section transform to apply before tokenization.
// Transforms run in the order they are added and can modify the section tree.
//
// Can be called multiple times to add multiple transforms.
//
// Example:
//
//	chunker := New(
//	    WithChunkTokenBudget(1000),
//	    WithSectionTransform(builtin.CollapseBlankLinesTransform()),
//	)
func WithSectionTransform(transform section.Transform) Option {
	return func(opts *options) {
		opts.sectionTransforms = append(opts.sectionTransforms, transform)
	}
}

// WithSectionTransforms adds multiple section transforms at once.
// This is a convenience function equivalent to calling WithSectionTransform
// multiple times.
//
// Example:
//
//	chunker := New(
//	    WithChunkTokenBudget(1000),
//	    WithSectionTransforms(
//	        builtin.CollapseBlankLinesTransform(),
//	        builtin.NormalizeHardWrapsTransform(),
//	    ),
//	)
func WithSectionTransforms(transforms ...section.Transform) Option {
	return func(opts *options) {
		opts.sectionTransforms = append(opts.sectionTransforms, transforms...)
	}
}
