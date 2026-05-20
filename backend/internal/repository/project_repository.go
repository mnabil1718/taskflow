package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mnabil1718/taskflow/internal/model"
)

var ErrDuplicateMember = errors.New("user is already a project member")

type ProjectRepository interface {
	Create(ctx context.Context, p *model.Project) error
	GetByID(ctx context.Context, id string) (*model.Project, error)
	List(ctx context.Context, userID string, page, limit int) ([]*model.Project, int, error)
	Update(ctx context.Context, p *model.Project) error
	Delete(ctx context.Context, id string) error
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

func (r *projectRepository) Create(ctx context.Context, p *model.Project) error {
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

	return tx.Commit()
}

func (r *projectRepository) GetByID(ctx context.Context, id string) (*model.Project, error) {
	p := &model.Project{}
	var desc sql.NullString
	var deadline sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, status, deadline, owner_id, created_at, updated_at
		FROM projects WHERE id = $1
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
		WHERE pm.user_id = $1
	`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT p.id, p.name, p.description, p.status, p.deadline, p.owner_id, p.created_at, p.updated_at
		FROM projects p
		JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = $1
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
		WHERE id = $5
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
	result, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
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
			SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2
		)
	`, projectID, userID).Scan(&exists)
	return exists, err
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
