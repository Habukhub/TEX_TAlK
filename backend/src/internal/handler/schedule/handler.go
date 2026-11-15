package schedule

import (
	"tex/src/internal/dto"
	"tex/src/internal/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	scheduleUC usecase.ScheduleUseCase
}

func NewHandler(scheduleUC usecase.ScheduleUseCase) *Handler {
	return &Handler{scheduleUC: scheduleUC}
}

// GetTimeline godoc
// @Summary Get Schedule Timeline
// @Description Get the list of all schedule timeline slots
// @Tags schedule
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/v1/schedule [get]
func (h *Handler) GetTimeline(c fiber.Ctx) error {
	timeline, err := h.scheduleUC.GetTimeline(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve the schedule timeline."})
	}
	return c.JSON(fiber.Map{"data": timeline})
}

// AddSlot godoc
// @Summary Add Schedule Slot
// @Description Add a new schedule slot (Admin only)
// @Tags schedule
// @Accept  json
// @Produce  json
// @Param request body dto.ScheduleSlotRequest true "Schedule Slot Request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Bad Request (e.g., invalid JSON format or missing fields)"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/v1/admin/schedule [post]
func (h *Handler) AddSlot(c fiber.Ctx) error {
	var req dto.ScheduleSlotRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.scheduleUC.AddSlot(c.Context(), req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "slot added"})
}

// DeleteSlot godoc
// @Summary Delete Schedule Slot
// @Description Delete a schedule slot by ID (Admin only)
// @Tags schedule
// @Accept  json
// @Produce  json
// @Param id path string true "Schedule Slot ID"
// @Success 204
// @Failure 400 {object} map[string]interface{} "Bad Request (e.g., invalid UUID format)"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/v1/admin/schedule/{id} [delete]
func (h *Handler) DeleteSlot(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "The provided schedule ID is not a valid UUID format."})
	}

	if err := h.scheduleUC.DeleteSlot(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete the schedule slot."})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
