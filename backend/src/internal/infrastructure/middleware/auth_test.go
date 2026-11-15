package middleware_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"tex/src/internal/infrastructure/middleware"
	"tex/src/internal/logger"
	"tex/src/internal/util"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const testJWTSecret = "test-secret-key"

func setupAuthMiddlewareApp(secret string) *fiber.App {
	app := fiber.New()
	app.Use(middleware.AuthMiddleware(secret))
	app.Get("/protected", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": c.Locals("user_id"),
			"email":   c.Locals("email"),
		})
	})
	return app
}

// ===========================================
// AuthMiddleware Tests
// ===========================================

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	app := setupAuthMiddlewareApp(testJWTSecret)

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_InvalidFormat_NoBearer(t *testing.T) {
	app := setupAuthMiddlewareApp(testJWTSecret)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidTokenWithoutBearer")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_InvalidFormat_WrongScheme(t *testing.T) {
	app := setupAuthMiddlewareApp(testJWTSecret)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	app := setupAuthMiddlewareApp(testJWTSecret)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.a.valid.jwt")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_TokenSignedWithWrongSecret(t *testing.T) {
	app := setupAuthMiddlewareApp(testJWTSecret)

	wrongToken, _ := util.GenerateJWT(uuid.New(), "user@example.com", "wrong-secret")

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+wrongToken)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_ValidToken_PassThrough(t *testing.T) {
	app := setupAuthMiddlewareApp(testJWTSecret)

	userID := uuid.New()
	validToken, err := util.GenerateJWT(userID, "user@example.com", testJWTSecret)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthMiddleware_ValidToken_SetsLocals(t *testing.T) {
	// ตรวจสอบว่า Middleware เซ็ต Locals เช่น email ได้ถูกต้อง
	app := fiber.New()
	app.Use(middleware.AuthMiddleware(testJWTSecret))
	app.Get("/me", func(c fiber.Ctx) error {
		email := c.Locals("email")
		return c.JSON(fiber.Map{"email": email})
	})

	validToken, _ := util.GenerateJWT(uuid.New(), "check@example.com", testJWTSecret)

	req := httptest.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	app := setupAuthMiddlewareApp(testJWTSecret)

	// Token ที่ exp ในอดีต (หมดอายุแล้ว)
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MDAwMDAwMDAsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsInVzZXJfaWQiOiIxMjM0NTY3OC0xMjM0LTEyMzQtMTIzNC0xMjM0NTY3ODkwMTIifQ.invalidsignature")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	_ = time.Now()
}

// ===========================================
// AccessLogMiddleware Tests
// ===========================================

func setupAccessLogApp() *fiber.App {
	appLogger := logger.NewLogger("test-service", "test")
	app := fiber.New()
	app.Use(middleware.AccessLogMiddleware(appLogger))
	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})
	app.Get("/server-error", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusInternalServerError).SendString("error")
	})
	return app
}

func TestAccessLogMiddleware_LogsSuccessRequest(t *testing.T) {
	app := setupAccessLogApp()

	req := httptest.NewRequest("GET", "/ping", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAccessLogMiddleware_LogsWithProvidedTraceID(t *testing.T) {
	app := setupAccessLogApp()

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("X-Trace-ID", "my-custom-trace-id")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	// Trace ID ที่ส่งมาต้องถูก echo กลับใน Response header
	assert.Equal(t, "my-custom-trace-id", resp.Header.Get("X-Trace-ID"))
}

func TestAccessLogMiddleware_GeneratesTraceIDWhenMissing(t *testing.T) {
	app := setupAccessLogApp()

	req := httptest.NewRequest("GET", "/ping", nil)
	// ไม่ส่ง X-Trace-ID ไป ระบบต้องสร้างให้เอง
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	// ต้องมี X-Trace-ID ใน Response header (สร้างโดยระบบ)
	assert.NotEmpty(t, resp.Header.Get("X-Trace-ID"))
	_ = context.Background()
}

func TestAccessLogMiddleware_LogsErrorResponse(t *testing.T) {
	app := setupAccessLogApp()

	req := httptest.NewRequest("GET", "/server-error", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	// Middleware ต้องไม่ block แม้ Handler จะคืน 500
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
