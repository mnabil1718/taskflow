package bootstrap

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/mnabil1718/taskflow/internal/handler"
)

func NewApp(health *handler.HealthHandler, auth *handler.AuthHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadBufferSize: 16 * 1024,
	})

	app.Use(logger.New(logger.Config{
		Format: "${time} | ${status} | ${latency} | ${method} ${path}\n",
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/health"
		},
	}))

	registerRoutes(app, health, auth)
	return app
}

func registerRoutes(app *fiber.App, health *handler.HealthHandler, auth *handler.AuthHandler) {
	app.Get("/health", health.Check)

	v1 := app.Group("/api/v1")

	authGroup := v1.Group("/auth")
	authGroup.Post("/register", auth.Register)
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/logout", auth.Logout)
	authGroup.Post("/refresh", auth.Refresh)
}
