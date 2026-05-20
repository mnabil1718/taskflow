package bootstrap

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app  *fiber.App
	port string
}

func NewServer() *Server {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	return &Server{
		app:  newApp(),
		port: port,
	}
}

func (s *Server) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.app.Listen(":" + s.port); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	return s.app.ShutdownWithTimeout(10 * time.Second)
}
