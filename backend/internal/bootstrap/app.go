package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/handler"
)

func newApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ReadBufferSize: 16 * 1024,
	})

	registerRoutes(app)
	return app
}

func registerRoutes(app *fiber.App) {
	health := handler.NewHealthHandler()
	app.Get("/health", health.Check)
}
