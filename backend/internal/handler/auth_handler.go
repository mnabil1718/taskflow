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

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account and returns an access/refresh token pair.
// @Description  Validation rules: name >= 2 chars, valid email, password >= 8 chars.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.RegisterRequest true "Registration payload"
// @Success      201 {object} response.Body{data=model.TokenPair} "Registration successful"
// @Failure      400 {object} response.Body "Validation error or malformed JSON body"
// @Failure      409 {object} response.Body "Email already registered"
// @Failure      500 {object} response.Body "Internal server error"
// @Router       /auth/register [post]
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

// Login godoc
// @Summary      Log in with email and password
// @Description  Authenticates a user and returns an access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.LoginRequest true "Login payload"
// @Success      200 {object} response.Body{data=model.TokenPair} "Login successful"
// @Failure      400 {object} response.Body "Missing fields or malformed JSON body"
// @Failure      401 {object} response.Body "Invalid email or password"
// @Failure      500 {object} response.Body "Internal server error"
// @Router       /auth/login [post]
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

// Logout godoc
// @Summary      Log out the current session
// @Description  Invalidates the supplied refresh token. Idempotent: returns 200
// @Description  even if the token does not exist.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.RefreshRequest true "Refresh token to invalidate"
// @Success      200 {object} response.Body "Logged out successfully"
// @Failure      400 {object} response.Body "Missing refresh_token or malformed JSON body"
// @Failure      500 {object} response.Body "Internal server error"
// @Router       /auth/logout [post]
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

// Refresh godoc
// @Summary      Exchange a refresh token for a new token pair
// @Description  Rotates the refresh token: the supplied token is deleted and a
// @Description  new access/refresh pair is issued.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.RefreshRequest true "Current refresh token"
// @Success      200 {object} response.Body{data=model.TokenPair} "Token refreshed"
// @Failure      400 {object} response.Body "Missing refresh_token or malformed JSON body"
// @Failure      401 {object} response.Body "Invalid or expired refresh token"
// @Failure      500 {object} response.Body "Internal server error"
// @Router       /auth/refresh [post]
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
