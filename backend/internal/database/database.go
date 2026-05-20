package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mnabil1718/taskflow/internal/config"
)

func NewDB(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}

// RunMigrations uses its own connection so closing the migrator does not
// affect the application's *sql.DB pool.
func RunMigrations(cfg *config.Config, migrationsPath string) error {
	u := &url.URL{
		Scheme: "pgx5",
		User:   url.UserPassword(cfg.DB.User, cfg.DB.Password),
		Host:   cfg.DB.Host + ":" + strconv.Itoa(cfg.DB.Port),
		Path:   "/" + cfg.DB.Name,
	}
	q := url.Values{}
	q.Set("sslmode", cfg.DB.SSLMode)
	u.RawQuery = q.Encode()

	m, err := migrate.New("file://"+migrationsPath, u.String())
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	slog.Info("migrations applied successfully")
	return nil
}
