package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mnabil1718/taskflow/internal/model"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *model.Notification) error
	ListByUser(ctx context.Context, userID string, limit int) ([]*model.Notification, error)
	UnreadCount(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, userID, id string) error
	MarkAllRead(ctx context.Context, userID string) error
}

type notificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *model.Notification) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO notifications (user_id, type, task_id, project_id, title, reminder_window)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`,
		n.UserID,
		n.Type,
		n.TaskID,
		n.ProjectID,
		nullableString(n.Title),
		nullableString(n.ReminderWindow),
	).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (r *notificationRepository) ListByUser(ctx context.Context, userID string, limit int) ([]*model.Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, type, task_id, project_id, title, reminder_window, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	out := make([]*model.Notification, 0)
	for rows.Next() {
		n := &model.Notification{}
		var taskID, projectID, title, window sql.NullString
		var readAt sql.NullTime
		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&taskID,
			&projectID,
			&title,
			&window,
			&readAt,
			&n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		if taskID.Valid {
			n.TaskID = &taskID.String
		}
		if projectID.Valid {
			n.ProjectID = &projectID.String
		}
		if title.Valid {
			n.Title = title.String
		}
		if window.Valid {
			n.ReminderWindow = window.String
		}
		if readAt.Valid {
			n.ReadAt = &readAt.Time
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *notificationRepository) UnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unread count: %w", err)
	}
	return count, nil
}

func (r *notificationRepository) MarkRead(ctx context.Context, userID, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE notifications SET read_at = NOW()
		WHERE id = $1 AND user_id = $2 AND read_at IS NULL
	`, id, userID)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	return nil
}

func (r *notificationRepository) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE notifications SET read_at = NOW()
		WHERE user_id = $1 AND read_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}
