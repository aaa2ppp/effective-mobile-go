package logger

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// legacy wrapper for slog
type legacy slog.Logger

func Legacy(log *slog.Logger) *legacy {
	return (*legacy)(log)
}

func (log *legacy) Fatalf(f string, v ...any) {
	(*slog.Logger)(log).Log(context.Background(), slog.LevelError+2, fmt.Sprintf(strings.TrimSuffix(f, "\n"), v...))
}

func (log *legacy) Printf(f string, v ...any) {
	(*slog.Logger)(log).Log(context.Background(), slog.LevelInfo, fmt.Sprintf(strings.TrimSuffix(f, "\n"), v...))
}
