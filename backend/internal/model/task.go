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
	ID             string       `json:"id"                       example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	Title          string       `json:"title"                    example:"Wire up auth middleware"`
	Description    string       `json:"description,omitempty"    example:"Plug JWTProtected into the v1 router"`
	Status         TaskStatus   `json:"status"                   example:"todo"`
	Priority       TaskPriority `json:"priority"                 example:"medium"`
	Position       string       `json:"position"                 example:"00001000"`
	ProjectID      string       `json:"project_id"               example:"c303012a-6275-4aa3-adec-ebfb123f4567"`
	AssigneeID     *string      `json:"assignee_id,omitempty"    example:"f02c1d9c-1f73-4d3a-9b8c-aab0cf2d12cd"`
	AssigneeName   *string      `json:"assignee_name,omitempty"  example:"Jane Doe"`
	AssigneeEmail  *string      `json:"assignee_email,omitempty" example:"jane@example.com"`
	CreatedBy      *string      `json:"created_by,omitempty"     example:"6b3a0c0e-2cc1-4f3c-8d9c-1a1b2c3d4e5f"`
	DueDate        *time.Time   `json:"due_date,omitempty"       example:"2026-06-01T17:00:00Z"`
	CreatedAt      time.Time    `json:"created_at"               example:"2026-05-20T09:15:00Z"`
	UpdatedAt      time.Time    `json:"updated_at"               example:"2026-05-20T09:15:00Z"`
}

type TaskActivityLog struct {
	ID            string     `json:"id"                       example:"5b5a7c2c-1d3e-4d3a-9b8c-aab0cf2d99aa"`
	TaskID        string     `json:"task_id"                  example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	ChangedBy     *string    `json:"changed_by,omitempty"     example:"6b3a0c0e-2cc1-4f3c-8d9c-1a1b2c3d4e5f"`
	ChangedByName *string    `json:"changed_by_name,omitempty" example:"Jane Doe"`
	FromStatus    TaskStatus `json:"from_status"              example:"todo"`
	ToStatus      TaskStatus `json:"to_status"                example:"in_progress"`
	CreatedAt     time.Time  `json:"created_at"               example:"2026-05-20T10:00:00Z"`
}

// TaskActivityLogPage is the paginated envelope returned by the activity
// logs endpoint. HasMore lets the client know whether another "Load more"
// click would yield results without having to fetch and discover empty.
type TaskActivityLogPage struct {
	Items   []*TaskActivityLog `json:"items"`
	HasMore bool               `json:"has_more"  example:"true"`
}

// TaskActivityLogFilter narrows a paginated activity-log read.
// Before is a cursor — pass the CreatedAt of the oldest entry you have
// to fetch the next page.
type TaskActivityLogFilter struct {
	Before *time.Time
	Limit  int
}

type CreateTaskRequest struct {
	Title       string       `json:"title"              example:"Wire up auth middleware"`
	Description string       `json:"description"        example:"Plug JWTProtected into the v1 router"`
	Priority    TaskPriority `json:"priority"           example:"high"`
	// Optional Lexorank position. Send one only when creating from the Kanban
	// board and you need exact placement; the data-table create flow can omit
	// it and the server will default to end-of-column.
	Position   string     `json:"position,omitempty" example:"00009000"`
	AssigneeID *string    `json:"assignee_id"        example:"f02c1d9c-1f73-4d3a-9b8c-aab0cf2d12cd"`
	DueDate    *time.Time `json:"due_date"           example:"2026-06-01T17:00:00Z"`
}

type UpdateTaskRequest struct {
	Title       string       `json:"title"        example:"Wire up auth middleware (revised)"`
	Description string       `json:"description"  example:"Also cover refresh-token flow"`
	Status      TaskStatus   `json:"status"       example:"in_progress"`
	Priority    TaskPriority `json:"priority"     example:"high"`
	AssigneeID  *string      `json:"assignee_id"  example:"f02c1d9c-1f73-4d3a-9b8c-aab0cf2d12cd"`
	DueDate     *time.Time   `json:"due_date"     example:"2026-06-05T17:00:00Z"`
}

type TaskFilter struct {
	Statuses   []TaskStatus
	Priorities []TaskPriority
	AssigneeID string
	Search     string
	SortBy     string
	SortOrder  string
	Page       int
	Limit      int
}

type BulkDeleteTasksRequest struct {
	IDs []string `json:"ids"`
}

type BulkDeleteTasksResponse struct {
	DeletedCount int `json:"deleted_count"`
}

type AssignTaskRequest struct {
	AssigneeID *string `json:"assignee_id" example:"f02c1d9c-1f73-4d3a-9b8c-aab0cf2d12cd"`
}

// UpdateTaskStatusRequest is the payload for the member-allowed status
// endpoint. Splitting it out from UpdateTaskRequest keeps the RBAC rule
// "members can only change status" enforceable at the API surface — the
// endpoint can't accept a title or assignee field even if the client tries.
type UpdateTaskStatusRequest struct {
	Status TaskStatus `json:"status" example:"in_progress"`
}

// MoveTaskRequest is the payload sent by the Kanban board on drag & drop.
// The client computes the new position string (Lexorank) so the server stays
// a dumb persistence layer for ordering. Status may equal the current status
// (in-column reorder) or differ (cross-column move).
type MoveTaskRequest struct {
	Status   TaskStatus `json:"status"   example:"in_progress"`
	Position string     `json:"position" example:"0000h000"`
}

// BoardView is the response shape for the Kanban board endpoint:
// tasks bucketed by status, each bucket already sorted by position.
type BoardView struct {
	Todo       []*Task `json:"todo"`
	InProgress []*Task `json:"in_progress"`
	Done       []*Task `json:"done"`
}
