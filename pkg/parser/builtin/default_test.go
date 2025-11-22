package parser

import (
	"context"
	"strings"
	"testing"

	chunkyctx "github.com/wyvernzora/chunky/pkg/context"
	"github.com/wyvernzora/chunky/pkg/section"
)

func TestParserSimpleDocument(t *testing.T) {
	markdown := `# Hello World

This is a simple document.`

	ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Test Document"})
	root, fm, err := DefaultParser(ctx, []byte(markdown))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Title() != "Test Document" {
		t.Errorf("expected root title 'Test Document', got %q", root.Title())
	}
	if root.Level() != 0 {
		t.Errorf("expected root level 0, got %d", root.Level())
	}
	if fm == nil {
		t.Error("frontmatter should not be nil")
	}

	children := root.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	h1 := children[0]
	if h1.Title() != "Hello World" {
		t.Errorf("expected title 'Hello World', got %q", h1.Title())
	}
	if h1.Level() != 1 {
		t.Errorf("expected level 1, got %d", h1.Level())
	}

	content := strings.TrimSpace(h1.Content())
	if content != "This is a simple document." {
		t.Errorf("expected content 'This is a simple document.', got %q", content)
	}
}

func TestParserNestedHeadings(t *testing.T) {
	markdown := `# Chapter 1

Chapter 1 intro.

## Section 1.1

Section 1.1 content.

## Section 1.2

Section 1.2 content.

# Chapter 2

Chapter 2 intro.`

	ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Document"})
	root, _, err := DefaultParser(ctx, []byte(markdown))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	children := root.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(children))
	}

	ch1 := children[0]
	if ch1.Title() != "Chapter 1" {
		t.Errorf("expected 'Chapter 1', got %q", ch1.Title())
	}
	if ch1.Level() != 1 {
		t.Errorf("expected level 1, got %d", ch1.Level())
	}

	ch1Content := strings.TrimSpace(ch1.Content())
	if !strings.Contains(ch1Content, "Chapter 1 intro.") {
		t.Errorf("expected content to contain 'Chapter 1 intro.', got %q", ch1Content)
	}

	ch1Children := ch1.Children()
	if len(ch1Children) != 2 {
		t.Fatalf("expected 2 sections in chapter 1, got %d", len(ch1Children))
	}

	sec1_1 := ch1Children[0]
	if sec1_1.Title() != "Section 1.1" {
		t.Errorf("expected 'Section 1.1', got %q", sec1_1.Title())
	}
	if sec1_1.Level() != 2 {
		t.Errorf("expected level 2, got %d", sec1_1.Level())
	}

	sec1_2 := ch1Children[1]
	if sec1_2.Title() != "Section 1.2" {
		t.Errorf("expected 'Section 1.2', got %q", sec1_2.Title())
	}

	ch2 := children[1]
	if ch2.Title() != "Chapter 2" {
		t.Errorf("expected 'Chapter 2', got %q", ch2.Title())
	}
	if len(ch2.Children()) != 0 {
		t.Errorf("expected 0 children in chapter 2, got %d", len(ch2.Children()))
	}
}

func TestParserWithFrontmatter(t *testing.T) {
	markdown := `---
title: My Document
author: John Doe
tags:
  - markdown
  - test
---

# Introduction

This is the introduction.`

	ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Document"})
	root, fm, err := DefaultParser(ctx, []byte(markdown))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if fm == nil {
		t.Fatal("frontmatter is nil")
	}

	if title, ok := fm["title"].(string); !ok || title != "My Document" {
		t.Errorf("expected frontmatter title 'My Document', got %v", fm["title"])
	}

	if author, ok := fm["author"].(string); !ok || author != "John Doe" {
		t.Errorf("expected frontmatter author 'John Doe', got %v", fm["author"])
	}

	children := root.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	if children[0].Title() != "Introduction" {
		t.Errorf("expected title 'Introduction', got %q", children[0].Title())
	}
}

func TestParserEmptyDocument(t *testing.T) {
	ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Empty"})
	root, fm, err := DefaultParser(ctx, []byte(""))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if root == nil {
		t.Fatal("root is nil")
	}
	if len(root.Children()) != 0 {
		t.Errorf("expected 0 children, got %d", len(root.Children()))
	}
	if fm == nil {
		t.Error("frontmatter should not be nil")
	}
}

func TestParserNoHeadings(t *testing.T) {
	markdown := `This is a document with no headings.

Just some paragraphs of text.

And another one.`

	ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Document"})
	root, _, err := DefaultParser(ctx, []byte(markdown))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(root.Children()) != 0 {
		t.Errorf("expected 0 children, got %d", len(root.Children()))
	}

	content := strings.TrimSpace(root.Content())
	if !strings.Contains(content, "This is a document with no headings.") {
		t.Errorf("root content should contain document text, got %q", content)
	}
}

func TestParserDeepNesting(t *testing.T) {
	markdown := `# H1

H1 content

## H2

H2 content

### H3

H3 content

#### H4

H4 content

##### H5

H5 content

###### H6

H6 content`

	ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Document"})
	root, _, err := DefaultParser(ctx, []byte(markdown))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Navigate down the tree
	h1 := root.Children()[0]
	if h1.Level() != 1 || h1.Title() != "H1" {
		t.Errorf("expected H1 at level 1, got %d/%q", h1.Level(), h1.Title())
	}

	h2 := h1.Children()[0]
	if h2.Level() != 2 || h2.Title() != "H2" {
		t.Errorf("expected H2 at level 2, got %d/%q", h2.Level(), h2.Title())
	}

	h3 := h2.Children()[0]
	if h3.Level() != 3 || h3.Title() != "H3" {
		t.Errorf("expected H3 at level 3, got %d/%q", h3.Level(), h3.Title())
	}

	h4 := h3.Children()[0]
	if h4.Level() != 4 || h4.Title() != "H4" {
		t.Errorf("expected H4 at level 4, got %d/%q", h4.Level(), h4.Title())
	}

	h5 := h4.Children()[0]
	if h5.Level() != 5 || h5.Title() != "H5" {
		t.Errorf("expected H5 at level 5, got %d/%q", h5.Level(), h5.Title())
	}

	h6 := h5.Children()[0]
	if h6.Level() != 6 || h6.Title() != "H6" {
		t.Errorf("expected H6 at level 6, got %d/%q", h6.Level(), h6.Title())
	}
}

func TestParserSkippedLevels(t *testing.T) {
	// Test that skipping heading levels (H1 -> H3) works
	markdown := `# H1

H1 content

### H3

H3 content

## H2

H2 content`

	ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Document"})
	root, _, err := DefaultParser(ctx, []byte(markdown))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	h1 := root.Children()[0]
	if h1.Title() != "H1" {
		t.Errorf("expected H1, got %q", h1.Title())
	}

	// H3 should be a child of H1 even though we skipped H2
	h1Children := h1.Children()
	if len(h1Children) != 2 {
		t.Fatalf("expected 2 children of H1, got %d", len(h1Children))
	}

	h3 := h1Children[0]
	if h3.Title() != "H3" || h3.Level() != 3 {
		t.Errorf("expected H3 at level 3, got %q at level %d", h3.Title(), h3.Level())
	}

	h2 := h1Children[1]
	if h2.Title() != "H2" || h2.Level() != 2 {
		t.Errorf("expected H2 at level 2, got %q at level %d", h2.Title(), h2.Level())
	}
}

func TestParserInlineFormatting(t *testing.T) {
	markdown := `# **Bold** and *italic* and ` + "`code`" + `

Some content here.`

	ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Document"})
	root, _, err := DefaultParser(ctx, []byte(markdown))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	h1 := root.Children()[0]
	// Title should have inline formatting stripped to plain text
	if h1.Title() != "Bold and italic and code" {
		t.Errorf("expected plain text title, got %q", h1.Title())
	}
}

func TestParserReset(t *testing.T) {
	markdown1 := `# Document 1

Content 1`

	markdown2 := `# Document 2

Content 2`

	// Parse first document
	ctx1 := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Doc1"})
	root1, _, err := DefaultParser(ctx1, []byte(markdown1))
	if err != nil {
		t.Fatalf("Parse 1 failed: %v", err)
	}
	if root1.Children()[0].Title() != "Document 1" {
		t.Errorf("expected 'Document 1', got %q", root1.Children()[0].Title())
	}

	// Parse second document (each call creates a new worker instance)
	ctx2 := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Doc2"})
	root2, _, err := DefaultParser(ctx2, []byte(markdown2))
	if err != nil {
		t.Fatalf("Parse 2 failed: %v", err)
	}
	if root2.Children()[0].Title() != "Document 2" {
		t.Errorf("expected 'Document 2', got %q", root2.Children()[0].Title())
	}

	// Ensure they're different roots
	if root1 == root2 {
		t.Error("root1 and root2 should be different instances")
	}
}

func TestParserMultipleH1s(t *testing.T) {
	markdown := `# First H1

First content

# Second H1

Second content

# Third H1

Third content`

	ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Document"})
	root, _, err := DefaultParser(ctx, []byte(markdown))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	children := root.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 H1 sections, got %d", len(children))
	}

	titles := []string{"First H1", "Second H1", "Third H1"}
	for i, expected := range titles {
		if children[i].Title() != expected {
			t.Errorf("child %d: expected %q, got %q", i, expected, children[i].Title())
		}
		if children[i].Level() != 1 {
			t.Errorf("child %d: expected level 1, got %d", i, children[i].Level())
		}
	}
}

func TestParserContentBetweenSections(t *testing.T) {
	markdown := `Some preamble text before any headings.

# Heading 1

Text after heading 1.

# Heading 2

Text after heading 2.`

	ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Document"})
	root, _, err := DefaultParser(ctx, []byte(markdown))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Preamble should go into root content
	rootContent := strings.TrimSpace(root.Content())
	if !strings.Contains(rootContent, "Some preamble text") {
		t.Errorf("root content should contain preamble, got %q", rootContent)
	}

	children := root.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	h1Content := strings.TrimSpace(children[0].Content())
	if !strings.Contains(h1Content, "Text after heading 1") {
		t.Errorf("h1 content should contain its text, got %q", h1Content)
	}

	h2Content := strings.TrimSpace(children[1].Content())
	if !strings.Contains(h2Content, "Text after heading 2") {
		t.Errorf("h2 content should contain its text, got %q", h2Content)
	}
}

// --- Tests for internal helper functions ---

func TestSpliceText(t *testing.T) {
	src := []byte("Hello World")

	tests := []struct {
		name     string
		start    int
		stop     int
		expected string
		nextPos  int
	}{
		{"normal range", 0, 5, "Hello", 5},
		{"full range", 0, 11, "Hello World", 11},
		{"negative start", -5, 5, "Hello", 5},
		{"stop beyond length", 6, 100, "World", 11},
		{"empty range", 5, 5, "", 5},
		{"inverted range", 10, 5, "", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, next := spliceText(src, tt.start, tt.stop)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
			if next != tt.nextPos {
				t.Errorf("expected next position %d, got %d", tt.nextPos, next)
			}
		})
	}
}

func TestParentForLevel(t *testing.T) {
	root := section.NewRoot("Root")
	h1 := root.CreateChild("H1", 1, "")
	h2 := h1.CreateChild("H2", 2, "")
	h3 := h2.CreateChild("H3", 3, "")

	stack := []sectionFrame{
		{s: root},
		{s: h1},
		{s: h2},
		{s: h3},
	}

	tests := []struct {
		name        string
		targetLevel int
		expectedIdx int
		shouldError bool
	}{
		{"find parent for h1", 1, 0, false},
		{"find parent for h2", 2, 1, false},
		{"find parent for h3", 3, 2, false},
		{"find parent for h4", 4, 3, false},
		{"invalid nesting", 0, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, err := parentForLevel(stack, tt.targetLevel)
			if tt.shouldError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if idx != tt.expectedIdx {
					t.Errorf("expected index %d, got %d", tt.expectedIdx, idx)
				}
			}
		})
	}
}

func TestInlineTextExtraction(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			"plain text",
			"# Hello World",
			"Hello World",
		},
		{
			"bold text",
			"# **Bold Text**",
			"Bold Text",
		},
		{
			"italic text",
			"# *Italic Text*",
			"Italic Text",
		},
		{
			"code span",
			"# `code here`",
			"code here",
		},
		{
			"mixed formatting",
			"# **Bold** and *italic* and `code`",
			"Bold and italic and code",
		},
		{
			"link text",
			"# [Link Text](http://example.com)",
			"Link Text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "Test"})
			root, _, err := DefaultParser(ctx, []byte(tt.markdown))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			children := root.Children()
			if len(children) == 0 {
				t.Fatal("expected at least one heading")
			}

			title := children[0].Title()
			if title != tt.expected {
				t.Errorf("expected title %q, got %q", tt.expected, title)
			}
		})
	}
}

func TestParserStackTruncationOnLevelDecrease(t *testing.T) {
	markdown := `---
title: Doc
tags: [t]
---

Preamble before any heading.

# H1
Text under H1.

### H3 under H1
H3 body.

## H2 after H3
H2 body.

# Second H1
Second body.
Trailing line after second H1.
`

	ctx := chunkyctx.WithFileInfo(context.Background(), chunkyctx.FileInfo{Title: "TestDoc"})
	root, fm, err := DefaultParser(ctx, []byte(markdown))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if root == nil {
		t.Fatal("root is nil")
	}
	if fm == nil {
		t.Fatal("frontmatter is nil (should be parsed)")
	}
	if title, ok := fm["title"].(string); !ok || title != "Doc" {
		t.Fatalf("frontmatter title = %v, want 'Doc'", fm["title"])
	}

	// Root content should include preamble
	if rc := strings.TrimSpace(root.Content()); !strings.Contains(rc, "Preamble before any heading.") {
		t.Fatalf("root content missing preamble, got: %q", rc)
	}

	children := root.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 top-level sections (H1, Second H1), got %d", len(children))
	}

	h1 := children[0]
	if h1.Title() != "H1" || h1.Level() != 1 {
		t.Fatalf("first child should be H1 (L1), got %q (L%d)", h1.Title(), h1.Level())
	}
	if !strings.Contains(h1.Content(), "Text under H1.") {
		t.Fatalf("H1 content missing, got: %q", h1.Content())
	}

	h1Kids := h1.Children()
	if len(h1Kids) != 2 {
		t.Fatalf("expected H1 to have 2 children (H3, H2), got %d", len(h1Kids))
	}

	// Order should be H3 then H2
	h3 := h1Kids[0]
	if h3.Title() != "H3 under H1" || h3.Level() != 3 {
		t.Fatalf("expected first child of H1 to be 'H3 under H1' (L3), got %q (L%d)", h3.Title(), h3.Level())
	}
	if !strings.Contains(h3.Content(), "H3 body.") {
		t.Fatalf("H3 content missing, got: %q", h3.Content())
	}

	h2 := h1Kids[1]
	if h2.Title() != "H2 after H3" || h2.Level() != 2 {
		t.Fatalf("expected second child of H1 to be 'H2 after H3' (L2), got %q (L%d)", h2.Title(), h2.Level())
	}
	if !strings.Contains(h2.Content(), "H2 body.") {
		t.Fatalf("H2 content missing, got: %q", h2.Content())
	}

	// Ensure H2 did NOT accidentally nest under H3 (old bug)
	if len(h3.Children()) != 0 {
		t.Fatalf("H3 should not have H2 as a child; got %d children", len(h3.Children()))
	}

	second := children[1]
	if second.Title() != "Second H1" || second.Level() != 1 {
		t.Fatalf("second top-level section should be 'Second H1' (L1), got %q (L%d)", second.Title(), second.Level())
	}
	if sc := second.Content(); !strings.Contains(sc, "Second body.") || !strings.Contains(sc, "Trailing line after second H1.") {
		t.Fatalf("'Second H1' content missing trailing lines, got: %q", sc)
	}
}
