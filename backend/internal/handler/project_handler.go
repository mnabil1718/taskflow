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

// ProjectPage is the paginated envelope returned by the list endpoint.
type ProjectPage struct {
	Items      any `json:"items"`
	Total      int `json:"total"      example:"42"`
	Page       int `json:"page"       example:"1"`
	Limit      int `json:"limit"      example:"10"`
	TotalPages int `json:"total_pages" example:"5"`
}

// Create godoc
// @Summary      Create a new project
// @Description  Creates a project and automatically makes the caller the owner and first member.
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        request body     model.CreateProjectRequest                   true "Project payload"
// @Success      201     {object} response.Body{data=model.Project}            "Project created"
// @Failure      400     {object} response.Body                                "Validation error or malformed body"
// @Failure      401     {object} response.Body                                "Missing or invalid token"
// @Failure      500     {object} response.Body                                "Internal server error"
// @Security     BearerAuth
// @Router       /projects [post]
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

// List godoc
// @Summary      List projects
// @Description  Returns a paginated list of projects the caller owns or is a member of.
// @Tags         projects
// @Produce      json
// @Param        page  query    int                                          false "Page number (default 1)"
// @Param        limit query    int                                          false "Items per page, max 100 (default 10)"
// @Success      200   {object} response.Body{data=handler.ProjectPage}     "Projects retrieved"
// @Failure      401   {object} response.Body                               "Missing or invalid token"
// @Failure      500   {object} response.Body                               "Internal server error"
// @Security     BearerAuth
// @Router       /projects [get]
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

	return response.Success(c, fiber.StatusOK, "projects retrieved", ProjectPage{
		Items:      projects,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// GetByID godoc
// @Summary      Get a project
// @Description  Returns a single project by ID. Only accessible to project members.
// @Tags         projects
// @Produce      json
// @Param        id  path     string                                   true "Project UUID"
// @Success      200 {object} response.Body{data=model.Project}        "Project retrieved"
// @Failure      401 {object} response.Body                            "Missing or invalid token"
// @Failure      404 {object} response.Body                            "Project not found or caller is not a member"
// @Failure      500 {object} response.Body                            "Internal server error"
// @Security     BearerAuth
// @Router       /projects/{id} [get]
func (h *ProjectHandler) GetByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	project, err := h.svc.GetByID(c.Context(), userID, c.Params("id"))
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "project retrieved", project)
}

// Update godoc
// @Summary      Update a project
// @Description  Replaces name, description, status, and deadline. Only the owner can update.
// @Description  Status must be "active" or "archived".
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        id      path     string                                    true "Project UUID"
// @Param        request body     model.UpdateProjectRequest                true "Updated project fields"
// @Success      200     {object} response.Body{data=model.Project}         "Project updated"
// @Failure      400     {object} response.Body                             "Validation error or malformed body"
// @Failure      401     {object} response.Body                             "Missing or invalid token"
// @Failure      403     {object} response.Body                             "Caller is not the project owner"
// @Failure      404     {object} response.Body                             "Project not found"
// @Failure      500     {object} response.Body                             "Internal server error"
// @Security     BearerAuth
// @Router       /projects/{id} [put]
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

// Delete godoc
// @Summary      Delete a project
// @Description  Soft-deletes a project. Only the owner can delete.
// @Tags         projects
// @Produce      json
// @Param        id  path     string       true "Project UUID"
// @Success      200 {object} response.Body "Project deleted"
// @Failure      401 {object} response.Body "Missing or invalid token"
// @Failure      403 {object} response.Body "Caller is not the project owner"
// @Failure      404 {object} response.Body "Project not found"
// @Failure      500 {object} response.Body "Internal server error"
// @Security     BearerAuth
// @Router       /projects/{id} [delete]
func (h *ProjectHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.svc.Delete(c.Context(), userID, c.Params("id")); err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "project deleted", nil)
}

// BulkDelete godoc
// @Summary      Bulk soft-delete projects
// @Description  Soft-deletes every project the caller owns from the given list. IDs the caller does not own — or that don't exist — are silently skipped. Returns the number of projects actually deleted.
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        request body     model.BulkDeleteProjectsRequest                       true "IDs to delete"
// @Success      200     {object} response.Body{data=model.BulkDeleteProjectsResponse} "Projects deleted"
// @Failure      400     {object} response.Body                                         "Validation error or malformed body"
// @Failure      401     {object} response.Body                                         "Missing or invalid token"
// @Failure      500     {object} response.Body                                         "Internal server error"
// @Security     BearerAuth
// @Router       /projects/bulk-delete [post]
func (h *ProjectHandler) BulkDelete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req model.BulkDeleteProjectsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	count, err := h.svc.BulkDelete(c.Context(), userID, req.IDs)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "projects deleted", model.BulkDeleteProjectsResponse{
		DeletedCount: count,
	})
}

// AddMember godoc
// @Summary      Add a member to a project
// @Description  Adds a registered user to the project with the given role. Only the owner can add members.
// @Description  Valid roles are "admin" and "member". Defaults to "member" if omitted.
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        id      path     string                                          true "Project UUID"
// @Param        request body     model.AddMemberRequest                          true "Member payload"
// @Success      201     {object} response.Body{data=model.ProjectMember}         "Member added"
// @Failure      400     {object} response.Body                                   "Validation error — missing user_id or invalid role"
// @Failure      401     {object} response.Body                                   "Missing or invalid token"
// @Failure      403     {object} response.Body                                   "Caller is not the project owner"
// @Failure      404     {object} response.Body                                   "Project or user not found"
// @Failure      409     {object} response.Body                                   "User is already a member"
// @Failure      500     {object} response.Body                                   "Internal server error"
// @Security     BearerAuth
// @Router       /projects/{id}/members [post]
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

// RemoveMember godoc
// @Summary      Remove a member from a project
// @Description  Removes the specified user from the project. Only the owner can remove members.
// @Description  The owner cannot remove themselves.
// @Tags         projects
// @Produce      json
// @Param        id     path     string       true "Project UUID"
// @Param        userID path     string       true "UUID of the user to remove"
// @Success      200    {object} response.Body "Member removed"
// @Failure      400    {object} response.Body "Owner cannot remove themselves"
// @Failure      401    {object} response.Body "Missing or invalid token"
// @Failure      403    {object} response.Body "Caller is not the project owner"
// @Failure      404    {object} response.Body "Project or member not found"
// @Failure      500    {object} response.Body "Internal server error"
// @Security     BearerAuth
// @Router       /projects/{id}/members/{userID} [delete]
func (h *ProjectHandler) RemoveMember(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.svc.RemoveMember(c.Context(), userID, c.Params("id"), c.Params("userID")); err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "member removed", nil)
}

// GetMembers godoc
// @Summary      List project members
// @Description  Returns all members of the project. Accessible to any project member.
// @Tags         projects
// @Produce      json
// @Param        id  path     string                                              true "Project UUID"
// @Success      200 {object} response.Body{data=[]model.ProjectMember}           "Members retrieved"
// @Failure      401 {object} response.Body                                       "Missing or invalid token"
// @Failure      404 {object} response.Body                                       "Project not found or caller is not a member"
// @Failure      500 {object} response.Body                                       "Internal server error"
// @Security     BearerAuth
// @Router       /projects/{id}/members [get]
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
