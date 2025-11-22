package frontmatter

import (
	"context"
	"errors"
	"testing"
)

func TestApplyTransform_Empty(t *testing.T) {
	fm := EmptyFrontMatter()
	ctx := context.Background()

	// No transforms should succeed
	err := ApplyTransform(ctx, fm)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestApplyTransform_SingleTransform(t *testing.T) {
	fm := FrontMatter{"title": "Original"}
	ctx := context.Background()

	addAuthor := func(ctx context.Context, fm FrontMatter) error {
		fm["author"] = "John Doe"
		return nil
	}

	err := ApplyTransform(ctx, fm, addAuthor)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if fm["title"] != "Original" {
		t.Errorf("expected title to remain 'Original', got %v", fm["title"])
	}
	if fm["author"] != "John Doe" {
		t.Errorf("expected author 'John Doe', got %v", fm["author"])
	}
}

func TestApplyTransform_MultipleTransforms(t *testing.T) {
	fm := FrontMatter{"title": "My Document"}
	ctx := context.Background()

	addAuthor := func(ctx context.Context, fm FrontMatter) error {
		fm["author"] = "Jane Smith"
		return nil
	}

	addTags := func(ctx context.Context, fm FrontMatter) error {
		fm["tags"] = []string{"go", "markdown"}
		return nil
	}

	modifyTitle := func(ctx context.Context, fm FrontMatter) error {
		if title, ok := fm["title"].(string); ok {
			fm["title"] = title + " - Updated"
		}
		return nil
	}

	err := ApplyTransform(ctx, fm, addAuthor, addTags, modifyTitle)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if fm["title"] != "My Document - Updated" {
		t.Errorf("expected updated title, got %v", fm["title"])
	}
	if fm["author"] != "Jane Smith" {
		t.Errorf("expected author 'Jane Smith', got %v", fm["author"])
	}
	tags, ok := fm["tags"].([]string)
	if !ok || len(tags) != 2 {
		t.Errorf("expected tags array with 2 elements, got %v", fm["tags"])
	}
}

func TestApplyTransform_RemoveKey(t *testing.T) {
	fm := FrontMatter{
		"title":  "Test",
		"draft":  true,
		"author": "Someone",
	}
	ctx := context.Background()

	removeDraft := func(ctx context.Context, fm FrontMatter) error {
		delete(fm, "draft")
		return nil
	}

	err := ApplyTransform(ctx, fm, removeDraft)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, exists := fm["draft"]; exists {
		t.Error("expected 'draft' key to be removed")
	}
	if fm["title"] != "Test" {
		t.Errorf("expected title to remain, got %v", fm["title"])
	}
}

func TestApplyTransform_ErrorHandling(t *testing.T) {
	fm := FrontMatter{"title": "Test"}
	ctx := context.Background()

	expectedErr := errors.New("transform failed")

	successfulTransform := func(ctx context.Context, fm FrontMatter) error {
		fm["step1"] = "done"
		return nil
	}

	failingTransform := func(ctx context.Context, fm FrontMatter) error {
		return expectedErr
	}

	neverCalledTransform := func(ctx context.Context, fm FrontMatter) error {
		fm["step3"] = "done"
		return nil
	}

	err := ApplyTransform(ctx, fm, successfulTransform, failingTransform, neverCalledTransform)

	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	// First transform should have been applied
	if fm["step1"] != "done" {
		t.Error("expected first transform to have been applied")
	}

	// Third transform should NOT have been applied
	if _, exists := fm["step3"]; exists {
		t.Error("expected third transform to NOT have been applied after error")
	}
}

func TestApplyTransform_OrderMatters(t *testing.T) {
	ctx := context.Background()

	incrementCounter := func(ctx context.Context, fm FrontMatter) error {
		if counter, ok := fm["counter"].(int); ok {
			fm["counter"] = counter + 1
		}
		return nil
	}

	doubleCounter := func(ctx context.Context, fm FrontMatter) error {
		if counter, ok := fm["counter"].(int); ok {
			fm["counter"] = counter * 2
		}
		return nil
	}

	// Apply in order: increment then double
	fm1 := FrontMatter{"counter": 0}
	err := ApplyTransform(ctx, fm1, incrementCounter, doubleCounter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (0 + 1) * 2 = 2
	if fm1["counter"] != 2 {
		t.Errorf("expected counter 2, got %v", fm1["counter"])
	}

	// Apply in reverse order: double then increment
	fm2 := FrontMatter{"counter": 0}
	err = ApplyTransform(ctx, fm2, doubleCounter, incrementCounter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (0 * 2) + 1 = 1
	if fm2["counter"] != 1 {
		t.Errorf("expected counter 1, got %v", fm2["counter"])
	}
}

func TestApplyTransform_ContextPropagation(t *testing.T) {
	fm := EmptyFrontMatter()

	type contextKey string
	const userKey contextKey = "user"

	ctx := context.WithValue(context.Background(), userKey, "testuser")

	checkContext := func(ctx context.Context, fm FrontMatter) error {
		if user, ok := ctx.Value(userKey).(string); ok {
			fm["processed_by"] = user
			return nil
		}
		return errors.New("context value not found")
	}

	err := ApplyTransform(ctx, fm, checkContext)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if fm["processed_by"] != "testuser" {
		t.Errorf("expected processed_by 'testuser', got %v", fm["processed_by"])
	}
}

func TestApplyTransform_NilFrontMatter(t *testing.T) {
	var fm FrontMatter // nil map
	ctx := context.Background()

	addKey := func(ctx context.Context, fm FrontMatter) error {
		fm["test"] = "value"
		return nil
	}

	// This should panic when writing to nil map
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when writing to nil map")
		}
	}()

	ApplyTransform(ctx, fm, addKey)
}

func TestApplyTransform_EmptyFrontMatter(t *testing.T) {
	fm := EmptyFrontMatter()
	ctx := context.Background()

	populate := func(ctx context.Context, fm FrontMatter) error {
		fm["title"] = "New Title"
		fm["author"] = "Author Name"
		fm["tags"] = []string{"tag1", "tag2"}
		return nil
	}

	err := ApplyTransform(ctx, fm, populate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm) != 3 {
		t.Errorf("expected 3 keys in frontmatter, got %d", len(fm))
	}
}
