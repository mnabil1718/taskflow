package notifier

import (
	"sync"

	"github.com/mnabil1718/taskflow/internal/model"
)

// Hub is the in-memory fan-out for live notifications. It is pure SSE
// transport: persistence and the decision of *who* gets *what* live in the
// service layer, which calls Publish after writing the row. The streamed
// payload is the same model.Notification shape the REST list returns, so a
// client can dedupe a live item against the fetched list by id.

type Subscription struct {
	ch chan model.Notification
}

func (s *Subscription) Events() <-chan model.Notification { return s.ch }

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
	sub := &Subscription{ch: make(chan model.Notification, 16)}

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

// Publish delivers n to every subscription for each user in userIDs.
// Delivery is best-effort: if a subscriber's buffer is full, the event is
// dropped for that subscriber rather than blocking the publisher (the row is
// already persisted, so the client still sees it on next list/refresh).
func (h *Hub) Publish(userIDs []string, n model.Notification) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range userIDs {
		subs, ok := h.clients[uid]
		if !ok {
			continue
		}
		for sub := range subs {
			select {
			case sub.ch <- n:
			default:
			}
		}
	}
}
