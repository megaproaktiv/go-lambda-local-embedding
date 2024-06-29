package hugoembedding

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

const (
	LevelDebug = slog.Level(-4)
	LevelInfo  = slog.Level(0)
	LevelWarn  = slog.Level(4)
	LevelError = slog.Level(8)
)

func init() {
	handler := slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{Level: LevelDebug})
	Logger = slog.New(handler)
}
