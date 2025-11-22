package builtin

import (
	"context"
	"testing"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
)

func TestCollapseBlankLinesTransform(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no blank lines",
			input:    "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "single blank line preserved",
			input:    "line1\n\nline2",
			expected: "line1\n\nline2",
		},
		{
			name:     "multiple blank lines collapsed",
			input:    "line1\n\n\n\nline2",
			expected: "line1\n\nline2",
		},
		{
			name:     "mixed blanks collapsed",
			input:    "line1\n\n\nline2\n\n\n\nline3",
			expected: "line1\n\nline2\n\nline3",
		},
		{
			name:     "preserve blanks in code fence",
			input:    "text\n```\ncode\n\n\n\nmore code\n```\ntext",
			expected: "text\n```\ncode\n\n\n\nmore code\n```\ntext",
		},
		{
			name:     "collapse outside but not inside fence",
			input:    "text\n\n\n```\ncode\n\n\ncode\n```\n\n\ntext",
			expected: "text\n\n```\ncode\n\n\ncode\n```\n\ntext",
		},
		{
			name:     "tildes code fence",
			input:    "text\n~~~\ncode\n\n\n\ncode\n~~~\ntext",
			expected: "text\n~~~\ncode\n\n\n\ncode\n~~~\ntext",
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

			transform := CollapseBlankLinesTransform()
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
