package model

import "time"

type TaskStatus string
type TaskPriority string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"

	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

type Task struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	ProjectID   string       `json:"project_id"`
	AssigneeID  *string      `json:"assignee_id,omitempty"`
	CreatedBy   *string      `json:"created_by,omitempty"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type TaskActivityLog struct {
	ID         string     `json:"id"`
	TaskID     string     `json:"task_id"`
	ChangedBy  *string    `json:"changed_by,omitempty"`
	FromStatus TaskStatus `json:"from_status"`
	ToStatus   TaskStatus `json:"to_status"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateTaskRequest struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Priority    TaskPriority `json:"priority"`
	AssigneeID  *string      `json:"assignee_id"`
	DueDate     *time.Time   `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	AssigneeID  *string      `json:"assignee_id"`
	DueDate     *time.Time   `json:"due_date"`
}

type TaskFilter struct {
	Status     TaskStatus
	Priority   TaskPriority
	AssigneeID string
	Page       int
	Limit      int
}
