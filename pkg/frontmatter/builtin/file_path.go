package builtin

import (
	"context"
	"fmt"
	"log/slog"

	cctx "github.com/wyvernzora/chunky/pkg/context"
	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
)

// InjectFilePath returns a transform that injects the file path from context
// into the frontmatter at the specified key. If the key already exists in the
// frontmatter, the transform skips injection and logs a debug message.
//
// The file path is retrieved from the FileInfo in the context. If no FileInfo
// exists in the context or the path is empty, an error is returned.
//
// Parameters:
//   - key: The frontmatter key to use for the file path. If empty, defaults to "file_path".
//
// Returns:
//   - A Transform that injects the file path into frontmatter
//
// Example:
//
//	transform := InjectFilePath("source_file")
//	err := fm.ApplyTransform(ctx, frontmatter, transform)
func InjectFilePath(key string) fm.Transform {
	if key == "" {
		key = "file_path"
	}
	return func(ctx context.Context, frontmatter fm.FrontMatter) error {
		logger := cctx.Logger(ctx)

		if frontmatter == nil {
			logger.Error("frontmatter is nil")
			return fmt.Errorf("InjectFilePath: frontmatter cannot be nil")
		}

		if _, exists := frontmatter[key]; exists {
			logger.Debug("file path key already exists in frontmatter, skipping injection",
				slog.String("key", key))
			return nil
		}

		fi, ok := cctx.FileInfoFrom(ctx)
		if !ok || fi.Path == "" {
			logger.Error("file path not available in context")
			return fmt.Errorf("InjectFilePath: file path not found in context")
		}

		frontmatter[key] = fi.Path
		logger.Debug("injected file path into frontmatter",
			slog.String("key", key),
			slog.String("path", fi.Path))

		return nil
	}
}
