package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	debug bool
	log   *slog.Logger
}

func New() *Logger {
	debug := os.Getenv("DEBUG") == "true"

	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	return &Logger{
		debug: debug,
		log:   slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})),
	}
}

func (l *Logger) Debug(msg string, args ...any) {
	if l.debug {
		l.log.Debug(msg, args...)
	}
}
