//go:build wireinject
// +build wireinject

package main

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/mnabil1718/taskflow/internal/bootstrap"
	"github.com/mnabil1718/taskflow/internal/config"
	"github.com/mnabil1718/taskflow/internal/handler"
	"github.com/mnabil1718/taskflow/internal/repository"
	"github.com/mnabil1718/taskflow/internal/service"
)

func initServer(cfg *config.Config, db *sql.DB) *bootstrap.Server {
	wire.Build(
		repository.NewUserRepository,
		repository.NewTokenRepository,
		repository.NewProjectRepository,
		repository.NewTaskRepository,
		repository.NewDashboardRepository,
		service.NewAuthService,
		service.NewProjectService,
		service.NewTaskService,
		service.NewDashboardService,
		handler.NewHealthHandler,
		handler.NewAuthHandler,
		handler.NewProjectHandler,
		handler.NewTaskHandler,
		handler.NewDashboardHandler,
		bootstrap.NewApp,
		bootstrap.NewServer,
	)
	return nil
}
