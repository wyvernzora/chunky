package tokenizer

import (
	"testing"

	"github.com/wyvernzora/chunky/pkg/section"
)

func TestNewTokenizedSection(t *testing.T) {
	sec := section.NewRoot("Test")
	children := []*TokenizedSection{}

	ts := NewTokenizedSection(sec, 10, 10, children)

	if ts.GetSection() != sec {
		t.Error("section not set correctly")
	}
	if ts.GetContentTokens() != 10 {
		t.Errorf("expected ContentTokens 10, got %d", ts.GetContentTokens())
	}
	if ts.GetSubtreeTokens() != 10 {
		t.Errorf("expected SubtreeTokens 10, got %d", ts.GetSubtreeTokens())
	}
	if len(ts.GetChildren()) != 0 {
		t.Errorf("expected 0 children, got %d", len(ts.GetChildren()))
	}
}

func TestNewTokenizedSection_WithChildren(t *testing.T) {
	root := section.NewRoot("root")
	childSec := root.CreateChild("child", 1, "")

	child := NewTokenizedSection(childSec, 5, 5, nil)
	parent := NewTokenizedSection(root, 10, 15, []*TokenizedSection{child})

	if parent.GetContentTokens() != 10 {
		t.Errorf("parent ContentTokens: expected 10, got %d", parent.GetContentTokens())
	}
	if parent.GetSubtreeTokens() != 15 {
		t.Errorf("parent SubtreeTokens: expected 15, got %d", parent.GetSubtreeTokens())
	}
	if len(parent.GetChildren()) != 1 {
		t.Errorf("expected 1 child, got %d", len(parent.GetChildren()))
	}
	if parent.GetChildren()[0] != child {
		t.Error("child reference not preserved")
	}
}

func TestTokenizedSection_Render_Empty(t *testing.T) {
	sec := section.NewRoot("Test")
	ts := NewTokenizedSection(sec, 0, 0, nil)

	rendered := ts.Render()
	if rendered != "" {
		t.Errorf("expected empty string, got %q", rendered)
	}
}

func TestTokenizedSection_Render_Nil(t *testing.T) {
	var ts *TokenizedSection
	rendered := ts.Render()
	if rendered != "" {
		t.Errorf("expected empty string for nil, got %q", rendered)
	}
}

func TestTokenizedSection_Render_SingleNode(t *testing.T) {
	sec := section.NewRoot("Test")
	sec.SetContent("Hello, world!")

	ts := NewTokenizedSection(sec, 13, 13, nil)

	rendered := ts.Render()
	if rendered != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got %q", rendered)
	}
}

func TestTokenizedSection_Render_WithChildren(t *testing.T) {
	root := section.NewRoot("root")
	root.SetContent("Root content\n")

	child1 := root.CreateChild("child1", 1, "Child 1 content\n")
	child2 := root.CreateChild("child2", 1, "Child 2 content\n")

	// Build tokenized tree
	tc1 := NewTokenizedSection(child1, 16, 16, nil)
	tc2 := NewTokenizedSection(child2, 16, 16, nil)
	troot := NewTokenizedSection(root, 13, 45, []*TokenizedSection{tc1, tc2})

	rendered := troot.Render()
	expected := "Root content\nChild 1 content\nChild 2 content\n"

	if rendered != expected {
		t.Errorf("expected %q, got %q", expected, rendered)
	}
}

func TestTokenizedSection_Render_NestedChildren(t *testing.T) {
	root := section.NewRoot("root")
	root.SetContent("A")

	child := root.CreateChild("child", 1, "B")
	grandchild := child.CreateChild("grandchild", 2, "C")

	// Build tokenized tree
	tgrand := NewTokenizedSection(grandchild, 1, 1, nil)
	tchild := NewTokenizedSection(child, 1, 2, []*TokenizedSection{tgrand})
	troot := NewTokenizedSection(root, 1, 3, []*TokenizedSection{tchild})

	rendered := troot.Render()
	expected := "ABC"

	if rendered != expected {
		t.Errorf("expected %q, got %q", expected, rendered)
	}
}

func TestTokenizedSection_Render_ComplexTree(t *testing.T) {
	// Build tree:
	//   root ("R")
	//   ├── A ("A")
	//   │   ├── A1 ("A1")
	//   │   └── A2 ("A2")
	//   └── B ("B")
	root := section.NewRoot("root")
	root.SetContent("R")

	a := root.CreateChild("A", 1, "A")
	a.CreateChild("A1", 2, "A1")
	a.CreateChild("A2", 2, "A2")

	root.CreateChild("B", 1, "B")

	// Build tokenized tree (need to build from leaves up)
	ta1 := NewTokenizedSection(a.Children()[0], 2, 2, nil)
	ta2 := NewTokenizedSection(a.Children()[1], 2, 2, nil)
	ta := NewTokenizedSection(a, 1, 5, []*TokenizedSection{ta1, ta2})

	b := root.Children()[1]
	tb := NewTokenizedSection(b, 1, 1, nil)

	troot := NewTokenizedSection(root, 1, 7, []*TokenizedSection{ta, tb})

	rendered := troot.Render()
	expected := "RAA1A2B"

	if rendered != expected {
		t.Errorf("expected %q, got %q", expected, rendered)
	}
}

func TestTokenizedSection_Render_WithNilChild(t *testing.T) {
	root := section.NewRoot("root")
	root.SetContent("Root")

	// Create tokenized section with nil child in slice
	troot := NewTokenizedSection(root, 4, 4, []*TokenizedSection{nil})

	rendered := troot.Render()
	if rendered != "Root" {
		t.Errorf("expected 'Root', got %q", rendered)
	}
}

func TestTokenizedSection_Render_MultipleNilChildren(t *testing.T) {
	root := section.NewRoot("root")
	root.SetContent("Test")

	child1Sec := root.CreateChild("child1", 1, "A")
	child2Sec := root.CreateChild("child2", 1, "B")

	tc1 := NewTokenizedSection(child1Sec, 1, 1, nil)
	tc2 := NewTokenizedSection(child2Sec, 1, 1, nil)

	// Mix nil and non-nil children
	troot := NewTokenizedSection(root, 4, 6, []*TokenizedSection{nil, tc1, nil, tc2, nil})

	rendered := troot.Render()
	expected := "TestAB"

	if rendered != expected {
		t.Errorf("expected %q, got %q", expected, rendered)
	}
}

func TestTokenizedSection_Render_PreservesWhitespace(t *testing.T) {
	root := section.NewRoot("root")
	root.SetContent("Line 1\n\nLine 2\n")

	ts := NewTokenizedSection(root, 15, 15, nil)

	rendered := ts.Render()
	expected := "Line 1\n\nLine 2\n"

	if rendered != expected {
		t.Errorf("whitespace not preserved:\nexpected: %q\ngot:      %q", expected, rendered)
	}
}

func TestTokenizedSection_Render_EmptyChildren(t *testing.T) {
	root := section.NewRoot("root")
	root.SetContent("Root")

	child := root.CreateChild("child", 1, "") // Empty content

	tchild := NewTokenizedSection(child, 0, 0, nil)
	troot := NewTokenizedSection(root, 4, 4, []*TokenizedSection{tchild})

	rendered := troot.Render()
	expected := "Root"

	if rendered != expected {
		t.Errorf("expected %q, got %q", expected, rendered)
	}
}

func TestTokenizedSection_Render_UnicodeContent(t *testing.T) {
	root := section.NewRoot("root")
	root.SetContent("Hello 世界 🌍")

	ts := NewTokenizedSection(root, 10, 10, nil)

	rendered := ts.Render()
	expected := "Hello 世界 🌍"

	if rendered != expected {
		t.Errorf("unicode not preserved:\nexpected: %q\ngot:      %q", expected, rendered)
	}
}

func TestTokenizedSection_SubtreeTokensInvariant(t *testing.T) {
	// Test that SubtreeTokens = ContentTokens + sum(children.GetSubtreeTokens())
	root := section.NewRoot("root")
	child1 := root.CreateChild("child1", 1, "")
	child2 := root.CreateChild("child2", 1, "")

	tc1 := NewTokenizedSection(child1, 10, 10, nil)
	tc2 := NewTokenizedSection(child2, 15, 15, nil)
	troot := NewTokenizedSection(root, 5, 30, []*TokenizedSection{tc1, tc2})

	// Verify invariant
	expectedSubtree := troot.GetContentTokens() + tc1.GetSubtreeTokens() + tc2.GetSubtreeTokens()
	if troot.GetSubtreeTokens() != expectedSubtree {
		t.Errorf("subtree invariant violated: expected %d, got %d", expectedSubtree, troot.GetSubtreeTokens())
	}
}
