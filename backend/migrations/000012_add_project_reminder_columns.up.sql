ALTER TABLE projects
    ADD COLUMN reminder_3d_sent_at TIMESTAMPTZ,
    ADD COLUMN reminder_1d_sent_at TIMESTAMPTZ;

CREATE INDEX idx_projects_reminder_3d_due
    ON projects (deadline)
    WHERE reminder_3d_sent_at IS NULL AND deadline IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX idx_projects_reminder_1d_due
    ON projects (deadline)
    WHERE reminder_1d_sent_at IS NULL AND deadline IS NOT NULL AND deleted_at IS NULL;
