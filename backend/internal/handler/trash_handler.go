package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/response"
	"github.com/mnabil1718/taskflow/internal/service"
)

type TrashHandler struct {
	svc service.TrashService
}

func NewTrashHandler(svc service.TrashService) *TrashHandler {
	return &TrashHandler{svc: svc}
}

// List godoc
// @Summary      List trashed items
// @Description  Returns soft-deleted projects (owner-scoped) and soft-deleted
// @Description  tasks (project-owner or task-creator scoped, and only when the
// @Description  task's project itself is still active) in a single feed
// @Description  ordered by deleted_at descending.
// @Tags         trash
// @Produce      json
// @Success      200 {object} response.Body{data=[]model.TrashItem}  "Trash retrieved"
// @Failure      401 {object} response.Body                          "Missing or invalid token"
// @Failure      500 {object} response.Body                          "Internal server error"
// @Security     BearerAuth
// @Router       /trash [get]
func (h *TrashHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	items, err := h.svc.List(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}
	return response.Success(c, fiber.StatusOK, "trash retrieved", items)
}

// Restore godoc
// @Summary      Restore items from trash
// @Description  Clears deleted_at on the listed projects/tasks the caller is
// @Description  allowed to restore. Items missing the right permissions or
// @Description  whose parent project is itself trashed are silently skipped.
// @Tags         trash
// @Accept       json
// @Produce      json
// @Param        request body     model.BulkTrashRequest                          true "IDs to restore"
// @Success      200     {object} response.Body{data=model.BulkTrashResponse}     "Items restored"
// @Failure      400     {object} response.Body                                   "Validation error"
// @Failure      401     {object} response.Body                                   "Missing or invalid token"
// @Failure      500     {object} response.Body                                   "Internal server error"
// @Security     BearerAuth
// @Router       /trash/restore [post]
func (h *TrashHandler) Restore(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req model.BulkTrashRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	resp, err := h.svc.Restore(c.Context(), userID, &req)
	if err != nil {
		return h.handleServiceError(c, err)
	}
	return response.Success(c, fiber.StatusOK, "items restored", resp)
}

// Purge godoc
// @Summary      Permanently delete items from trash
// @Description  Deletes the listed projects/tasks for good. Purging a project
// @Description  cascades to its tasks via the FK.
// @Tags         trash
// @Accept       json
// @Produce      json
// @Param        request body     model.BulkTrashRequest                          true "IDs to purge"
// @Success      200     {object} response.Body{data=model.BulkTrashResponse}     "Items purged"
// @Failure      400     {object} response.Body                                   "Validation error"
// @Failure      401     {object} response.Body                                   "Missing or invalid token"
// @Failure      500     {object} response.Body                                   "Internal server error"
// @Security     BearerAuth
// @Router       /trash/purge [post]
func (h *TrashHandler) Purge(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req model.BulkTrashRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	resp, err := h.svc.Purge(c.Context(), userID, &req)
	if err != nil {
		return h.handleServiceError(c, err)
	}
	return response.Success(c, fiber.StatusOK, "items purged", resp)
}

// EmptyAll godoc
// @Summary      Empty the trash
// @Description  Permanently deletes every trashed project (owner-scoped) and
// @Description  every trashed task (project-owner or task-creator scoped) in
// @Description  one shot. No body required.
// @Tags         trash
// @Produce      json
// @Success      200 {object} response.Body{data=model.BulkTrashResponse}    "Trash emptied"
// @Failure      401 {object} response.Body                                  "Missing or invalid token"
// @Failure      500 {object} response.Body                                  "Internal server error"
// @Security     BearerAuth
// @Router       /trash [delete]
func (h *TrashHandler) EmptyAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	resp, err := h.svc.EmptyAll(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}
	return response.Success(c, fiber.StatusOK, "trash emptied", resp)
}

func (h *TrashHandler) handleServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrValidation):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}
}
