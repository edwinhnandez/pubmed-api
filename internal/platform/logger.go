package platform

import (
	"log/slog"
	"os"
)

// NewLogger creates a new structured logger
func NewLogger(level string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.Level(GetLogLevel(level)),
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

