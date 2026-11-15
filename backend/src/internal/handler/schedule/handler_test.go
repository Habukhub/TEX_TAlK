package schedule_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"tex/src/internal/domain"
	"tex/src/internal/dto"
	schedulehandler "tex/src/internal/handler/schedule"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock UseCase ---
type mockScheduleUseCase struct {
	mock.Mock
}

func (m *mockScheduleUseCase) GetTimeline(ctx context.Context) ([]domain.Schedule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Schedule), args.Error(1)
}
func (m *mockScheduleUseCase) AddSlot(ctx context.Context, req dto.ScheduleSlotRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}
func (m *mockScheduleUseCase) DeleteSlot(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// --- Tests ---

func TestGetTimelineHandler_Success(t *testing.T) {
	mockUC := new(mockScheduleUseCase)
	h := schedulehandler.NewHandler(mockUC)

	app := fiber.New()
	app.Get("/api/v1/schedule", h.GetTimeline)

	expected := []domain.Schedule{
		{ID: uuid.New(), ScheduleType: domain.ScheduleTypeSpeaker, StartTime: "09:00", EndTime: "10:00"},
	}
	mockUC.On("GetTimeline", mock.Anything).Return(expected, nil)

	req := httptest.NewRequest("GET", "/api/v1/schedule", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetTimelineHandler_InternalError(t *testing.T) {
	mockUC := new(mockScheduleUseCase)
	h := schedulehandler.NewHandler(mockUC)

	app := fiber.New()
	app.Get("/api/v1/schedule", h.GetTimeline)

	mockUC.On("GetTimeline", mock.Anything).Return([]domain.Schedule{}, errors.New("db error"))

	req := httptest.NewRequest("GET", "/api/v1/schedule", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestAddSlotHandler_Success(t *testing.T) {
	mockUC := new(mockScheduleUseCase)
	h := schedulehandler.NewHandler(mockUC)

	app := fiber.New()
	app.Post("/api/v1/admin/schedule", h.AddSlot)

	submissionID := uuid.New()
	body, _ := json.Marshal(dto.ScheduleSlotRequest{
		SubmissionID: &submissionID,
		ScheduleType: "SPEAKER",
		StartTime:    "09:00",
		EndTime:      "10:00",
	})

	mockUC.On("AddSlot", mock.Anything, mock.AnythingOfType("dto.ScheduleSlotRequest")).Return(nil)

	req := httptest.NewRequest("POST", "/api/v1/admin/schedule", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestAddSlotHandler_BadRequest(t *testing.T) {
	mockUC := new(mockScheduleUseCase)
	h := schedulehandler.NewHandler(mockUC)

	app := fiber.New()
	app.Post("/api/v1/admin/schedule", h.AddSlot)

	req := httptest.NewRequest("POST", "/api/v1/admin/schedule", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDeleteSlotHandler_Success(t *testing.T) {
	mockUC := new(mockScheduleUseCase)
	h := schedulehandler.NewHandler(mockUC)

	app := fiber.New()
	app.Delete("/api/v1/admin/schedule/:id", h.DeleteSlot)

	id := uuid.New()
	mockUC.On("DeleteSlot", mock.Anything, id).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/admin/schedule/"+id.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestDeleteSlotHandler_InvalidUUID(t *testing.T) {
	mockUC := new(mockScheduleUseCase)
	h := schedulehandler.NewHandler(mockUC)

	app := fiber.New()
	app.Delete("/api/v1/admin/schedule/:id", h.DeleteSlot)

	req := httptest.NewRequest("DELETE", "/api/v1/admin/schedule/not-a-uuid", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDeleteSlotHandler_NotFound(t *testing.T) {
	mockUC := new(mockScheduleUseCase)
	h := schedulehandler.NewHandler(mockUC)

	app := fiber.New()
	app.Delete("/api/v1/admin/schedule/:id", h.DeleteSlot)

	id := uuid.New()
	mockUC.On("DeleteSlot", mock.Anything, id).Return(errors.New("schedule slot not found"))

	req := httptest.NewRequest("DELETE", "/api/v1/admin/schedule/"+id.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
