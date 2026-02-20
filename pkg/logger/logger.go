package logger

import (
	"context"
	"log/slog"
	"os"
)

const (
	LoggerKey   = "logger"
	RequestID   = "requestID"
	ServiceName = "service"
)

type Logger interface {
	Info(ctx context.Context, msg string, keysAndValues ...any)
	Warn(ctx context.Context, msg string, keysAndValues ...any)
	Error(ctx context.Context, msg string, keysAndValues ...any)
	Debug(ctx context.Context, msg string, keysAndValues ...any)
	With(args ...any) Logger
}

type logger struct {
	log *slog.Logger
}

func (l *logger) With(args ...any) Logger {
	return &logger{log: l.log.With(args...)}
}

func (l *logger) Info(ctx context.Context, msg string, keysAndValues ...any) {
	l.log.InfoContext(ctx, msg, keysAndValues...)
}

func (l *logger) Warn(ctx context.Context, msg string, keysAndValues ...any) {
	l.log.WarnContext(ctx, msg, keysAndValues...)
}

func (l *logger) Error(ctx context.Context, msg string, keysAndValues ...any) {
	l.log.ErrorContext(ctx, msg, keysAndValues...)
}

func (l *logger) Debug(ctx context.Context, msg string, keysAndValues ...any) {
	l.log.DebugContext(ctx, msg, keysAndValues...)
}

func New(serviceName string) Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	log = log.With(ServiceName, serviceName)
	return &logger{log: log}
}

func GetLoggerFromCtx(ctx context.Context) Logger {
	if lg, ok := ctx.Value(LoggerKey).(Logger); ok {
		return lg
	}
	return New("default")
}
