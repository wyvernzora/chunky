package parser

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/adrg/frontmatter"
	cctx "github.com/wyvernzora/chunky/pkg/context"
	cfm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// DefaultParser is the default implementation of pkg.Parser.
//
// It operates in four stages:
//  1. Extract YAML frontmatter from the document header (delimited by "---")
//  2. Parse the remaining Markdown into an AST using goldmark
//  3. Walk the AST to identify heading locations, levels, and titles
//  4. Fold the headings and intervening text into a nested Section structure
//
// The root section title is derived from context using chunkyctx.FileInfoFrom().
// If no FileInfo is present in context, the title defaults to "Untitled".
//
// Each call creates a fresh internal worker instance, making the function
// safe for concurrent use.
func DefaultParser(ctx context.Context, markdown []byte) (*section.Section, cfm.FrontMatter, error) {
	w := &worker{ctx: ctx}
	return w.parse(markdown)
}

// worker is the internal parser implementation that holds state during parsing.
type worker struct {
	ctx    context.Context
	src    []byte           // source bytes (frontmatter removed)
	doc    ast.Node         // goldmark AST root
	spans  []headingSpan    // ordered headings extracted from AST
	cursor int              // current byte position during section folding
	stack  []sectionFrame   // section stack for nesting logic
	root   *section.Section // final parsed section tree
}

func (w *worker) parse(markdown []byte) (*section.Section, cfm.FrontMatter, error) {
	logger := cctx.Logger(w.ctx)

	// Derive title from context
	title := "Untitled"
	if fi, ok := cctx.FileInfoFrom(w.ctx); ok {
		title = fi.Title
	}

	logger.Debug("starting document parse",
		slog.String("title", title),
		slog.Int("markdown_size", len(markdown)))

	// 1) extract frontmatter
	var fm cfm.FrontMatter
	body, err := frontmatter.Parse(bytes.NewReader(markdown), &fm)
	if err != nil {
		logger.Error("frontmatter parsing failed", slog.Any("error", err))
		return nil, nil, err
	}
	if fm == nil {
		fm = cfm.EmptyFrontMatter()
	}
	w.src = []byte(body)
	logger.Debug("frontmatter extracted",
		slog.Int("frontmatter_keys", len(fm)),
		slog.Int("body_size", len(w.src)))

	// 2) parse Markdown AST using goldmark
	if err := w.parseDoc(); err != nil {
		logger.Error("markdown AST parsing failed", slog.Any("error", err))
		return nil, nil, err
	}
	logger.Debug("markdown AST parsed")

	// 3) collect heading spans (offsets + titles)
	w.extractHeadings()
	logger.Debug("heading spans extracted", slog.Int("heading_count", len(w.spans)))

	// 4) fold spans + raw text into a Section tree
	if err := w.fold(title); err != nil {
		logger.Error("section folding failed", slog.Any("error", err))
		return nil, nil, err
	}
	logger.Debug("section tree folded successfully",
		slog.Int("root_children", len(w.root.Children())))

	return w.root, fm, nil
}

// --- Internal data structures ------------------------------------------------

type sectionFrame struct{ s *section.Section }

type headingSpan struct {
	Node  *ast.Heading // goldmark AST node
	Start int          // byte offset where heading line begins
	End   int          // byte offset where heading line ends
	Level int          // nesting depth (1=h1, 2=h2, etc.)
	Title string       // rendered heading text with inline formatting stripped
}

// --- Stage 1: parse the document AST ----------------------------------------

func (w *worker) parseDoc() error {
	md := goldmark.New(
		goldmark.WithParserOptions(
			gparser.WithAutoHeadingID(), // OK to keep; not strictly required
		),
	)
	w.doc = md.Parser().Parse(text.NewReader(w.src))
	if w.doc == nil {
		return errors.New("goldmark: empty document root")
	}
	return nil
}

// --- Stage 2: collect heading spans -----------------------------------------

func (w *worker) extractHeadings() {
	logger := cctx.Logger(w.ctx)
	var spans []headingSpan

	ast.Walk(w.doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		lines := h.Lines()
		if lines.Len() == 0 {
			return ast.WalkContinue, nil
		}
		seg := lines.At(0)

		title := inlineText(h, w.src)
		spans = append(spans, headingSpan{
			Node:  h,
			Start: seg.Start,
			End:   seg.Stop,
			Level: h.Level,
			Title: title,
		})
		logger.Debug("heading discovered",
			slog.Int("level", h.Level),
			slog.String("title", title),
			slog.Int("start", seg.Start),
			slog.Int("end", seg.Stop))
		return ast.WalkContinue, nil
	})

	w.spans = spans
}

func inlineText(h *ast.Heading, src []byte) string {
	var buf bytes.Buffer
	for n := h.FirstChild(); n != nil; n = n.NextSibling() {
		switch t := n.(type) {
		case *ast.Text:
			buf.Write(t.Segment.Value(src))
		default:
			buf.WriteString(extractInlineText(t, src))
		}
	}
	return buf.String()
}

func extractInlineText(n ast.Node, src []byte) string {
	var buf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch t := c.(type) {
		case *ast.Text:
			buf.Write(t.Segment.Value(src))
		default:
			buf.WriteString(extractInlineText(t, src))
		}
	}
	return buf.String()
}

// --- Stage 3: fold spans into a Section tree --------------------------------

func (w *worker) fold(docTitle string) error {
	logger := cctx.Logger(w.ctx)
	w.root = section.NewRoot(docTitle)
	w.stack = []sectionFrame{{s: w.root}}
	w.cursor = 0

	logger.Debug("starting section folding", slog.String("root_title", docTitle))

	for i, h := range w.spans {
		// append pre-heading text to current section
		if h.Start > w.cursor {
			pre, next := spliceText(w.src, w.cursor, h.Start)
			w.stack[len(w.stack)-1].s.AppendContent(pre)
			logger.Debug("appended pre-heading content",
				slog.Int("heading_index", i),
				slog.Int("content_length", len(pre)))
			w.cursor = next
		}

		// find parent section for this heading level
		pi, err := parentForLevel(w.stack, h.Level)
		if err != nil {
			logger.Error("failed to find parent section",
				slog.Int("heading_index", i),
				slog.String("heading_title", h.Title),
				slog.Int("heading_level", h.Level),
				slog.Any("error", err))
			return fmt.Errorf("invalid section stack at heading %d (%q): %w", i, h.Title, err)
		}

		// **IMPORTANT**: truncate stack to parent before pushing child
		w.stack = w.stack[:pi+1]
		parent := w.stack[pi].s

		// create new section under parent
		sec := parent.CreateChild(h.Title, h.Level, "")
		w.stack = append(w.stack, sectionFrame{s: sec})
		logger.Debug("created section",
			slog.String("title", h.Title),
			slog.Int("level", h.Level),
			slog.String("parent", parent.Title()),
			slog.Int("stack_depth", len(w.stack)))

		// move cursor to end of heading line
		w.cursor = h.End
	}

	// final trailing content
	if w.cursor < len(w.src) {
		pre, _ := spliceText(w.src, w.cursor, len(w.src))
		w.stack[len(w.stack)-1].s.AppendContent(pre)
		logger.Debug("appended trailing content", slog.Int("content_length", len(pre)))
	}

	return nil
}

// --- Pure helpers ------------------------------------------------------------

func spliceText(src []byte, start, stop int) (string, int) {
	if start < 0 {
		start = 0
	}
	if stop > len(src) {
		stop = len(src)
	}
	if stop <= start {
		return "", start
	}
	return string(src[start:stop]), stop
}

func parentForLevel(stack []sectionFrame, target int) (int, error) {
	i := len(stack) - 1
	for i >= 0 && stack[i].s.Level() >= target {
		i--
	}
	if i < 0 {
		return -1, errors.New("no valid parent section")
	}
	return i, nil
}
