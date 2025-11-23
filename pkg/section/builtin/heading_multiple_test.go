package builtin

import (
	"context"
	"strings"
	"testing"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
)

// TestMultipleSectionsWithBothTransforms tests the exact scenario from the
// bug report: multiple sections in a document with both transforms applied
// in the order they appear in the chunker (HeadingPrefix then HeadingPathComment).
func TestMultipleSectionsWithBothTransforms(t *testing.T) {
	// Create a section tree similar to what the parser would create
	root := section.NewRoot("4. Ending D - Deer in Headlights")

	// First child section - "Core Themes & Conflicts"
	child1 := root.CreateChild("Core Themes & Conflicts", 2, "Some content about themes")

	// Second child section - "Hope as Hazard" (this is where the bug shows up)
	child2 := child1.CreateChild("Hope as Hazard", 3, "Content about hope")

	fmView := fm.EmptyFrontMatter().View()
	ctx := context.Background()

	// Apply transforms in the same order as the chunker
	// 1. HeadingPrefixTransform
	prefixTransform := HeadingPrefixTransform()
	if err := prefixTransform(ctx, fmView, child1); err != nil {
		t.Fatalf("prefix transform failed on child1: %v", err)
	}
	if err := prefixTransform(ctx, fmView, child2); err != nil {
		t.Fatalf("prefix transform failed on child2: %v", err)
	}

	t.Logf("After HeadingPrefixTransform:")
	t.Logf("child1 content:\n%s", child1.Content())
	t.Logf("child2 content:\n%s", child2.Content())

	// 2. HeadingPathCommentTransform
	pathTransform := HeadingPathCommentTransform()
	if err := pathTransform(ctx, fmView, child1); err != nil {
		t.Fatalf("path transform failed on child1: %v", err)
	}
	if err := pathTransform(ctx, fmView, child2); err != nil {
		t.Fatalf("path transform failed on child2: %v", err)
	}

	t.Logf("\nAfter HeadingPathCommentTransform:")
	t.Logf("child1 content:\n%s", child1.Content())
	t.Logf("child2 content:\n%s", child2.Content())

	// Check child2 (where the bug was reported)
	child2Content := child2.Content()
	lines := strings.Split(child2Content, "\n")

	// The first line should be the path comment
	if !strings.HasPrefix(lines[0], "<!-- path:") {
		t.Errorf("line 0 should be path comment, got: %q", lines[0])
	}

	// The path comment should NOT have ## prefix
	if strings.HasPrefix(lines[0], "##") {
		t.Errorf("BUG REPRODUCED: line 0 should NOT start with ##, got: %q", lines[0])
	}

	// The second non-empty line should be the heading
	if !strings.HasPrefix(strings.TrimSpace(lines[1]), "###") {
		t.Errorf("line 1 should be heading, got: %q", lines[1])
	}
}
