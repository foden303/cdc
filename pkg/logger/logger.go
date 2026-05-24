package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"

	"github.com/foden/cdc/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// loggerKey is the context key for storing a logger instance.
type loggerKey struct{}

// globalLogger holds the current logger instance, swappable at runtime.
var globalLogger atomic.Pointer[slog.Logger]

// levelVar allows dynamic log level changes at runtime.
var levelVar slog.LevelVar

func init() {
	// Initialize with a default text logger at info level.
	levelVar.Set(slog.LevelInfo)
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: &levelVar}))
	globalLogger.Store(l)
}

// Init configures the global logger based on the provided LogConfig.
// It sets up the handler mode (json/text), log level, and optional file rotation.
func Init(cfg config.LogConfig) {
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	if cfg.FilePath != "" {
		maxSize := cfg.MaxSizeMB
		if maxSize <= 0 {
			maxSize = 100
		}
		maxBackups := cfg.MaxBackups
		if maxBackups <= 0 {
			maxBackups = 5
		}
		writers = append(writers, &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			Compress:   false,
		})
	}

	w := io.MultiWriter(writers...)

	levelVar.Set(parseLevel(cfg.Level))

	opts := &slog.HandlerOptions{
		Level:     &levelVar,
		AddSource: strings.EqualFold(cfg.Level, "debug"),
	}

	var handler slog.Handler
	if strings.EqualFold(cfg.Mode, "json") {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	l := slog.New(handler)
	globalLogger.Store(l)
	slog.SetDefault(l)
}

// L returns the global logger instance.
func L() *slog.Logger {
	return globalLogger.Load()
}

// From extracts a logger from the context. If no logger is stored in the context,
// it returns the global logger.
func From(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return L()
}

// With creates a child logger with additional structured fields.
// Useful for component-scoped loggers (e.g., flow_id, source_table).
func With(args ...any) *slog.Logger {
	return L().With(args...)
}

// WithContext stores a logger in the context for downstream extraction via From().
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// SetLevel dynamically changes the log level at runtime.
// Accepts: "debug", "info", "warn", "error" (case-insensitive).
func SetLevel(level string) {
	levelVar.Set(parseLevel(level))
}

// LogAttrs is a convenience wrapper for hot-path logging that avoids allocations
// by using slog.LogAttrs directly.
func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	From(ctx).LogAttrs(ctx, level, msg, attrs...)
}

// parseLevel converts a string level name to slog.Level.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
