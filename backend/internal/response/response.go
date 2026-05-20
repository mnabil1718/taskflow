package response

import "github.com/gofiber/fiber/v2"

// Body is the standard envelope returned by every API endpoint.
// On success, Data holds the payload and Message describes the outcome.
// On failure, Error holds a human-readable message and Data/Message are empty.
type Body struct {
	Data    any    `json:"data,omitempty" swaggertype:"object"`
	Message string `json:"message,omitempty" example:"operation successful"`
	Error   string `json:"error,omitempty" example:"validation error: password must be at least 8 characters"`
}

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(Body{Data: data, Message: message})
}

func Error(c *fiber.Ctx, status int, err string) error {
	return c.Status(status).JSON(Body{Error: err})
}
