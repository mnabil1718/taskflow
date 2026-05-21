package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mnabil1718/taskflow/internal/model"
)

type DashboardRepository interface {
	ProjectTaskCounts(ctx context.Context, userID string) ([]*model.ProjectTaskCounts, error)
	UpcomingTasksForUser(ctx context.Context, userID string, days int) ([]*model.UpcomingTask, error)
}

type dashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) ProjectTaskCounts(ctx context.Context, userID string) ([]*model.ProjectTaskCounts, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			p.id,
			p.name,
			COALESCE(SUM(CASE WHEN t.status = 'todo'        THEN 1 ELSE 0 END), 0) AS todo,
			COALESCE(SUM(CASE WHEN t.status = 'in_progress' THEN 1 ELSE 0 END), 0) AS in_progress,
			COALESCE(SUM(CASE WHEN t.status = 'done'        THEN 1 ELSE 0 END), 0) AS done,
			COUNT(t.id) AS total
		FROM projects p
		JOIN project_members pm ON pm.project_id = p.id
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE pm.user_id = $1 AND p.deleted_at IS NULL
		GROUP BY p.id, p.name, p.created_at
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("project task counts: %w", err)
	}
	defer rows.Close()

	out := make([]*model.ProjectTaskCounts, 0)
	for rows.Next() {
		c := &model.ProjectTaskCounts{}
		if err := rows.Scan(&c.ProjectID, &c.ProjectName, &c.Todo, &c.InProgress, &c.Done, &c.Total); err != nil {
			return nil, fmt.Errorf("scan project task counts: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *dashboardRepository) UpcomingTasksForUser(ctx context.Context, userID string, days int) ([]*model.UpcomingTask, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.title, t.status, t.priority, t.due_date, t.project_id, p.name
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		WHERE t.assignee_id = $1
		  AND p.deleted_at IS NULL
		  AND t.status != 'done'
		  AND t.due_date IS NOT NULL
		  AND t.due_date >= NOW()
		  AND t.due_date <= NOW() + ($2::int * INTERVAL '1 day')
		ORDER BY t.due_date ASC
	`, userID, days)
	if err != nil {
		return nil, fmt.Errorf("upcoming tasks: %w", err)
	}
	defer rows.Close()

	out := make([]*model.UpcomingTask, 0)
	for rows.Next() {
		ut := &model.UpcomingTask{}
		if err := rows.Scan(&ut.ID, &ut.Title, &ut.Status, &ut.Priority, &ut.DueDate, &ut.ProjectID, &ut.ProjectName); err != nil {
			return nil, fmt.Errorf("scan upcoming task: %w", err)
		}
		out = append(out, ut)
	}
	return out, rows.Err()
}
