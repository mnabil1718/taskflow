DROP INDEX IF EXISTS idx_projects_reminder_1d_due;
DROP INDEX IF EXISTS idx_projects_reminder_3d_due;

ALTER TABLE projects
    DROP COLUMN IF EXISTS reminder_1d_sent_at,
    DROP COLUMN IF EXISTS reminder_3d_sent_at;
