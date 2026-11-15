package logger

import (
	"log/slog"
	"os"
	"time"
)

type Logger struct {
	*slog.Logger
}

func NewLogger(serviceName, environment string) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				return slog.String("@timestamp", time.Now().UTC().Format(time.RFC3339Nano))
			}
			return attr
		},
	})

	return &Logger{
		Logger: slog.New(handler).With(
			"service.name", serviceName,
			"environment", environment,
		),
	}
}
