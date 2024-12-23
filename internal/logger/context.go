package logger

import (
	"context"
	"log/slog"
)

type contextKey int

func ContextWithLogger(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey(1), log)
}

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	if v := ctx.Value(contextKey(1)); v != nil {
		return v.(*slog.Logger)
	}
	return nil
}

func GetLoggerFromContextOrDefault(ctx context.Context) *slog.Logger {
	if v := ctx.Value(contextKey(1)); v != nil {
		return v.(*slog.Logger)
	}
	return slog.Default()
}
