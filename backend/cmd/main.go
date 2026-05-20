package main

import (
	"log/slog"
	"os"

	"github.com/mnabil1718/taskflow/internal/config"
	"github.com/mnabil1718/taskflow/internal/database"
	"github.com/mnabil1718/taskflow/internal/logger"
	"github.com/mnabil1718/taskflow/internal/seeder"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Init(cfg.App.Env)

	db, err := database.NewDB(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.RunMigrations(cfg, cfg.App.MigrationsPath); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	if cfg.App.Env == "development" {
		if err := seeder.Run(db); err != nil {
			slog.Error("failed to seed database", "error", err)
			os.Exit(1)
		}
	}

	server := initServer(cfg, db)

	if err := server.Run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}
