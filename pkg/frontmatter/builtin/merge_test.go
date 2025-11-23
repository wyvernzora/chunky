package builtin

import (
	"context"
	"testing"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
)

func TestMergeFrontMatter_EmptyData(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"existing": "value",
	}

	transform := MergeFrontMatter(fm.FrontMatter{})
	err := transform(context.Background(), frontmatter)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(frontmatter) != 1 {
		t.Errorf("expected frontmatter unchanged, got %v", frontmatter)
	}
}

func TestMergeFrontMatter_AddNewKeys(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"existing": "value",
	}

	transform := MergeFrontMatter(fm.FrontMatter{
		"new_key":  "new_value",
		"another":  123,
		"bool_val": true,
	})

	err := transform(context.Background(), frontmatter)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(frontmatter) != 4 {
		t.Errorf("expected 4 keys, got %d", len(frontmatter))
	}

	if frontmatter["existing"] != "value" {
		t.Errorf("existing key modified")
	}

	if frontmatter["new_key"] != "new_value" {
		t.Errorf("new_key not added correctly")
	}

	if frontmatter["another"] != 123 {
		t.Errorf("another not added correctly")
	}

	if frontmatter["bool_val"] != true {
		t.Errorf("bool_val not added correctly")
	}
}

func TestMergeFrontMatter_NoOverwrite(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"key1": "original",
		"key2": 42,
	}

	transform := MergeFrontMatter(fm.FrontMatter{
		"key1": "should_not_overwrite",
		"key2": 999,
		"key3": "new_value",
	})

	err := transform(context.Background(), frontmatter)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if frontmatter["key1"] != "original" {
		t.Errorf("key1 was overwritten: got %v", frontmatter["key1"])
	}

	if frontmatter["key2"] != 42 {
		t.Errorf("key2 was overwritten: got %v", frontmatter["key2"])
	}

	if frontmatter["key3"] != "new_value" {
		t.Errorf("key3 not added: got %v", frontmatter["key3"])
	}
}

func TestMergeFrontMatter_EmptyFrontMatter(t *testing.T) {
	frontmatter := fm.FrontMatter{}

	transform := MergeFrontMatter(fm.FrontMatter{
		"key1": "value1",
		"key2": "value2",
	})

	err := transform(context.Background(), frontmatter)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(frontmatter) != 2 {
		t.Errorf("expected 2 keys, got %d", len(frontmatter))
	}

	if frontmatter["key1"] != "value1" {
		t.Errorf("key1 not added")
	}

	if frontmatter["key2"] != "value2" {
		t.Errorf("key2 not added")
	}
}

func TestMergeFrontMatter_ComplexTypes(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"existing": "value",
	}

	nestedMap := map[string]interface{}{
		"nested_key": "nested_value",
	}

	slice := []string{"a", "b", "c"}

	transform := MergeFrontMatter(fm.FrontMatter{
		"map":   nestedMap,
		"slice": slice,
	})

	err := transform(context.Background(), frontmatter)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, ok := frontmatter["map"]; !ok {
		t.Errorf("nested map not added")
	}

	if _, ok := frontmatter["slice"]; !ok {
		t.Errorf("slice not added")
	}
}

func TestMergeFrontMatter_NilData(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"existing": "value",
	}

	transform := MergeFrontMatter(nil)
	err := transform(context.Background(), frontmatter)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(frontmatter) != 1 {
		t.Errorf("frontmatter modified when data is nil")
	}
}

func TestMergeFrontMatter_MultipleApplications(t *testing.T) {
	frontmatter := fm.FrontMatter{
		"key1": "value1",
	}

	transform1 := MergeFrontMatter(fm.FrontMatter{
		"key2": "value2",
	})

	transform2 := MergeFrontMatter(fm.FrontMatter{
		"key3": "value3",
		"key1": "should_not_overwrite",
	})

	err := transform1(context.Background(), frontmatter)
	if err != nil {
		t.Fatalf("transform1 failed: %v", err)
	}

	err = transform2(context.Background(), frontmatter)
	if err != nil {
		t.Fatalf("transform2 failed: %v", err)
	}

	if len(frontmatter) != 3 {
		t.Errorf("expected 3 keys, got %d", len(frontmatter))
	}

	if frontmatter["key1"] != "value1" {
		t.Errorf("key1 was overwritten")
	}

	if frontmatter["key2"] != "value2" {
		t.Errorf("key2 not present")
	}

	if frontmatter["key3"] != "value3" {
		t.Errorf("key3 not present")
	}
}
