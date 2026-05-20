DROP INDEX IF EXISTS idx_tasks_reminder_1d_due;
DROP INDEX IF EXISTS idx_tasks_reminder_3d_due;

ALTER TABLE tasks
    DROP COLUMN IF EXISTS reminder_1d_sent_at,
    DROP COLUMN IF EXISTS reminder_3d_sent_at;
