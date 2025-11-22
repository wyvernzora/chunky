package context

import "context"

type fiKeyType struct{}

var fiKey fiKeyType

// FileInfo carries lightweight file metadata through the processing pipeline.
// It's safe to store in context and provides information about the document
// being parsed.
type FileInfo struct {
	Path  string // File path (full or repo-relative)
	Title string // Document title derived for the root section
}

// WithFileInfo returns a child context carrying file metadata.
func WithFileInfo(ctx context.Context, fi FileInfo) context.Context {
	return context.WithValue(ctx, fiKey, fi)
}

// FileInfoFrom returns the file metadata if present.
func FileInfoFrom(ctx context.Context) (FileInfo, bool) {
	if v := ctx.Value(fiKey); v != nil {
		if fi, ok := v.(FileInfo); ok {
			return fi, true
		}
	}
	return FileInfo{}, false
}

// MustFileInfo returns file metadata or a zero value if missing.
func MustFileInfo(ctx context.Context) FileInfo {
	fi, _ := FileInfoFrom(ctx)
	return fi
}
