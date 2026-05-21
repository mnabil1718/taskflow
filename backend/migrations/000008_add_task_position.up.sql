-- Add a per-column ordering key for the Kanban board.
-- Stored as text so the client's Lexorank algorithm can pick a value
-- strictly between two neighbors without ever renumbering siblings.
ALTER TABLE tasks
    ADD COLUMN position TEXT;

-- Backfill: per (project_id, status) bucket, lay existing tasks down on
-- a sparse numeric grid (`00001000`, `00002000`, ...) ordered by creation
-- time. Lexicographic order matches numeric order because every value has
-- the same width, and the step of 1000 leaves room for the client to
-- insert between any two rows many times before space runs out.
UPDATE tasks t
SET position = lpad((rn * 1000)::text, 8, '0')
FROM (
    SELECT id,
           row_number() OVER (PARTITION BY project_id, status ORDER BY created_at, id) AS rn
    FROM tasks
) ranked
WHERE t.id = ranked.id;

ALTER TABLE tasks
    ALTER COLUMN position SET NOT NULL;

-- Composite index covering the exact board read pattern:
-- WHERE project_id = ? ORDER BY status, position.
CREATE INDEX idx_tasks_project_status_position ON tasks(project_id, status, position);
