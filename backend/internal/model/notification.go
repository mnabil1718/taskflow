package model

import "time"

// Notification types. These are the only kinds the system emits: a task
// being assigned to a user, and the 3-day / 1-day deadline warnings for a
// task or a project. Stored as plain strings (VARCHAR) so adding a kind
// doesn't require an enum migration.
const (
	NotificationTaskAssigned     = "task.assigned"
	NotificationTaskDeadline     = "task.deadline_reminder"
	NotificationProjectDeadline  = "project.deadline_reminder"
)

// Notification is one row of a user's notification feed. It is both the
// persisted record (REST list endpoint) and the live SSE payload, so the
// frontend can dedupe a streamed item against the fetched list by id.
// task_id / project_id let the client deep-link; title is denormalised so
// the feed renders without extra joins or lookups for deleted entities.
type Notification struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Type           string     `json:"type"                       example:"task.assigned"`
	TaskID         *string    `json:"task_id,omitempty"`
	ProjectID      *string    `json:"project_id,omitempty"`
	Title          string     `json:"title,omitempty"`
	ReminderWindow string     `json:"reminder_window,omitempty"  example:"3d"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// NotificationPage is the list response: the most-recent notifications plus
// the unread total so the bell badge doesn't need a second request.
type NotificationPage struct {
	Items       []*Notification `json:"items"`
	UnreadCount int             `json:"unread_count"`
}
