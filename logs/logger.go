package logs

import (
	"bytes"
	"log/slog"
	"os"
)

// GetLevelFromString Initialize a logLevel (default is INFO)
func GetLevelFromString(strLevel string) slog.Level {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(strLevel)); err != nil {
		return slog.LevelInfo
	}
	return logLevel
}

// GetLoggerFromLevel Initialize a logger from a log level
func GetLoggerFromLevel(logLevel slog.Level) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{Level: logLevel}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, handlerOpts))
	slog.SetDefault(logger)
	return logger
}

// GetLoggerFromBufferWithLogger Initialize a logger from a log level & a buffer
func GetLoggerFromBufferWithLogger(buf *bytes.Buffer, logLevel slog.Level) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{Level: logLevel}
	logger := slog.New(slog.NewJSONHandler(buf, handlerOpts))
	slog.SetDefault(logger)
	return logger
}

// GetLoggerFromString Initialize a logger from a string
func GetLoggerFromString(strLevel string) *slog.Logger {
	return GetLoggerFromLevel(GetLevelFromString(strLevel))
}
