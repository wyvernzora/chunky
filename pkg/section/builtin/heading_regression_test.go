package builtin

import (
	"context"
	"strings"
	"testing"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
)

// TestHeadingPrefixWithPathComment_Regression tests that the heading prefix
// is not incorrectly added to the path comment line.
// This is a regression test for the bug where output showed:
//
//	## <!-- path: ... -->
//	## Hope as Hazard
func TestHeadingPrefixWithPathComment_Regression(t *testing.T) {
	// Simulate what happens when both transforms run in sequence
	root := section.NewRoot("Document")
	child := root.CreateChild("Hope as Hazard", 2, "Content here")

	// First, path comment transform runs
	pathTransform := HeadingPathCommentTransform()
	err := pathTransform(context.Background(), fm.EmptyFrontMatter().View(), child)
	if err != nil {
		t.Fatalf("path transform failed: %v", err)
	}

	// Then, heading prefix transform runs
	prefixTransform := HeadingPrefixTransform()
	err = prefixTransform(context.Background(), fm.EmptyFrontMatter().View(), child)
	if err != nil {
		t.Fatalf("prefix transform failed: %v", err)
	}

	result := child.Content()

	// The result should be:
	// <!-- path: Document / Hope as Hazard -->
	// ## Hope as Hazard
	//
	// Content here

	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d: %q", len(lines), result)
	}

	// Line 0 should be the path comment WITHOUT any ## prefix
	if !strings.HasPrefix(lines[0], "<!--") {
		t.Errorf("line 0 should start with <!--, got: %q", lines[0])
	}
	if strings.HasPrefix(lines[0], "##") {
		t.Errorf("BUG: line 0 should NOT start with ##, got: %q", lines[0])
	}

	// Line 1 should be the heading
	if !strings.HasPrefix(lines[1], "##") {
		t.Errorf("line 1 should start with ##, got: %q", lines[1])
	}
	if strings.Contains(lines[1], "<!--") {
		t.Errorf("line 1 should NOT contain <!--, got: %q", lines[1])
	}

	// Expected format
	expectedPrefix := "<!-- path: Document / Hope as Hazard -->\n## Hope as Hazard\n\nContent here"
	if result != expectedPrefix {
		t.Errorf("unexpected result:\nexpected: %q\ngot: %q", expectedPrefix, result)
	}
}

// TestHeadingPrefixWithExistingHeading_Regression tests the case where
// the section content already contains a heading line.
func TestHeadingPrefixWithExistingHeading_Regression(t *testing.T) {
	root := section.NewRoot("Document")
	// Content already has the heading line - abnormal case
	child := root.CreateChild("Hope as Hazard", 2, "## Hope as Hazard\nContent here")

	// Path comment transform runs first
	pathTransform := HeadingPathCommentTransform()
	err := pathTransform(context.Background(), fm.EmptyFrontMatter().View(), child)
	if err != nil {
		t.Fatalf("path transform failed: %v", err)
	}

	// Then heading prefix transform runs
	prefixTransform := HeadingPrefixTransform()
	err = prefixTransform(context.Background(), fm.EmptyFrontMatter().View(), child)
	if err != nil {
		t.Fatalf("prefix transform failed: %v", err)
	}

	result := child.Content()
	lines := strings.Split(result, "\n")

	// Check that ## is NOT added to the comment line
	if strings.HasPrefix(lines[0], "##") {
		t.Errorf("BUG REPRODUCED: line 0 should NOT start with ##, got: %q", lines[0])
	}

	t.Logf("Result:\n%s", result)
}
