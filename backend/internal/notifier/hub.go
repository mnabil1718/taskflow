package notifier

import (
	"sync"
	"time"

	"github.com/mnabil1718/taskflow/internal/model"
)

type EventType string

const (
	EventTaskCreated          EventType = "task.created"
	EventTaskUpdated          EventType = "task.updated"
	EventTaskDeleted          EventType = "task.deleted"
	EventTaskAssigned         EventType = "task.assigned"
	EventTaskMoved            EventType = "task.moved"
	EventTaskDeadlineReminder EventType = "task.deadline_reminder"
)

type Event struct {
	Type           EventType   `json:"type"                     example:"task.created"`
	TaskID         string      `json:"task_id"                  example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	ProjectID      string      `json:"project_id"               example:"c303012a-6275-4aa3-adec-ebfb123f4567"`
	Task           *model.Task `json:"task,omitempty"`
	ReminderWindow string      `json:"reminder_window,omitempty" example:"3d"`
	Timestamp      time.Time   `json:"timestamp"                example:"2026-05-20T16:00:00Z"`
}

type Subscription struct {
	ch chan Event
}

func (s *Subscription) Events() <-chan Event { return s.ch }

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Subscription]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*Subscription]struct{})}
}

// Subscribe registers a new subscription for userID. The returned unsubscribe
// func must be called when the consumer is done (e.g. SSE connection closed).
func (h *Hub) Subscribe(userID string) (*Subscription, func()) {
	sub := &Subscription{ch: make(chan Event, 16)}

	h.mu.Lock()
	if _, ok := h.clients[userID]; !ok {
		h.clients[userID] = make(map[*Subscription]struct{})
	}
	h.clients[userID][sub] = struct{}{}
	h.mu.Unlock()

	return sub, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if subs, ok := h.clients[userID]; ok {
			if _, exists := subs[sub]; exists {
				delete(subs, sub)
				close(sub.ch)
				if len(subs) == 0 {
					delete(h.clients, userID)
				}
			}
		}
	}
}

// Publish delivers event to every subscription for each user in userIDs.
// Delivery is best-effort: if a subscriber's buffer is full, the event is
// dropped for that subscriber rather than blocking the publisher.
func (h *Hub) Publish(userIDs []string, event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range userIDs {
		subs, ok := h.clients[uid]
		if !ok {
			continue
		}
		for sub := range subs {
			select {
			case sub.ch <- event:
			default:
			}
		}
	}
}
