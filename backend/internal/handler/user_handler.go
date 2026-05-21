package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/response"
	"github.com/mnabil1718/taskflow/internal/service"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Search godoc
// @Summary      Search users
// @Description  Returns up to 10 users whose name or email contains the query.
// @Description  The caller is excluded from results — invite pickers never suggest the user to themselves.
// @Tags         users
// @Produce      json
// @Param        q query    string                                  true "Name or email substring (min 1 char)"
// @Success      200 {object} response.Body{data=[]model.User}       "Users matched"
// @Failure      400 {object} response.Body                          "Empty query"
// @Failure      401 {object} response.Body                          "Missing or invalid token"
// @Failure      500 {object} response.Body                          "Internal server error"
// @Security     BearerAuth
// @Router       /users/search [get]
func (h *UserHandler) Search(c *fiber.Ctx) error {
	callerID := c.Locals("user_id").(string)
	q := c.Query("q")

	users, err := h.svc.Search(c.Context(), callerID, q)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}

	return response.Success(c, fiber.StatusOK, "users retrieved", users)
}
