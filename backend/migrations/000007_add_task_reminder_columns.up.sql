ALTER TABLE tasks
    ADD COLUMN reminder_3d_sent_at TIMESTAMPTZ,
    ADD COLUMN reminder_1d_sent_at TIMESTAMPTZ;

CREATE INDEX idx_tasks_reminder_3d_due
    ON tasks (due_date)
    WHERE reminder_3d_sent_at IS NULL AND status != 'done' AND assignee_id IS NOT NULL;

CREATE INDEX idx_tasks_reminder_1d_due
    ON tasks (due_date)
    WHERE reminder_1d_sent_at IS NULL AND status != 'done' AND assignee_id IS NOT NULL;
