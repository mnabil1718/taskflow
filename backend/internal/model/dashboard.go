package model

import "time"

type ProjectTaskCounts struct {
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	Todo        int    `json:"todo"`
	InProgress  int    `json:"in_progress"`
	Done        int    `json:"done"`
	Total       int    `json:"total"`
}

type UpcomingTask struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	DueDate     time.Time    `json:"due_date"`
	ProjectID   string       `json:"project_id"`
	ProjectName string       `json:"project_name"`
}
