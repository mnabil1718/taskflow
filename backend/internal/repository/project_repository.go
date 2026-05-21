package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mnabil1718/taskflow/internal/model"
)

var ErrDuplicateMember = errors.New("user is already a project member")

type ProjectRepository interface {
	Create(ctx context.Context, p *model.Project, invites []model.ProjectMemberInvite) error
	GetByID(ctx context.Context, id string) (*model.Project, error)
	List(ctx context.Context, userID string, page, limit int) ([]*model.Project, int, error)
	Update(ctx context.Context, p *model.Project) error
	Delete(ctx context.Context, id string) error
	BulkSoftDelete(ctx context.Context, ownerID string, ids []string) (int, error)
	AddMember(ctx context.Context, projectID, userID string, role model.ProjectRole) error
	RemoveMember(ctx context.Context, projectID, userID string) error
	GetMember(ctx context.Context, projectID, userID string) (*model.ProjectMember, error)
	GetMembers(ctx context.Context, projectID string) ([]*model.ProjectMember, error)
	IsMember(ctx context.Context, projectID, userID string) (bool, error)
}

type projectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &projectRepository{db: db}
}

// Create inserts a project, the owner membership, and any extra invites in a
// single transaction. If any invitee references a missing user (FK 23503)
// the whole project is rolled back so the caller can't end up with a
// half-populated team.
func (r *projectRepository) Create(ctx context.Context, p *model.Project, invites []model.ProjectMemberInvite) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO projects (name, description, status, deadline, owner_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, p.Name, nullableString(p.Description), p.Status, p.Deadline, p.OwnerID).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, p.ID, p.OwnerID)
	if err != nil {
		return fmt.Errorf("add owner to members: %w", err)
	}

	for _, inv := range invites {
		// Skip self-invites silently — the owner row is already in.
		if inv.UserID == p.OwnerID {
			continue
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO project_members (project_id, user_id, role)
			VALUES ($1, $2, $3)
			ON CONFLICT (project_id, user_id) DO NOTHING
		`, p.ID, inv.UserID, inv.Role)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return ErrNotFound
			}
			return fmt.Errorf("add invitee %s: %w", inv.UserID, err)
		}
	}

	return tx.Commit()
}

func (r *projectRepository) GetByID(ctx context.Context, id string) (*model.Project, error) {
	p := &model.Project{}
	var desc sql.NullString
	var deadline sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, status, deadline, owner_id, created_at, updated_at
		FROM projects WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&p.ID, &p.Name, &desc, &p.Status, &deadline, &p.OwnerID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get project: %w", err)
	}

	if desc.Valid {
		p.Description = desc.String
	}
	if deadline.Valid {
		p.Deadline = &deadline.Time
	}

	return p, nil
}

func (r *projectRepository) List(ctx context.Context, userID string, page, limit int) ([]*model.Project, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT p.id)
		FROM projects p
		JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = $1 AND p.deleted_at IS NULL
	`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT p.id, p.name, p.description, p.status, p.deadline, p.owner_id, p.created_at, p.updated_at
		FROM projects p
		JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]*model.Project, 0)
	for rows.Next() {
		p := &model.Project{}
		var desc sql.NullString
		var deadline sql.NullTime
		if err := rows.Scan(&p.ID, &p.Name, &desc, &p.Status, &deadline, &p.OwnerID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan project: %w", err)
		}
		if desc.Valid {
			p.Description = desc.String
		}
		if deadline.Valid {
			p.Deadline = &deadline.Time
		}
		projects = append(projects, p)
	}

	return projects, total, rows.Err()
}

func (r *projectRepository) Update(ctx context.Context, p *model.Project) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE projects
		SET name = $1, description = $2, status = $3, deadline = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
	`, p.Name, nullableString(p.Description), p.Status, p.Deadline, p.ID)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *projectRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE projects SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// BulkSoftDelete marks every project the caller owns and that hasn't already
// been deleted as deleted_at = NOW(). Projects the caller doesn't own — or
// IDs that don't exist — are silently skipped; the returned count tells the
// caller how many rows were actually affected. This keeps the operation
// idempotent under concurrent deletes without requiring a per-ID round trip.
func (r *projectRepository) BulkSoftDelete(ctx context.Context, ownerID string, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, ownerID)
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		UPDATE projects SET deleted_at = NOW()
		WHERE owner_id = $1
		  AND deleted_at IS NULL
		  AND id IN (%s)
	`, strings.Join(placeholders, ","))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("bulk soft delete projects: %w", err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

func (r *projectRepository) AddMember(ctx context.Context, projectID, userID string, role model.ProjectRole) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)
	`, projectID, userID, role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return ErrDuplicateMember
			case "23503":
				return ErrNotFound
			}
		}
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (r *projectRepository) RemoveMember(ctx context.Context, projectID, userID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM project_members WHERE project_id = $1 AND user_id = $2
	`, projectID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *projectRepository) GetMember(ctx context.Context, projectID, userID string) (*model.ProjectMember, error) {
	m := &model.ProjectMember{}
	err := r.db.QueryRowContext(ctx, `
		SELECT pm.project_id, pm.user_id, u.name, u.email, pm.role, pm.joined_at
		FROM project_members pm
		JOIN users u ON pm.user_id = u.id
		WHERE pm.project_id = $1 AND pm.user_id = $2
	`, projectID, userID).Scan(&m.ProjectID, &m.UserID, &m.Name, &m.Email, &m.Role, &m.JoinedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get member: %w", err)
	}
	return m, nil
}

func (r *projectRepository) GetMembers(ctx context.Context, projectID string) ([]*model.ProjectMember, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pm.project_id, pm.user_id, u.name, u.email, pm.role, pm.joined_at
		FROM project_members pm
		JOIN users u ON pm.user_id = u.id
		WHERE pm.project_id = $1
		ORDER BY pm.joined_at
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	defer rows.Close()

	members := make([]*model.ProjectMember, 0)
	for rows.Next() {
		m := &model.ProjectMember{}
		if err := rows.Scan(&m.ProjectID, &m.UserID, &m.Name, &m.Email, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}

	return members, rows.Err()
}

func (r *projectRepository) IsMember(ctx context.Context, projectID, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM project_members pm
			JOIN projects p ON p.id = pm.project_id
			WHERE pm.project_id = $1 AND pm.user_id = $2 AND p.deleted_at IS NULL
		)
	`, projectID, userID).Scan(&exists)
	return exists, err
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
