DROP INDEX IF EXISTS idx_projects_active;
ALTER TABLE projects DROP COLUMN IF EXISTS deleted_at;
