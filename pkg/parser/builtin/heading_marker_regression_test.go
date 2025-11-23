package parser

import (
	"context"
	"strings"
	"testing"

	cctx "github.com/wyvernzora/chunky/pkg/context"
	"github.com/wyvernzora/chunky/pkg/section"
)

// TestHeadingMarkersNotInContent_Regression tests that heading markers
// (like "# " or "## ") do not appear in section content after parsing.
//
// This was a bug where goldmark's segment.Start pointed to the beginning
// of the heading TEXT, not the beginning of the LINE, causing the parser
// to include the heading markers in the previous section's content.
//
// Example of the bug:
//
//	Root section content: "# "
//	Child section content: "Some text\n## "
//
// After the fix, section content should never include heading markers.
func TestHeadingMarkersNotInContent_Regression(t *testing.T) {
	markdown := `# Document Root

Some intro content.

## Core Themes & Conflicts

This section has some content about the core themes.

### Hope as Hazard

This is the section where the bug might appear.

Content about hope as a hazard.
`

	ctx := cctx.WithFileInfo(context.Background(), cctx.FileInfo{
		Path:  "test.md",
		Title: "test",
	})

	root, _, err := DefaultParser(ctx, []byte(markdown))
	if err != nil {
		t.Fatalf("parser failed: %v", err)
	}

	// Check that root content does not contain "# "
	rootContent := root.Content()
	if strings.Contains(rootContent, "# ") {
		t.Errorf("root section content should not contain heading marker '# ', got: %q", rootContent)
	}
	if strings.Contains(rootContent, "## ") {
		t.Errorf("root section content should not contain heading marker '## ', got: %q", rootContent)
	}

	// Check all child sections recursively
	var checkSection func(*testing.T, string, string)
	checkSection = func(t *testing.T, path string, content string) {
		// Content should not start with heading markers
		trimmed := strings.TrimSpace(content)
		if strings.HasPrefix(trimmed, "# ") {
			t.Errorf("section %q content should not start with '# ', got: %q", path, trimmed[:min(20, len(trimmed))])
		}
		if strings.HasPrefix(trimmed, "## ") {
			t.Errorf("section %q content should not start with '## ', got: %q", path, trimmed[:min(20, len(trimmed))])
		}
		if strings.HasPrefix(trimmed, "### ") {
			t.Errorf("section %q content should not start with '### ', got: %q", path, trimmed[:min(20, len(trimmed))])
		}

		// Content should not end with heading markers (except in actual text)
		// Check for standalone heading markers at the end
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			// Check if line is ONLY heading markers (the bug pattern)
			if trimmedLine == "#" || trimmedLine == "##" || trimmedLine == "###" ||
				trimmedLine == "# " || trimmedLine == "## " || trimmedLine == "### " {
				t.Errorf("section %q has line with only heading markers: %q", path, trimmedLine)
			}
		}
	}

	checkSection(t, "root", rootContent)

	// Walk all descendants
	var walk func(t *testing.T, s *section.Section, parentPath string)
	walk = func(t *testing.T, s *section.Section, parentPath string) {
		path := parentPath + " / " + s.Title()
		checkSection(t, path, s.Content())
		for _, child := range s.Children() {
			walk(t, child, path)
		}
	}

	for _, child := range root.Children() {
		walk(t, child, "root")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
