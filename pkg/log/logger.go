package log

import (
	"context"
	"io"
	"log/slog"
)

type loggerKey struct{}

var key loggerKey

// WithLogger stores a slog.Logger in the context.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	if l == nil {
		return ctx
	}
	return context.WithValue(ctx, key, l)
}

// Logger retrieves a slog.Logger from context, falling back to slog.Default().
func Logger(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if v := ctx.Value(key); v != nil {
		if l, ok := v.(*slog.Logger); ok && l != nil {
			return l
		}
	}
	return slog.Default()
}

// WithAttrs returns a new context with logger attributes added.
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	l := Logger(ctx).With(attrsToAny(attrs)...)
	return WithLogger(ctx, l)
}

// WithKV lets you pass raw key/value pairs (same contract as slog.With / Logger.With).
func WithKV(ctx context.Context, kv ...any) context.Context {
	l := Logger(ctx).With(kv...)
	return WithLogger(ctx, l)
}

// WithGroup returns a context whose logger is grouped under the given name.
func WithGroup(ctx context.Context, name string) context.Context {
	l := Logger(ctx).WithGroup(name)
	return WithLogger(ctx, l)
}

// Nop returns a logger that discards all output.
func Nop() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

// attrsToAny adapts []slog.Attr to []any for Logger.With(...any).
func attrsToAny(attrs []slog.Attr) []any {
	out := make([]any, len(attrs))
	for i, a := range attrs {
		out[i] = a
	}
	return out
}
