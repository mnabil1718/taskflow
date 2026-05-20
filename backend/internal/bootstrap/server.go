package bootstrap

import (
	"os"

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
	return s.app.Listen(":" + s.port)
}
