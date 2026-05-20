package bootstrap

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/config"
)

type Server struct {
	app  *fiber.App
	port string
}

func NewServer(cfg *config.Config, app *fiber.App) *Server {
	return &Server{
		app:  app,
		port: cfg.App.Port,
	}
}

func (s *Server) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.app.Listen(":" + s.port); err != nil {
			slog.Error("server error", "error", err)
		}
	}()

	<-quit
	slog.Info("shutting down server...")

	if err := s.app.ShutdownWithTimeout(10 * time.Second); err != nil {
		slog.Warn("shutdown timeout exceeded, forcing exit", "error", err)
		return err
	}

	slog.Info("server shutdown complete")
	return nil
}
