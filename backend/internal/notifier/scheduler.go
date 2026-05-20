package notifier

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/mnabil1718/taskflow/internal/repository"
)

// defaultSchedulerInterval controls how often the scheduler scans for
// tasks whose 3-day or 1-day reminder is due. One minute is precise
// enough that a freshly-created task with a soon deadline will get its
// reminder within a minute, while keeping the per-tick DB load trivial.
const defaultSchedulerInterval = time.Minute

// reminderKinds is the fixed set of windows the scheduler emits, in the
// order they're scanned each tick. Adding a new window means adding a
// new column + migration plus an entry here.
var reminderKinds = []repository.ReminderKind{
	repository.ReminderThreeDay,
	repository.ReminderOneDay,
}

// DeadlineScheduler periodically scans the tasks table for upcoming
// deadlines and publishes task.deadline_reminder events to the assignee.
// Persistence of "already sent" lives in tasks.reminder_*_sent_at so
// reminders survive restarts and are never sent twice for the same window.
type DeadlineScheduler struct {
	repo     repository.TaskRepository
	hub      *Hub
	interval time.Duration

	stopOnce sync.Once
	stop     chan struct{}
	done     chan struct{}
}

func NewDeadlineScheduler(repo repository.TaskRepository, hub *Hub) *DeadlineScheduler {
	return &DeadlineScheduler{
		repo:     repo,
		hub:      hub,
		interval: defaultSchedulerInterval,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
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
		s.dispatch(ctx, kind)
	}
}

func (s *DeadlineScheduler) dispatch(ctx context.Context, kind repository.ReminderKind) {
	tasks, err := s.repo.PendingReminders(ctx, kind)
	if err != nil {
		slog.Error("scheduler: pending reminders", "kind", kind, "error", err)
		return
	}

	for _, t := range tasks {
		if t.AssigneeID == nil {
			continue
		}

		s.hub.Publish([]string{*t.AssigneeID}, Event{
			Type:           EventTaskDeadlineReminder,
			TaskID:         t.ID,
			ProjectID:      t.ProjectID,
			Task:           t,
			ReminderWindow: string(kind),
		})

		if err := s.repo.MarkReminderSent(ctx, t.ID, kind); err != nil {
			slog.Error("scheduler: mark reminder sent",
				"task_id", t.ID,
				"kind", kind,
				"error", err)
		}
	}
}

