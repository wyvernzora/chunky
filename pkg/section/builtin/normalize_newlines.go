package builtin

import (
	"context"
	"strings"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/section"
)

// NormalizeNewlinesTransform converts all "\r\n" and "\r" to "\n" within section.Content.
// This transform is idempotent by definition.
func NormalizeNewlinesTransform() section.Transform {
	return func(ctx context.Context, _ fm.FrontMatterView, s *section.Section) error {
		content := s.Content()

		// Replace \r\n first, then \r
		content = strings.ReplaceAll(content, "\r\n", "\n")
		content = strings.ReplaceAll(content, "\r", "\n")

		s.SetContent(content)
		return nil
	}
}
