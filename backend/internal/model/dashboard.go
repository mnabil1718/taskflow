package model

import "time"

type ProjectTaskCounts struct {
	ProjectID   string `json:"project_id"   example:"c303012a-6275-4aa3-adec-ebfb123f4567"`
	ProjectName string `json:"project_name" example:"Q4 Roadmap"`
	Todo        int    `json:"todo"         example:"5"`
	InProgress  int    `json:"in_progress"  example:"3"`
	Done        int    `json:"done"         example:"12"`
	Total       int    `json:"total"        example:"20"`
}

type UpcomingTask struct {
	ID          string       `json:"id"           example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	Title       string       `json:"title"        example:"Ship release candidate"`
	Status      TaskStatus   `json:"status"       example:"in_progress"`
	Priority    TaskPriority `json:"priority"     example:"high"`
	DueDate     time.Time    `json:"due_date"     example:"2026-05-22T17:00:00Z"`
	ProjectID   string       `json:"project_id"   example:"c303012a-6275-4aa3-adec-ebfb123f4567"`
	ProjectName string       `json:"project_name" example:"Q4 Roadmap"`
}
