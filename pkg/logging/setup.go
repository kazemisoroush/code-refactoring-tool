// Package logging provides centralized logging configuration for the application.
package logging

import (
	"log/slog"
	"os"
	"strings"
)

// SetupLogger configures slog with JSON output and the specified log level
func SetupLogger(level string) {
	var logLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Create JSON handler with specified log level
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	// Set as default logger
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
