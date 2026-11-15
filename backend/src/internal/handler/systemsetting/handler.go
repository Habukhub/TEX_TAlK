package systemsetting

import (
	"tex/src/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	settingUC usecase.SystemSettingUseCase
}

func NewHandler(settingUC usecase.SystemSettingUseCase) *Handler {
	return &Handler{settingUC: settingUC}
}

// GetSettings godoc
// @Summary Get system settings
// @Description Get current public system settings
// @Tags settings
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/settings [get]
func (h *Handler) GetSettings(c fiber.Ctx) error {
	enabled, err := h.settingUC.IsPresentationUploadEnabled(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve settings"})
	}

	return c.JSON(fiber.Map{
		"presentation_upload_enabled": enabled,
	})
}

// UpdatePresentationUploadEnabled godoc
// @Summary Update presentation upload setting
// @Description Toggle the presentation upload feature on or off (Admin only)
// @Tags settings
// @Accept json
// @Produce json
// @Param request body map[string]bool true "Setting value"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security ApiKeyAuth
// @Router /api/v1/admin/settings/presentation_upload_enabled [put]
func (h *Handler) UpdatePresentationUploadEnabled(c fiber.Ctx) error {
	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.settingUC.SetPresentationUploadEnabled(c.Context(), req.Enabled); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update setting"})
	}

	return c.JSON(fiber.Map{"message": "Setting updated successfully"})
}
