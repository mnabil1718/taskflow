package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/mnabil1718/taskflow/internal/repository"
)

// defaultSchedulerInterval controls how often the scheduler scans for tasks
// and projects whose 3-day or 1-day reminder is due. One minute is precise
// enough that a freshly-created item with a soon deadline gets its reminder
// within a minute, while keeping the per-tick DB load trivial.
const defaultSchedulerInterval = time.Minute

// reminderKinds is the fixed set of windows the scheduler emits, in the order
// they're scanned each tick. Adding a window means adding a column + migration
// plus an entry here.
var reminderKinds = []repository.ReminderKind{
	repository.ReminderThreeDay,
	repository.ReminderOneDay,
}

// DeadlineScheduler periodically scans for upcoming task and project deadlines
// and creates notifications for them. "Already sent" state is persisted in the
// tasks/projects reminder_*_sent_at columns so reminders survive restarts and
// are never sent twice for the same (entity, window). It lives in the service
// layer (not the notifier package) because it now persists notifications via
// NotificationService — keeping it here avoids a service<->notifier import
// cycle and keeps the notifier package as pure SSE transport.
type DeadlineScheduler struct {
	taskRepo    repository.TaskRepository
	projectRepo repository.ProjectRepository
	notif       NotificationService
	interval    time.Duration

	stopOnce sync.Once
	stop     chan struct{}
	done     chan struct{}
}

func NewDeadlineScheduler(taskRepo repository.TaskRepository, projectRepo repository.ProjectRepository, notif NotificationService) *DeadlineScheduler {
	return &DeadlineScheduler{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		notif:       notif,
		interval:    defaultSchedulerInterval,
		stop:        make(chan struct{}),
		done:        make(chan struct{}),
	}
}

// Start launches the scheduler goroutine. Safe to call once.
func (s *DeadlineScheduler) Start() {
	go s.run()
}

// Stop signals the scheduler to exit and waits for the goroutine to drain.
// Safe to call multiple times.
func (s *DeadlineScheduler) Stop() {
	s.stopOnce.Do(func() { close(s.stop) })
	<-s.done
}

func (s *DeadlineScheduler) run() {
	defer close(s.done)

	s.tick()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *DeadlineScheduler) tick() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, kind := range reminderKinds {
		s.dispatchTasks(ctx, kind)
		s.dispatchProjects(ctx, kind)
	}
}

func (s *DeadlineScheduler) dispatchTasks(ctx context.Context, kind repository.ReminderKind) {
	tasks, err := s.taskRepo.PendingReminders(ctx, kind)
	if err != nil {
		slog.Error("scheduler: pending task reminders", "kind", kind, "error", err)
		return
	}

	for _, t := range tasks {
		if t.AssigneeID == nil {
			continue
		}

		s.notif.NotifyTaskDeadline(ctx, *t.AssigneeID, t, string(kind))

		if err := s.taskRepo.MarkReminderSent(ctx, t.ID, kind); err != nil {
			slog.Error("scheduler: mark task reminder sent", "task_id", t.ID, "kind", kind, "error", err)
		}
	}
}

func (s *DeadlineScheduler) dispatchProjects(ctx context.Context, kind repository.ReminderKind) {
	projects, err := s.projectRepo.PendingProjectReminders(ctx, kind)
	if err != nil {
		slog.Error("scheduler: pending project reminders", "kind", kind, "error", err)
		return
	}

	for _, p := range projects {
		members, err := s.projectRepo.GetMembers(ctx, p.ID)
		if err != nil {
			slog.Error("scheduler: project members", "project_id", p.ID, "error", err)
			continue
		}

		for _, m := range members {
			s.notif.NotifyProjectDeadline(ctx, m.UserID, p, string(kind))
		}

		if err := s.projectRepo.MarkProjectReminderSent(ctx, p.ID, kind); err != nil {
			slog.Error("scheduler: mark project reminder sent", "project_id", p.ID, "kind", kind, "error", err)
		}
	}
}
