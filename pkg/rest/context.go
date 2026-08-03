package rest

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	requestIdContextKey contextKey = "request_id"
	loggerContextKey    contextKey = "toolkit_logger"
)

func GetContextLogger(ctx context.Context) *slog.Logger {
	val := ctx.Value(loggerContextKey)
	if log, ok := val.(*slog.Logger); ok && log != nil {
		return log
	}

	return nil
}

func WithContextLogger(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, log)
}

func WithContextRequestId(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, requestIdContextKey, requestId)
}

func RequestIdFromContext(ctx context.Context) (string, bool) {
	val := ctx.Value(requestIdContextKey)
	s, ok := val.(string)
	return s, ok && s != ""
}
