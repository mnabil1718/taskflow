package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/response"
	"github.com/mnabil1718/taskflow/internal/service"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	tokens, err := h.svc.Register(c.Context(), &req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "registration successful", tokens)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	tokens, err := h.svc.Login(c.Context(), &req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "login successful", tokens)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	if req.RefreshToken == "" {
		return response.Error(c, fiber.StatusBadRequest, "refresh_token is required")
	}

	if err := h.svc.Logout(c.Context(), req.RefreshToken); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "logout failed")
	}

	return response.Success(c, fiber.StatusOK, "logged out successfully", nil)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	if req.RefreshToken == "" {
		return response.Error(c, fiber.StatusBadRequest, "refresh_token is required")
	}

	tokens, err := h.svc.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "token refreshed", tokens)
}

func (h *AuthHandler) handleServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrValidation):
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrEmailTaken):
		return response.Error(c, fiber.StatusConflict, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		return response.Error(c, fiber.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrTokenInvalid):
		return response.Error(c, fiber.StatusUnauthorized, err.Error())
	default:
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}
}
