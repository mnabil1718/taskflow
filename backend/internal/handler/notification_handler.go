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
// @Description  Opens a long-lived SSE connection that pushes task events to the
// @Description  caller. The set of events the caller receives depends on the event
// @Description  type — see "Routing" below.
// @Description
// @Description  Event types:
// @Description    - task.created           — a new task was created in a project the caller is a member of
// @Description    - task.updated           — a task in such a project was updated
// @Description    - task.deleted           — a task in such a project was deleted
// @Description    - task.assigned          — a task's assignee was changed (assign or unassign)
// @Description    - task.deadline_reminder — the caller is the assignee of an open task whose due_date is approaching
// @Description
// @Description  Routing:
// @Description    - task.created / updated / deleted / assigned are fanned out to **every member** of the affected project (team feed).
// @Description    - task.deadline_reminder is sent **only to the assignee** of the task.
// @Description
// @Description  Deadline reminders fire in two windows per task: when the deadline
// @Description  is within 3 days, and again when it is within 1 day. The
// @Description  `reminder_window` field on the event payload is "3d" or "1d"
// @Description  accordingly. Each window fires at most once per (task, window) pair;
// @Description  changing the task's due_date or assignee resets the windows so the
// @Description  new schedule / new assignee gets fresh warnings. Done tasks and
// @Description  unassigned tasks never receive reminders.
// @Description
// @Description  Wire format: text/event-stream. Each event arrives as two lines
// @Description  followed by a blank line:
// @Description    event: <type>
// @Description    data:  <json-encoded notifier.Event>
// @Description
// @Description  A `: ping` comment line is sent every 15 seconds to keep the
// @Description  connection alive through proxies. Native browser EventSource does
// @Description  not support custom headers; clients can either use fetch-based
// @Description  streaming or an EventSource polyfill that allows headers.
// @Tags         notifications
// @Produce      text/event-stream
// @Success      200 {object} notifier.Event "Schema of the JSON payload carried in each event's `data:` line. `reminder_window` is present only on task.deadline_reminder events."
// @Failure      401 {object} response.Body  "Missing or invalid token"
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
