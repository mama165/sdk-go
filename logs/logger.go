package logs

import (
	"log/slog"
	"os"
)

// GetLoggerFromString Build a logger
// Fallback as INFO by default
func GetLoggerFromString(logLevel string) *slog.Logger {
	var level slog.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		level = slog.LevelInfo
	}
	return GetLoggerFromLevel(level)
}

// GetLoggerFromLevel Build a logger
// Fallback as INFO by default
func GetLoggerFromLevel(logLevel slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	return slog.New(handler)
}
