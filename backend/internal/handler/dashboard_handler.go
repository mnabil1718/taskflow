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

// ProjectTaskCounts godoc
// @Summary      Per-project task counts
// @Description  Returns a count of tasks grouped by status (todo, in_progress, done) for every
// @Description  project the caller is a member of, plus the total task count per project.
// @Description  Projects with no tasks are still included with zeroed counters.
// @Tags         dashboard
// @Produce      json
// @Success      200 {object} response.Body{data=[]model.ProjectTaskCounts} "Project task counts retrieved"
// @Failure      401 {object} response.Body                                 "Missing or invalid token"
// @Failure      500 {object} response.Body                                 "Internal server error"
// @Security     BearerAuth
// @Router       /dashboard/project-task-counts [get]
func (h *DashboardHandler) ProjectTaskCounts(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	counts, err := h.svc.ProjectTaskCounts(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}

	return response.Success(c, fiber.StatusOK, "project task counts retrieved", counts)
}

// UpcomingTasks godoc
// @Summary      Upcoming tasks assigned to the caller
// @Description  Returns tasks assigned to the caller whose due_date falls within the next 3 days
// @Description  and whose status is not "done". Ordered by soonest due_date first.
// @Tags         dashboard
// @Produce      json
// @Success      200 {object} response.Body{data=[]model.UpcomingTask} "Upcoming tasks retrieved"
// @Failure      401 {object} response.Body                            "Missing or invalid token"
// @Failure      500 {object} response.Body                            "Internal server error"
// @Security     BearerAuth
// @Router       /dashboard/upcoming-tasks [get]
func (h *DashboardHandler) UpcomingTasks(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	tasks, err := h.svc.UpcomingTasks(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}

	return response.Success(c, fiber.StatusOK, "upcoming tasks retrieved", tasks)
}
