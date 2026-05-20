package seeder

import (
	"database/sql"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

// Fixed UUIDs keep the seed deterministic and idempotency check simple.
const (
	user1ID    = "00000000-0000-0000-0000-000000000001"
	user2ID    = "00000000-0000-0000-0000-000000000002"
	project1ID = "00000000-0000-0000-0000-000000000011"
	project2ID = "00000000-0000-0000-0000-000000000012"
	task1ID    = "00000000-0000-0000-0000-000000000021"
	task2ID    = "00000000-0000-0000-0000-000000000022"
	task3ID    = "00000000-0000-0000-0000-000000000023"
	task4ID    = "00000000-0000-0000-0000-000000000024"
	task5ID    = "00000000-0000-0000-0000-000000000025"
)

func Run(db *sql.DB) error {
	var count int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE id IN ($1, $2)", user1ID, user2ID,
	).Scan(&count); err != nil {
		return fmt.Errorf("seeder check: %w", err)
	}
	if count > 0 {
		slog.Info("seed data already present, skipping")
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	pw := string(hash)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback()

	if err := seedUsers(tx, pw); err != nil {
		return err
	}
	if err := seedProjects(tx); err != nil {
		return err
	}
	if err := seedProjectMembers(tx); err != nil {
		return err
	}
	if err := seedTasks(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed tx: %w", err)
	}

	slog.Info("seed data inserted successfully")
	return nil
}

func seedUsers(tx *sql.Tx, passwordHash string) error {
	_, err := tx.Exec(`
		INSERT INTO users (id, name, email, password_hash) VALUES
		($1, 'Alice Smith', 'alice@example.com', $3),
		($2, 'Bob Jones',   'bob@example.com',   $3)
	`, user1ID, user2ID, passwordHash)
	if err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	return nil
}

func seedProjects(tx *sql.Tx) error {
	_, err := tx.Exec(`
		INSERT INTO projects (id, name, description, status, deadline, owner_id) VALUES
		($1, 'Project Alpha', 'Backend API and infrastructure work',
		 'active', NOW() + INTERVAL '30 days', $3),
		($2, 'Project Beta',  'Frontend and dashboard features',
		 'active', NOW() + INTERVAL '60 days', $4)
	`, project1ID, project2ID, user1ID, user2ID)
	if err != nil {
		return fmt.Errorf("seed projects: %w", err)
	}
	return nil
}

func seedProjectMembers(tx *sql.Tx) error {
	_, err := tx.Exec(`
		INSERT INTO project_members (project_id, user_id, role) VALUES
		($1, $3, 'owner'),  ($1, $4, 'member'),
		($2, $4, 'owner'),  ($2, $3, 'member')
	`, project1ID, project2ID, user1ID, user2ID)
	if err != nil {
		return fmt.Errorf("seed project_members: %w", err)
	}
	return nil
}

func seedTasks(tx *sql.Tx) error {
	_, err := tx.Exec(`
		INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, created_by) VALUES
		($1, 'Set up CI/CD pipeline',      'Configure GitHub Actions workflows',   'todo',        'high',   $6, $8, $8),
		($2, 'Design database schema',     'Create ER diagram and migrations',     'done',        'high',   $6, $8, $8),
		($3, 'Implement auth endpoints',   'Register, login, refresh token flow',  'in_progress', 'high',   $6, $9, $8),
		($4, 'Build dashboard UI',         'Task summary charts per project',      'todo',        'medium', $7, $9, $9),
		($5, 'Write API documentation',    'Document all REST endpoints',          'todo',        'low',    $7, $8, $9)
	`, task1ID, task2ID, task3ID, task4ID, task5ID,
		project1ID, project2ID, user1ID, user2ID)
	if err != nil {
		return fmt.Errorf("seed tasks: %w", err)
	}
	return nil
}
