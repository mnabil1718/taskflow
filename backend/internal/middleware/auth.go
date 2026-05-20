package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/response"
	"github.com/mnabil1718/taskflow/internal/service"
)

func JWTProtected(svc service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return response.Error(c, fiber.StatusUnauthorized, "missing or invalid authorization header")
		}

		claims, err := svc.ValidateAccessToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "invalid or expired token")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		return c.Next()
	}
}
