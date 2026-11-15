package submission

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"tex/src/internal/dto"
	"tex/src/internal/infrastructure/firebase"
	"tex/src/internal/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	submissionUC  usecase.SubmissionUseCase
	systemSetting usecase.SystemSettingUseCase
	firebaser     *firebase.StorageService
}

func NewHandler(submissionUC usecase.SubmissionUseCase, systemSetting usecase.SystemSettingUseCase, firebaser *firebase.StorageService) *Handler {
	return &Handler{submissionUC: submissionUC, systemSetting: systemSetting, firebaser: firebaser}
}

// Submit godoc
// @Summary Submit Application
// @Description Submit a new application
// @Tags submission
// @Accept  json
// @Produce  json
// @Param request body dto.SubmitRequest true "Submit Request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Bad Request (e.g., missing required fields)"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/v1/submissions [post]
func (h *Handler) Submit(c fiber.Ctx) error {
	var req dto.SubmitRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload. Please ensure all required fields are provided correctly."})
	}
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.submissionUC.Submit(c.Context(), req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to submit the application. Please try again later."})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "submitted successfully"})
}

// GetPending godoc
// @Summary Get Pending Submissions
// @Description Get a list of pending submissions (Admin only)
// @Tags submission
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/v1/admin/submissions/pending [get]
func (h *Handler) GetPending(c fiber.Ctx) error {
	submissions, err := h.submissionUC.GetPending(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve the list of pending submissions."})
	}
	return c.JSON(fiber.Map{"data": submissions})
}

// GetApproved godoc
// @Summary Get Approved Submissions
// @Description Get a list of approved submissions (Admin only)
// @Tags submission
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/v1/admin/submissions/approved [get]
func (h *Handler) GetApproved(c fiber.Ctx) error {
	submissions, err := h.submissionUC.GetApproved(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve the list of approved submissions."})
	}
	return c.JSON(fiber.Map{"data": submissions})
}

// GetPublicApproved godoc
// @Summary Get Public Approved Submissions
// @Description Get a list of approved submissions for the public schedule (no sensitive data)
// @Tags submission
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/v1/speakers [get]
func (h *Handler) GetPublicApproved(c fiber.Ctx) error {
	submissions, err := h.submissionUC.GetApproved(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve the list of approved speakers."})
	}

	var publicSpeakers []dto.PublicSubmissionResponse
	for _, sub := range submissions {
		publicSpeakers = append(publicSpeakers, dto.PublicSubmissionResponse{
			ID:            sub.ID.String(),
			FullName:      sub.FullName,
			TopicTitle:    sub.TopicTitle,
			Status:        string(sub.Status),
			SlideFileName: sub.SlideFileName,
		})
	}

	if publicSpeakers == nil {
		publicSpeakers = []dto.PublicSubmissionResponse{}
	}

	return c.JSON(fiber.Map{"data": publicSpeakers})
}

// Approve godoc
// @Summary Approve Submission
// @Description Approve a submission by ID (Admin only)
// @Tags submission
// @Accept  json
// @Produce  json
// @Param id path string true "Submission ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Bad Request (e.g., invalid UUID format)"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/v1/admin/submissions/{id}/approve [put]
func (h *Handler) Approve(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "The provided submission ID is not a valid UUID format."})
	}

	if err := h.submissionUC.Approve(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to approve the submission."})
	}
	return c.JSON(fiber.Map{"message": "approved"})
}

// Reject godoc
// @Summary Reject Submission
// @Description Reject a submission by ID (Admin only)
// @Tags submission
// @Accept  json
// @Produce  json
// @Param id path string true "Submission ID"
// @Success 204
// @Failure 400 {object} map[string]interface{} "Bad Request (e.g., invalid UUID format)"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/v1/admin/submissions/{id} [delete]
func (h *Handler) Reject(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "The provided submission ID is not a valid UUID format."})
	}

	if err := h.submissionUC.Reject(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reject the submission."})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateInfo godoc
// @Summary Update Submission Info
// @Description Update the applicant name and topic title of a submission (Admin only)
// @Tags submission
// @Accept  json
// @Produce  json
// @Param id path string true "Submission ID"
// @Param request body dto.UpdateSubmissionInfoRequest true "Update Info Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/v1/admin/submissions/{id}/info [put]
func (h *Handler) UpdateInfo(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "The provided submission ID is not a valid UUID format."})
	}

	var req dto.UpdateSubmissionInfoRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload."})
	}
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.submissionUC.UpdateInfo(c.Context(), id, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update submission info."})
	}
	return c.JSON(fiber.Map{"message": "submission info updated successfully"})
}

// UploadSlide godoc
// @Summary Upload Presentation Slide
// @Description Speaker uploads their presentation slide (PDF/PPTX) to the backend, which proxies it to Firebase.
// @Tags submission
// @Accept  multipart/form-data
// @Produce  json
// @Param id path string true "Submission ID"
// @Param file formData file true "Slide file"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/submissions/{id}/slide [put]
func (h *Handler) UploadSlide(c fiber.Ctx) error {
	// Security Check: Verify if upload is currently enabled
	enabled, err := h.systemSetting.IsPresentationUploadEnabled(c.Context())
	if err != nil {
		log.Printf("Error checking upload setting: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify upload permissions."})
	}
	if !enabled {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Presentation upload is currently closed."})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid submission ID."})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to get file from request."})
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("Error opening file: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file."})
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read file."})
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileName := fileHeader.Filename
	ext := strings.ToLower(filepath.Ext(fileName))
	allowedExts := map[string]bool{".pdf": true, ".ppt": true, ".pptx": true}
	if !allowedExts[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Only PDF, PPT, and PPTX files are allowed."})
	}
	objectPath := fmt.Sprintf("slides_textalk/%s%s", id.String(), ext)

	// Get existing submission to check for old slide
	existing, err := h.submissionUC.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Submission not found."})
	}

	if h.firebaser == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Firebase storage is not configured."})
	}

	if err := h.firebaser.UploadFile(c.Context(), objectPath, fileBytes, contentType); err != nil {
		log.Printf("Error uploading to Firebase: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upload file to storage."})
	}

	if err := h.submissionUC.UpdateSlide(c.Context(), id, objectPath, fileName); err != nil {
		log.Printf("Error updating DB: %v\n", err)
		if errors.Is(err, usecase.ErrSubmissionNotApproved) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Submission is not approved."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update slide info in database."})
	}

	// If update success, and old file exists with DIFFERENT path (e.g. extension change), delete old file
	if existing.SlideURL != nil && *existing.SlideURL != "" && *existing.SlideURL != objectPath {
		log.Printf("Deleting old slide file: %s\n", *existing.SlideURL)
		_ = h.firebaser.DeleteFile(context.Background(), *existing.SlideURL)
	}

	return c.JSON(fiber.Map{"message": "slide uploaded successfully"})
}

// DownloadSlide godoc
// @Summary Download a submission's slide (Admin only)
// @Tags submission
// @Produce  application/octet-stream
// @Param id path string true "Submission ID"
// @Success 200
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security ApiKeyAuth
// @Router /api/v1/admin/submissions/{id}/slide [get]
func (h *Handler) DownloadSlide(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid submission ID."})
	}

	sub, err := h.submissionUC.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Submission not found."})
	}
	if sub.SlideURL == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No slide uploaded for this submission."})
	}
	if h.firebaser == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Firebase storage is not configured."})
	}

	fileData, err := h.firebaser.DownloadFile(c.Context(), *sub.SlideURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to download file from storage."})
	}

	fileName := "slide"
	if sub.SlideFileName != nil && *sub.SlideFileName != "" {
		fileName = fmt.Sprintf("%s_%s", sub.FullName, *sub.SlideFileName)
	}

	c.Set("Content-Type", "application/octet-stream")
	// Use filename* for UTF-8 support (RFC 6266)
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, fileName, url.PathEscape(fileName)))
	return c.Send(fileData)
}

// DownloadZip godoc
// @Summary Download ZIP of selected approved slides (Admin only)
// @Tags submission
// @Accept  json
// @Produce  application/zip
// @Param request body dto.DownloadZipRequest true "List of Submission IDs"
// @Success 200
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security ApiKeyAuth
// @Router /api/v1/admin/submissions/slides/zip [post]
func (h *Handler) DownloadZip(c fiber.Ctx) error {
	if h.firebaser == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Firebase storage is not configured."})
	}

	var req dto.DownloadZipRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body."})
	}
	if len(req.IDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No IDs provided."})
	}
	if h.firebaser == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Firebase storage is not configured."})
	}

	var validIDs []uuid.UUID
	for _, rawID := range req.IDs {
		id, err := uuid.Parse(rawID)
		if err != nil {
			continue
		}
		validIDs = append(validIDs, id)
	}

	subs, err := h.submissionUC.GetByIDs(c.Context(), validIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve submissions."})
	}

	entries := make(map[string]string) // zipName -> GCS objectPath
	for _, sub := range subs {
		if sub.SlideURL == nil {
			continue
		}
		fileName := fmt.Sprintf("%s_%s", sub.FullName, *sub.SlideFileName)
		entries[fileName] = *sub.SlideURL
	}

	if len(entries) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No slides found for the provided IDs."})
	}

	zipData, err := h.firebaser.BuildZip(c.Context(), entries)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to build ZIP archive."})
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", "attachment; filename=\"slides.zip\"")
	return c.Send(zipData)
}
