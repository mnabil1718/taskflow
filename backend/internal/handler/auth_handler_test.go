//go:build integration

package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mnabil1718/taskflow/internal/config"
	"github.com/mnabil1718/taskflow/internal/handler"
	"github.com/mnabil1718/taskflow/internal/repository"
	"github.com/mnabil1718/taskflow/internal/service"
	"github.com/mnabil1718/taskflow/internal/testutil"
)

var integrationDB *testutil.TestDB

func TestMain(m *testing.M) {
	var err error
	integrationDB, err = testutil.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testutil.Open: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	integrationDB.Close()
	os.Exit(code)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func integrationCfg() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:  "test-access-secret-at-least-32-chars",
			RefreshSecret: "test-refresh-secret-at-least-32-ch",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
	}
}

func buildApp(db *sql.DB) *fiber.App {
	cfg := integrationCfg()
	svc := service.NewAuthService(
		repository.NewUserRepository(db),
		repository.NewTokenRepository(db),
		cfg,
	)
	h := handler.NewAuthHandler(svc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	auth := app.Group("/api/v1/auth")
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/logout", h.Logout)
	auth.Post("/refresh", h.Refresh)
	return app
}

type apiResp struct {
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
	Error   string          `json:"error"`
}

func post(app *fiber.App, path string, body any) *http.Response {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, 10000)
	return resp
}

func decode(t *testing.T, resp *http.Response) apiResp {
	t.Helper()
	var r apiResp
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&r))
	return r
}

func tokenPair(t *testing.T, r apiResp) (access, refresh string) {
	t.Helper()
	var pair struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	require.NoError(t, json.Unmarshal(r.Data, &pair))
	return pair.AccessToken, pair.RefreshToken
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func TestRegisterHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	resp := post(app, "/api/v1/auth/register", map[string]string{
		"name": "Alice", "email": "alice@example.com", "password": "password123",
	})

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	r := decode(t, resp)
	assert.Equal(t, "registration successful", r.Message)
	access, refresh := tokenPair(t, r)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
}

func TestRegisterHandler_ValidationErrors(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	cases := []struct {
		name string
		body map[string]string
	}{
		{"short name", map[string]string{"name": "A", "email": "a@b.com", "password": "password123"}},
		{"invalid email", map[string]string{"name": "Alice", "email": "not-an-email", "password": "password123"}},
		{"short password", map[string]string{"name": "Alice", "email": "a@b.com", "password": "short"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := post(app, "/api/v1/auth/register", tc.body)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestRegisterHandler_DuplicateEmail(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	body := map[string]string{"name": "Alice", "email": "alice@example.com", "password": "password123"}
	post(app, "/api/v1/auth/register", body)

	resp := post(app, "/api/v1/auth/register", body)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestRegisterHandler_InvalidBody(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, 5000)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Login
// ---------------------------------------------------------------------------

func TestLoginHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	post(app, "/api/v1/auth/register", map[string]string{
		"name": "Alice", "email": "alice@example.com", "password": "password123",
	})

	resp := post(app, "/api/v1/auth/login", map[string]string{
		"email": "alice@example.com", "password": "password123",
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	r := decode(t, resp)
	access, refresh := tokenPair(t, r)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	post(app, "/api/v1/auth/register", map[string]string{
		"name": "Alice", "email": "alice@example.com", "password": "password123",
	})

	resp := post(app, "/api/v1/auth/login", map[string]string{
		"email": "alice@example.com", "password": "wrongpassword",
	})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLoginHandler_UnknownEmail(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	resp := post(app, "/api/v1/auth/login", map[string]string{
		"email": "nobody@example.com", "password": "password123",
	})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLoginHandler_MissingFields(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	cases := []struct {
		name string
		body map[string]string
	}{
		{"missing email", map[string]string{"email": "", "password": "password123"}},
		{"missing password", map[string]string{"email": "a@b.com", "password": ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := post(app, "/api/v1/auth/login", tc.body)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

// ---------------------------------------------------------------------------
// Logout
// ---------------------------------------------------------------------------

func TestLogoutHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	r1 := decode(t, post(app, "/api/v1/auth/register", map[string]string{
		"name": "Alice", "email": "alice@example.com", "password": "password123",
	}))
	_, refreshToken := tokenPair(t, r1)

	resp := post(app, "/api/v1/auth/logout", map[string]string{"refresh_token": refreshToken})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLogoutHandler_MissingToken(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	resp := post(app, "/api/v1/auth/logout", map[string]string{})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Refresh
// ---------------------------------------------------------------------------

func TestRefreshHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	r1 := decode(t, post(app, "/api/v1/auth/register", map[string]string{
		"name": "Alice", "email": "alice@example.com", "password": "password123",
	}))
	_, firstRefresh := tokenPair(t, r1)

	r2 := decode(t, post(app, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": firstRefresh,
	}))

	newAccess, newRefresh := tokenPair(t, r2)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)
	assert.NotEqual(t, firstRefresh, newRefresh)
}

func TestRefreshHandler_InvalidToken(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	resp := post(app, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": "not-a-real-token",
	})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRefreshHandler_MissingToken(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildApp(integrationDB.DB)

	resp := post(app, "/api/v1/auth/refresh", map[string]string{})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
