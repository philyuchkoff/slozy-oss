// Package logger provides structured logging for the application.
package logger

import (
	"log/slog"
	"os"
)

var defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// SetDefault sets the default logger.
func SetDefault(l *slog.Logger) {
	defaultLogger = l
}

// Info logs at Info level.
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs at Warn level.
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs at Error level.
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Fatal logs at Error level and exits with code 1.
func Fatal(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}
