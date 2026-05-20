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
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Deadline    *time.Time `json:"deadline"`
}

type UpdateProjectRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      ProjectStatus `json:"status"`
	Deadline    *time.Time    `json:"deadline"`
}

type AddMemberRequest struct {
	UserID string      `json:"user_id"`
	Role   ProjectRole `json:"role"`
}
