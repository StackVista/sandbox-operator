package logr

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

type logKeyMarker struct{}

var logKey = &logKeyMarker{}

func WithContext(ctx context.Context, logger logr.Logger) context.Context {
	return context.WithValue(ctx, logKey, logger)
}

func Ctx(ctx context.Context) logr.Logger {
	l, ok := ctx.Value(logKey).(logr.Logger)
	if !ok || l == nil {
		return zapr.NewLogger(zap.NewNop())
	}

	return l
}
