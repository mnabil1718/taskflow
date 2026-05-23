package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/notifier"
	"github.com/mnabil1718/taskflow/internal/repository"
)

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrAssigneeNotMember = errors.New("assignee is not a member of this project")
)

type TaskService interface {
	Create(ctx context.Context, userID, projectID string, req *model.CreateTaskRequest) (*model.Task, error)
	GetByID(ctx context.Context, userID, taskID string) (*model.Task, error)
	List(ctx context.Context, userID, projectID string, filter model.TaskFilter) ([]*model.Task, int, error)
	ListAll(ctx context.Context, userID string, filter model.TaskFilter) ([]*model.Task, int, error)
	BulkDelete(ctx context.Context, userID string, ids []string) (int, error)
	Board(ctx context.Context, userID, projectID string) (*model.BoardView, error)
	Update(ctx context.Context, userID, taskID string, req *model.UpdateTaskRequest) (*model.Task, error)
	Move(ctx context.Context, userID, taskID string, req *model.MoveTaskRequest) (*model.Task, error)
	Delete(ctx context.Context, userID, taskID string) error
	Assign(ctx context.Context, userID, taskID string, req *model.AssignTaskRequest) (*model.Task, error)
	GetActivityLogs(ctx context.Context, userID, taskID string) ([]*model.TaskActivityLog, error)
}

type taskService struct {
	taskRepo    repository.TaskRepository
	projectRepo repository.ProjectRepository
	hub         *notifier.Hub
}

func NewTaskService(taskRepo repository.TaskRepository, projectRepo repository.ProjectRepository, hub *notifier.Hub) TaskService {
	return &taskService{taskRepo: taskRepo, projectRepo: projectRepo, hub: hub}
}

// notify fans out a task event to every member of the task's project.
// Failures here are swallowed so that a notifier hiccup never breaks a
// mutation that has already committed.
func (s *taskService) notify(ctx context.Context, projectID string, ev notifier.Event) {
	members, err := s.projectRepo.GetMembers(ctx, projectID)
	if err != nil {
		return
	}
	recipients := make([]string, 0, len(members))
	for _, m := range members {
		recipients = append(recipients, m.UserID)
	}
	s.hub.Publish(recipients, ev)
}

func (s *taskService) Create(ctx context.Context, userID, projectID string, req *model.CreateTaskRequest) (*model.Task, error) {
	if _, err := s.projectRepo.GetByID(ctx, projectID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	isMember, err := s.projectRepo.IsMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrProjectNotFound
	}

	if err := validateCreateTask(req); err != nil {
		return nil, err
	}

	if req.AssigneeID != nil && *req.AssigneeID != "" {
		ok, err := s.projectRepo.IsMember(ctx, projectID, *req.AssigneeID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrAssigneeNotMember
		}
	}

	priority := req.Priority
	if priority == "" {
		priority = model.TaskPriorityMedium
	}

	// position is required at the column level, but the client only needs
	// to compute one when it cares about exact placement on the Kanban board.
	// For the primary data-table flow we default to a monotone time-based
	// string so the new task naturally sorts at the end of its column.
	position := req.Position
	if position == "" {
		position = defaultPosition()
	}

	creator := userID
	t := &model.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskStatusTodo,
		Priority:    priority,
		Position:    position,
		ProjectID:   projectID,
		AssigneeID:  normalizeUUIDPtr(req.AssigneeID),
		CreatedBy:   &creator,
		DueDate:     req.DueDate,
	}

	if err := s.taskRepo.Create(ctx, t); err != nil {
		return nil, err
	}

	s.notify(ctx, t.ProjectID, notifier.Event{
		Type:      notifier.EventTaskCreated,
		TaskID:    t.ID,
		ProjectID: t.ProjectID,
		Task:      t,
	})

	return t, nil
}

func (s *taskService) GetByID(ctx context.Context, userID, taskID string) (*model.Task, error) {
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	isMember, err := s.projectRepo.IsMember(ctx, t.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrTaskNotFound
	}

	return t, nil
}

func (s *taskService) List(ctx context.Context, userID, projectID string, filter model.TaskFilter) ([]*model.Task, int, error) {
	if _, err := s.projectRepo.GetByID(ctx, projectID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, 0, ErrProjectNotFound
		}
		return nil, 0, err
	}

	isMember, err := s.projectRepo.IsMember(ctx, projectID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !isMember {
		return nil, 0, ErrProjectNotFound
	}

	if filter.Status != "" && !isValidTaskStatus(filter.Status) {
		return nil, 0, fmt.Errorf("%w: status must be 'todo', 'in_progress', or 'done'", ErrValidation)
	}
	if filter.Priority != "" && !isValidTaskPriority(filter.Priority) {
		return nil, 0, fmt.Errorf("%w: priority must be 'low', 'medium', or 'high'", ErrValidation)
	}

	return s.taskRepo.List(ctx, projectID, filter)
}

func (s *taskService) ListAll(ctx context.Context, userID string, filter model.TaskFilter) ([]*model.Task, int, error) {
	if filter.Status != "" && !isValidTaskStatus(filter.Status) {
		return nil, 0, fmt.Errorf("%w: status must be 'todo', 'in_progress', or 'done'", ErrValidation)
	}
	if filter.Priority != "" && !isValidTaskPriority(filter.Priority) {
		return nil, 0, fmt.Errorf("%w: priority must be 'low', 'medium', or 'high'", ErrValidation)
	}
	return s.taskRepo.ListAll(ctx, userID, filter)
}

const bulkDeleteTasksMaxIDs = 100

func (s *taskService) BulkDelete(ctx context.Context, userID string, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	if len(ids) > bulkDeleteTasksMaxIDs {
		return 0, fmt.Errorf("%w: at most %d ids per request", ErrValidation, bulkDeleteTasksMaxIDs)
	}
	return s.taskRepo.BulkDelete(ctx, userID, ids)
}

func (s *taskService) Board(ctx context.Context, userID, projectID string) (*model.BoardView, error) {
	if _, err := s.projectRepo.GetByID(ctx, projectID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	isMember, err := s.projectRepo.IsMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrProjectNotFound
	}

	tasks, err := s.taskRepo.BoardList(ctx, projectID)
	if err != nil {
		return nil, err
	}

	view := &model.BoardView{
		Todo:       make([]*model.Task, 0),
		InProgress: make([]*model.Task, 0),
		Done:       make([]*model.Task, 0),
	}
	for _, t := range tasks {
		switch t.Status {
		case model.TaskStatusTodo:
			view.Todo = append(view.Todo, t)
		case model.TaskStatusInProgress:
			view.InProgress = append(view.InProgress, t)
		case model.TaskStatusDone:
			view.Done = append(view.Done, t)
		}
	}
	return view, nil
}

func (s *taskService) Move(ctx context.Context, userID, taskID string, req *model.MoveTaskRequest) (*model.Task, error) {
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	isMember, err := s.projectRepo.IsMember(ctx, t.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrTaskNotFound
	}

	if !isValidTaskStatus(req.Status) {
		return nil, fmt.Errorf("%w: status must be 'todo', 'in_progress', or 'done'", ErrValidation)
	}
	if !isValidPosition(req.Position) {
		return nil, fmt.Errorf("%w: position is required and must be ASCII printable", ErrValidation)
	}

	updated, prevStatus, err := s.taskRepo.Move(ctx, taskID, req.Status, req.Position)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	if prevStatus != updated.Status {
		changedBy := userID
		if err := s.taskRepo.LogStatusChange(ctx, updated.ID, &changedBy, prevStatus, updated.Status); err != nil {
			return nil, err
		}
	}

	s.notify(ctx, updated.ProjectID, notifier.Event{
		Type:      notifier.EventTaskMoved,
		TaskID:    updated.ID,
		ProjectID: updated.ProjectID,
		Task:      updated,
	})

	return updated, nil
}

func (s *taskService) Update(ctx context.Context, userID, taskID string, req *model.UpdateTaskRequest) (*model.Task, error) {
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	isMember, err := s.projectRepo.IsMember(ctx, t.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrTaskNotFound
	}

	if err := validateUpdateTask(req); err != nil {
		return nil, err
	}

	if req.AssigneeID != nil && *req.AssigneeID != "" {
		ok, err := s.projectRepo.IsMember(ctx, t.ProjectID, *req.AssigneeID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrAssigneeNotMember
		}
	}

	prevStatus := t.Status
	prevDue := t.DueDate

	t.Title = req.Title
	t.Description = req.Description
	t.Status = req.Status
	t.Priority = req.Priority
	t.AssigneeID = normalizeUUIDPtr(req.AssigneeID)
	t.DueDate = req.DueDate

	if err := s.taskRepo.Update(ctx, t); err != nil {
		return nil, err
	}

	if prevStatus != t.Status {
		changedBy := userID
		if err := s.taskRepo.LogStatusChange(ctx, t.ID, &changedBy, prevStatus, t.Status); err != nil {
			return nil, err
		}
	}

	// A new due_date invalidates any reminder that was already fired against
	// the old schedule — let the scheduler send fresh ones.
	if !sameDueDate(prevDue, t.DueDate) {
		_ = s.taskRepo.ClearReminders(ctx, t.ID)
	}

	s.notify(ctx, t.ProjectID, notifier.Event{
		Type:      notifier.EventTaskUpdated,
		TaskID:    t.ID,
		ProjectID: t.ProjectID,
		Task:      t,
	})

	return t, nil
}

func (s *taskService) Delete(ctx context.Context, userID, taskID string) error {
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrTaskNotFound
		}
		return err
	}

	p, err := s.projectRepo.GetByID(ctx, t.ProjectID)
	if err != nil {
		return err
	}

	isMember, err := s.projectRepo.IsMember(ctx, t.ProjectID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrTaskNotFound
	}

	if p.OwnerID != userID && (t.CreatedBy == nil || *t.CreatedBy != userID) {
		return ErrForbidden
	}

	if err := s.taskRepo.Delete(ctx, taskID); err != nil {
		return err
	}

	s.notify(ctx, t.ProjectID, notifier.Event{
		Type:      notifier.EventTaskDeleted,
		TaskID:    taskID,
		ProjectID: t.ProjectID,
	})

	return nil
}

func (s *taskService) Assign(ctx context.Context, userID, taskID string, req *model.AssignTaskRequest) (*model.Task, error) {
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	isMember, err := s.projectRepo.IsMember(ctx, t.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrTaskNotFound
	}

	assignee := normalizeUUIDPtr(req.AssigneeID)
	if assignee != nil {
		ok, err := s.projectRepo.IsMember(ctx, t.ProjectID, *assignee)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrAssigneeNotMember
		}
	}

	prevAssignee := t.AssigneeID

	if err := s.taskRepo.UpdateAssignee(ctx, taskID, assignee); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	t.AssigneeID = assignee

	// Assignment change means the new assignee hasn't been warned about the
	// deadline yet — reset the per-task reminder flags so they fire again.
	if !sameAssignee(prevAssignee, t.AssigneeID) {
		_ = s.taskRepo.ClearReminders(ctx, t.ID)
	}

	s.notify(ctx, t.ProjectID, notifier.Event{
		Type:      notifier.EventTaskAssigned,
		TaskID:    t.ID,
		ProjectID: t.ProjectID,
		Task:      t,
	})

	return t, nil
}

func (s *taskService) GetActivityLogs(ctx context.Context, userID, taskID string) ([]*model.TaskActivityLog, error) {
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	isMember, err := s.projectRepo.IsMember(ctx, t.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrTaskNotFound
	}

	return s.taskRepo.GetActivityLogs(ctx, taskID)
}

func validateCreateTask(req *model.CreateTaskRequest) error {
	if len(req.Title) == 0 {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if len(req.Title) > 500 {
		return fmt.Errorf("%w: title must be at most 500 characters", ErrValidation)
	}
	if req.Priority != "" && !isValidTaskPriority(req.Priority) {
		return fmt.Errorf("%w: priority must be 'low', 'medium', or 'high'", ErrValidation)
	}
	// Position is optional on create — server defaults it. But if the client
	// does send one (Kanban-create flow), it must be well-formed so it can't
	// poison the (project_id, status, position) ordering.
	if req.Position != "" && !isValidPosition(req.Position) {
		return fmt.Errorf("%w: position must be ASCII printable", ErrValidation)
	}
	return nil
}

func validateUpdateTask(req *model.UpdateTaskRequest) error {
	if len(req.Title) == 0 {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if len(req.Title) > 500 {
		return fmt.Errorf("%w: title must be at most 500 characters", ErrValidation)
	}
	if !isValidTaskStatus(req.Status) {
		return fmt.Errorf("%w: status must be 'todo', 'in_progress', or 'done'", ErrValidation)
	}
	if !isValidTaskPriority(req.Priority) {
		return fmt.Errorf("%w: priority must be 'low', 'medium', or 'high'", ErrValidation)
	}
	return nil
}

func isValidTaskStatus(s model.TaskStatus) bool {
	switch s {
	case model.TaskStatusTodo, model.TaskStatusInProgress, model.TaskStatusDone:
		return true
	}
	return false
}

func isValidTaskPriority(p model.TaskPriority) bool {
	switch p {
	case model.TaskPriorityLow, model.TaskPriorityMedium, model.TaskPriorityHigh:
		return true
	}
	return false
}

// isValidPosition guards the position string at the API boundary. The client
// computes Lexorank, but the server still rejects empty or non-printable
// values so a malformed payload can't poison the (project_id, status, position)
// index ordering.
func isValidPosition(p string) bool {
	if p == "" || len(p) > 255 {
		return false
	}
	for _, r := range p {
		if r < 0x21 || r > 0x7e {
			return false
		}
	}
	return true
}

// defaultPosition produces a monotone time-based position string for tasks
// created without an explicit position. Nanoseconds since epoch zero-padded
// to a fixed width sorts lexicographically the same as chronologically, so
// new tasks land at the end of their column. The Kanban client's Lexorank
// can still insert *between* two of these (e.g. by adding suffix chars).
func defaultPosition() string {
	return fmt.Sprintf("%020d", time.Now().UTC().UnixNano())
}

func normalizeUUIDPtr(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	v := *s
	return &v
}

func sameDueDate(a, b *time.Time) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return a.Equal(*b)
	}
}

func sameAssignee(a, b *string) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}
