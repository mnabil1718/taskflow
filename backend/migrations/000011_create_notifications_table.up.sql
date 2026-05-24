CREATE TABLE notifications (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID        NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    type            VARCHAR(64) NOT NULL,
    task_id         UUID        REFERENCES tasks(id)             ON DELETE CASCADE,
    project_id      UUID        REFERENCES projects(id)          ON DELETE CASCADE,
    title           TEXT,
    reminder_window VARCHAR(8),
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_created ON notifications (user_id, created_at DESC);
CREATE INDEX idx_notifications_user_unread  ON notifications (user_id) WHERE read_at IS NULL;
