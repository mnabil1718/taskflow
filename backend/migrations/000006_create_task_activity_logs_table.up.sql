CREATE TABLE task_activity_logs (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id     UUID        NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    changed_by  UUID        REFERENCES users(id) ON DELETE SET NULL,
    from_status task_status NOT NULL,
    to_status   task_status NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_activity_logs_task_id ON task_activity_logs(task_id);
