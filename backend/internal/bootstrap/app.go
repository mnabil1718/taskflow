package bootstrap

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/mnabil1718/taskflow/internal/config"
	"github.com/mnabil1718/taskflow/internal/handler"
	"github.com/mnabil1718/taskflow/internal/middleware"
	"github.com/mnabil1718/taskflow/internal/response"
	"github.com/mnabil1718/taskflow/internal/service"
)

func NewApp(cfg *config.Config, authSvc service.AuthService, health *handler.HealthHandler, auth *handler.AuthHandler, project *handler.ProjectHandler, task *handler.TaskHandler, dashboard *handler.DashboardHandler, notif *handler.NotificationHandler, user *handler.UserHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadBufferSize: 16 * 1024,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.App.CORSAllowOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Authorization",
		AllowCredentials: true,
	}))

	app.Use(logger.New(logger.Config{
		Format: "${time} | ${status} | ${latency} | ${method} ${path}\n",
		Next: func(c *fiber.Ctx) bool {
			p := c.Path()
			return p == "/health" || len(p) >= 9 && p[:9] == "/swagger/"
		},
	}))

	registerRoutes(app, authSvc, health, auth, project, task, dashboard, notif, user)
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

func registerRoutes(app *fiber.App, authSvc service.AuthService, health *handler.HealthHandler, auth *handler.AuthHandler, project *handler.ProjectHandler, task *handler.TaskHandler, dashboard *handler.DashboardHandler, notif *handler.NotificationHandler, user *handler.UserHandler) {
	app.Get("/health", health.Check)
	app.Get("/swagger/*", swagger.HandlerDefault)

	// 100 requests/min per IP for all API routes
	v1 := app.Group("/api/v1", rateLimiter(100, time.Minute))

	// 10 requests/min per IP on auth to prevent brute-force
	authGroup := v1.Group("/auth", rateLimiter(10, time.Minute))
	authGroup.Post("/register", auth.Register)
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/logout", auth.Logout)
	authGroup.Post("/refresh", auth.Refresh)

	protected := v1.Group("", middleware.JWTProtected(authSvc))

	projects := protected.Group("/projects")
	projects.Post("", project.Create)
	projects.Get("", project.List)
	projects.Post("/bulk-delete", project.BulkDelete)
	projects.Get("/:id", project.GetByID)
	projects.Put("/:id", project.Update)
	projects.Delete("/:id", project.Delete)
	projects.Get("/:id/members", project.GetMembers)
	projects.Post("/:id/members", project.AddMember)
	projects.Delete("/:id/members/:userID", project.RemoveMember)

	projects.Post("/:id/tasks", task.Create)
	projects.Get("/:id/tasks", task.List)
	projects.Get("/:id/board", task.Board)

	tasks := protected.Group("/tasks")
	tasks.Get("", task.ListAll)
	tasks.Get("/:taskID", task.GetByID)
	tasks.Put("/:taskID", task.Update)
	tasks.Delete("/:taskID", task.Delete)
	tasks.Patch("/:taskID/assign", task.Assign)
	tasks.Patch("/:taskID/move", task.Move)
	tasks.Get("/:taskID/activity", task.GetActivityLogs)

	dash := protected.Group("/dashboard")
	dash.Get("/project-task-counts", dashboard.ProjectTaskCounts)
	dash.Get("/upcoming-tasks", dashboard.UpcomingTasks)

	notifications := protected.Group("/notifications")
	notifications.Get("/stream", notif.Stream)

	users := protected.Group("/users")
	users.Get("/search", user.Search)
}
