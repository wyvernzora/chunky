package builtin

import (
	"context"
	"testing"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
)

func TestNormalizeNewlinesTransform(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unix newlines unchanged",
			input:    "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "windows newlines converted",
			input:    "line1\r\nline2\r\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "mac newlines converted",
			input:    "line1\rline2\rline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "mixed newlines normalized",
			input:    "line1\r\nline2\nline3\rline4",
			expected: "line1\nline2\nline3\nline4",
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

			transform := NormalizeNewlinesTransform()
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
