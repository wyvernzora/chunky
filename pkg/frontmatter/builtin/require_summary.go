package builtin

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	cctx "github.com/wyvernzora/chunky/pkg/context"
	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
)

// RequireSummary returns a transform that validates the presence and content
// of a "summary" field in the frontmatter. The summary must exist, be a string,
// and contain non-whitespace content.
//
// This transform is useful for enforcing documentation standards where every
// document must have a meaningful summary.
//
// Returns:
//   - A Transform that validates the summary field
//   - An error if:
//   - frontmatter is nil
//   - the "summary" key is missing
//   - the summary value is not a string
//   - the summary string is empty or contains only whitespace
//
// Example:
//
//	transform := RequireSummary()
//	err := fm.ApplyTransform(ctx, frontmatter, transform)
//	if err != nil {
//	    // Handle missing or invalid summary
//	}
func RequireSummary() fm.Transform {
	return func(ctx context.Context, frontmatter fm.FrontMatter) error {
		logger := cctx.Logger(ctx)

		if frontmatter == nil {
			logger.Error("frontmatter is nil")
			return fmt.Errorf("RequireSummary: frontmatter cannot be nil")
		}

		raw, ok := frontmatter["summary"]
		if !ok {
			logger.Error("summary field missing from frontmatter")
			return fmt.Errorf("RequireSummary: frontmatter missing required 'summary' field")
		}

		s, ok := raw.(string)
		if !ok {
			logger.Error("summary field is not a string",
				slog.String("type", fmt.Sprintf("%T", raw)))
			return fmt.Errorf("RequireSummary: 'summary' field must be a string, got %T", raw)
		}

		if strings.TrimSpace(s) == "" {
			logger.Error("summary field is empty or contains only whitespace")
			return fmt.Errorf("RequireSummary: 'summary' field cannot be empty")
		}

		logger.Debug("summary field validated successfully",
			slog.Int("length", len(s)))

		return nil
	}
}
