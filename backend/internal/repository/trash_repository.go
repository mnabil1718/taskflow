package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mnabil1718/taskflow/internal/model"
)

type TrashRepository interface {
	// List returns soft-deleted projects (owner-scoped) and soft-deleted
	// tasks (project-owner-or-creator-scoped, and only when the task's
	// project itself is still active — tasks under a trashed project are
	// represented by the project row, not as individual entries).
	List(ctx context.Context, userID string) ([]*model.TrashItem, error)

	// RestoreProjects clears deleted_at on projects owned by the caller.
	// Returns the count of rows actually restored.
	RestoreProjects(ctx context.Context, userID string, ids []string) (int, error)

	// RestoreTasks clears deleted_at on tasks whose project is active and
	// whose deletion the caller is allowed to undo (project owner or task
	// creator). Returns the count of rows actually restored.
	RestoreTasks(ctx context.Context, userID string, ids []string) (int, error)

	// PurgeProjects permanently deletes projects owned by the caller. The
	// FK ON DELETE CASCADE removes child tasks automatically.
	PurgeProjects(ctx context.Context, userID string, ids []string) (int, error)

	// PurgeTasks permanently deletes tasks the caller is allowed to remove.
	PurgeTasks(ctx context.Context, userID string, ids []string) (int, error)

	// EmptyAll purges every soft-deleted item the caller can see. Returns
	// the per-kind counts so the UI can report what was cleared.
	EmptyAll(ctx context.Context, userID string) (purgedProjects, purgedTasks int, err error)
}

type trashRepository struct {
	db *sql.DB
}

func NewTrashRepository(db *sql.DB) TrashRepository {
	return &trashRepository{db: db}
}

func (r *trashRepository) List(ctx context.Context, userID string) ([]*model.TrashItem, error) {
	// UNION ALL keeps the two scans independent (no de-dup work), and the
	// outer ORDER BY interleaves them by deleted_at so the most recently
	// deleted things surface first regardless of kind. The project branch
	// also reports how many tasks were swept in by the cascade soft-delete
	// so the row can advertise "Auth rework — 5 tasks" instead of leaving
	// the user wondering where the tasks went.
	rows, err := r.db.QueryContext(ctx, `
		SELECT kind, id, title, project_id, project_name, task_count, deleted_at
		FROM (
			SELECT 'project'::text AS kind,
			       p.id::text       AS id,
			       p.name           AS title,
			       NULL::text       AS project_id,
			       NULL::text       AS project_name,
			       (SELECT COUNT(*) FROM tasks
			        WHERE project_id = p.id
			          AND deleted_at = p.deleted_at)::int AS task_count,
			       p.deleted_at     AS deleted_at
			FROM projects p
			WHERE p.owner_id = $1 AND p.deleted_at IS NOT NULL

			UNION ALL

			SELECT 'task'::text     AS kind,
			       t.id::text       AS id,
			       t.title          AS title,
			       t.project_id::text AS project_id,
			       p.name           AS project_name,
			       NULL::int        AS task_count,
			       t.deleted_at     AS deleted_at
			FROM tasks t
			JOIN projects p ON p.id = t.project_id
			WHERE t.deleted_at IS NOT NULL
			  AND p.deleted_at IS NULL
			  AND (p.owner_id = $1 OR t.created_by = $1)
		) trash
		ORDER BY deleted_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list trash: %w", err)
	}
	defer rows.Close()

	out := make([]*model.TrashItem, 0)
	for rows.Next() {
		item := &model.TrashItem{}
		var projectID, projectName sql.NullString
		var taskCount sql.NullInt64
		if err := rows.Scan(&item.Kind, &item.ID, &item.Title, &projectID, &projectName, &taskCount, &item.DeletedAt); err != nil {
			return nil, fmt.Errorf("scan trash item: %w", err)
		}
		if projectID.Valid {
			item.ProjectID = &projectID.String
		}
		if projectName.Valid {
			item.ProjectName = &projectName.String
		}
		if taskCount.Valid {
			n := int(taskCount.Int64)
			item.TaskCount = &n
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// RestoreProjects clears deleted_at on projects owned by the caller AND on
// the tasks that were soft-deleted *as part of* those project deletions
// (matched by identical deleted_at timestamp). Tasks individually trashed
// before or after the project was deleted have a different timestamp and
// stay in trash for the user to handle separately.
func (r *trashRepository) RestoreProjects(ctx context.Context, userID string, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	placeholders, args := bulkIDArgs(userID, ids)
	// `old_state` snapshots the (id, deleted_at) pairs before the project
	// UPDATE runs. PostgreSQL modifying CTEs share one snapshot, so the
	// task UPDATE still sees the original timestamps even though
	// rest_projects is busy clearing them.
	query := fmt.Sprintf(`
		WITH old_state AS (
			SELECT id, deleted_at
			FROM projects
			WHERE id IN (%s)
			  AND owner_id = $1
			  AND deleted_at IS NOT NULL
		), rest_tasks AS (
			UPDATE tasks t
			SET deleted_at = NULL, updated_at = NOW()
			FROM old_state os
			WHERE t.project_id = os.id
			  AND t.deleted_at = os.deleted_at
			RETURNING 1
		), rest_projects AS (
			UPDATE projects p
			SET deleted_at = NULL, updated_at = NOW()
			FROM old_state os
			WHERE p.id = os.id
			RETURNING 1
		)
		SELECT COUNT(*) FROM rest_projects
	`, placeholders)
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("restore projects: %w", err)
	}
	return count, nil
}

func (r *trashRepository) RestoreTasks(ctx context.Context, userID string, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	placeholders, args := bulkIDArgs(userID, ids)
	// Subquery filters out tasks whose project is trashed or which the
	// caller has no business restoring; the outer UPDATE only touches
	// rows that pass both gates.
	query := fmt.Sprintf(`
		UPDATE tasks t
		SET deleted_at = NULL, updated_at = NOW()
		FROM projects p
		WHERE t.project_id = p.id
		  AND t.id IN (%s)
		  AND t.deleted_at IS NOT NULL
		  AND p.deleted_at IS NULL
		  AND (p.owner_id = $1 OR t.created_by = $1)
	`, placeholders)
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("restore tasks: %w", err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

func (r *trashRepository) PurgeProjects(ctx context.Context, userID string, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	placeholders, args := bulkIDArgs(userID, ids)
	// Permanent delete — the FK ON DELETE CASCADE on tasks/project_members
	// fans out the removal so we don't have to clean those up manually.
	query := fmt.Sprintf(`
		DELETE FROM projects
		WHERE id IN (%s)
		  AND owner_id = $1
		  AND deleted_at IS NOT NULL
	`, placeholders)
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("purge projects: %w", err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

func (r *trashRepository) PurgeTasks(ctx context.Context, userID string, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	placeholders, args := bulkIDArgs(userID, ids)
	query := fmt.Sprintf(`
		DELETE FROM tasks t
		USING projects p
		WHERE t.project_id = p.id
		  AND t.id IN (%s)
		  AND t.deleted_at IS NOT NULL
		  AND (p.owner_id = $1 OR t.created_by = $1)
	`, placeholders)
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("purge tasks: %w", err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

func (r *trashRepository) EmptyAll(ctx context.Context, userID string) (int, int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("empty trash: begin tx: %w", err)
	}
	defer tx.Rollback()

	// Purge tasks first. Doing projects first would have the FK cascade
	// sweep some of these rows away too, masking the per-kind count.
	taskResult, err := tx.ExecContext(ctx, `
		DELETE FROM tasks t
		USING projects p
		WHERE t.project_id = p.id
		  AND t.deleted_at IS NOT NULL
		  AND p.deleted_at IS NULL
		  AND (p.owner_id = $1 OR t.created_by = $1)
	`, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("empty trash: purge tasks: %w", err)
	}
	taskCount, _ := taskResult.RowsAffected()

	projectResult, err := tx.ExecContext(ctx, `
		DELETE FROM projects
		WHERE owner_id = $1 AND deleted_at IS NOT NULL
	`, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("empty trash: purge projects: %w", err)
	}
	projectCount, _ := projectResult.RowsAffected()

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("empty trash: commit: %w", err)
	}
	return int(projectCount), int(taskCount), nil
}

// bulkIDArgs builds the placeholder list and args slice for a query that
// takes userID as $1 and a variadic list of ids starting at $2.
func bulkIDArgs(userID string, ids []string) (string, []any) {
	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, userID)
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}
	return strings.Join(placeholders, ","), args
}
