package service

import (
	"context"
	"log/slog"

	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/notifier"
	"github.com/mnabil1718/taskflow/internal/repository"
)

const (
	notificationDefaultLimit = 50
	notificationMaxLimit     = 100
)

// NotificationService owns notification creation and reads. The Notify*
// helpers persist a row and then publish it to the live SSE hub, in that
// order, so a streamed item always has a durable backing row (the client
// dedupes by id). Persistence failures are logged and swallowed — a
// notifier hiccup must never break the mutation that triggered it.
type NotificationService interface {
	List(ctx context.Context, userID string, limit int) (*model.NotificationPage, error)
	MarkRead(ctx context.Context, userID, id string) error
	MarkAllRead(ctx context.Context, userID string) error

	NotifyTaskAssigned(ctx context.Context, assigneeID string, t *model.Task)
	NotifyTaskDeadline(ctx context.Context, assigneeID string, t *model.Task, window string)
	NotifyProjectDeadline(ctx context.Context, userID string, p *model.Project, window string)
}

type notificationService struct {
	repo repository.NotificationRepository
	hub  *notifier.Hub
}

func NewNotificationService(repo repository.NotificationRepository, hub *notifier.Hub) NotificationService {
	return &notificationService{repo: repo, hub: hub}
}

func (s *notificationService) List(ctx context.Context, userID string, limit int) (*model.NotificationPage, error) {
	if limit < 1 {
		limit = notificationDefaultLimit
	}
	if limit > notificationMaxLimit {
		limit = notificationMaxLimit
	}

	items, err := s.repo.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	unread, err := s.repo.UnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &model.NotificationPage{Items: items, UnreadCount: unread}, nil
}

func (s *notificationService) MarkRead(ctx context.Context, userID, id string) error {
	return s.repo.MarkRead(ctx, userID, id)
}

func (s *notificationService) MarkAllRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllRead(ctx, userID)
}

func (s *notificationService) NotifyTaskAssigned(ctx context.Context, assigneeID string, t *model.Task) {
	s.emit(ctx, &model.Notification{
		UserID:    assigneeID,
		Type:      model.NotificationTaskAssigned,
		TaskID:    &t.ID,
		ProjectID: &t.ProjectID,
		Title:     t.Title,
	})
}

func (s *notificationService) NotifyTaskDeadline(ctx context.Context, assigneeID string, t *model.Task, window string) {
	s.emit(ctx, &model.Notification{
		UserID:         assigneeID,
		Type:           model.NotificationTaskDeadline,
		TaskID:         &t.ID,
		ProjectID:      &t.ProjectID,
		Title:          t.Title,
		ReminderWindow: window,
	})
}

func (s *notificationService) NotifyProjectDeadline(ctx context.Context, userID string, p *model.Project, window string) {
	s.emit(ctx, &model.Notification{
		UserID:         userID,
		Type:           model.NotificationProjectDeadline,
		ProjectID:      &p.ID,
		Title:          p.Name,
		ReminderWindow: window,
	})
}

// emit persists the notification then pushes it to any live SSE subscriber.
// Publishing only after a successful insert guarantees the streamed item is
// already durable, so a client that later refetches the list sees the same id.
func (s *notificationService) emit(ctx context.Context, n *model.Notification) {
	if err := s.repo.Create(ctx, n); err != nil {
		slog.Error("notification persist", "type", n.Type, "user_id", n.UserID, "error", err)
		return
	}
	s.hub.Publish([]string{n.UserID}, *n)
}
