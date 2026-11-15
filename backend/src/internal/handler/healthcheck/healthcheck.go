package healthcheck

import (
	"os"

	"tex/src/internal/constant"
	"tex/src/internal/dto"

	"github.com/gofiber/fiber/v3"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// GetHealthz godoc
// @Summary Check API Health
// @Description get the status of server
// @Tags healthcheck
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.HealthCheckResponse
// @Router /api/v1/healthz [get]
func (h *Handler) GetHealthz(c fiber.Ctx) error {
	return c.JSON(dto.HealthCheckResponse{
		Version: os.Getenv(constant.WordVersion),
	})
}
