package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/repository"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrForbidden       = errors.New("access denied")
	ErrAlreadyMember   = errors.New("user is already a member of this project")
	ErrMemberNotFound  = errors.New("member not found")
	ErrUserNotFound    = errors.New("user not found")
)

type ProjectService interface {
	Create(ctx context.Context, ownerID string, req *model.CreateProjectRequest) (*model.Project, error)
	GetByID(ctx context.Context, userID, projectID string) (*model.Project, error)
	List(ctx context.Context, userID string, page, limit int) ([]*model.Project, int, error)
	Update(ctx context.Context, userID, projectID string, req *model.UpdateProjectRequest) (*model.Project, error)
	Delete(ctx context.Context, userID, projectID string) error
	BulkDelete(ctx context.Context, userID string, ids []string) (int, error)
	AddMember(ctx context.Context, ownerID, projectID string, req *model.AddMemberRequest) (*model.ProjectMember, error)
	RemoveMember(ctx context.Context, ownerID, projectID, targetUserID string) error
	GetMembers(ctx context.Context, userID, projectID string) ([]*model.ProjectMember, error)
}

const bulkDeleteMaxIDs = 100

type projectService struct {
	projectRepo repository.ProjectRepository
	userRepo    repository.UserRepository
}

func NewProjectService(projectRepo repository.ProjectRepository, userRepo repository.UserRepository) ProjectService {
	return &projectService{projectRepo: projectRepo, userRepo: userRepo}
}

func (s *projectService) Create(ctx context.Context, ownerID string, req *model.CreateProjectRequest) (*model.Project, error) {
	if err := validateCreateProject(req); err != nil {
		return nil, err
	}

	invites, err := s.normalizeInvites(ctx, ownerID, req.Members)
	if err != nil {
		return nil, err
	}

	p := &model.Project{
		Name:        req.Name,
		Description: req.Description,
		Status:      model.ProjectStatusActive,
		Deadline:    req.Deadline,
		OwnerID:     ownerID,
	}

	if err := s.projectRepo.Create(ctx, p, invites); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return p, nil
}

// normalizeInvites validates each invite (role defaults to "member",
// duplicate user_ids collapse, owner is filtered out, and every referenced
// user must exist) so the repo layer can trust the slice it receives.
func (s *projectService) normalizeInvites(ctx context.Context, ownerID string, raw []model.ProjectMemberInvite) ([]model.ProjectMemberInvite, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(raw))
	out := make([]model.ProjectMemberInvite, 0, len(raw))

	for _, inv := range raw {
		if inv.UserID == "" {
			return nil, fmt.Errorf("%w: invite user_id is required", ErrValidation)
		}
		if inv.UserID == ownerID {
			continue
		}
		if _, dup := seen[inv.UserID]; dup {
			continue
		}
		seen[inv.UserID] = struct{}{}

		role := inv.Role
		if role == "" {
			role = model.ProjectRoleMember
		}
		if role != model.ProjectRoleAdmin && role != model.ProjectRoleMember {
			return nil, fmt.Errorf("%w: invite role must be 'admin' or 'member'", ErrValidation)
		}

		if _, err := s.userRepo.GetByID(ctx, inv.UserID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}

		out = append(out, model.ProjectMemberInvite{UserID: inv.UserID, Role: role})
	}
	return out, nil
}

func (s *projectService) GetByID(ctx context.Context, userID, projectID string) (*model.Project, error) {
	p, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	isMember, err := s.projectRepo.IsMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrProjectNotFound
	}

	return p, nil
}

func (s *projectService) List(ctx context.Context, userID string, page, limit int) ([]*model.Project, int, error) {
	return s.projectRepo.List(ctx, userID, page, limit)
}

func (s *projectService) Update(ctx context.Context, userID, projectID string, req *model.UpdateProjectRequest) (*model.Project, error) {
	p, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	if p.OwnerID != userID {
		return nil, ErrForbidden
	}

	if err := validateUpdateProject(req); err != nil {
		return nil, err
	}

	prevDeadline := p.Deadline

	p.Name = req.Name
	p.Description = req.Description
	p.Status = req.Status
	p.Deadline = req.Deadline

	if err := s.projectRepo.Update(ctx, p); err != nil {
		return nil, err
	}

	// A changed deadline invalidates any reminder already sent against the
	// old schedule — let the scheduler warn members again for the new one.
	if !sameDueDate(prevDeadline, p.Deadline) {
		_ = s.projectRepo.ClearProjectReminders(ctx, p.ID)
	}

	return p, nil
}

func (s *projectService) Delete(ctx context.Context, userID, projectID string) error {
	p, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrProjectNotFound
		}
		return err
	}

	if p.OwnerID != userID {
		return ErrForbidden
	}

	return s.projectRepo.Delete(ctx, projectID)
}

// BulkDelete soft-deletes every project in ids that the caller owns. IDs
// belonging to projects the caller does not own — or that don't exist — are
// silently skipped so the operation stays idempotent. Returns the number of
// projects actually deleted.
func (s *projectService) BulkDelete(ctx context.Context, userID string, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("%w: ids is required", ErrValidation)
	}
	if len(ids) > bulkDeleteMaxIDs {
		return 0, fmt.Errorf("%w: at most %d ids per request", ErrValidation, bulkDeleteMaxIDs)
	}
	for _, id := range ids {
		if id == "" {
			return 0, fmt.Errorf("%w: ids must not contain empty values", ErrValidation)
		}
	}
	return s.projectRepo.BulkSoftDelete(ctx, userID, ids)
}

func (s *projectService) AddMember(ctx context.Context, ownerID, projectID string, req *model.AddMemberRequest) (*model.ProjectMember, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrValidation)
	}

	p, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	if p.OwnerID != ownerID {
		return nil, ErrForbidden
	}

	role := req.Role
	if role == "" {
		role = model.ProjectRoleMember
	}
	if role != model.ProjectRoleAdmin && role != model.ProjectRoleMember {
		return nil, fmt.Errorf("%w: role must be 'admin' or 'member'", ErrValidation)
	}

	if _, err := s.userRepo.GetByID(ctx, req.UserID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := s.projectRepo.AddMember(ctx, projectID, req.UserID, role); err != nil {
		if errors.Is(err, repository.ErrDuplicateMember) {
			return nil, ErrAlreadyMember
		}
		return nil, err
	}

	return s.projectRepo.GetMember(ctx, projectID, req.UserID)
}

func (s *projectService) RemoveMember(ctx context.Context, ownerID, projectID, targetUserID string) error {
	p, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrProjectNotFound
		}
		return err
	}

	if p.OwnerID != ownerID {
		return ErrForbidden
	}

	if targetUserID == ownerID {
		return fmt.Errorf("%w: owner cannot be removed from the project", ErrValidation)
	}

	if err := s.projectRepo.RemoveMember(ctx, projectID, targetUserID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrMemberNotFound
		}
		return err
	}

	return nil
}

func (s *projectService) GetMembers(ctx context.Context, userID, projectID string) ([]*model.ProjectMember, error) {
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	isMember, err := s.projectRepo.IsMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrProjectNotFound
	}

	return s.projectRepo.GetMembers(ctx, projectID)
}

func validateCreateProject(req *model.CreateProjectRequest) error {
	if len(req.Name) == 0 {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if len(req.Name) > 255 {
		return fmt.Errorf("%w: name must be at most 255 characters", ErrValidation)
	}
	return nil
}

func validateUpdateProject(req *model.UpdateProjectRequest) error {
	if len(req.Name) == 0 {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if len(req.Name) > 255 {
		return fmt.Errorf("%w: name must be at most 255 characters", ErrValidation)
	}
	if req.Status != model.ProjectStatusActive && req.Status != model.ProjectStatusArchived {
		return fmt.Errorf("%w: status must be 'active' or 'archived'", ErrValidation)
	}
	return nil
}
