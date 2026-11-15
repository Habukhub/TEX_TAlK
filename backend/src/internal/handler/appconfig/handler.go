package appconfig

import (
	"tex/src/internal/configs"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	conf *configs.Config
}

func NewHandler(conf *configs.Config) *Handler {
	return &Handler{conf: conf}
}

// GetConfig godoc
// @Summary Get public app configuration
// @Description Returns runtime feature flags for the frontend
// @Tags config
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/config [get]
func (h *Handler) GetConfig(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"schedule_polling_enabled": h.conf.App.SchedulePollingEnabled,
	})
}
