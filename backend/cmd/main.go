package main

import (
	"log/slog"
	"os"

	"github.com/mnabil1718/taskflow/internal/config"
	"github.com/mnabil1718/taskflow/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Init(cfg.App.Env)

	server := initServer(cfg)

	if err := server.Run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}
