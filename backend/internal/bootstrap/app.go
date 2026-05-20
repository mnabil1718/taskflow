package bootstrap

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/mnabil1718/taskflow/internal/handler"
	"github.com/mnabil1718/taskflow/internal/response"
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

func rateLimiter(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.Error(c, fiber.StatusTooManyRequests, "too many requests, please try again later")
		},
	})
}

func registerRoutes(app *fiber.App, health *handler.HealthHandler, auth *handler.AuthHandler) {
	app.Get("/health", health.Check)

	// 100 requests/min per IP for all API routes
	v1 := app.Group("/api/v1", rateLimiter(100, time.Minute))

	// 10 requests/min per IP on auth to prevent brute-force
	authGroup := v1.Group("/auth", rateLimiter(10, time.Minute))
	authGroup.Post("/register", auth.Register)
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/logout", auth.Logout)
	authGroup.Post("/refresh", auth.Refresh)
}
