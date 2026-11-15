package submission_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"tex/src/internal/domain"
	"tex/src/internal/dto"
	submissionhandler "tex/src/internal/handler/submission"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock UseCase ---
type mockSubmissionUseCase struct {
	mock.Mock
}

func (m *mockSubmissionUseCase) Submit(ctx context.Context, req dto.SubmitRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}
func (m *mockSubmissionUseCase) GetPending(ctx context.Context) ([]domain.Submission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Submission), args.Error(1)
}
func (m *mockSubmissionUseCase) GetApproved(ctx context.Context) ([]domain.Submission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Submission), args.Error(1)
}
func (m *mockSubmissionUseCase) Approve(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockSubmissionUseCase) Reject(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockSubmissionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Submission), args.Error(1)
}
func (m *mockSubmissionUseCase) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Submission, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]domain.Submission), args.Error(1)
}
func (m *mockSubmissionUseCase) UpdateSlide(ctx context.Context, id uuid.UUID, slideURL string, slideFileName string) error {
	args := m.Called(ctx, id, slideURL, slideFileName)
	return args.Error(0)
}
func (m *mockSubmissionUseCase) UpdateInfo(ctx context.Context, id uuid.UUID, req dto.UpdateSubmissionInfoRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

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

// --- Tests ---

func TestSubmitHandler_Success(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Post("/api/v1/submissions", h.Submit)

	body, _ := json.Marshal(dto.SubmitRequest{
		FullName:   "Somchai Jaidee",
		Email:      "somchai@example.com",
		TopicTitle: "Test topic",
	})

	mockUC.On("Submit", mock.Anything, mock.AnythingOfType("dto.SubmitRequest")).Return(nil)

	req := httptest.NewRequest("POST", "/api/v1/submissions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	mockUC.AssertExpectations(t)
}

func TestSubmitHandler_BadRequest_InvalidJSON(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Post("/api/v1/submissions", h.Submit)

	req := httptest.NewRequest("POST", "/api/v1/submissions", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	mockUC.AssertNotCalled(t, "Submit")
}

func TestSubmitHandler_InternalServerError(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Post("/api/v1/submissions", h.Submit)

	body, _ := json.Marshal(dto.SubmitRequest{FullName: "Test", Email: "test@e.com", TopicTitle: "Topic"})
	mockUC.On("Submit", mock.Anything, mock.AnythingOfType("dto.SubmitRequest")).Return(errors.New("db error"))

	req := httptest.NewRequest("POST", "/api/v1/submissions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestGetPendingHandler_Success(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Get("/api/v1/admin/submissions/pending", h.GetPending)

	expected := []domain.Submission{{ID: uuid.New(), Status: domain.StatusPending}}
	mockUC.On("GetPending", mock.Anything).Return(expected, nil)

	req := httptest.NewRequest("GET", "/api/v1/admin/submissions/pending", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestApproveHandler_Success(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Put("/api/v1/admin/submissions/:id/approve", h.Approve)

	id := uuid.New()
	mockUC.On("Approve", mock.Anything, id).Return(nil)

	req := httptest.NewRequest("PUT", "/api/v1/admin/submissions/"+id.String()+"/approve", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestApproveHandler_InvalidUUID(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Put("/api/v1/admin/submissions/:id/approve", h.Approve)

	req := httptest.NewRequest("PUT", "/api/v1/admin/submissions/not-a-uuid/approve", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestRejectHandler_Success(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Delete("/api/v1/admin/submissions/:id", h.Reject)

	id := uuid.New()
	mockUC.On("Reject", mock.Anything, id).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/admin/submissions/"+id.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestRejectHandler_InvalidUUID(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Delete("/api/v1/admin/submissions/:id", h.Reject)

	req := httptest.NewRequest("DELETE", "/api/v1/admin/submissions/not-a-uuid", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetApprovedHandler_Success(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Get("/api/v1/admin/submissions/approved", h.GetApproved)

	expected := []domain.Submission{{ID: uuid.New(), Status: domain.StatusApproved}}
	mockUC.On("GetApproved", mock.Anything).Return(expected, nil)

	req := httptest.NewRequest("GET", "/api/v1/admin/submissions/approved", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetApprovedHandler_InternalError(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Get("/api/v1/admin/submissions/approved", h.GetApproved)

	mockUC.On("GetApproved", mock.Anything).Return([]domain.Submission{}, errors.New("db error"))

	req := httptest.NewRequest("GET", "/api/v1/admin/submissions/approved", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestGetPendingHandler_InternalError(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Get("/api/v1/admin/submissions/pending", h.GetPending)

	mockUC.On("GetPending", mock.Anything).Return([]domain.Submission{}, errors.New("db error"))

	req := httptest.NewRequest("GET", "/api/v1/admin/submissions/pending", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestApproveHandler_InternalError(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Put("/api/v1/admin/submissions/:id/approve", h.Approve)

	id := uuid.New()
	mockUC.On("Approve", mock.Anything, id).Return(errors.New("db error"))

	req := httptest.NewRequest("PUT", "/api/v1/admin/submissions/"+id.String()+"/approve", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestRejectHandler_InternalError(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Delete("/api/v1/admin/submissions/:id", h.Reject)

	id := uuid.New()
	mockUC.On("Reject", mock.Anything, id).Return(errors.New("db error"))

	req := httptest.NewRequest("DELETE", "/api/v1/admin/submissions/"+id.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestDownloadSlideHandler_StorageNotConfigured(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Get("/api/v1/admin/submissions/:id/slide", h.DownloadSlide)

	id := uuid.New()
	slideURL := "slides_textalk/example.pdf"
	mockUC.On("GetByID", mock.Anything, id).Return(&domain.Submission{
		ID:       id,
		SlideURL: &slideURL,
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/admin/submissions/"+id.String()+"/slide", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	mockUC.AssertExpectations(t)
}

func TestSubmitHandler_ValidationError_EmptyName(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Post("/api/v1/submissions", h.Submit)

	body, _ := json.Marshal(dto.SubmitRequest{FullName: "", Email: "a@b.com", TopicTitle: "Topic"})
	req := httptest.NewRequest("POST", "/api/v1/submissions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	mockUC.AssertNotCalled(t, "Submit")
}

func TestSubmitHandler_ValidationError_InvalidEmail(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Post("/api/v1/submissions", h.Submit)

	body, _ := json.Marshal(dto.SubmitRequest{FullName: "Name", Email: "not-an-email", TopicTitle: "Topic"})
	req := httptest.NewRequest("POST", "/api/v1/submissions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	mockUC.AssertNotCalled(t, "Submit")
}

func TestSubmitHandler_ValidationError_EmptyTopic(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Post("/api/v1/submissions", h.Submit)

	body, _ := json.Marshal(dto.SubmitRequest{FullName: "Name", Email: "a@b.com", TopicTitle: ""})
	req := httptest.NewRequest("POST", "/api/v1/submissions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	mockUC.AssertNotCalled(t, "Submit")
}

func TestDownloadZipHandler_StorageNotConfigured(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Post("/api/v1/admin/submissions/slides/zip", h.DownloadZip)

	body, _ := json.Marshal(dto.DownloadZipRequest{IDs: []string{uuid.NewString()}})
	req := httptest.NewRequest("POST", "/api/v1/admin/submissions/slides/zip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	mockUC.AssertNotCalled(t, "GetByID")
}

func TestUpdateInfoHandler_Success(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Put("/api/v1/admin/submissions/:id/info", h.UpdateInfo)

	id := uuid.New()
	reqBody := dto.UpdateSubmissionInfoRequest{FullName: "New Name", TopicTitle: "New Topic"}
	mockUC.On("UpdateInfo", mock.Anything, id, reqBody).Return(nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/admin/submissions/"+id.String()+"/info", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockUC.AssertExpectations(t)
}

func TestUpdateInfoHandler_InvalidUUID(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Put("/api/v1/admin/submissions/:id/info", h.UpdateInfo)

	body, _ := json.Marshal(dto.UpdateSubmissionInfoRequest{FullName: "Name", TopicTitle: "Topic"})
	req := httptest.NewRequest("PUT", "/api/v1/admin/submissions/not-a-uuid/info", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUpdateInfoHandler_ValidationError_EmptyFields(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Put("/api/v1/admin/submissions/:id/info", h.UpdateInfo)

	id := uuid.New()
	body, _ := json.Marshal(dto.UpdateSubmissionInfoRequest{FullName: "", TopicTitle: ""})
	req := httptest.NewRequest("PUT", "/api/v1/admin/submissions/"+id.String()+"/info", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	mockUC.AssertNotCalled(t, "UpdateInfo")
}

func TestUpdateInfoHandler_InternalError(t *testing.T) {
	mockUC := new(mockSubmissionUseCase)
	h := submissionhandler.NewHandler(mockUC, new(mockSystemSettingUseCase), nil)

	app := fiber.New()
	app.Put("/api/v1/admin/submissions/:id/info", h.UpdateInfo)

	id := uuid.New()
	reqBody := dto.UpdateSubmissionInfoRequest{FullName: "Name", TopicTitle: "Topic"}
	mockUC.On("UpdateInfo", mock.Anything, id, reqBody).Return(errors.New("db error"))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/admin/submissions/"+id.String()+"/info", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
