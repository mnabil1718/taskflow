package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/notifier"
	"github.com/mnabil1718/taskflow/internal/response"
	"github.com/mnabil1718/taskflow/internal/service"
	"github.com/valyala/fasthttp"
)

const sseHeartbeatInterval = 15 * time.Second

type NotificationHandler struct {
	svc service.NotificationService
	hub *notifier.Hub
}

func NewNotificationHandler(svc service.NotificationService, hub *notifier.Hub) *NotificationHandler {
	return &NotificationHandler{svc: svc, hub: hub}
}

// List godoc
// @Summary      List the caller's notifications
// @Description  Returns the caller's most-recent notifications (newest first) plus the
// @Description  count of unread ones for the bell badge. Notifications are created only
// @Description  for: a task assigned to the caller, and task/project deadline reminders
// @Description  (3 days then 1 day before the deadline).
// @Tags         notifications
// @Produce      json
// @Param        limit query int false "Max items to return (default 50, max 100)"
// @Success      200 {object} response.Body{data=model.NotificationPage} "Notifications retrieved"
// @Failure      401 {object} response.Body "Missing or invalid token"
// @Failure      500 {object} response.Body "Internal server error"
// @Security     BearerAuth
// @Router       /notifications [get]
func (h *NotificationHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	page, err := h.svc.List(c.Context(), userID, c.QueryInt("limit", 0))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}
	return response.Success(c, fiber.StatusOK, "notifications retrieved", page)
}

// MarkRead godoc
// @Summary      Mark one notification as read
// @Tags         notifications
// @Produce      json
// @Param        id path string true "Notification ID"
// @Success      200 {object} response.Body "Notification marked read"
// @Failure      401 {object} response.Body "Missing or invalid token"
// @Failure      500 {object} response.Body "Internal server error"
// @Security     BearerAuth
// @Router       /notifications/{id}/read [post]
func (h *NotificationHandler) MarkRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.svc.MarkRead(c.Context(), userID, c.Params("id")); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}
	return response.Success(c, fiber.StatusOK, "notification marked read", nil)
}

// MarkAllRead godoc
// @Summary      Mark all the caller's notifications as read
// @Tags         notifications
// @Produce      json
// @Success      200 {object} response.Body "Notifications marked read"
// @Failure      401 {object} response.Body "Missing or invalid token"
// @Failure      500 {object} response.Body "Internal server error"
// @Security     BearerAuth
// @Router       /notifications/read-all [post]
func (h *NotificationHandler) MarkAllRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.svc.MarkAllRead(c.Context(), userID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}
	return response.Success(c, fiber.StatusOK, "notifications marked read", nil)
}

// Stream godoc
// @Summary      Stream notifications (Server-Sent Events)
// @Description  Opens a long-lived SSE connection that pushes the caller's notifications
// @Description  live as they are created. Each pushed item is the same model.Notification
// @Description  shape as the list endpoint (matched by id), so a client merges live items
// @Description  into the fetched list without duplicates.
// @Description
// @Description  Notification types:
// @Description    - task.assigned             — a task was assigned to the caller
// @Description    - task.deadline_reminder    — a task assigned to the caller is approaching its due date
// @Description    - project.deadline_reminder — a project the caller is a member of is approaching its deadline
// @Description
// @Description  Deadline reminders fire in two windows (3 days then 1 day before); the
// @Description  `reminder_window` field is "3d" or "1d" accordingly.
// @Description
// @Description  Wire format: text/event-stream. Each event arrives as two lines followed
// @Description  by a blank line:
// @Description    event: <type>
// @Description    data:  <json-encoded model.Notification>
// @Description
// @Description  A `: ping` comment is sent every 15 seconds to keep the connection alive.
// @Tags         notifications
// @Produce      text/event-stream
// @Success      200 {object} model.Notification "JSON payload carried in each event's data: line"
// @Failure      401 {object} response.Body      "Missing or invalid token"
// @Security     BearerAuth
// @Router       /notifications/stream [get]
func (h *NotificationHandler) Stream(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	sub, unsubscribe := h.hub.Subscribe(userID)

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		defer unsubscribe()

		if _, err := fmt.Fprint(w, ": connected\n\n"); err != nil {
			return
		}
		if err := w.Flush(); err != nil {
			return
		}

		ticker := time.NewTicker(sseHeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case n, ok := <-sub.Events():
				if !ok {
					return
				}
				payload, err := json.Marshal(n)
				if err != nil {
					continue
				}
				if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", n.Type, payload); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			case <-ticker.C:
				if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	}))

	return nil
}
