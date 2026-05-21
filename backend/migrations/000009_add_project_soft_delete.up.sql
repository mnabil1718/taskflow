ALTER TABLE projects ADD COLUMN deleted_at TIMESTAMPTZ;

-- Partial index keeps lookups on active rows cheap and lets the
-- soft-delete filter (deleted_at IS NULL) stay sargable.
CREATE INDEX idx_projects_active ON projects(id) WHERE deleted_at IS NULL;
