package model

import "time"

// TrashKind distinguishes between the two entity types that can land in the
// unified trash list. Kept as a small string union so it serialises cleanly
// to JSON and reads naturally in URL params.
type TrashKind string

const (
	TrashKindProject TrashKind = "project"
	TrashKindTask    TrashKind = "task"
)

// TrashItem is a single row in the unified trash list. The shape unifies
// soft-deleted projects and soft-deleted tasks so the frontend can render
// them in one table; ProjectID/ProjectName are populated only for tasks,
// and TaskCount only for projects (so the row can advertise how many
// tasks were bundled in via the cascade soft-delete).
type TrashItem struct {
	Kind        TrashKind `json:"kind"                    example:"task"`
	ID          string    `json:"id"                      example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	Title       string    `json:"title"                   example:"Wire up auth middleware"`
	ProjectID   *string   `json:"project_id,omitempty"    example:"c303012a-6275-4aa3-adec-ebfb123f4567"`
	ProjectName *string   `json:"project_name,omitempty"  example:"Auth rework"`
	TaskCount   *int      `json:"task_count,omitempty"    example:"5"`
	DeletedAt   time.Time `json:"deleted_at"              example:"2026-05-23T09:15:00Z"`
}

// BulkTrashRequest is the payload for restore / purge endpoints. Splitting
// the ids by kind avoids a discriminator-per-row and keeps each ids slice
// small enough for an IN-list parameter expansion.
type BulkTrashRequest struct {
	ProjectIDs []string `json:"project_ids"`
	TaskIDs    []string `json:"task_ids"`
}

// BulkTrashResponse reports how many rows were actually affected per kind.
// Items the caller wasn't authorised to touch are silently skipped, so
// these counts can be lower than what was requested.
type BulkTrashResponse struct {
	RestoredProjects int `json:"restored_projects,omitempty"`
	RestoredTasks    int `json:"restored_tasks,omitempty"`
	PurgedProjects   int `json:"purged_projects,omitempty"`
	PurgedTasks      int `json:"purged_tasks,omitempty"`
}
