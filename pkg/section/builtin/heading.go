package builtin

import (
	"context"
	"fmt"
	"strings"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
)

// HeadingPathCommentTransform computes a breadcrumb path from root to the
// current section and inserts it as an HTML comment at the very top of
// section.Content. The format is: <!-- path: Root / Section1 / Section2 -->
//
// If the section has no content, this transform does nothing (skips the comment).
// This typically happens when a section has no body text and no children.
//
// If a path comment is already present and matches, it does nothing.
// If present but stale, it replaces the comment.
// This transform is idempotent.
func HeadingPathCommentTransform() section.Transform {
	return func(ctx context.Context, _ fm.FrontMatterView, s *section.Section) error {
		content := s.Content()

		// Skip if content is empty or whitespace-only
		if strings.TrimSpace(content) == "" {
			return nil
		}

		// Compute breadcrumb path
		path := computePath(s)
		expectedComment := fmt.Sprintf("<!-- path: %s -->\n", path)

		lines := strings.Split(content, "\n")

		// Check if first line is a path comment
		firstLine := lines[0]
		if strings.HasPrefix(firstLine, "<!-- path:") {
			// Path comment exists, check if it matches
			if firstLine == strings.TrimSuffix(expectedComment, "\n") {
				// Already correct, do nothing
				return nil
			}
			// Stale comment, replace it
			lines[0] = strings.TrimSuffix(expectedComment, "\n")
			s.SetContent(strings.Join(lines, "\n"))
			return nil
		}

		// No path comment, insert at top
		s.SetContent(expectedComment + content)
		return nil
	}
}

// computePath builds the breadcrumb path from root to this section.
func computePath(s *section.Section) string {
	var parts []string
	for current := s; current != nil; current = current.Parent() {
		parts = append([]string{current.Title()}, parts...)
	}
	return strings.Join(parts, " / ")
}

// HeadingPrefixTransform prepends the section's own Markdown heading line
// to the top of section.Content. Root sections (Level = 0) are skipped.
// The injected format is: "### Title\n\n"
//
// This transform is idempotent: if the first non-comment line already matches
// the exact heading for this section, it does nothing.
func HeadingPrefixTransform() section.Transform {
	return func(ctx context.Context, _ fm.FrontMatterView, s *section.Section) error {
		// Skip root sections
		if s.Level() == 0 {
			return nil
		}

		// Compute expected heading line
		expectedHeading := strings.Repeat("#", s.Level()) + " " + s.Title()

		content := s.Content()
		lines := strings.Split(content, "\n")

		// Find first non-comment line
		firstNonCommentIdx := -1
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			// Skip HTML comments
			if strings.HasPrefix(trimmed, "<!--") {
				continue
			}
			firstNonCommentIdx = i
			break
		}

		// Check if first non-comment line matches expected heading
		if firstNonCommentIdx >= 0 {
			if strings.TrimSpace(lines[firstNonCommentIdx]) == expectedHeading {
				// Already present, idempotent
				return nil
			}
		}

		// Insert heading at the first non-comment position or at top
		headingToInsert := expectedHeading + "\n\n"

		if firstNonCommentIdx < 0 {
			// No non-comment lines, append to end
			if content == "" {
				s.SetContent(headingToInsert)
			} else {
				s.SetContent(content + headingToInsert)
			}
			return nil
		}

		// Insert before first non-comment line
		before := strings.Join(lines[:firstNonCommentIdx], "\n")
		if before != "" {
			before += "\n"
		}
		after := strings.Join(lines[firstNonCommentIdx:], "\n")

		s.SetContent(before + headingToInsert + after)
		return nil
	}
}
