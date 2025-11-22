package builtin

import (
	"context"
	"testing"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
)

func TestHeadingPathCommentTransform(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *section.Section
		expected string
	}{
		{
			name: "root section",
			setup: func() *section.Section {
				return section.NewRoot("Root")
			},
			expected: "<!-- path: Root -->\n",
		},
		{
			name: "nested section",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				child := root.CreateChild("Chapter 1", 1, "")
				return child
			},
			expected: "<!-- path: Root / Chapter 1 -->\n",
		},
		{
			name: "deeply nested",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				ch1 := root.CreateChild("Chapter 1", 1, "")
				sec := ch1.CreateChild("Section 1.1", 2, "")
				return sec
			},
			expected: "<!-- path: Root / Chapter 1 / Section 1.1 -->\n",
		},
		{
			name: "with existing content",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				root.SetContent("Some content here")
				return root
			},
			expected: "<!-- path: Root -->\nSome content here",
		},
		{
			name: "idempotent - correct path exists",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				root.SetContent("<!-- path: Root -->\nContent")
				return root
			},
			expected: "<!-- path: Root -->\nContent",
		},
		{
			name: "replace stale path",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				root.SetContent("<!-- path: Old / Path -->\nContent")
				return root
			},
			expected: "<!-- path: Root -->\nContent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			transform := HeadingPathCommentTransform()
			err := transform(context.Background(), fm.EmptyFrontMatter().View(), s)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if s.Content() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, s.Content())
			}

			// Test idempotency
			err = transform(context.Background(), fm.EmptyFrontMatter().View(), s)
			if err != nil {
				t.Fatalf("unexpected error on second pass: %v", err)
			}
			if s.Content() != tt.expected {
				t.Errorf("not idempotent: expected %q, got %q", tt.expected, s.Content())
			}
		})
	}
}

func TestHeadingPrefixTransform(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *section.Section
		expected string
	}{
		{
			name: "skip root",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				root.SetContent("Content")
				return root
			},
			expected: "Content",
		},
		{
			name: "add heading to empty section",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				return root.CreateChild("Chapter 1", 1, "")
			},
			expected: "# Chapter 1\n\n",
		},
		{
			name: "add heading before content",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				child := root.CreateChild("Chapter 1", 1, "Some content")
				return child
			},
			expected: "# Chapter 1\n\nSome content",
		},
		{
			name: "level 2 heading",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				child := root.CreateChild("Section", 2, "Content")
				return child
			},
			expected: "## Section\n\nContent",
		},
		{
			name: "level 3 heading",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				child := root.CreateChild("Subsection", 3, "Content")
				return child
			},
			expected: "### Subsection\n\nContent",
		},
		{
			name: "idempotent - heading exists",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				child := root.CreateChild("Chapter", 1, "# Chapter\n\nContent")
				return child
			},
			expected: "# Chapter\n\nContent",
		},
		{
			name: "insert after path comment",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				child := root.CreateChild("Chapter", 1, "<!-- path: Root / Chapter -->\nContent")
				return child
			},
			expected: "<!-- path: Root / Chapter -->\n# Chapter\n\nContent",
		},
		{
			name: "idempotent with path comment",
			setup: func() *section.Section {
				root := section.NewRoot("Root")
				child := root.CreateChild("Chapter", 1, "<!-- path: Root / Chapter -->\n# Chapter\n\nContent")
				return child
			},
			expected: "<!-- path: Root / Chapter -->\n# Chapter\n\nContent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			transform := HeadingPrefixTransform()
			err := transform(context.Background(), fm.EmptyFrontMatter().View(), s)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if s.Content() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, s.Content())
			}

			// Test idempotency
			err = transform(context.Background(), fm.EmptyFrontMatter().View(), s)
			if err != nil {
				t.Fatalf("unexpected error on second pass: %v", err)
			}
			if s.Content() != tt.expected {
				t.Errorf("not idempotent: expected %q, got %q", tt.expected, s.Content())
			}
		})
	}
}
