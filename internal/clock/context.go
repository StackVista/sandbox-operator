package clock

import (
	"context"

	clk "github.com/benbjohnson/clock"
)

type clockKeyMarker struct{}

var clockKey = &clockKeyMarker{}

func Ctx(ctx context.Context) clk.Clock {
	c, ok := ctx.Value(clockKey).(clk.Clock)
	if !ok || c == nil {
		return clk.New()
	}

	return c
}

func WithContext(ctx context.Context, c clk.Clock) context.Context {
	return context.WithValue(ctx, clockKey, c)
}
