package systemsetting_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"tex/src/internal/handler/systemsetting"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSystemSettingUseCase struct {
	mock.Mock
}

func (m *mockSystemSettingUseCase) IsPresentationUploadEnabled(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *mockSystemSettingUseCase) SetPresentationUploadEnabled(ctx context.Context, enabled bool) error {
	args := m.Called(ctx, enabled)
	return args.Error(0)
}

func TestGetSettings_Success(t *testing.T) {
	mockUC := new(mockSystemSettingUseCase)
	h := systemsetting.NewHandler(mockUC)

	app := fiber.New()
	app.Get("/settings", h.GetSettings)

	mockUC.On("IsPresentationUploadEnabled", mock.Anything).Return(true, nil)

	req := httptest.NewRequest("GET", "/settings", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockUC.AssertExpectations(t)
}

func TestGetSettings_Error(t *testing.T) {
	mockUC := new(mockSystemSettingUseCase)
	h := systemsetting.NewHandler(mockUC)

	app := fiber.New()
	app.Get("/settings", h.GetSettings)

	mockUC.On("IsPresentationUploadEnabled", mock.Anything).Return(false, errors.New("db error"))

	req := httptest.NewRequest("GET", "/settings", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestUpdateSettings_Success(t *testing.T) {
	mockUC := new(mockSystemSettingUseCase)
	h := systemsetting.NewHandler(mockUC)

	app := fiber.New()
	app.Put("/settings", h.UpdatePresentationUploadEnabled)

	body, _ := json.Marshal(map[string]bool{"enabled": true})
	mockUC.On("SetPresentationUploadEnabled", mock.Anything, true).Return(nil)

	req := httptest.NewRequest("PUT", "/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockUC.AssertExpectations(t)
}

func TestUpdateSettings_BadRequest(t *testing.T) {
	mockUC := new(mockSystemSettingUseCase)
	h := systemsetting.NewHandler(mockUC)

	app := fiber.New()
	app.Put("/settings", h.UpdatePresentationUploadEnabled)

	req := httptest.NewRequest("PUT", "/settings", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
