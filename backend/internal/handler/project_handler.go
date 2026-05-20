package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/response"
	"github.com/mnabil1718/taskflow/internal/service"
)

type ProjectHandler struct {
	svc service.ProjectService
}

func NewProjectHandler(svc service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

type paginatedResponse struct {
	Items      any `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

func (h *ProjectHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req model.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	project, err := h.svc.Create(c.Context(), userID, &req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "project created", project)
}

func (h *ProjectHandler) GetByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	project, err := h.svc.GetByID(c.Context(), userID, c.Params("id"))
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "project retrieved", project)
}

func (h *ProjectHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	projects, total, err := h.svc.List(c.Context(), userID, page, limit)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	return response.Success(c, fiber.StatusOK, "projects retrieved", paginatedResponse{
		Items:      projects,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *ProjectHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req model.UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	project, err := h.svc.Update(c.Context(), userID, c.Params("id"), &req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "project updated", project)
}

func (h *ProjectHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.svc.Delete(c.Context(), userID, c.Params("id")); err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "project deleted", nil)
}

func (h *ProjectHandler) AddMember(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req model.AddMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	member, err := h.svc.AddMember(c.Context(), userID, c.Params("id"), &req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "member added", member)
}

func (h *ProjectHandler) RemoveMember(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.svc.RemoveMember(c.Context(), userID, c.Params("id"), c.Params("userID")); err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "member removed", nil)
}

func (h *ProjectHandler) GetMembers(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	members, err := h.svc.GetMembers(c.Context(), userID, c.Params("id"))
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "members retrieved", members)
}

func (h *ProjectHandler) handleServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrValidation):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrProjectNotFound):
		return response.Error(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrForbidden):
		return response.Error(c, fiber.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrAlreadyMember):
		return response.Error(c, fiber.StatusConflict, err.Error())
	case errors.Is(err, service.ErrMemberNotFound):
		return response.Error(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		return response.Error(c, fiber.StatusNotFound, err.Error())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}
}
