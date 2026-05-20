package testutil

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestDB wraps a *sql.DB connected to the integration test database.
type TestDB struct {
	DB *sql.DB
}

// Open connects to the test database (TEST_DB_* env vars, defaulting to the
// db-test docker-compose service on port 5434) and runs all pending migrations.
func Open() (*TestDB, error) {
	host := getenv("TEST_DB_HOST", "localhost")
	port := getenv("TEST_DB_PORT", "5434")
	user := getenv("TEST_DB_USER", "taskflow_test")
	password := getenv("TEST_DB_PASSWORD", "taskflow_test")
	name := getenv("TEST_DB_NAME", "taskflow_test")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, name,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open test db: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping test db: %w", err)
	}

	tdb := &TestDB{DB: db}
	if err := tdb.runMigrations(host, port, user, password, name); err != nil {
		db.Close()
		return nil, err
	}

	return tdb, nil
}

// Truncate removes all rows from every table so each test starts clean.
func (tdb *TestDB) Truncate(t *testing.T) {
	t.Helper()
	const q = `TRUNCATE TABLE task_activity_logs, refresh_tokens, tasks,
		project_members, projects, users RESTART IDENTITY CASCADE`
	if _, err := tdb.DB.Exec(q); err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

// Close closes the underlying database connection.
func (tdb *TestDB) Close() {
	tdb.DB.Close()
}

func (tdb *TestDB) runMigrations(host, port, user, password, name string) error {
	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "../../migrations")

	u := &url.URL{
		Scheme:   "pgx5",
		User:     url.UserPassword(user, password),
		Host:     host + ":" + port,
		Path:     "/" + name,
		RawQuery: "sslmode=disable",
	}

	m, err := migrate.New("file://"+migrationsPath, u.String())
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
