package tokenizer

import (
	"context"
	"errors"
	"testing"

	"github.com/wyvernzora/chunky/pkg/section"
)

func TestMakeTokenizer(t *testing.T) {
	counter := func(text string) (int, error) {
		return len(text), nil
	}

	tok := MakeTokenizer(counter)
	if tok == nil {
		t.Fatal("MakeTokenizer returned nil")
	}

	count, err := tok.Count("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestTokenizer_Count_Success(t *testing.T) {
	// Simple word counter
	counter := func(text string) (int, error) {
		if text == "" {
			return 0, nil
		}
		words := 1
		for _, ch := range text {
			if ch == ' ' {
				words++
			}
		}
		return words, nil
	}

	tok := MakeTokenizer(counter)

	testCases := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty", "", 0},
		{"single word", "hello", 1},
		{"two words", "hello world", 2},
		{"multiple words", "the quick brown fox", 4},
		{"with punctuation", "hello, world!", 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := tok.Count(tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, count)
			}
		})
	}
}

func TestTokenizer_Count_Error(t *testing.T) {
	expectedErr := errors.New("counting failed")
	counter := func(text string) (int, error) {
		if text == "fail" {
			return 0, expectedErr
		}
		return len(text), nil
	}

	tok := MakeTokenizer(counter)

	count, err := tok.Count("fail")
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if count != 0 {
		t.Errorf("expected count 0 on error, got %d", count)
	}
}

func TestTokenizer_Tokenize_SingleNode(t *testing.T) {
	counter := func(text string) (int, error) {
		return len(text), nil
	}

	tok := MakeTokenizer(counter)
	root := section.NewRoot("Test")
	root.SetContent("hello")

	ctx := context.Background()
	result, err := tok.Tokenize(ctx, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.GetContentTokens() != 5 {
		t.Errorf("expected ContentTokens 5, got %d", result.GetContentTokens())
	}
	if result.GetSubtreeTokens() != 5 {
		t.Errorf("expected SubtreeTokens 5, got %d", result.GetSubtreeTokens())
	}
	if len(result.GetChildren()) != 0 {
		t.Errorf("expected no children, got %d", len(result.GetChildren()))
	}
}

func TestTokenizer_Tokenize_WithChildren(t *testing.T) {
	// Character counter
	counter := func(text string) (int, error) {
		return len(text), nil
	}

	tok := MakeTokenizer(counter)

	// Build tree: root (5) -> child1 (6) -> grandchild (10)
	//                      -> child2 (4)
	root := section.NewRoot("root")
	root.SetContent("hello") // 5 chars

	child1 := root.CreateChild("child1", 1, "world!") // 6 chars
	child1.CreateChild("grandchild", 2, "0123456789") // 10 chars

	root.CreateChild("child2", 1, "test") // 4 chars

	ctx := context.Background()
	result, err := tok.Tokenize(ctx, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Root: 5 + (6 + 10) + 4 = 25
	if result.GetContentTokens() != 5 {
		t.Errorf("root ContentTokens: expected 5, got %d", result.GetContentTokens())
	}
	if result.GetSubtreeTokens() != 25 {
		t.Errorf("root SubtreeTokens: expected 25, got %d", result.GetSubtreeTokens())
	}

	if len(result.GetChildren()) != 2 {
		t.Fatalf("expected 2 children, got %d", len(result.GetChildren()))
	}

	// Child1: 6 + 10 = 16
	child1Result := result.GetChildren()[0]
	if child1Result.GetContentTokens() != 6 {
		t.Errorf("child1 ContentTokens: expected 6, got %d", child1Result.GetContentTokens())
	}
	if child1Result.GetSubtreeTokens() != 16 {
		t.Errorf("child1 SubtreeTokens: expected 16, got %d", child1Result.GetSubtreeTokens())
	}

	// Grandchild: 10
	if len(child1Result.GetChildren()) != 1 {
		t.Fatalf("expected 1 grandchild, got %d", len(child1Result.GetChildren()))
	}
	grandchild := child1Result.GetChildren()[0]
	if grandchild.GetContentTokens() != 10 {
		t.Errorf("grandchild ContentTokens: expected 10, got %d", grandchild.GetContentTokens())
	}
	if grandchild.GetSubtreeTokens() != 10 {
		t.Errorf("grandchild SubtreeTokens: expected 10, got %d", grandchild.GetSubtreeTokens())
	}

	// Child2: 4
	child2Result := result.GetChildren()[1]
	if child2Result.GetContentTokens() != 4 {
		t.Errorf("child2 ContentTokens: expected 4, got %d", child2Result.GetContentTokens())
	}
	if child2Result.GetSubtreeTokens() != 4 {
		t.Errorf("child2 SubtreeTokens: expected 4, got %d", child2Result.GetSubtreeTokens())
	}
}

func TestTokenizer_Tokenize_EmptyContent(t *testing.T) {
	counter := func(text string) (int, error) {
		return len(text), nil
	}

	tok := MakeTokenizer(counter)
	root := section.NewRoot("Empty")
	// No content set

	ctx := context.Background()
	result, err := tok.Tokenize(ctx, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.GetContentTokens() != 0 {
		t.Errorf("expected ContentTokens 0, got %d", result.GetContentTokens())
	}
	if result.GetSubtreeTokens() != 0 {
		t.Errorf("expected SubtreeTokens 0, got %d", result.GetSubtreeTokens())
	}
}

func TestTokenizer_Tokenize_Error(t *testing.T) {
	expectedErr := errors.New("counting failed")
	counter := func(text string) (int, error) {
		if text == "fail" {
			return 0, expectedErr
		}
		return len(text), nil
	}

	tok := MakeTokenizer(counter)
	root := section.NewRoot("Test")
	root.SetContent("fail")

	ctx := context.Background()
	result, err := tok.Tokenize(ctx, root)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}
}

func TestTokenizer_Tokenize_ErrorInChild(t *testing.T) {
	expectedErr := errors.New("child counting failed")
	counter := func(text string) (int, error) {
		if text == "fail" {
			return 0, expectedErr
		}
		return len(text), nil
	}

	tok := MakeTokenizer(counter)
	root := section.NewRoot("Test")
	root.SetContent("ok")
	root.CreateChild("child", 1, "fail")

	ctx := context.Background()
	result, err := tok.Tokenize(ctx, root)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}
}

func TestTokenizer_Tokenize_DeepTree(t *testing.T) {
	counter := func(text string) (int, error) {
		return 1, nil // Each node counts as 1
	}

	tok := MakeTokenizer(counter)

	// Build chain: root -> l1 -> l2 -> l3 -> l4 -> l5
	root := section.NewRoot("root")
	root.SetContent("0")
	current := root
	for i := 1; i <= 5; i++ {
		current = current.CreateChild("level", i, "x")
	}

	ctx := context.Background()
	result, err := tok.Tokenize(ctx, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 6 nodes total (root + 5 children)
	if result.GetSubtreeTokens() != 6 {
		t.Errorf("expected SubtreeTokens 6, got %d", result.GetSubtreeTokens())
	}
}

func TestTokenizer_Tokenize_WideTree(t *testing.T) {
	counter := func(text string) (int, error) {
		return 1, nil // Each node counts as 1
	}

	tok := MakeTokenizer(counter)

	// Build tree: root with 10 children
	root := section.NewRoot("root")
	root.SetContent("0")
	for i := 1; i <= 10; i++ {
		root.CreateChild("child", 1, "x")
	}

	ctx := context.Background()
	result, err := tok.Tokenize(ctx, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 11 nodes total (root + 10 children)
	if result.GetSubtreeTokens() != 11 {
		t.Errorf("expected SubtreeTokens 11, got %d", result.GetSubtreeTokens())
	}

	if len(result.GetChildren()) != 10 {
		t.Errorf("expected 10 children, got %d", len(result.GetChildren()))
	}
}

func TestTokenizer_Tokenize_PreservesStructure(t *testing.T) {
	counter := func(text string) (int, error) {
		return len(text), nil
	}

	tok := MakeTokenizer(counter)

	root := section.NewRoot("Root Title")
	child := root.CreateChild("Child Title", 1, "content")

	ctx := context.Background()
	result, err := tok.Tokenize(ctx, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify section references are preserved
	if result.GetSection() != root {
		t.Error("root section reference not preserved")
	}

	if len(result.GetChildren()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(result.GetChildren()))
	}

	if result.GetChildren()[0].GetSection() != child {
		t.Error("child section reference not preserved")
	}

	// Verify titles are accessible
	if result.GetSection().Title() != "Root Title" {
		t.Errorf("expected root title 'Root Title', got %q", result.GetSection().Title())
	}

	if result.GetChildren()[0].GetSection().Title() != "Child Title" {
		t.Errorf("expected child title 'Child Title', got %q", result.GetChildren()[0].GetSection().Title())
	}
}
