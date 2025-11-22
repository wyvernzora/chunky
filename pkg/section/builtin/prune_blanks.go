package builtin

import (
	"context"
	"strings"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
)

// PruneLeadingBlankLinesTransform removes leading blank lines at the start
// of section.Content but allows up to maxKeep blank lines to remain.
// Example: maxKeep=0 removes all; maxKeep=1 keeps at most one.
// This transform is idempotent with the same maxKeep value.
func PruneLeadingBlankLinesTransform(maxKeep int) section.Transform {
	return func(ctx context.Context, _ fm.FrontMatterView, s *section.Section) error {
		content := s.Content()
		if content == "" {
			return nil
		}

		lines := strings.Split(content, "\n")

		// Count leading blank lines
		leadingBlanks := 0
		for i, line := range lines {
			if strings.TrimSpace(line) == "" {
				leadingBlanks++
			} else {
				// Stop at first non-blank line
				break
			}
			if i == len(lines)-1 {
				// All lines are blank
				leadingBlanks = len(lines)
			}
		}

		// Determine how many to remove
		toRemove := leadingBlanks - maxKeep
		if toRemove <= 0 {
			return nil // Nothing to do
		}

		// Remove the excess leading blanks
		result := strings.Join(lines[toRemove:], "\n")
		s.SetContent(result)
		return nil
	}
}

// PruneTrailingBlankLinesTransform removes trailing blank lines at the end
// of section.Content but allows up to maxKeep blank lines to remain.
// Example: maxKeep=0 removes all; maxKeep=1 keeps at most one.
// This transform is idempotent with the same maxKeep value.
func PruneTrailingBlankLinesTransform(maxKeep int) section.Transform {
	return func(ctx context.Context, _ fm.FrontMatterView, s *section.Section) error {
		content := s.Content()
		if content == "" {
			return nil
		}

		lines := strings.Split(content, "\n")

		// Count trailing blank lines
		trailingBlanks := 0
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.TrimSpace(lines[i]) == "" {
				trailingBlanks++
			} else {
				break
			}
		}

		// Determine how many to remove
		toRemove := trailingBlanks - maxKeep
		if toRemove <= 0 {
			return nil // Nothing to do
		}

		// Remove the excess trailing blanks
		newLen := len(lines) - toRemove
		result := strings.Join(lines[:newLen], "\n")
		s.SetContent(result)
		return nil
	}
}
