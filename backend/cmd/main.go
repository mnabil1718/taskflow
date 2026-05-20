package main

import (
	"log/slog"
	"os"

	_ "github.com/mnabil1718/taskflow/docs"
	"github.com/mnabil1718/taskflow/internal/config"
	"github.com/mnabil1718/taskflow/internal/database"
	"github.com/mnabil1718/taskflow/internal/logger"
	"github.com/mnabil1718/taskflow/internal/seeder"
)

// @title           TaskFlow API
// @version         1.0
// @description     Task management REST API with JWT authentication.
// @description     All responses follow the shape { data, message, error }.
// @termsOfService  http://swagger.io/terms/

// @contact.name   TaskFlow Maintainers
// @contact.url    https://github.com/mnabil1718/taskflow

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1
// @schemes   http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT access token.
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
