package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/repository"
	"github.com/mnabil1718/taskflow/internal/service"
)

// ---------------------------------------------------------------------------
// mockProjectRepo
// ---------------------------------------------------------------------------

type mockProjectRepo struct {
	projects        map[string]*model.Project
	members         map[string]map[string]*model.ProjectMember // projectID → userID → member
	createErr       error
	addMemberErr    error
	removeMemberErr error
}

func newMockProjectRepo() *mockProjectRepo {
	return &mockProjectRepo{
		projects: make(map[string]*model.Project),
		members:  make(map[string]map[string]*model.ProjectMember),
	}
}

func (m *mockProjectRepo) Create(_ context.Context, p *model.Project, invites []model.ProjectMemberInvite) error {
	if m.createErr != nil {
		return m.createErr
	}
	p.ID = fmt.Sprintf("project-%d", len(m.projects)+1)
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.projects[p.ID] = p

	// mirror real repo: auto-add owner to project_members
	if m.members[p.ID] == nil {
		m.members[p.ID] = make(map[string]*model.ProjectMember)
	}
	m.members[p.ID][p.OwnerID] = &model.ProjectMember{
		ProjectID: p.ID,
		UserID:    p.OwnerID,
		Role:      model.ProjectRoleOwner,
		JoinedAt:  time.Now(),
	}

	for _, inv := range invites {
		if inv.UserID == p.OwnerID {
			continue
		}
		if _, exists := m.members[p.ID][inv.UserID]; exists {
			continue
		}
		m.members[p.ID][inv.UserID] = &model.ProjectMember{
			ProjectID: p.ID,
			UserID:    inv.UserID,
			Role:      inv.Role,
			JoinedAt:  time.Now(),
		}
	}
	return nil
}

func (m *mockProjectRepo) GetByID(_ context.Context, id string) (*model.Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return p, nil
}

func (m *mockProjectRepo) List(_ context.Context, userID string, page, limit int) ([]*model.Project, int, error) {
	var all []*model.Project
	for id, p := range m.projects {
		if memMap, ok := m.members[id]; ok {
			if _, member := memMap[userID]; member {
				all = append(all, p)
			}
		}
	}
	total := len(all)
	start := (page - 1) * limit
	if start >= total {
		return []*model.Project{}, 0, nil
	}
	end := start + limit
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (m *mockProjectRepo) Update(_ context.Context, p *model.Project) error {
	if _, ok := m.projects[p.ID]; !ok {
		return repository.ErrNotFound
	}
	p.UpdatedAt = time.Now()
	m.projects[p.ID] = p
	return nil
}

func (m *mockProjectRepo) Delete(_ context.Context, id string) error {
	if _, ok := m.projects[id]; !ok {
		return repository.ErrNotFound
	}
	delete(m.projects, id)
	delete(m.members, id)
	return nil
}

func (m *mockProjectRepo) BulkSoftDelete(_ context.Context, ownerID string, ids []string) (int, error) {
	count := 0
	for _, id := range ids {
		p, ok := m.projects[id]
		if !ok || p.OwnerID != ownerID {
			continue
		}
		delete(m.projects, id)
		delete(m.members, id)
		count++
	}
	return count, nil
}

func (m *mockProjectRepo) AddMember(_ context.Context, projectID, userID string, role model.ProjectRole) error {
	if m.addMemberErr != nil {
		return m.addMemberErr
	}
	if m.members[projectID] == nil {
		m.members[projectID] = make(map[string]*model.ProjectMember)
	}
	if _, exists := m.members[projectID][userID]; exists {
		return repository.ErrDuplicateMember
	}
	m.members[projectID][userID] = &model.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
		JoinedAt:  time.Now(),
	}
	return nil
}

func (m *mockProjectRepo) RemoveMember(_ context.Context, projectID, userID string) error {
	if m.removeMemberErr != nil {
		return m.removeMemberErr
	}
	members, ok := m.members[projectID]
	if !ok {
		return repository.ErrNotFound
	}
	if _, exists := members[userID]; !exists {
		return repository.ErrNotFound
	}
	delete(members, userID)
	return nil
}

func (m *mockProjectRepo) GetMember(_ context.Context, projectID, userID string) (*model.ProjectMember, error) {
	members, ok := m.members[projectID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	mem, ok := members[userID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return mem, nil
}

func (m *mockProjectRepo) GetMembers(_ context.Context, projectID string) ([]*model.ProjectMember, error) {
	members, ok := m.members[projectID]
	if !ok {
		return []*model.ProjectMember{}, nil
	}
	result := make([]*model.ProjectMember, 0, len(members))
	for _, mem := range members {
		result = append(result, mem)
	}
	return result, nil
}

func (m *mockProjectRepo) IsMember(_ context.Context, projectID, userID string) (bool, error) {
	members, ok := m.members[projectID]
	if !ok {
		return false, nil
	}
	_, exists := members[userID]
	return exists, nil
}

func (m *mockProjectRepo) PendingProjectReminders(_ context.Context, _ repository.ReminderKind) ([]*model.Project, error) {
	return nil, nil
}
func (m *mockProjectRepo) MarkProjectReminderSent(_ context.Context, _ string, _ repository.ReminderKind) error {
	return nil
}
func (m *mockProjectRepo) ClearProjectReminders(_ context.Context, _ string) error { return nil }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// seedProject directly writes a project + owner membership into the mock repo.
func seedProject(pr *mockProjectRepo, id, ownerID, name string) *model.Project {
	p := &model.Project{
		ID:        id,
		Name:      name,
		Status:    model.ProjectStatusActive,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	pr.projects[id] = p
	if pr.members[id] == nil {
		pr.members[id] = make(map[string]*model.ProjectMember)
	}
	pr.members[id][ownerID] = &model.ProjectMember{
		ProjectID: id,
		UserID:    ownerID,
		Role:      model.ProjectRoleOwner,
		JoinedAt:  time.Now(),
	}
	return p
}

func newProjectSvc(pr *mockProjectRepo, ur *mockUserRepo) service.ProjectService {
	return service.NewProjectService(pr, ur)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestCreateProject_Success(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	svc := newProjectSvc(pr, ur)

	p, err := svc.Create(context.Background(), "u1", &model.CreateProjectRequest{
		Name:        "My Project",
		Description: "A description",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, p.ID)
	assert.Equal(t, "My Project", p.Name)
	assert.Equal(t, model.ProjectStatusActive, p.Status)
	assert.Equal(t, "u1", p.OwnerID)

	// owner must be auto-added to project_members
	isMember, _ := pr.IsMember(context.Background(), p.ID, "u1")
	assert.True(t, isMember)
}

func TestCreateProject_ValidationErrors(t *testing.T) {
	svc := newProjectSvc(newMockProjectRepo(), newMockUserRepo())
	ctx := context.Background()

	cases := []struct {
		name string
		req  model.CreateProjectRequest
	}{
		{"empty name", model.CreateProjectRequest{Name: ""}},
		{"name too long", model.CreateProjectRequest{Name: string(make([]byte, 256))}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(ctx, "u1", &tc.req)
			require.Error(t, err)
			assert.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestGetByID_Success(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	p, err := svc.GetByID(context.Background(), "u1", "p1")

	require.NoError(t, err)
	assert.Equal(t, "p1", p.ID)
}

func TestGetByID_ProjectNotFound(t *testing.T) {
	svc := newProjectSvc(newMockProjectRepo(), newMockUserRepo())

	_, err := svc.GetByID(context.Background(), "u1", "nonexistent")

	assert.ErrorIs(t, err, service.ErrProjectNotFound)
}

func TestGetByID_NonMemberCannotAccess(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha") // u1 is owner; u2 is not a member
	svc := newProjectSvc(pr, ur)

	_, err := svc.GetByID(context.Background(), "u2", "p1")

	// must not leak existence to non-members
	assert.ErrorIs(t, err, service.ErrProjectNotFound)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestList_ReturnsOnlyUserProjects(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	seedProject(pr, "p2", "u1", "Beta")
	seedProject(pr, "p3", "u2", "Gamma") // belongs to u2 only
	svc := newProjectSvc(pr, ur)

	projects, total, err := svc.List(context.Background(), "u1", 1, 10)

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, projects, 2)
}

func TestList_EmptyWhenNoProjects(t *testing.T) {
	svc := newProjectSvc(newMockProjectRepo(), newMockUserRepo())

	projects, total, err := svc.List(context.Background(), "u1", 1, 10)

	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, projects)
}

func TestList_Pagination(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	for i := 1; i <= 5; i++ {
		seedProject(pr, fmt.Sprintf("p%d", i), "u1", fmt.Sprintf("Project %d", i))
	}
	svc := newProjectSvc(pr, ur)

	page1, total, err := svc.List(context.Background(), "u1", 1, 3)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, page1, 3)

	page2, _, err := svc.List(context.Background(), "u1", 2, 3)
	require.NoError(t, err)
	assert.Len(t, page2, 2)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdate_Success(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	updated, err := svc.Update(context.Background(), "u1", "p1", &model.UpdateProjectRequest{
		Name:   "Alpha Renamed",
		Status: model.ProjectStatusArchived,
	})

	require.NoError(t, err)
	assert.Equal(t, "Alpha Renamed", updated.Name)
	assert.Equal(t, model.ProjectStatusArchived, updated.Status)
}

func TestUpdate_Forbidden(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	_, err := svc.Update(context.Background(), "u2", "p1", &model.UpdateProjectRequest{
		Name:   "Hijacked",
		Status: model.ProjectStatusActive,
	})

	assert.ErrorIs(t, err, service.ErrForbidden)
}

func TestUpdate_ProjectNotFound(t *testing.T) {
	svc := newProjectSvc(newMockProjectRepo(), newMockUserRepo())

	_, err := svc.Update(context.Background(), "u1", "nonexistent", &model.UpdateProjectRequest{
		Name:   "X",
		Status: model.ProjectStatusActive,
	})

	assert.ErrorIs(t, err, service.ErrProjectNotFound)
}

func TestUpdate_ValidationErrors(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)
	ctx := context.Background()

	cases := []struct {
		name string
		req  model.UpdateProjectRequest
	}{
		{"empty name", model.UpdateProjectRequest{Name: "", Status: model.ProjectStatusActive}},
		{"invalid status", model.UpdateProjectRequest{Name: "Alpha", Status: "pending"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Update(ctx, "u1", "p1", &tc.req)
			assert.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDelete_Success(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	err := svc.Delete(context.Background(), "u1", "p1")

	require.NoError(t, err)
	assert.NotContains(t, pr.projects, "p1")
}

func TestDelete_Forbidden(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	err := svc.Delete(context.Background(), "u2", "p1")

	assert.ErrorIs(t, err, service.ErrForbidden)
}

func TestDelete_ProjectNotFound(t *testing.T) {
	svc := newProjectSvc(newMockProjectRepo(), newMockUserRepo())

	err := svc.Delete(context.Background(), "u1", "nonexistent")

	assert.ErrorIs(t, err, service.ErrProjectNotFound)
}

// ---------------------------------------------------------------------------
// AddMember
// ---------------------------------------------------------------------------

func TestAddMember_Success(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	member, err := svc.AddMember(context.Background(), "u1", "p1", &model.AddMemberRequest{
		UserID: "u2",
		Role:   model.ProjectRoleMember,
	})

	require.NoError(t, err)
	require.NotNil(t, member)
	assert.Equal(t, "u2", member.UserID)
	assert.Equal(t, model.ProjectRoleMember, member.Role)

	isMember, _ := pr.IsMember(context.Background(), "p1", "u2")
	assert.True(t, isMember)
}

func TestAddMember_DefaultsToMemberRole(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	member, err := svc.AddMember(context.Background(), "u1", "p1", &model.AddMemberRequest{
		UserID: "u2",
		// Role intentionally omitted
	})

	require.NoError(t, err)
	assert.Equal(t, model.ProjectRoleMember, member.Role)
}

func TestAddMember_MissingUserID(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	_, err := svc.AddMember(context.Background(), "u1", "p1", &model.AddMemberRequest{UserID: ""})

	assert.ErrorIs(t, err, service.ErrValidation)
}

func TestAddMember_InvalidRole(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	_, err := svc.AddMember(context.Background(), "u1", "p1", &model.AddMemberRequest{
		UserID: "u2",
		Role:   "owner", // cannot assign owner role
	})

	assert.ErrorIs(t, err, service.ErrValidation)
}

func TestAddMember_Forbidden(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	seedUser(ur, "u3", "Carol", "carol@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	_, err := svc.AddMember(context.Background(), "u2", "p1", &model.AddMemberRequest{
		UserID: "u3",
		Role:   model.ProjectRoleMember,
	})

	assert.ErrorIs(t, err, service.ErrForbidden)
}

func TestAddMember_ProjectNotFound(t *testing.T) {
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	svc := newProjectSvc(newMockProjectRepo(), ur)

	_, err := svc.AddMember(context.Background(), "u1", "nonexistent", &model.AddMemberRequest{
		UserID: "u2",
		Role:   model.ProjectRoleMember,
	})

	assert.ErrorIs(t, err, service.ErrProjectNotFound)
}

func TestAddMember_TargetUserNotFound(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	_, err := svc.AddMember(context.Background(), "u1", "p1", &model.AddMemberRequest{
		UserID: "ghost-user",
		Role:   model.ProjectRoleMember,
	})

	assert.ErrorIs(t, err, service.ErrUserNotFound)
}

func TestAddMember_AlreadyMember(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	// manually add u2 as existing member
	pr.members["p1"]["u2"] = &model.ProjectMember{
		ProjectID: "p1", UserID: "u2", Role: model.ProjectRoleMember, JoinedAt: time.Now(),
	}
	svc := newProjectSvc(pr, ur)

	_, err := svc.AddMember(context.Background(), "u1", "p1", &model.AddMemberRequest{
		UserID: "u2",
		Role:   model.ProjectRoleMember,
	})

	assert.ErrorIs(t, err, service.ErrAlreadyMember)
}

// ---------------------------------------------------------------------------
// RemoveMember
// ---------------------------------------------------------------------------

func TestRemoveMember_Success(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	p := seedProject(pr, "p1", "u1", "Alpha")
	pr.members[p.ID]["u2"] = &model.ProjectMember{
		ProjectID: p.ID, UserID: "u2", Role: model.ProjectRoleMember, JoinedAt: time.Now(),
	}
	svc := newProjectSvc(pr, ur)

	err := svc.RemoveMember(context.Background(), "u1", "p1", "u2")

	require.NoError(t, err)
	isMember, _ := pr.IsMember(context.Background(), "p1", "u2")
	assert.False(t, isMember)
}

func TestRemoveMember_Forbidden(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	seedUser(ur, "u3", "Carol", "carol@example.com", "password123")
	p := seedProject(pr, "p1", "u1", "Alpha")
	pr.members[p.ID]["u2"] = &model.ProjectMember{
		ProjectID: p.ID, UserID: "u2", Role: model.ProjectRoleMember, JoinedAt: time.Now(),
	}
	svc := newProjectSvc(pr, ur)

	err := svc.RemoveMember(context.Background(), "u2", "p1", "u3")

	assert.ErrorIs(t, err, service.ErrForbidden)
}

func TestRemoveMember_OwnerCannotRemoveSelf(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	err := svc.RemoveMember(context.Background(), "u1", "p1", "u1")

	assert.ErrorIs(t, err, service.ErrValidation)
}

func TestRemoveMember_MemberNotFound(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	err := svc.RemoveMember(context.Background(), "u1", "p1", "u2")

	assert.ErrorIs(t, err, service.ErrMemberNotFound)
}

func TestRemoveMember_ProjectNotFound(t *testing.T) {
	svc := newProjectSvc(newMockProjectRepo(), newMockUserRepo())

	err := svc.RemoveMember(context.Background(), "u1", "nonexistent", "u2")

	assert.ErrorIs(t, err, service.ErrProjectNotFound)
}

// ---------------------------------------------------------------------------
// GetMembers
// ---------------------------------------------------------------------------

func TestGetMembers_Success(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	p := seedProject(pr, "p1", "u1", "Alpha")
	pr.members[p.ID]["u2"] = &model.ProjectMember{
		ProjectID: p.ID, UserID: "u2", Role: model.ProjectRoleMember, JoinedAt: time.Now(),
	}
	svc := newProjectSvc(pr, ur)

	members, err := svc.GetMembers(context.Background(), "u1", "p1")

	require.NoError(t, err)
	assert.Len(t, members, 2) // owner + u2
}

func TestGetMembers_NonMemberForbidden(t *testing.T) {
	pr := newMockProjectRepo()
	ur := newMockUserRepo()
	seedUser(ur, "u1", "Alice", "alice@example.com", "password123")
	seedUser(ur, "u2", "Bob", "bob@example.com", "password123")
	seedProject(pr, "p1", "u1", "Alpha")
	svc := newProjectSvc(pr, ur)

	_, err := svc.GetMembers(context.Background(), "u2", "p1")

	assert.ErrorIs(t, err, service.ErrProjectNotFound)
}

func TestGetMembers_ProjectNotFound(t *testing.T) {
	svc := newProjectSvc(newMockProjectRepo(), newMockUserRepo())

	_, err := svc.GetMembers(context.Background(), "u1", "nonexistent")

	assert.ErrorIs(t, err, service.ErrProjectNotFound)
}
