package builtin

import (
	"context"
	"testing"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
)

func TestPruneLeadingBlankLinesTransform(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxKeep  int
		expected string
	}{
		{
			name:     "remove all leading blanks",
			input:    "\n\n\nline1\nline2",
			maxKeep:  0,
			expected: "line1\nline2",
		},
		{
			name:     "keep one leading blank",
			input:    "\n\n\nline1\nline2",
			maxKeep:  1,
			expected: "\nline1\nline2",
		},
		{
			name:     "keep two leading blanks",
			input:    "\n\n\nline1\nline2",
			maxKeep:  2,
			expected: "\n\nline1\nline2",
		},
		{
			name:     "no leading blanks",
			input:    "line1\nline2",
			maxKeep:  0,
			expected: "line1\nline2",
		},
		{
			name:     "maxKeep exceeds blanks",
			input:    "\n\nline1",
			maxKeep:  5,
			expected: "\n\nline1",
		},
		{
			name:     "empty content",
			input:    "",
			maxKeep:  0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := section.NewRoot("Test")
			s.SetContent(tt.input)

			transform := PruneLeadingBlankLinesTransform(tt.maxKeep)
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

func TestPruneTrailingBlankLinesTransform(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxKeep  int
		expected string
	}{
		{
			name:     "remove all trailing blanks",
			input:    "line1\nline2\n\n\n",
			maxKeep:  0,
			expected: "line1\nline2",
		},
		{
			name:     "keep one trailing blank",
			input:    "line1\nline2\n\n\n",
			maxKeep:  1,
			expected: "line1\nline2\n",
		},
		{
			name:     "keep two trailing blanks",
			input:    "line1\nline2\n\n\n",
			maxKeep:  2,
			expected: "line1\nline2\n\n",
		},
		{
			name:     "no trailing blanks",
			input:    "line1\nline2",
			maxKeep:  0,
			expected: "line1\nline2",
		},
		{
			name:     "maxKeep exceeds blanks",
			input:    "line1\n\n",
			maxKeep:  5,
			expected: "line1\n\n",
		},
		{
			name:     "empty content",
			input:    "",
			maxKeep:  0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := section.NewRoot("Test")
			s.SetContent(tt.input)

			transform := PruneTrailingBlankLinesTransform(tt.maxKeep)
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
