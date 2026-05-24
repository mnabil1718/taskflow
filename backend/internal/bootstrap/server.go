package bootstrap

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/config"
	"github.com/mnabil1718/taskflow/internal/service"
)

type Server struct {
	app       *fiber.App
	port      string
	scheduler *service.DeadlineScheduler
}

func NewServer(cfg *config.Config, app *fiber.App, scheduler *service.DeadlineScheduler) *Server {
	return &Server{
		app:       app,
		port:      cfg.App.Port,
		scheduler: scheduler,
	}
}

func (s *Server) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	s.scheduler.Start()
	slog.Info("deadline reminder scheduler started")

	go func() {
		if err := s.app.Listen(":" + s.port); err != nil {
			slog.Error("server error", "error", err)
		}
	}()

	<-quit
	slog.Info("shutting down server...")

	s.scheduler.Stop()

	if err := s.app.ShutdownWithTimeout(10 * time.Second); err != nil {
		slog.Warn("shutdown timeout exceeded, forcing exit", "error", err)
		return err
	}

	slog.Info("server shutdown complete")
	return nil
}
