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

type TaskRepository interface {
	Create(ctx context.Context, t *model.Task) error
	GetByID(ctx context.Context, id string) (*model.Task, error)
	List(ctx context.Context, projectID string, filter model.TaskFilter) ([]*model.Task, int, error)
	Update(ctx context.Context, t *model.Task) error
	UpdateAssignee(ctx context.Context, id string, assigneeID *string) error
	Delete(ctx context.Context, id string) error
	LogStatusChange(ctx context.Context, taskID string, changedBy *string, from, to model.TaskStatus) error
	GetActivityLogs(ctx context.Context, taskID string) ([]*model.TaskActivityLog, error)
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, t *model.Task) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, created_by, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`,
		t.Title,
		nullableString(t.Description),
		t.Status,
		t.Priority,
		t.ProjectID,
		nullableUUID(t.AssigneeID),
		nullableUUID(t.CreatedBy),
		t.DueDate,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrNotFound
		}
		return fmt.Errorf("create task: %w", err)
	}
	return nil
}

func (r *taskRepository) GetByID(ctx context.Context, id string) (*model.Task, error) {
	t := &model.Task{}
	var desc sql.NullString
	var assignee, creator sql.NullString
	var dueDate sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, title, description, status, priority, project_id, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks WHERE id = $1
	`, id).Scan(
		&t.ID,
		&t.Title,
		&desc,
		&t.Status,
		&t.Priority,
		&t.ProjectID,
		&assignee,
		&creator,
		&dueDate,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	if desc.Valid {
		t.Description = desc.String
	}
	if assignee.Valid {
		t.AssigneeID = &assignee.String
	}
	if creator.Valid {
		t.CreatedBy = &creator.String
	}
	if dueDate.Valid {
		t.DueDate = &dueDate.Time
	}

	return t, nil
}

func (r *taskRepository) List(ctx context.Context, projectID string, filter model.TaskFilter) ([]*model.Task, int, error) {
	page := filter.Page
	limit := filter.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	conds := []string{"project_id = $1"}
	args := []any{projectID}
	idx := 2

	if filter.Status != "" {
		conds = append(conds, fmt.Sprintf("status = $%d", idx))
		args = append(args, filter.Status)
		idx++
	}
	if filter.Priority != "" {
		conds = append(conds, fmt.Sprintf("priority = $%d", idx))
		args = append(args, filter.Priority)
		idx++
	}
	if filter.AssigneeID != "" {
		conds = append(conds, fmt.Sprintf("assignee_id = $%d", idx))
		args = append(args, filter.AssigneeID)
		idx++
	}

	where := strings.Join(conds, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE %s", where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	sortCol, sortDir := taskSortClause(filter.SortBy, filter.SortOrder)

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, title, description, status, priority, project_id, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks
		WHERE %s
		ORDER BY %s %s NULLS LAST, created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, idx, idx+1)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*model.Task, 0)
	for rows.Next() {
		t := &model.Task{}
		var desc sql.NullString
		var assignee, creator sql.NullString
		var dueDate sql.NullTime
		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&desc,
			&t.Status,
			&t.Priority,
			&t.ProjectID,
			&assignee,
			&creator,
			&dueDate,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		if desc.Valid {
			t.Description = desc.String
		}
		if assignee.Valid {
			t.AssigneeID = &assignee.String
		}
		if creator.Valid {
			t.CreatedBy = &creator.String
		}
		if dueDate.Valid {
			t.DueDate = &dueDate.Time
		}
		tasks = append(tasks, t)
	}

	return tasks, total, rows.Err()
}

func (r *taskRepository) Update(ctx context.Context, t *model.Task) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET title = $1,
		    description = $2,
		    status = $3,
		    priority = $4,
		    assignee_id = $5,
		    due_date = $6,
		    updated_at = NOW()
		WHERE id = $7
	`,
		t.Title,
		nullableString(t.Description),
		t.Status,
		t.Priority,
		nullableUUID(t.AssigneeID),
		t.DueDate,
		t.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrNotFound
		}
		return fmt.Errorf("update task: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) UpdateAssignee(ctx context.Context, id string, assigneeID *string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET assignee_id = $1, updated_at = NOW()
		WHERE id = $2
	`, nullableUUID(assigneeID), id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrNotFound
		}
		return fmt.Errorf("update task assignee: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) LogStatusChange(ctx context.Context, taskID string, changedBy *string, from, to model.TaskStatus) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO task_activity_logs (task_id, changed_by, from_status, to_status)
		VALUES ($1, $2, $3, $4)
	`, taskID, nullableUUID(changedBy), from, to)
	if err != nil {
		return fmt.Errorf("log status change: %w", err)
	}
	return nil
}

func (r *taskRepository) GetActivityLogs(ctx context.Context, taskID string) ([]*model.TaskActivityLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, task_id, changed_by, from_status, to_status, created_at
		FROM task_activity_logs
		WHERE task_id = $1
		ORDER BY created_at DESC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("get activity logs: %w", err)
	}
	defer rows.Close()

	logs := make([]*model.TaskActivityLog, 0)
	for rows.Next() {
		l := &model.TaskActivityLog{}
		var changedBy sql.NullString
		if err := rows.Scan(&l.ID, &l.TaskID, &changedBy, &l.FromStatus, &l.ToStatus, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan activity log: %w", err)
		}
		if changedBy.Valid {
			l.ChangedBy = &changedBy.String
		}
		logs = append(logs, l)
	}

	return logs, rows.Err()
}

func taskSortClause(sortBy, sortOrder string) (string, string) {
	col := "created_at"
	switch strings.ToLower(sortBy) {
	case "status":
		col = "status"
	case "priority":
		col = "priority"
	case "assignee", "assignee_id":
		col = "assignee_id"
	case "due_date":
		col = "due_date"
	case "updated_at":
		col = "updated_at"
	case "title":
		col = "title"
	case "created_at", "":
		col = "created_at"
	}

	dir := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		dir = "ASC"
	}
	return col, dir
}

func nullableUUID(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
