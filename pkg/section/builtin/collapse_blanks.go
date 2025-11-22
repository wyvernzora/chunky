package builtin

import (
	"bytes"
	"context"
	"regexp"
	"sort"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// CollapseBlankLinesTransform returns a transform that collapses 3+ consecutive
// blank lines into exactly two blank lines, preserving verbatim content inside
// fenced/code blocks. It uses goldmark to accurately identify code block boundaries.
//
// The transform is idempotent: applying it multiple times produces the same result.
//
// Notes:
//   - "Blank line" means a line containing only spaces/tabs or nothing.
//   - For CRLF normalization, run NormalizeNewlinesTransform first.
func CollapseBlankLinesTransform() section.Transform {
	return func(ctx context.Context, _ fm.FrontMatterView, s *section.Section) error {
		return collapseBlankLinesImpl(ctx, s)
	}
}

func collapseBlankLinesImpl(_ context.Context, s *section.Section) error {
	src := []byte(s.Content())
	if len(src) == 0 {
		return nil
	}

	// Parse with goldmark to obtain accurate block boundaries.
	md := goldmark.New(goldmark.WithParserOptions(
		gparser.WithAutoHeadingID(), // harmless here
	))
	doc := md.Parser().Parse(text.NewReader(src))

	// Exclusion ranges (byte offsets) for blocks where we must NOT edit.
	exclude := collectNoEditRanges(doc)

	// Compute editable ranges as the complement of exclusions over [0, len(src)).
	editables := complementRanges(exclude, len(src))
	if len(editables) == 0 {
		return nil
	}

	// For each editable range, collapse 3+ blank lines to exactly 2.
	re := regexp.MustCompile(`(?m)(?:\n[ \t]*){3,}`)
	edits := make([]edit, 0, len(editables))

	for _, r := range editables {
		segment := src[r.start:r.end]
		replaced := re.ReplaceAll(segment, []byte("\n\n"))
		if !bytes.Equal(segment, replaced) {
			edits = append(edits, edit{
				start: r.start,
				end:   r.end,
				repl:  replaced,
			})
		}
	}

	if len(edits) == 0 {
		return nil
	}

	// Apply edits in reverse source order to avoid index drift.
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })

	buf := make([]byte, len(src))
	copy(buf, src)
	for _, e := range edits {
		buf = append(buf[:e.start], append(e.repl, buf[e.end:]...)...)
	}

	s.SetContent(string(buf))
	return nil
}

// ---- helpers ----

type span struct{ start, end int }

type edit struct {
	start int
	end   int
	repl  []byte
}

// collectNoEditRanges returns byte spans for blocks that must be left intact
// (fenced/code blocks). It uses node.Lines() to capture exact source ranges.
func collectNoEditRanges(doc ast.Node) []span {
	var out []span

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n.Kind() {
		case ast.KindFencedCodeBlock, ast.KindCodeBlock:
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				seg := lines.At(i)
				out = append(out, span{start: seg.Start, end: seg.Stop})
			}
		default:
			// leave other blocks editable
		}
		return ast.WalkContinue, nil
	})

	// Merge overlaps just in case (defensive; code blocks usually already non-overlapping)
	sort.Slice(out, func(i, j int) bool { return out[i].start < out[j].start })
	out = mergeSpans(out)
	return out
}

func mergeSpans(in []span) []span {
	if len(in) == 0 {
		return in
	}
	out := make([]span, 0, len(in))
	cur := in[0]
	for i := 1; i < len(in); i++ {
		if in[i].start <= cur.end {
			if in[i].end > cur.end {
				cur.end = in[i].end
			}
			continue
		}
		out = append(out, cur)
		cur = in[i]
	}
	out = append(out, cur)
	return out
}

// complementRanges returns the editable regions given excluded spans.
func complementRanges(exclude []span, total int) []span {
	if total <= 0 {
		return nil
	}
	if len(exclude) == 0 {
		return []span{{0, total}}
	}
	var out []span
	cursor := 0
	for _, e := range exclude {
		if e.start > cursor {
			out = append(out, span{cursor, e.start})
		}
		if e.end > cursor {
			cursor = e.end
		}
	}
	if cursor < total {
		out = append(out, span{cursor, total})
	}
	return out
}
