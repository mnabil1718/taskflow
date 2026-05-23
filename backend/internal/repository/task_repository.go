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

// ReminderKind is a whitelisted identifier for a deadline-reminder window.
// The repository maps it to the corresponding reminder_*_sent_at column.
type ReminderKind string

const (
	ReminderThreeDay ReminderKind = "3d"
	ReminderOneDay   ReminderKind = "1d"
)

type TaskRepository interface {
	Create(ctx context.Context, t *model.Task) error
	GetByID(ctx context.Context, id string) (*model.Task, error)
	List(ctx context.Context, projectID string, filter model.TaskFilter) ([]*model.Task, int, error)
	ListAll(ctx context.Context, userID string, filter model.TaskFilter) ([]*model.Task, int, error)
	BoardList(ctx context.Context, projectID string) ([]*model.Task, error)
	Update(ctx context.Context, t *model.Task) error
	UpdateAssignee(ctx context.Context, id string, assigneeID *string) error
	Move(ctx context.Context, id string, newStatus model.TaskStatus, newPosition string) (*model.Task, model.TaskStatus, error)
	Delete(ctx context.Context, id string) error
	LogStatusChange(ctx context.Context, taskID string, changedBy *string, from, to model.TaskStatus) error
	GetActivityLogs(ctx context.Context, taskID string) ([]*model.TaskActivityLog, error)

	PendingReminders(ctx context.Context, kind ReminderKind) ([]*model.Task, error)
	MarkReminderSent(ctx context.Context, taskID string, kind ReminderKind) error
	ClearReminders(ctx context.Context, taskID string) error
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, t *model.Task) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO tasks (title, description, status, priority, position, project_id, assignee_id, created_by, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`,
		t.Title,
		nullableString(t.Description),
		t.Status,
		t.Priority,
		t.Position,
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
		SELECT id, title, description, status, priority, position, project_id, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks WHERE id = $1
	`, id).Scan(
		&t.ID,
		&t.Title,
		&desc,
		&t.Status,
		&t.Priority,
		&t.Position,
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
	if filter.Search != "" {
		conds = append(conds, fmt.Sprintf("title ILIKE $%d", idx))
		args = append(args, "%"+filter.Search+"%")
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
		SELECT id, title, description, status, priority, position, project_id, assignee_id, created_by, due_date, created_at, updated_at
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
		t, err := scanTask(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, total, rows.Err()
}

// ListAll returns a paginated list of tasks across every project the given
// user is a member of. Supports the same filters as List.
func (r *taskRepository) ListAll(ctx context.Context, userID string, filter model.TaskFilter) ([]*model.Task, int, error) {
	page := filter.Page
	limit := filter.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	conds := []string{"pm.user_id = $1"}
	args := []any{userID}
	idx := 2

	if filter.Status != "" {
		conds = append(conds, fmt.Sprintf("t.status = $%d", idx))
		args = append(args, filter.Status)
		idx++
	}
	if filter.Priority != "" {
		conds = append(conds, fmt.Sprintf("t.priority = $%d", idx))
		args = append(args, filter.Priority)
		idx++
	}
	if filter.AssigneeID != "" {
		conds = append(conds, fmt.Sprintf("t.assignee_id = $%d", idx))
		args = append(args, filter.AssigneeID)
		idx++
	}
	if filter.Search != "" {
		conds = append(conds, fmt.Sprintf("t.title ILIKE $%d", idx))
		args = append(args, "%"+filter.Search+"%")
		idx++
	}

	where := strings.Join(conds, " AND ")

	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM tasks t
		JOIN project_members pm ON pm.project_id = t.project_id
		WHERE %s
	`, where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count all tasks: %w", err)
	}

	sortCol, sortDir := taskSortClause(filter.SortBy, filter.SortOrder)

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)
	listQuery := fmt.Sprintf(`
		SELECT t.id, t.title, t.description, t.status, t.priority, t.position,
		       t.project_id, t.assignee_id, t.created_by, t.due_date, t.created_at, t.updated_at
		FROM tasks t
		JOIN project_members pm ON pm.project_id = t.project_id
		WHERE %s
		ORDER BY t.%s %s NULLS LAST, t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, sortCol, sortDir, idx, idx+1)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list all tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*model.Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, total, rows.Err()
}

// BoardList returns every task in the project ordered by (status, position)
// for the Kanban board. No pagination — boards are meant to render in one
// shot, and the composite index (project_id, status, position) covers the
// read.
func (r *taskRepository) BoardList(ctx context.Context, projectID string) ([]*model.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, status, priority, position, project_id, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks
		WHERE project_id = $1
		ORDER BY status, position
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("board list: %w", err)
	}
	defer rows.Close()

	tasks := make([]*model.Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
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

// Move updates status and position atomically and returns the refreshed
// row along with the previous status — callers need the old status to
// decide whether to write an activity log entry.
func (r *taskRepository) Move(ctx context.Context, id string, newStatus model.TaskStatus, newPosition string) (*model.Task, model.TaskStatus, error) {
	t := &model.Task{}
	var desc sql.NullString
	var assignee, creator sql.NullString
	var dueDate sql.NullTime
	var prevStatus model.TaskStatus

	// Snapshot the pre-update status in a CTE so the caller can detect a
	// cross-column move without a second round trip. The plain RETURNING
	// clause only sees post-update values.
	err := r.db.QueryRowContext(ctx, `
		WITH old AS (
		    SELECT status FROM tasks WHERE id = $3
		), upd AS (
		    UPDATE tasks
		    SET status = $1,
		        position = $2,
		        updated_at = NOW()
		    WHERE id = $3
		    RETURNING id, title, description, status, priority, position, project_id, assignee_id, created_by, due_date, created_at, updated_at
		)
		SELECT upd.id, upd.title, upd.description, upd.status, upd.priority, upd.position, upd.project_id, upd.assignee_id, upd.created_by, upd.due_date, upd.created_at, upd.updated_at, old.status
		FROM upd, old
	`, newStatus, newPosition, id).Scan(
		&t.ID,
		&t.Title,
		&desc,
		&t.Status,
		&t.Priority,
		&t.Position,
		&t.ProjectID,
		&assignee,
		&creator,
		&dueDate,
		&t.CreatedAt,
		&t.UpdatedAt,
		&prevStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrNotFound
		}
		return nil, "", fmt.Errorf("move task: %w", err)
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

	return t, prevStatus, nil
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

func (r *taskRepository) PendingReminders(ctx context.Context, kind ReminderKind) ([]*model.Task, error) {
	col, interval, err := reminderColumn(kind)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, status, priority, position, project_id, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks
		WHERE assignee_id IS NOT NULL
		  AND status != 'done'
		  AND due_date IS NOT NULL
		  AND due_date > NOW()
		  AND due_date <= NOW() + INTERVAL '%s'
		  AND %s IS NULL
		ORDER BY due_date ASC
	`, interval, col)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("pending reminders: %w", err)
	}
	defer rows.Close()

	tasks := make([]*model.Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan pending reminder: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// scanTask reads one row from a SELECT that follows the canonical task
// column order. Centralised so the column list and the nullable handling
// stay in lockstep across List, BoardList, and PendingReminders.
func scanTask(rows *sql.Rows) (*model.Task, error) {
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
		&t.Position,
		&t.ProjectID,
		&assignee,
		&creator,
		&dueDate,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		return nil, err
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

func (r *taskRepository) MarkReminderSent(ctx context.Context, taskID string, kind ReminderKind) error {
	col, _, err := reminderColumn(kind)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`UPDATE tasks SET %s = NOW() WHERE id = $1`, col)
	if _, err := r.db.ExecContext(ctx, query, taskID); err != nil {
		return fmt.Errorf("mark reminder sent: %w", err)
	}
	return nil
}

func (r *taskRepository) ClearReminders(ctx context.Context, taskID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET reminder_3d_sent_at = NULL, reminder_1d_sent_at = NULL
		WHERE id = $1
	`, taskID)
	if err != nil {
		return fmt.Errorf("clear reminders: %w", err)
	}
	return nil
}

// reminderColumn maps a ReminderKind to its column and the INTERVAL window
// it covers. Returning the interval as a literal string is safe because
// kind is whitelisted by the switch — no caller input ever reaches the SQL.
func reminderColumn(kind ReminderKind) (string, string, error) {
	switch kind {
	case ReminderThreeDay:
		return "reminder_3d_sent_at", "3 days", nil
	case ReminderOneDay:
		return "reminder_1d_sent_at", "1 day", nil
	default:
		return "", "", fmt.Errorf("unknown reminder kind: %q", kind)
	}
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
