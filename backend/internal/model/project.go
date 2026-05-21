package model

import "time"

type ProjectStatus string
type ProjectRole string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"

	ProjectRoleOwner  ProjectRole = "owner"
	ProjectRoleAdmin  ProjectRole = "admin"
	ProjectRoleMember ProjectRole = "member"
)

type Project struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Status      ProjectStatus `json:"status"`
	Deadline    *time.Time    `json:"deadline,omitempty"`
	OwnerID     string        `json:"owner_id"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type ProjectMember struct {
	ProjectID string      `json:"project_id"`
	UserID    string      `json:"user_id"`
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	Role      ProjectRole `json:"role"`
	JoinedAt  time.Time   `json:"joined_at"`
}

type CreateProjectRequest struct {
	Name        string                 `json:"name"        example:"Q4 Roadmap"`
	Description string                 `json:"description" example:"All tasks for Q4 planning"`
	Deadline    *time.Time             `json:"deadline"    example:"2026-12-31T23:59:59Z"`
	Members     []ProjectMemberInvite  `json:"members"     example:"[]"`
}

// ProjectMemberInvite identifies a user that should be added to a project at
// creation time, alongside the role they should hold.
type ProjectMemberInvite struct {
	UserID string      `json:"user_id" example:"c303012a-6275-4aa3-adec-ebfb123f4567"`
	Role   ProjectRole `json:"role"    example:"member"`
}

type UpdateProjectRequest struct {
	Name        string        `json:"name"        example:"Q4 Roadmap (revised)"`
	Description string        `json:"description" example:"Updated scope for Q4"`
	Status      ProjectStatus `json:"status"      example:"archived"`
	Deadline    *time.Time    `json:"deadline"    example:"2026-12-31T23:59:59Z"`
}

type AddMemberRequest struct {
	UserID string      `json:"user_id" example:"c303012a-6275-4aa3-adec-ebfb123f4567"`
	Role   ProjectRole `json:"role"    example:"member"`
}

type BulkDeleteProjectsRequest struct {
	IDs []string `json:"ids" example:"c303012a-6275-4aa3-adec-ebfb123f4567,d404123b-7386-5bb4-bcfe-fcfc234e5678"`
}

type BulkDeleteProjectsResponse struct {
	DeletedCount int `json:"deleted_count" example:"3"`
}
