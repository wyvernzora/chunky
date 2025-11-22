package builtin

import (
	"bytes"
	"context"
	"sort"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// NormalizeHardWrapsTransform returns a transform that collapses single newlines
// inside paragraph blocks into a single space, preserving blank-line paragraph breaks
// and leaving code blocks (and any non-paragraph blocks) untouched. It uses goldmark
// to accurately identify paragraph boundaries.
//
// The transform is idempotent: applying it multiple times produces the same result.
//
// Examples (inside a paragraph):
//   - "Line 1\nLine 2"           -> "Line 1 Line 2"
//   - "Line 1\n   Line 2"        -> "Line 1 Line 2"
//   - "Line 1\n\nLine 2"         -> unchanged (paragraph break)
//
// Outside paragraphs (e.g., fenced code), content is not modified.
func NormalizeHardWrapsTransform() section.Transform {
	return func(ctx context.Context, _ fm.FrontMatterView, s *section.Section) error {
		return normalizeHardWrapsImpl(ctx, s)
	}
}

func normalizeHardWrapsImpl(_ context.Context, s *section.Section) error {
	src := []byte(s.Content())
	if len(src) == 0 {
		return nil
	}

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(src))

	// Collect paragraph spans (byte ranges) to edit.
	type span struct{ start, end int }
	var paraSpans []span

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Kind() == ast.KindParagraph {
			lines := n.Lines()
			if lines.Len() > 0 {
				first := lines.At(0)
				last := lines.At(lines.Len() - 1)
				paraSpans = append(paraSpans, span{start: first.Start, end: last.Stop})
			}
		}
		return ast.WalkContinue, nil
	})

	if len(paraSpans) == 0 {
		return nil
	}

	// Produce edits for each paragraph: replace single newlines (not \n\n) with a space.
	type edit struct {
		start int
		end   int
		repl  []byte
	}
	edits := make([]edit, 0, len(paraSpans))

	for _, sp := range paraSpans {
		seg := src[sp.start:sp.end]
		joined := joinSingleNewlines(seg)
		if !bytes.Equal(seg, joined) {
			edits = append(edits, edit{start: sp.start, end: sp.end, repl: joined})
		}
	}

	if len(edits) == 0 {
		return nil
	}

	// Apply edits in reverse order to avoid index drift.
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })

	buf := make([]byte, len(src))
	copy(buf, src)
	for _, e := range edits {
		buf = append(buf[:e.start], append(e.repl, buf[e.end:]...)...)
	}

	s.SetContent(string(buf))
	return nil
}

// joinSingleNewlines replaces single '\n' (optionally followed by spaces/tabs) that
// are NOT part of a blank-line paragraph break (i.e., not '\n\n') with a single space.
//
// No regex lookbehind in Go, so we do a single pass with local context.
func joinSingleNewlines(b []byte) []byte {
	out := make([]byte, 0, len(b))

	i := 0
	for i < len(b) {
		ch := b[i]
		if ch != '\n' {
			out = append(out, ch)
			i++
			continue
		}

		// ch == '\n': look at previous and next to decide.
		prevIsNL := len(out) > 0 && out[len(out)-1] == '\n'

		// Count following whitespace after this newline.
		j := i + 1
		for j < len(b) && (b[j] == ' ' || b[j] == '\t') {
			j++
		}
		nextIsNL := j < len(b) && b[j] == '\n'

		if prevIsNL || nextIsNL {
			// We're in a blank line region (at least one adjacent newline). Preserve this newline.
			out = append(out, '\n')
			i++
			continue
		}

		// Single hard wrap inside a paragraph: replace newline + following indent with a space.
		// Emit one space (avoid doubling if there is already a space at end).
		if len(out) == 0 || out[len(out)-1] != ' ' {
			out = append(out, ' ')
		}
		i = j // skip newline and any following spaces/tabs
	}

	return out
}
