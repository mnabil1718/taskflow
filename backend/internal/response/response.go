package response

import "github.com/gofiber/fiber/v2"

type Body struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(Body{Data: data, Message: message})
}

func Error(c *fiber.Ctx, status int, err string) error {
	return c.Status(status).JSON(Body{Error: err})
}
