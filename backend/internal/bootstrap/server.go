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

func NewServer(cfg *config.Config) *Server {
	return &Server{
		app:  newApp(),
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

	return s.app.ShutdownWithTimeout(10 * time.Second)
}
