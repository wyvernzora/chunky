package builtin

import (
	"context"
	"log/slog"

	cctx "github.com/wyvernzora/chunky/pkg/context"
	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
)

// MergeFrontMatter creates a transform that merges additional metadata into frontmatter.
// The provided data is merged into the frontmatter, but existing keys are NOT overwritten.
//
// This is useful for adding default metadata, contextual information, or supplementary
// fields without modifying user-provided frontmatter.
//
// Merge behavior:
//   - Keys that don't exist in frontmatter are added
//   - Keys that already exist are left unchanged (no overwrite)
//   - The merge happens in-place, modifying the frontmatter map
//
// Parameters:
//   - data: Map of key-value pairs to merge into frontmatter
//
// Example:
//
//	// Add version and environment metadata
//	transform := MergeFrontMatter(fm.FrontMatter{
//	    "version": "1.0",
//	    "environment": "production",
//	})
//
//	// If frontmatter already has "version", it won't be overwritten
//	// If "environment" is missing, it will be added
func MergeFrontMatter(data fm.FrontMatter) fm.Transform {
	return func(ctx context.Context, frontmatter fm.FrontMatter) error {
		logger := cctx.Logger(ctx)

		if len(data) == 0 {
			logger.Debug("merge frontmatter: no data to merge")
			return nil
		}

		merged := 0
		skipped := 0

		for key, value := range data {
			if _, exists := frontmatter[key]; !exists {
				frontmatter[key] = value
				merged++
				logger.Debug("merge frontmatter: added key",
					slog.String("key", key))
			} else {
				skipped++
				logger.Debug("merge frontmatter: skipped existing key",
					slog.String("key", key))
			}
		}

		logger.Debug("merge frontmatter: completed",
			slog.Int("merged", merged),
			slog.Int("skipped", skipped))

		return nil
	}
}
