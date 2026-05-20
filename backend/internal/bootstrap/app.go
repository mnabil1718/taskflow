package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/handler"
)

func NewApp(health *handler.HealthHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadBufferSize: 16 * 1024,
	})

	registerRoutes(app, health)
	return app
}

func registerRoutes(app *fiber.App, health *handler.HealthHandler) {
	app.Get("/health", health.Check)

	v1 := app.Group("/api/v1")
	_ = v1 // auth, project, task route groups attach here
}
