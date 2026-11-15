package middleware

import (
	"context"
	"time"

	"tex/src/internal/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func AccessLogMiddleware(appLogger *logger.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		startedAt := time.Now()
		traceID := c.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}

		c.Set("X-Trace-ID", traceID)

		err := c.Next()

		ctx := logger.WithTraceID(context.Background(), traceID)
		appLogger.InfoContext(
			ctx,
			"request completed",
			"trace.id", traceID,
			"method", c.Method(),
			"path", c.OriginalURL(),
			"status", c.Response().StatusCode(),
			"latency_ms", time.Since(startedAt).Milliseconds(),
			"client_ip", c.IP(),
		)

		return err
	}
}
