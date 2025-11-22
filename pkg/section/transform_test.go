package section

import (
	"context"
	"errors"
	"strings"
	"testing"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
)

func TestApplyTransform_Empty(t *testing.T) {
	root := NewRoot("Test")
	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	// No transforms should succeed
	err := ApplyTransform(ctx, frontmatter, root)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestApplyTransform_SingleTransform(t *testing.T) {
	root := NewRoot("Test")
	root.SetContent("original content")
	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	appendSuffix := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		s.SetContent(s.Content() + " [modified]")
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, appendSuffix)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if root.Content() != "original content [modified]" {
		t.Errorf("expected modified content, got %q", root.Content())
	}
}

func TestApplyTransform_MultipleTransforms(t *testing.T) {
	root := NewRoot("Test")
	root.SetContent("hello")
	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	uppercase := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		s.SetContent(strings.ToUpper(s.Content()))
		return nil
	}

	addPrefix := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		s.SetContent("PREFIX: " + s.Content())
		return nil
	}

	addSuffix := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		s.SetContent(s.Content() + " :SUFFIX")
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, uppercase, addPrefix, addSuffix)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "PREFIX: HELLO :SUFFIX"
	if root.Content() != expected {
		t.Errorf("expected %q, got %q", expected, root.Content())
	}
}

func TestApplyTransform_DepthFirstWalk(t *testing.T) {
	// Build tree:
	//   root
	//   ├── child1
	//   │   └── grandchild1
	//   └── child2
	root := NewRoot("root")
	child1 := root.CreateChild("child1", 1, "")
	child1.CreateChild("grandchild1", 2, "")
	root.CreateChild("child2", 1, "")

	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	var visitOrder []string

	recordVisit := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		visitOrder = append(visitOrder, s.Title())
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, recordVisit)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Expect depth-first pre-order traversal
	expected := []string{"root", "child1", "grandchild1", "child2"}
	if len(visitOrder) != len(expected) {
		t.Fatalf("expected %d visits, got %d", len(expected), len(visitOrder))
	}

	for i, title := range expected {
		if visitOrder[i] != title {
			t.Errorf("visit %d: expected %q, got %q", i, title, visitOrder[i])
		}
	}
}

func TestApplyTransform_AllNodesModified(t *testing.T) {
	// Build tree: root -> child1 -> grandchild
	root := NewRoot("root")
	root.SetContent("A")
	child1 := root.CreateChild("child1", 1, "B")
	grandchild := child1.CreateChild("grandchild", 2, "C")

	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	appendX := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		s.SetContent(s.Content() + "X")
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, appendX)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if root.Content() != "AX" {
		t.Errorf("root: expected 'AX', got %q", root.Content())
	}
	if child1.Content() != "BX" {
		t.Errorf("child1: expected 'BX', got %q", child1.Content())
	}
	if grandchild.Content() != "CX" {
		t.Errorf("grandchild: expected 'CX', got %q", grandchild.Content())
	}
}

func TestApplyTransform_ErrorHandling(t *testing.T) {
	root := NewRoot("root")
	root.CreateChild("child1", 1, "")
	root.CreateChild("child2", 1, "")

	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	expectedErr := errors.New("transform failed")
	var visitedSections []string

	recordVisit := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		visitedSections = append(visitedSections, s.Title())
		return nil
	}

	failOnChild1 := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		if s.Title() == "child1" {
			return expectedErr
		}
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, recordVisit, failOnChild1)

	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	// Should have visited root and child1 before failing
	if len(visitedSections) != 2 {
		t.Errorf("expected 2 visits before error, got %d", len(visitedSections))
	}

	// child2 should not have been visited
	for _, title := range visitedSections {
		if title == "child2" {
			t.Error("child2 should not have been visited after error")
		}
	}
}

func TestApplyTransform_ErrorStopsWalk(t *testing.T) {
	// Build tree: root -> child -> grandchild
	root := NewRoot("root")
	child := root.CreateChild("child", 1, "")
	child.CreateChild("grandchild", 2, "")

	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	var visitCount int
	expectedErr := errors.New("stop")

	countAndFail := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		visitCount++
		if s.Title() == "child" {
			return expectedErr
		}
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, countAndFail)

	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	// Should have visited root and child only (grandchild not reached)
	if visitCount != 2 {
		t.Errorf("expected 2 visits, got %d", visitCount)
	}
}

func TestApplyTransform_OrderMatters(t *testing.T) {
	root := NewRoot("Test")
	root.SetContent("test")
	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	uppercase := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		s.SetContent(strings.ToUpper(s.Content()))
		return nil
	}

	addBrackets := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		s.SetContent("[" + s.Content() + "]")
		return nil
	}

	// Apply uppercase then brackets
	root1 := NewRoot("Test")
	root1.SetContent("test")
	err := ApplyTransform(ctx, frontmatter, root1, uppercase, addBrackets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root1.Content() != "[TEST]" {
		t.Errorf("expected '[TEST]', got %q", root1.Content())
	}

	// Apply brackets then uppercase
	root2 := NewRoot("Test")
	root2.SetContent("test")
	err = ApplyTransform(ctx, frontmatter, root2, addBrackets, uppercase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root2.Content() != "[TEST]" {
		t.Errorf("expected '[TEST]', got %q", root2.Content())
	}
}

func TestApplyTransform_FrontMatterView(t *testing.T) {
	root := NewRoot("Test")
	frontmatter := fm.FrontMatter{
		"title":  "My Document",
		"author": "Jane Doe",
	}
	ctx := context.Background()

	useFrontMatter := func(ctx context.Context, fmView fm.FrontMatterView, s *Section) error {
		if title, ok := fmView.Get("title"); ok {
			s.SetContent("Title: " + title.(string))
		}
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, useFrontMatter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if root.Content() != "Title: My Document" {
		t.Errorf("expected 'Title: My Document', got %q", root.Content())
	}
}

func TestApplyTransform_FrontMatterReadOnly(t *testing.T) {
	root := NewRoot("Test")
	frontmatter := fm.FrontMatter{
		"title": "Original",
	}
	ctx := context.Background()

	// This transform attempts to modify frontmatter through the view
	// The view should provide read-only access (deep copies)
	attemptModify := func(ctx context.Context, fmView fm.FrontMatterView, s *Section) error {
		// Get returns a deep copy, so modifying it shouldn't affect the original
		if title, ok := fmView.Get("title"); ok {
			if titleStr, ok := title.(string); ok {
				// This modifies the copy, not the original
				_ = titleStr + " Modified"
			}
		}
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, attemptModify)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Original frontmatter should be unchanged
	if frontmatter["title"] != "Original" {
		t.Errorf("frontmatter was modified: got %q", frontmatter["title"])
	}
}

func TestApplyTransform_ContextPropagation(t *testing.T) {
	root := NewRoot("root")
	root.CreateChild("child", 1, "")

	frontmatter := fm.EmptyFrontMatter()

	type contextKey string
	const userKey contextKey = "user"
	ctx := context.WithValue(context.Background(), userKey, "testuser")

	var seenValues []string

	checkContext := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		if user, ok := ctx.Value(userKey).(string); ok {
			seenValues = append(seenValues, user)
			return nil
		}
		return errors.New("context value not found")
	}

	err := ApplyTransform(ctx, frontmatter, root, checkContext)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Both root and child should have seen the context value
	if len(seenValues) != 2 {
		t.Errorf("expected 2 context checks, got %d", len(seenValues))
	}

	for _, val := range seenValues {
		if val != "testuser" {
			t.Errorf("expected 'testuser', got %q", val)
		}
	}
}

func TestApplyTransform_EmptyTree(t *testing.T) {
	root := NewRoot("Empty")
	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	var visitCount int

	countVisits := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		visitCount++
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, countVisits)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should visit just the root
	if visitCount != 1 {
		t.Errorf("expected 1 visit, got %d", visitCount)
	}
}

func TestApplyTransform_DeepTree(t *testing.T) {
	// Build a chain: root -> l1 -> l2 -> l3 -> l4 -> l5
	root := NewRoot("root")
	current := root
	depth := 5

	for i := 1; i <= depth; i++ {
		current = current.CreateChild("level"+string(rune('0'+i)), i, "")
	}

	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	var visitCount int

	countVisits := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		visitCount++
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, countVisits)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should visit root + 5 children = 6 nodes
	expectedVisits := depth + 1
	if visitCount != expectedVisits {
		t.Errorf("expected %d visits, got %d", expectedVisits, visitCount)
	}
}

func TestApplyTransform_WideTree(t *testing.T) {
	// Build tree: root with 10 children
	root := NewRoot("root")
	numChildren := 10

	for i := 1; i <= numChildren; i++ {
		root.CreateChild("child"+string(rune('0'+i)), 1, "")
	}

	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	var visitCount int

	countVisits := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		visitCount++
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, countVisits)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should visit root + 10 children = 11 nodes
	expectedVisits := numChildren + 1
	if visitCount != expectedVisits {
		t.Errorf("expected %d visits, got %d", expectedVisits, visitCount)
	}
}

func TestApplyTransform_ComplexTree(t *testing.T) {
	// Build complex tree:
	//   root
	//   ├── A
	//   │   ├── A1
	//   │   └── A2
	//   │       └── A2a
	//   └── B
	//       ├── B1
	//       └── B2
	root := NewRoot("root")

	a := root.CreateChild("A", 1, "")
	a.CreateChild("A1", 2, "")
	a2 := a.CreateChild("A2", 2, "")
	a2.CreateChild("A2a", 3, "")

	b := root.CreateChild("B", 1, "")
	b.CreateChild("B1", 2, "")
	b.CreateChild("B2", 2, "")

	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	var visitOrder []string

	recordVisit := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		visitOrder = append(visitOrder, s.Title())
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, recordVisit)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Expect depth-first pre-order: root, A, A1, A2, A2a, B, B1, B2
	expected := []string{"root", "A", "A1", "A2", "A2a", "B", "B1", "B2"}
	if len(visitOrder) != len(expected) {
		t.Fatalf("expected %d visits, got %d: %v", len(expected), len(visitOrder), visitOrder)
	}

	for i, title := range expected {
		if visitOrder[i] != title {
			t.Errorf("visit %d: expected %q, got %q", i, title, visitOrder[i])
		}
	}
}

func TestApplyTransform_ModifyingChildren(t *testing.T) {
	// Test that transforms can modify content of children
	root := NewRoot("root")
	child1 := root.CreateChild("child1", 1, "original1")
	child2 := root.CreateChild("child2", 1, "original2")

	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	appendTitle := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		s.SetContent(s.Content() + ":" + s.Title())
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, appendTitle)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if child1.Content() != "original1:child1" {
		t.Errorf("child1: expected 'original1:child1', got %q", child1.Content())
	}
	if child2.Content() != "original2:child2" {
		t.Errorf("child2: expected 'original2:child2', got %q", child2.Content())
	}
}

func TestApplyTransform_MultipleTransformsAllNodes(t *testing.T) {
	// Build tree with 3 nodes: root -> child -> grandchild
	root := NewRoot("root")
	root.SetContent("A")
	child := root.CreateChild("child", 1, "B")
	grandchild := child.CreateChild("grandchild", 2, "C")

	frontmatter := fm.EmptyFrontMatter()
	ctx := context.Background()

	addPrefix := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		s.SetContent("P" + s.Content())
		return nil
	}

	addSuffix := func(ctx context.Context, _ fm.FrontMatterView, s *Section) error {
		s.SetContent(s.Content() + "S")
		return nil
	}

	err := ApplyTransform(ctx, frontmatter, root, addPrefix, addSuffix)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Each node should have both transforms applied
	if root.Content() != "PAS" {
		t.Errorf("root: expected 'PAS', got %q", root.Content())
	}
	if child.Content() != "PBS" {
		t.Errorf("child: expected 'PBS', got %q", child.Content())
	}
	if grandchild.Content() != "PCS" {
		t.Errorf("grandchild: expected 'PCS', got %q", grandchild.Content())
	}
}
