package localrepo

import (
	"context"
	"log/slog"

	"effective-mobile-go/internal/logger"
)

const loggerGroup = "localrepo"

type helper struct {
	ctx context.Context
	op  string
	log *slog.Logger
}

func newHelper(ctx context.Context, op string) helper {
	return helper{ctx: ctx, op: op}
}

func (x *helper) Log() *slog.Logger {
	if x.log == nil {
		x.log = logger.GetLoggerFromContextOrDefault(x.ctx).WithGroup(loggerGroup).With("op", x.op)
	}
	return x.log
}
