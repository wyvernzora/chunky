package section

import (
	"testing"
)

func TestNewRoot(t *testing.T) {
	root := NewRoot("Test Document")

	if root.Title() != "Test Document" {
		t.Errorf("expected title 'Test Document', got %q", root.Title())
	}
	if root.Level() != 0 {
		t.Errorf("expected level 0, got %d", root.Level())
	}
	if !root.IsRoot() {
		t.Error("expected IsRoot() to be true")
	}
	if root.Parent() != nil {
		t.Error("expected Parent() to be nil for root")
	}
	if root.Content() != "" {
		t.Errorf("expected empty content, got %q", root.Content())
	}
	if len(root.Children()) != 0 {
		t.Errorf("expected 0 children, got %d", len(root.Children()))
	}
}

func TestSectionGetters(t *testing.T) {
	root := NewRoot("Root")
	child := root.CreateChild("Child", 1, "Initial content")

	if child.Title() != "Child" {
		t.Errorf("expected title 'Child', got %q", child.Title())
	}
	if child.Level() != 1 {
		t.Errorf("expected level 1, got %d", child.Level())
	}
	if child.Content() != "Initial content" {
		t.Errorf("expected content 'Initial content', got %q", child.Content())
	}
	if child.Parent() != root {
		t.Error("expected parent to be root section")
	}
	if child.IsRoot() {
		t.Error("expected IsRoot() to be false for child")
	}
}

func TestAppendContent(t *testing.T) {
	s := NewRoot("Test")

	s.AppendContent("First line\n")
	if s.Content() != "First line\n" {
		t.Errorf("expected 'First line\\n', got %q", s.Content())
	}

	s.AppendContent("Second line\n")
	if s.Content() != "First line\nSecond line\n" {
		t.Errorf("expected concatenated content, got %q", s.Content())
	}

	// Empty content should not change anything
	s.AppendContent("")
	if s.Content() != "First line\nSecond line\n" {
		t.Errorf("empty append should not change content, got %q", s.Content())
	}
}

func TestCreateChild(t *testing.T) {
	root := NewRoot("Root")

	child1 := root.CreateChild("Chapter 1", 1, "")
	child2 := root.CreateChild("Chapter 2", 1, "Some content")

	children := root.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	if children[0] != child1 {
		t.Error("first child should be child1")
	}
	if children[1] != child2 {
		t.Error("second child should be child2")
	}

	if child1.Parent() != root {
		t.Error("child1 parent should be root")
	}
	if child2.Parent() != root {
		t.Error("child2 parent should be root")
	}

	if child1.Content() != "" {
		t.Errorf("child1 should have empty content, got %q", child1.Content())
	}
	if child2.Content() != "Some content" {
		t.Errorf("child2 should have 'Some content', got %q", child2.Content())
	}
}

func TestChildrenIsolation(t *testing.T) {
	root := NewRoot("Root")
	root.CreateChild("Child 1", 1, "")
	root.CreateChild("Child 2", 1, "")

	// Get children and modify the slice
	children := root.Children()
	children[0] = nil
	children = append(children, NewRoot("Fake"))

	// Original children should be unchanged
	actualChildren := root.Children()
	if len(actualChildren) != 2 {
		t.Errorf("expected 2 children after modification, got %d", len(actualChildren))
	}
	if actualChildren[0] == nil {
		t.Error("first child should not be nil")
	}
	if actualChildren[0].Title() != "Child 1" {
		t.Errorf("first child title should be 'Child 1', got %q", actualChildren[0].Title())
	}
}

func TestNestedSections(t *testing.T) {
	root := NewRoot("Document")
	ch1 := root.CreateChild("Chapter 1", 1, "Ch1 intro")
	sec1_1 := ch1.CreateChild("Section 1.1", 2, "Sec 1.1 content")
	sec1_2 := ch1.CreateChild("Section 1.2", 2, "Sec 1.2 content")
	ch2 := root.CreateChild("Chapter 2", 1, "Ch2 intro")

	// Verify root structure
	rootChildren := root.Children()
	if len(rootChildren) != 2 {
		t.Fatalf("root should have 2 children, got %d", len(rootChildren))
	}
	if rootChildren[0] != ch1 || rootChildren[1] != ch2 {
		t.Error("root children order incorrect")
	}

	// Verify ch1 structure
	ch1Children := ch1.Children()
	if len(ch1Children) != 2 {
		t.Fatalf("ch1 should have 2 children, got %d", len(ch1Children))
	}
	if ch1Children[0] != sec1_1 || ch1Children[1] != sec1_2 {
		t.Error("ch1 children order incorrect")
	}

	// Verify ch2 has no children
	if len(ch2.Children()) != 0 {
		t.Errorf("ch2 should have no children, got %d", len(ch2.Children()))
	}

	// Verify parent relationships
	if sec1_1.Parent() != ch1 {
		t.Error("sec1_1 parent should be ch1")
	}
	if sec1_2.Parent() != ch1 {
		t.Error("sec1_2 parent should be ch1")
	}
	if ch1.Parent() != root {
		t.Error("ch1 parent should be root")
	}
	if ch2.Parent() != root {
		t.Error("ch2 parent should be root")
	}

	// Verify levels
	if sec1_1.Level() != 2 {
		t.Errorf("sec1_1 level should be 2, got %d", sec1_1.Level())
	}
	if ch1.Level() != 1 {
		t.Errorf("ch1 level should be 1, got %d", ch1.Level())
	}
}

func TestSectionContent(t *testing.T) {
	s := NewRoot("Test")

	// Multiple appends
	s.AppendContent("Line 1\n")
	s.AppendContent("Line 2\n")
	s.AppendContent("Line 3")

	expected := "Line 1\nLine 2\nLine 3"
	if s.Content() != expected {
		t.Errorf("expected %q, got %q", expected, s.Content())
	}
}

func TestSectionLevels(t *testing.T) {
	root := NewRoot("Root")
	h1 := root.CreateChild("H1", 1, "")
	h2 := h1.CreateChild("H2", 2, "")
	h3 := h2.CreateChild("H3", 3, "")
	h4 := h3.CreateChild("H4", 4, "")
	h5 := h4.CreateChild("H5", 5, "")
	h6 := h5.CreateChild("H6", 6, "")

	levels := []struct {
		section *Section
		level   int
	}{
		{root, 0},
		{h1, 1},
		{h2, 2},
		{h3, 3},
		{h4, 4},
		{h5, 5},
		{h6, 6},
	}

	for _, tc := range levels {
		if tc.section.Level() != tc.level {
			t.Errorf("section %q should have level %d, got %d",
				tc.section.Title(), tc.level, tc.section.Level())
		}
	}
}

func TestSetContent(t *testing.T) {
	s := NewRoot("Test")

	// Set initial content
	s.SetContent("Initial content")
	if s.Content() != "Initial content" {
		t.Errorf("expected 'Initial content', got %q", s.Content())
	}

	// Replace content
	s.SetContent("Replaced content")
	if s.Content() != "Replaced content" {
		t.Errorf("expected 'Replaced content', got %q", s.Content())
	}

	// Set to empty string
	s.SetContent("")
	if s.Content() != "" {
		t.Errorf("expected empty content, got %q", s.Content())
	}
}

func TestResetContent(t *testing.T) {
	s := NewRoot("Test")

	// Add some content
	s.AppendContent("Some content\n")
	s.AppendContent("More content\n")
	if s.Content() == "" {
		t.Fatal("expected content before reset")
	}

	// Reset should clear all content
	s.ResetContent()
	if s.Content() != "" {
		t.Errorf("expected empty content after reset, got %q", s.Content())
	}

	// Reset on already empty content should be safe
	s.ResetContent()
	if s.Content() != "" {
		t.Errorf("expected empty content after second reset, got %q", s.Content())
	}
}

func TestPrependContent(t *testing.T) {
	s := NewRoot("Test")

	// Prepend to empty content
	s.PrependContent("First\n")
	if s.Content() != "First\n" {
		t.Errorf("expected 'First\\n', got %q", s.Content())
	}

	// Prepend to existing content
	s.PrependContent("Prepended\n")
	if s.Content() != "Prepended\nFirst\n" {
		t.Errorf("expected 'Prepended\\nFirst\\n', got %q", s.Content())
	}

	// Prepend empty string
	s.PrependContent("")
	if s.Content() != "Prepended\nFirst\n" {
		t.Errorf("empty prepend should not change content, got %q", s.Content())
	}
}

func TestContentManipulation(t *testing.T) {
	s := NewRoot("Test")

	// Mix of append, prepend, set, and reset
	s.AppendContent("Middle")
	s.PrependContent("Start ")
	s.AppendContent(" End")

	expected := "Start Middle End"
	if s.Content() != expected {
		t.Errorf("expected %q, got %q", expected, s.Content())
	}

	// SetContent should replace everything
	s.SetContent("Completely New")
	if s.Content() != "Completely New" {
		t.Errorf("expected 'Completely New', got %q", s.Content())
	}

	// ResetContent should clear
	s.ResetContent()
	if s.Content() != "" {
		t.Errorf("expected empty content, got %q", s.Content())
	}

	// Should be able to continue using after reset
	s.AppendContent("After reset")
	if s.Content() != "After reset" {
		t.Errorf("expected 'After reset', got %q", s.Content())
	}
}
