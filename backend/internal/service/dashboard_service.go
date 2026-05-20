package service

import (
	"context"

	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/repository"
)

const upcomingTaskWindowDays = 3

type DashboardService interface {
	ProjectTaskCounts(ctx context.Context, userID string) ([]*model.ProjectTaskCounts, error)
	UpcomingTasks(ctx context.Context, userID string) ([]*model.UpcomingTask, error)
}

type dashboardService struct {
	repo repository.DashboardRepository
}

func NewDashboardService(repo repository.DashboardRepository) DashboardService {
	return &dashboardService{repo: repo}
}

func (s *dashboardService) ProjectTaskCounts(ctx context.Context, userID string) ([]*model.ProjectTaskCounts, error) {
	return s.repo.ProjectTaskCounts(ctx, userID)
}

func (s *dashboardService) UpcomingTasks(ctx context.Context, userID string) ([]*model.UpcomingTask, error) {
	return s.repo.UpcomingTasksForUser(ctx, userID, upcomingTaskWindowDays)
}
