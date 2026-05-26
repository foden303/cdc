package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/foden/cdc/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// levelVar allows dynamic log level changes at runtime.
var levelVar slog.LevelVar

func init() {
	levelVar.Set(slog.LevelInfo)
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
	slog.SetDefault(l)
}

// SetLevel dynamically changes the log level at runtime.
// Accepts: "debug", "info", "warn", "error" (case-insensitive).
func SetLevel(level string) {
	levelVar.Set(parseLevel(level))
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
