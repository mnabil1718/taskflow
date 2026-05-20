package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/response"
	"github.com/mnabil1718/taskflow/internal/service"
)

type TaskHandler struct {
	svc service.TaskService
}

func NewTaskHandler(svc service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

// TaskPage is the paginated envelope returned by the list endpoint.
type TaskPage struct {
	Items      any `json:"items"`
	Total      int `json:"total"       example:"42"`
	Page       int `json:"page"        example:"1"`
	Limit      int `json:"limit"       example:"10"`
	TotalPages int `json:"total_pages" example:"5"`
}

// Create godoc
// @Summary      Create a task in a project
// @Description  Creates a task scoped to a project. The caller must be a project member.
// @Description  If assignee_id is set, the assignee must also be a project member.
// @Description  Priority defaults to "medium" when omitted; status is always "todo" on creation.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id      path     string                                  true "Project UUID"
// @Param        request body     model.CreateTaskRequest                 true "Task payload"
// @Success      201     {object} response.Body{data=model.Task}          "Task created"
// @Failure      400     {object} response.Body                           "Validation error, invalid priority, or assignee not a project member"
// @Failure      401     {object} response.Body                           "Missing or invalid token"
// @Failure      404     {object} response.Body                           "Project not found or caller is not a member"
// @Failure      500     {object} response.Body                           "Internal server error"
// @Security     BearerAuth
// @Router       /projects/{id}/tasks [post]
func (h *TaskHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req model.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	task, err := h.svc.Create(c.Context(), userID, c.Params("id"), &req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "task created", task)
}

// List godoc
// @Summary      List tasks in a project
// @Description  Returns a paginated list of tasks in the project. The caller must be a project member.
// @Description  Supports filtering by status, priority, and assignee, and sorting by a whitelisted column.
// @Tags         tasks
// @Produce      json
// @Param        id           path     string                                  true  "Project UUID"
// @Param        status       query    string                                  false "Filter by status: todo, in_progress, done"
// @Param        priority     query    string                                  false "Filter by priority: low, medium, high"
// @Param        assignee_id  query    string                                  false "Filter by assignee UUID"
// @Param        sort_by      query    string                                  false "Sort column: status, priority, assignee_id, due_date, updated_at, title, created_at (default created_at)"
// @Param        sort_order   query    string                                  false "Sort direction: asc or desc (default desc)"
// @Param        page         query    int                                     false "Page number (default 1)"
// @Param        limit        query    int                                     false "Items per page, max 100 (default 10)"
// @Success      200          {object} response.Body{data=handler.TaskPage}    "Tasks retrieved"
// @Failure      400          {object} response.Body                           "Invalid filter value"
// @Failure      401          {object} response.Body                           "Missing or invalid token"
// @Failure      404          {object} response.Body                           "Project not found or caller is not a member"
// @Failure      500          {object} response.Body                           "Internal server error"
// @Security     BearerAuth
// @Router       /projects/{id}/tasks [get]
func (h *TaskHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	filter := model.TaskFilter{
		Status:     model.TaskStatus(c.Query("status")),
		Priority:   model.TaskPriority(c.Query("priority")),
		AssigneeID: c.Query("assignee_id"),
		SortBy:     c.Query("sort_by"),
		SortOrder:  c.Query("sort_order"),
		Page:       page,
		Limit:      limit,
	}

	tasks, total, err := h.svc.List(c.Context(), userID, c.Params("id"), filter)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	return response.Success(c, fiber.StatusOK, "tasks retrieved", TaskPage{
		Items:      tasks,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// GetByID godoc
// @Summary      Get a task
// @Description  Returns a single task by ID. The caller must be a member of the task's project.
// @Tags         tasks
// @Produce      json
// @Param        taskID path     string                          true "Task UUID"
// @Success      200    {object} response.Body{data=model.Task}  "Task retrieved"
// @Failure      401    {object} response.Body                   "Missing or invalid token"
// @Failure      404    {object} response.Body                   "Task not found or caller is not a project member"
// @Failure      500    {object} response.Body                   "Internal server error"
// @Security     BearerAuth
// @Router       /tasks/{taskID} [get]
func (h *TaskHandler) GetByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	task, err := h.svc.GetByID(c.Context(), userID, c.Params("taskID"))
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "task retrieved", task)
}

// Update godoc
// @Summary      Update a task
// @Description  Replaces title, description, status, priority, assignee, and due_date.
// @Description  The caller must be a project member. If assignee_id is set, the assignee must also be a project member.
// @Description  A status change is recorded automatically in the task activity log.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        taskID  path     string                                  true "Task UUID"
// @Param        request body     model.UpdateTaskRequest                 true "Updated task fields"
// @Success      200     {object} response.Body{data=model.Task}          "Task updated"
// @Failure      400     {object} response.Body                           "Validation error, invalid enum, or assignee not a project member"
// @Failure      401     {object} response.Body                           "Missing or invalid token"
// @Failure      404     {object} response.Body                           "Task not found or caller is not a project member"
// @Failure      500     {object} response.Body                           "Internal server error"
// @Security     BearerAuth
// @Router       /tasks/{taskID} [put]
func (h *TaskHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req model.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	task, err := h.svc.Update(c.Context(), userID, c.Params("taskID"), &req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "task updated", task)
}

// Delete godoc
// @Summary      Delete a task
// @Description  Permanently deletes a task. Only the project owner or the task creator can delete.
// @Tags         tasks
// @Produce      json
// @Param        taskID path     string       true "Task UUID"
// @Success      200    {object} response.Body "Task deleted"
// @Failure      401    {object} response.Body "Missing or invalid token"
// @Failure      403    {object} response.Body "Caller is not the project owner or task creator"
// @Failure      404    {object} response.Body "Task not found or caller is not a project member"
// @Failure      500    {object} response.Body "Internal server error"
// @Security     BearerAuth
// @Router       /tasks/{taskID} [delete]
func (h *TaskHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.svc.Delete(c.Context(), userID, c.Params("taskID")); err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "task deleted", nil)
}

// Assign godoc
// @Summary      Assign or unassign a task
// @Description  Updates only the assignee of a task. The caller must be a project member.
// @Description  Set assignee_id to a project member UUID to assign, or to null to unassign.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        taskID  path     string                                  true "Task UUID"
// @Param        request body     model.AssignTaskRequest                 true "Assignee payload (null to unassign)"
// @Success      200     {object} response.Body{data=model.Task}          "Task assignee updated"
// @Failure      400     {object} response.Body                           "Malformed body or assignee not a project member"
// @Failure      401     {object} response.Body                           "Missing or invalid token"
// @Failure      404     {object} response.Body                           "Task not found or caller is not a project member"
// @Failure      500     {object} response.Body                           "Internal server error"
// @Security     BearerAuth
// @Router       /tasks/{taskID}/assign [patch]
func (h *TaskHandler) Assign(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req model.AssignTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	task, err := h.svc.Assign(c.Context(), userID, c.Params("taskID"), &req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "task assignee updated", task)
}

// GetActivityLogs godoc
// @Summary      List a task's status-change history
// @Description  Returns the task's status-transition log in reverse chronological order.
// @Description  The caller must be a member of the task's project.
// @Tags         tasks
// @Produce      json
// @Param        taskID path     string                                          true "Task UUID"
// @Success      200    {object} response.Body{data=[]model.TaskActivityLog}     "Activity logs retrieved"
// @Failure      401    {object} response.Body                                   "Missing or invalid token"
// @Failure      404    {object} response.Body                                   "Task not found or caller is not a project member"
// @Failure      500    {object} response.Body                                   "Internal server error"
// @Security     BearerAuth
// @Router       /tasks/{taskID}/activity [get]
func (h *TaskHandler) GetActivityLogs(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	logs, err := h.svc.GetActivityLogs(c.Context(), userID, c.Params("taskID"))
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "activity logs retrieved", logs)
}

func (h *TaskHandler) handleServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrValidation):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrAssigneeNotMember):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrTaskNotFound):
		return response.Error(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrProjectNotFound):
		return response.Error(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrForbidden):
		return response.Error(c, fiber.StatusForbidden, err.Error())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}
}
