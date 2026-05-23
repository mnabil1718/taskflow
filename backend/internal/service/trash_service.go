package service

import (
	"context"
	"fmt"

	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/repository"
)

// trashBulkMax caps the size of restore/purge payloads so an oversized
// id list can't exhaust the parameter limit on the prepared statement.
const trashBulkMax = 100

type TrashService interface {
	List(ctx context.Context, userID string) ([]*model.TrashItem, error)
	Restore(ctx context.Context, userID string, req *model.BulkTrashRequest) (*model.BulkTrashResponse, error)
	Purge(ctx context.Context, userID string, req *model.BulkTrashRequest) (*model.BulkTrashResponse, error)
	EmptyAll(ctx context.Context, userID string) (*model.BulkTrashResponse, error)
}

type trashService struct {
	trashRepo repository.TrashRepository
}

func NewTrashService(trashRepo repository.TrashRepository) TrashService {
	return &trashService{trashRepo: trashRepo}
}

func (s *trashService) List(ctx context.Context, userID string) ([]*model.TrashItem, error) {
	return s.trashRepo.List(ctx, userID)
}

func (s *trashService) Restore(ctx context.Context, userID string, req *model.BulkTrashRequest) (*model.BulkTrashResponse, error) {
	if err := validateBulkTrash(req); err != nil {
		return nil, err
	}

	projectCount, err := s.trashRepo.RestoreProjects(ctx, userID, req.ProjectIDs)
	if err != nil {
		return nil, err
	}
	taskCount, err := s.trashRepo.RestoreTasks(ctx, userID, req.TaskIDs)
	if err != nil {
		return nil, err
	}
	return &model.BulkTrashResponse{
		RestoredProjects: projectCount,
		RestoredTasks:    taskCount,
	}, nil
}

func (s *trashService) Purge(ctx context.Context, userID string, req *model.BulkTrashRequest) (*model.BulkTrashResponse, error) {
	if err := validateBulkTrash(req); err != nil {
		return nil, err
	}

	taskCount, err := s.trashRepo.PurgeTasks(ctx, userID, req.TaskIDs)
	if err != nil {
		return nil, err
	}
	projectCount, err := s.trashRepo.PurgeProjects(ctx, userID, req.ProjectIDs)
	if err != nil {
		return nil, err
	}
	return &model.BulkTrashResponse{
		PurgedProjects: projectCount,
		PurgedTasks:    taskCount,
	}, nil
}

func (s *trashService) EmptyAll(ctx context.Context, userID string) (*model.BulkTrashResponse, error) {
	projectCount, taskCount, err := s.trashRepo.EmptyAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &model.BulkTrashResponse{
		PurgedProjects: projectCount,
		PurgedTasks:    taskCount,
	}, nil
}

func validateBulkTrash(req *model.BulkTrashRequest) error {
	if len(req.ProjectIDs) == 0 && len(req.TaskIDs) == 0 {
		return fmt.Errorf("%w: at least one id is required", ErrValidation)
	}
	if len(req.ProjectIDs) > trashBulkMax || len(req.TaskIDs) > trashBulkMax {
		return fmt.Errorf("%w: at most %d ids per kind", ErrValidation, trashBulkMax)
	}
	return nil
}
