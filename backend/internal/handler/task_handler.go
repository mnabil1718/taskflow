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

type TaskPage struct {
	Items      any `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

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

func (h *TaskHandler) GetByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	task, err := h.svc.GetByID(c.Context(), userID, c.Params("taskID"))
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "task retrieved", task)
}

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

func (h *TaskHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.svc.Delete(c.Context(), userID, c.Params("taskID")); err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "task deleted", nil)
}

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
