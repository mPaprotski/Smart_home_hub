// Package logging produces allowlisted JSON logs. Raw configuration/payloads are not logged.
package logging

import (
	"io"
	"log/slog"
)

func New(w io.Writer, level string) *slog.Logger {
	levels := map[string]slog.Level{"DEBUG": slog.LevelDebug, "INFO": slog.LevelInfo, "WARN": slog.LevelWarn, "ERROR": slog.LevelError}
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: levels[level]}))
}
