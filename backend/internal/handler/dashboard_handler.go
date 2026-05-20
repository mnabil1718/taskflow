package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/response"
	"github.com/mnabil1718/taskflow/internal/service"
)

type DashboardHandler struct {
	svc service.DashboardService
}

func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) ProjectTaskCounts(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	counts, err := h.svc.ProjectTaskCounts(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}

	return response.Success(c, fiber.StatusOK, "project task counts retrieved", counts)
}

func (h *DashboardHandler) UpcomingTasks(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	tasks, err := h.svc.UpcomingTasks(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}

	return response.Success(c, fiber.StatusOK, "upcoming tasks retrieved", tasks)
}
