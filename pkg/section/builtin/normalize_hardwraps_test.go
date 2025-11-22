package builtin

import (
	"context"
	"testing"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
)

func TestNormalizeHardWrapsTransform(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "merge prose paragraph",
			input:    "This is a long\nsentence that was\nhard wrapped.",
			expected: "This is a long sentence that was hard wrapped.",
		},
		{
			name:     "preserve blank line paragraph breaks",
			input:    "Paragraph 1 with\nwrapped lines.\n\nParagraph 2 with\nmore wraps.",
			expected: "Paragraph 1 with wrapped lines.\n\nParagraph 2 with more wraps.",
		},
		{
			name:     "preserve ATX headings",
			input:    "# Heading\nSome text\nwrapped.",
			expected: "# Heading\nSome text wrapped.",
		},
		{
			name:     "preserve blockquotes",
			input:    "> Quote line\n> continues here\n\nRegular text\nwrapped.",
			expected: "> Quote line > continues here\n\nRegular text wrapped.",
		},
		{
			name:     "preserve unordered list",
			input:    "- Item 1\n- Item 2\n\nText after\nwrapped.",
			expected: "- Item 1\n- Item 2\n\nText after wrapped.",
		},
		{
			name:     "preserve ordered list",
			input:    "1. First\n2. Second\n\nText after\nwrapped.",
			expected: "1. First\n2. Second\n\nText after wrapped.",
		},
		{
			name:     "preserve code fences",
			input:    "Text before.\n```\ncode line 1\ncode line 2\n```\nText after\nwrapped.",
			expected: "Text before.\n```\ncode line 1\ncode line 2\n```\nText after wrapped.",
		},
		{
			name:     "preserve HTML comments",
			input:    "<!-- comment -->\nText after\nwrapped.",
			expected: "<!-- comment -->\nText after wrapped.",
		},
		{
			name:     "don't merge when followed by blank",
			input:    "Line 1\nLine 2\n\nLine 3",
			expected: "Line 1 Line 2\n\nLine 3",
		},
		{
			name:     "don't merge when followed by special",
			input:    "Text line\n# Heading",
			expected: "Text line\n# Heading",
		},
		{
			name:     "empty content",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := section.NewRoot("Test")
			s.SetContent(tt.input)

			transform := NormalizeHardWrapsTransform()
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
