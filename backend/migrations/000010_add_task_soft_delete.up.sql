ALTER TABLE tasks ADD COLUMN deleted_at TIMESTAMPTZ;

-- Partial index keeps queries on active rows cheap and lets the soft-delete
-- filter (deleted_at IS NULL) stay sargable on every task read path.
CREATE INDEX idx_tasks_active ON tasks(id) WHERE deleted_at IS NULL;
