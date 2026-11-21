package platform

import (
	"log/slog"
	"os"
)

// NewLogger returns JSON logs in production (machine-readable for log tools)
// and human-readable text logs in development.
func NewLogger(production bool) *slog.Logger {
	if production {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}