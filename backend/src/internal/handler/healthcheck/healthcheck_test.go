package healthcheck_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"tex/src/internal/handler/healthcheck"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func setupHealthcheckApp() *fiber.App {
	app := fiber.New()
	h := healthcheck.NewHandler()
	app.Get("/api/v1/healthz", h.GetHealthz)
	return app
}

func TestGetHealthz_ReturnsOK(t *testing.T) {
	app := setupHealthcheckApp()

	req := httptest.NewRequest("GET", "/api/v1/healthz", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetHealthz_ReturnsVersionField(t *testing.T) {
	app := setupHealthcheckApp()

	req := httptest.NewRequest("GET", "/api/v1/healthz", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)
	// ต้องมีฟิลด์ version ใน Response
	_, exists := body["version"]
	assert.True(t, exists)
}
