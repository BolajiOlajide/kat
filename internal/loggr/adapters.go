package loggr

import (
	"context"
	"log"
	"log/slog"
)

// StdLogger adapts Go's standard log package
type StdLogger struct {
	logger *log.Logger
}

func NewStdLogger(logger *log.Logger) *StdLogger {
	return &StdLogger{logger: logger}
}

func (l *StdLogger) Debug(msg string, keyvals ...any) {
	l.logger.Printf("[DEBUG] %s %v", msg, keyvals)
}

func (l *StdLogger) Info(msg string, keyvals ...any) {
	l.logger.Printf("[INFO] %s %v", msg, keyvals)
}

func (l *StdLogger) Warn(msg string, keyvals ...any) {
	l.logger.Printf("[WARN] %s %v", msg, keyvals)
}

func (l *StdLogger) Error(msg string, keyvals ...any) {
	l.logger.Printf("[ERROR] %s %v", msg, keyvals)
}

func (l *StdLogger) With(keyvals ...any) Logger {
	return l // Standard logger doesn't support structured logging
}

func (l *StdLogger) WithContext(ctx context.Context) Logger {
	return l // Standard logger doesn't support context
}

// SlogLogger adapts Go's slog package
type SlogLogger struct {
	logger *slog.Logger
}

func NewSlogLogger(logger *slog.Logger) *SlogLogger {
	return &SlogLogger{logger: logger}
}

func (l *SlogLogger) Debug(msg string, keyvals ...any) {
	l.logger.Debug(msg, keyvals...)
}

func (l *SlogLogger) Info(msg string, keyvals ...any) {
	l.logger.Info(msg, keyvals...)
}

func (l *SlogLogger) Warn(msg string, keyvals ...any) {
	l.logger.Warn(msg, keyvals...)
}

func (l *SlogLogger) Error(msg string, keyvals ...any) {
	l.logger.Error(msg, keyvals...)
}

func (l *SlogLogger) With(keyvals ...any) Logger {
	return &SlogLogger{logger: l.logger.With(keyvals...)}
}

func (l *SlogLogger) WithContext(ctx context.Context) Logger {
	return &SlogLogger{logger: l.logger.With(slog.Any("context", ctx))}
}
