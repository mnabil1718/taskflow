package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/notifier"
	"github.com/mnabil1718/taskflow/internal/repository"
	"github.com/mnabil1718/taskflow/internal/service"
)

// ---------------------------------------------------------------------------
// mockTaskRepo
// ---------------------------------------------------------------------------

type statusChange struct {
	taskID    string
	changedBy *string
	from, to  model.TaskStatus
}

type mockTaskRepo struct {
	tasks         map[string]*model.Task
	logs          []statusChange
	moveErr       error
	boardListErr  error
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{tasks: make(map[string]*model.Task)}
}

func (m *mockTaskRepo) Create(_ context.Context, t *model.Task) error {
	t.ID = fmt.Sprintf("task-%d", len(m.tasks)+1)
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	m.tasks[t.ID] = t
	return nil
}

func (m *mockTaskRepo) GetByID(_ context.Context, id string) (*model.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *mockTaskRepo) List(_ context.Context, projectID string, _ model.TaskFilter) ([]*model.Task, int, error) {
	out := make([]*model.Task, 0)
	for _, t := range m.tasks {
		if t.ProjectID == projectID {
			cp := *t
			out = append(out, &cp)
		}
	}
	return out, len(out), nil
}

func (m *mockTaskRepo) BoardList(_ context.Context, projectID string) ([]*model.Task, error) {
	if m.boardListErr != nil {
		return nil, m.boardListErr
	}
	out := make([]*model.Task, 0)
	for _, t := range m.tasks {
		if t.ProjectID == projectID {
			cp := *t
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *mockTaskRepo) Update(_ context.Context, t *model.Task) error {
	if _, ok := m.tasks[t.ID]; !ok {
		return repository.ErrNotFound
	}
	t.UpdatedAt = time.Now()
	cp := *t
	m.tasks[t.ID] = &cp
	return nil
}

func (m *mockTaskRepo) UpdateAssignee(_ context.Context, id string, assigneeID *string) error {
	t, ok := m.tasks[id]
	if !ok {
		return repository.ErrNotFound
	}
	t.AssigneeID = assigneeID
	t.UpdatedAt = time.Now()
	return nil
}

func (m *mockTaskRepo) Move(_ context.Context, id string, newStatus model.TaskStatus, newPosition string) (*model.Task, model.TaskStatus, error) {
	if m.moveErr != nil {
		return nil, "", m.moveErr
	}
	t, ok := m.tasks[id]
	if !ok {
		return nil, "", repository.ErrNotFound
	}
	prev := t.Status
	t.Status = newStatus
	t.Position = newPosition
	t.UpdatedAt = time.Now()
	cp := *t
	return &cp, prev, nil
}

func (m *mockTaskRepo) Delete(_ context.Context, id string) error {
	if _, ok := m.tasks[id]; !ok {
		return repository.ErrNotFound
	}
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepo) LogStatusChange(_ context.Context, taskID string, changedBy *string, from, to model.TaskStatus) error {
	m.logs = append(m.logs, statusChange{taskID: taskID, changedBy: changedBy, from: from, to: to})
	return nil
}

func (m *mockTaskRepo) GetActivityLogs(_ context.Context, taskID string) ([]*model.TaskActivityLog, error) {
	out := make([]*model.TaskActivityLog, 0, len(m.logs))
	for _, l := range m.logs {
		if l.taskID == taskID {
			out = append(out, &model.TaskActivityLog{TaskID: l.taskID, FromStatus: l.from, ToStatus: l.to})
		}
	}
	return out, nil
}

func (m *mockTaskRepo) PendingReminders(_ context.Context, _ repository.ReminderKind) ([]*model.Task, error) {
	return nil, nil
}
func (m *mockTaskRepo) MarkReminderSent(_ context.Context, _ string, _ repository.ReminderKind) error {
	return nil
}
func (m *mockTaskRepo) ClearReminders(_ context.Context, _ string) error { return nil }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTaskSvc(tr *mockTaskRepo, pr *mockProjectRepo) service.TaskService {
	return service.NewTaskService(tr, pr, notifier.NewHub())
}

// seedTask directly writes a task into the mock repo.
func seedTask(tr *mockTaskRepo, id, projectID string, status model.TaskStatus, position string) *model.Task {
	t := &model.Task{
		ID:        id,
		Title:     "test task",
		Status:    status,
		Priority:  model.TaskPriorityMedium,
		Position:  position,
		ProjectID: projectID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	tr.tasks[id] = t
	return t
}

// ---------------------------------------------------------------------------
// Create — position validation
// ---------------------------------------------------------------------------

func TestCreateTask_DefaultsPositionWhenOmitted(t *testing.T) {
	pr := newMockProjectRepo()
	tr := newMockTaskRepo()
	seedProject(pr, "p1", "u1", "Project 1")
	svc := newTaskSvc(tr, pr)

	got, err := svc.Create(context.Background(), "u1", "p1", &model.CreateTaskRequest{
		Title: "no position",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, got.Position, "server must default position when client omits it")
}

func TestCreateTask_RejectsMalformedPosition(t *testing.T) {
	pr := newMockProjectRepo()
	tr := newMockTaskRepo()
	seedProject(pr, "p1", "u1", "Project 1")
	svc := newTaskSvc(tr, pr)

	_, err := svc.Create(context.Background(), "u1", "p1", &model.CreateTaskRequest{
		Title:    "bad position",
		Position: "has space",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, service.ErrValidation)
}

func TestCreateTask_StoresPosition(t *testing.T) {
	pr := newMockProjectRepo()
	tr := newMockTaskRepo()
	seedProject(pr, "p1", "u1", "Project 1")
	svc := newTaskSvc(tr, pr)

	got, err := svc.Create(context.Background(), "u1", "p1", &model.CreateTaskRequest{
		Title:    "with position",
		Position: "00001000",
	})

	require.NoError(t, err)
	assert.Equal(t, "00001000", got.Position)
	assert.Equal(t, model.TaskStatusTodo, got.Status)
}

// ---------------------------------------------------------------------------
// Move
// ---------------------------------------------------------------------------

func TestMoveTask_SameColumn_NoActivityLog(t *testing.T) {
	pr := newMockProjectRepo()
	tr := newMockTaskRepo()
	seedProject(pr, "p1", "u1", "Project 1")
	seedTask(tr, "t1", "p1", model.TaskStatusTodo, "00001000")
	svc := newTaskSvc(tr, pr)

	got, err := svc.Move(context.Background(), "u1", "t1", &model.MoveTaskRequest{
		Status:   model.TaskStatusTodo,
		Position: "00002000",
	})

	require.NoError(t, err)
	assert.Equal(t, model.TaskStatusTodo, got.Status)
	assert.Equal(t, "00002000", got.Position)
	assert.Empty(t, tr.logs, "in-column reorder must not write an activity log")
}

func TestMoveTask_CrossColumn_WritesActivityLog(t *testing.T) {
	pr := newMockProjectRepo()
	tr := newMockTaskRepo()
	seedProject(pr, "p1", "u1", "Project 1")
	seedTask(tr, "t1", "p1", model.TaskStatusTodo, "00001000")
	svc := newTaskSvc(tr, pr)

	got, err := svc.Move(context.Background(), "u1", "t1", &model.MoveTaskRequest{
		Status:   model.TaskStatusInProgress,
		Position: "00001500",
	})

	require.NoError(t, err)
	assert.Equal(t, model.TaskStatusInProgress, got.Status)
	require.Len(t, tr.logs, 1)
	assert.Equal(t, model.TaskStatusTodo, tr.logs[0].from)
	assert.Equal(t, model.TaskStatusInProgress, tr.logs[0].to)
	require.NotNil(t, tr.logs[0].changedBy)
	assert.Equal(t, "u1", *tr.logs[0].changedBy)
}

func TestMoveTask_NotMember_ReturnsTaskNotFound(t *testing.T) {
	pr := newMockProjectRepo()
	tr := newMockTaskRepo()
	seedProject(pr, "p1", "u1", "Project 1")
	seedTask(tr, "t1", "p1", model.TaskStatusTodo, "00001000")
	svc := newTaskSvc(tr, pr)

	_, err := svc.Move(context.Background(), "intruder", "t1", &model.MoveTaskRequest{
		Status:   model.TaskStatusInProgress,
		Position: "00002000",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, service.ErrTaskNotFound)
}

func TestMoveTask_InvalidStatus(t *testing.T) {
	pr := newMockProjectRepo()
	tr := newMockTaskRepo()
	seedProject(pr, "p1", "u1", "Project 1")
	seedTask(tr, "t1", "p1", model.TaskStatusTodo, "00001000")
	svc := newTaskSvc(tr, pr)

	_, err := svc.Move(context.Background(), "u1", "t1", &model.MoveTaskRequest{
		Status:   "archived",
		Position: "00002000",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, service.ErrValidation)
}

func TestMoveTask_EmptyPosition(t *testing.T) {
	pr := newMockProjectRepo()
	tr := newMockTaskRepo()
	seedProject(pr, "p1", "u1", "Project 1")
	seedTask(tr, "t1", "p1", model.TaskStatusTodo, "00001000")
	svc := newTaskSvc(tr, pr)

	_, err := svc.Move(context.Background(), "u1", "t1", &model.MoveTaskRequest{
		Status:   model.TaskStatusTodo,
		Position: "",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, service.ErrValidation)
}

func TestMoveTask_TaskNotFound(t *testing.T) {
	pr := newMockProjectRepo()
	tr := newMockTaskRepo()
	seedProject(pr, "p1", "u1", "Project 1")
	svc := newTaskSvc(tr, pr)

	_, err := svc.Move(context.Background(), "u1", "missing", &model.MoveTaskRequest{
		Status:   model.TaskStatusDone,
		Position: "00001000",
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrTaskNotFound))
}

// ---------------------------------------------------------------------------
// Board
// ---------------------------------------------------------------------------

func TestBoard_GroupsTasksByStatus(t *testing.T) {
	pr := newMockProjectRepo()
	tr := newMockTaskRepo()
	seedProject(pr, "p1", "u1", "Project 1")
	seedTask(tr, "t1", "p1", model.TaskStatusTodo, "00001000")
	seedTask(tr, "t2", "p1", model.TaskStatusInProgress, "00001000")
	seedTask(tr, "t3", "p1", model.TaskStatusDone, "00001000")
	seedTask(tr, "t4", "p1", model.TaskStatusTodo, "00002000")
	svc := newTaskSvc(tr, pr)

	board, err := svc.Board(context.Background(), "u1", "p1")

	require.NoError(t, err)
	assert.Len(t, board.Todo, 2)
	assert.Len(t, board.InProgress, 1)
	assert.Len(t, board.Done, 1)
}

func TestBoard_NotMember_ReturnsProjectNotFound(t *testing.T) {
	pr := newMockProjectRepo()
	tr := newMockTaskRepo()
	seedProject(pr, "p1", "u1", "Project 1")
	svc := newTaskSvc(tr, pr)

	_, err := svc.Board(context.Background(), "intruder", "p1")

	require.Error(t, err)
	assert.ErrorIs(t, err, service.ErrProjectNotFound)
}
