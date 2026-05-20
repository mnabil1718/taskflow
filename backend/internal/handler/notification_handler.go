package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/notifier"
	"github.com/valyala/fasthttp"
)

const sseHeartbeatInterval = 15 * time.Second

type NotificationHandler struct {
	hub *notifier.Hub
}

func NewNotificationHandler(hub *notifier.Hub) *NotificationHandler {
	return &NotificationHandler{hub: hub}
}

// Stream godoc
// @Summary      Stream task notifications (Server-Sent Events)
// @Description  Opens a long-lived SSE connection that pushes task events for every
// @Description  project the caller is a member of. Event types: task.created,
// @Description  task.updated, task.deleted, task.assigned. A heartbeat comment line
// @Description  is sent every 15 seconds to keep the connection alive through proxies.
// @Description  The response media type is text/event-stream and each event has the
// @Description  shape: event: <type>\ndata: <json>\n\n.
// @Tags         notifications
// @Produce      text/event-stream
// @Success      200 {string} string "SSE stream of notifier.Event payloads"
// @Failure      401 {object} response.Body "Missing or invalid token"
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
			case ev, ok := <-sub.Events():
				if !ok {
					return
				}
				payload, err := json.Marshal(ev)
				if err != nil {
					continue
				}
				if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, payload); err != nil {
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
