//go:build integration

package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mnabil1718/taskflow/internal/handler"
	"github.com/mnabil1718/taskflow/internal/middleware"
	"github.com/mnabil1718/taskflow/internal/repository"
	"github.com/mnabil1718/taskflow/internal/service"
)

// ---------------------------------------------------------------------------
// App builder
// ---------------------------------------------------------------------------

func buildProjectApp(db *sql.DB) *fiber.App {
	cfg := integrationCfg()
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	authSvc := service.NewAuthService(userRepo, tokenRepo, cfg)
	projectSvc := service.NewProjectService(repository.NewProjectRepository(db), userRepo)

	authH := handler.NewAuthHandler(authSvc)
	projectH := handler.NewProjectHandler(projectSvc)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	auth := app.Group("/api/v1/auth")
	auth.Post("/register", authH.Register)
	auth.Post("/login", authH.Login)

	projects := app.Group("/api/v1/projects", middleware.JWTProtected(authSvc))
	projects.Post("", projectH.Create)
	projects.Get("", projectH.List)
	projects.Post("/bulk-delete", projectH.BulkDelete)
	projects.Get("/:id", projectH.GetByID)
	projects.Put("/:id", projectH.Update)
	projects.Delete("/:id", projectH.Delete)
	projects.Get("/:id/members", projectH.GetMembers)
	projects.Post("/:id/members", projectH.AddMember)
	projects.Delete("/:id/members/:userID", projectH.RemoveMember)

	return app
}

// ---------------------------------------------------------------------------
// Request helpers
// ---------------------------------------------------------------------------

func authGet(app *fiber.App, path, token string) *http.Response {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := app.Test(req, 10000)
	return resp
}

func authPost(app *fiber.App, path string, body any, token string) *http.Response {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := app.Test(req, 10000)
	return resp
}

func authPut(app *fiber.App, path string, body any, token string) *http.Response {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := app.Test(req, 10000)
	return resp
}

func authDel(app *fiber.App, path, token string) *http.Response {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := app.Test(req, 10000)
	return resp
}

// ---------------------------------------------------------------------------
// Decode helpers
// ---------------------------------------------------------------------------

type projectResp struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	OwnerID string `json:"owner_id"`
}

type projectListResp struct {
	Items      []projectResp `json:"items"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

type memberResp struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type membersListResp []memberResp

func extractProject(t *testing.T, r apiResp) projectResp {
	t.Helper()
	var p projectResp
	require.NoError(t, json.Unmarshal(r.Data, &p))
	return p
}

func extractProjectList(t *testing.T, r apiResp) projectListResp {
	t.Helper()
	var pl projectListResp
	require.NoError(t, json.Unmarshal(r.Data, &pl))
	return pl
}

func extractMember(t *testing.T, r apiResp) memberResp {
	t.Helper()
	var m memberResp
	require.NoError(t, json.Unmarshal(r.Data, &m))
	return m
}

func extractMembersList(t *testing.T, r apiResp) membersListResp {
	t.Helper()
	var ms membersListResp
	require.NoError(t, json.Unmarshal(r.Data, &ms))
	return ms
}

// ---------------------------------------------------------------------------
// Setup helpers
// ---------------------------------------------------------------------------

func registerUser(t *testing.T, app *fiber.App, name, email, password string) string {
	t.Helper()
	r := decode(t, post(app, "/api/v1/auth/register", map[string]string{
		"name": name, "email": email, "password": password,
	}))
	require.Empty(t, r.Error, "registration failed: %s", r.Error)
	access, _ := tokenPair(t, r)
	return access
}

func extractUserID(t *testing.T, accessToken string) string {
	t.Helper()
	cfg := integrationCfg()
	tok, err := jwt.Parse(accessToken, func(t *jwt.Token) (any, error) {
		return []byte(cfg.JWT.AccessSecret), nil
	})
	require.NoError(t, err)
	return tok.Claims.(jwt.MapClaims)["user_id"].(string)
}

func createProject(t *testing.T, app *fiber.App, token, name string) projectResp {
	t.Helper()
	r := decode(t, authPost(app, "/api/v1/projects", map[string]string{"name": name}, token))
	require.Empty(t, r.Error, "createProject failed: %s", r.Error)
	return extractProject(t, r)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestCreateProjectHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")

	resp := authPost(app, "/api/v1/projects", map[string]string{
		"name": "Alpha", "description": "first project",
	}, token)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	r := decode(t, resp)
	assert.Equal(t, "project created", r.Message)
	p := extractProject(t, r)
	assert.NotEmpty(t, p.ID)
	assert.Equal(t, "Alpha", p.Name)
	assert.Equal(t, "active", p.Status)
}

func TestCreateProjectHandler_NoAuth(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(`{"name":"X"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateProjectHandler_ValidationErrors(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")

	cases := []struct {
		name string
		body map[string]string
	}{
		{"empty name", map[string]string{"name": ""}},
		{"name too long", map[string]string{"name": strings.Repeat("x", 256)}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := authPost(app, "/api/v1/projects", tc.body, token)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestCreateProjectHandler_InvalidBody(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestGetByIDHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")
	p := createProject(t, app, token, "Alpha")

	resp := authGet(app, "/api/v1/projects/"+p.ID, token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	got := extractProject(t, decode(t, resp))
	assert.Equal(t, p.ID, got.ID)
	assert.Equal(t, "Alpha", got.Name)
}

func TestGetByIDHandler_NotFound(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")

	resp := authGet(app, "/api/v1/projects/00000000-0000-0000-0000-000000000000", token)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetByIDHandler_NonMemberCannotAccess(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	p := createProject(t, app, aliceToken, "Alpha")

	resp := authGet(app, "/api/v1/projects/"+p.ID, bobToken)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetByIDHandler_NoAuth(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/some-id", nil)
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestListProjectsHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")
	createProject(t, app, token, "Alpha")
	createProject(t, app, token, "Beta")

	resp := authGet(app, "/api/v1/projects", token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	pl := extractProjectList(t, decode(t, resp))
	assert.Equal(t, 2, pl.Total)
	assert.Len(t, pl.Items, 2)
	assert.Equal(t, 1, pl.Page)
}

func TestListProjectsHandler_EmptyList(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")

	resp := authGet(app, "/api/v1/projects", token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	pl := extractProjectList(t, decode(t, resp))
	assert.Equal(t, 0, pl.Total)
	assert.Empty(t, pl.Items)
}

func TestListProjectsHandler_OnlyOwnProjects(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	createProject(t, app, aliceToken, "Alice's Project")
	createProject(t, app, bobToken, "Bob's Project")

	resp := authGet(app, "/api/v1/projects", aliceToken)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	pl := extractProjectList(t, decode(t, resp))
	assert.Equal(t, 1, pl.Total)
}

func TestListProjectsHandler_Pagination(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")
	for i := 1; i <= 5; i++ {
		createProject(t, app, token, "Project")
	}

	resp := authGet(app, "/api/v1/projects?page=1&limit=2", token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	pl := extractProjectList(t, decode(t, resp))
	assert.Equal(t, 5, pl.Total)
	assert.Len(t, pl.Items, 2)
	assert.Equal(t, 3, pl.TotalPages)
}

func TestListProjectsHandler_NoAuth(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdateProjectHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")
	p := createProject(t, app, token, "Alpha")

	resp := authPut(app, "/api/v1/projects/"+p.ID, map[string]string{
		"name": "Alpha Renamed", "status": "archived",
	}, token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	got := extractProject(t, decode(t, resp))
	assert.Equal(t, "Alpha Renamed", got.Name)
	assert.Equal(t, "archived", got.Status)
}

func TestUpdateProjectHandler_Forbidden(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	p := createProject(t, app, aliceToken, "Alpha")

	resp := authPut(app, "/api/v1/projects/"+p.ID, map[string]string{
		"name": "Hijacked", "status": "active",
	}, bobToken)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestUpdateProjectHandler_NotFound(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")

	resp := authPut(app, "/api/v1/projects/00000000-0000-0000-0000-000000000000",
		map[string]string{"name": "X", "status": "active"}, token)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUpdateProjectHandler_ValidationErrors(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")
	p := createProject(t, app, token, "Alpha")

	cases := []struct {
		name string
		body map[string]string
	}{
		{"empty name", map[string]string{"name": "", "status": "active"}},
		{"invalid status", map[string]string{"name": "Alpha", "status": "pending"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := authPut(app, "/api/v1/projects/"+p.ID, tc.body, token)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestUpdateProjectHandler_NoAuth(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/some-id", strings.NewReader(`{"name":"X","status":"active"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDeleteProjectHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")
	p := createProject(t, app, token, "Alpha")

	resp := authDel(app, "/api/v1/projects/"+p.ID, token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// subsequent GET must return 404
	resp2 := authGet(app, "/api/v1/projects/"+p.ID, token)
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestDeleteProjectHandler_Forbidden(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	p := createProject(t, app, aliceToken, "Alpha")

	resp := authDel(app, "/api/v1/projects/"+p.ID, bobToken)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteProjectHandler_NotFound(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")

	resp := authDel(app, "/api/v1/projects/00000000-0000-0000-0000-000000000000", token)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteProjectHandler_NoAuth(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/some-id", nil)
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// AddMember
// ---------------------------------------------------------------------------

func TestAddMemberHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	bobID := extractUserID(t, bobToken)
	p := createProject(t, app, aliceToken, "Alpha")

	resp := authPost(app, "/api/v1/projects/"+p.ID+"/members",
		map[string]string{"user_id": bobID, "role": "member"}, aliceToken)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	m := extractMember(t, decode(t, resp))
	assert.Equal(t, bobID, m.UserID)
	assert.Equal(t, "member", m.Role)
}

func TestAddMemberHandler_DefaultsToMemberRole(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	bobID := extractUserID(t, bobToken)
	p := createProject(t, app, aliceToken, "Alpha")

	resp := authPost(app, "/api/v1/projects/"+p.ID+"/members",
		map[string]string{"user_id": bobID}, aliceToken)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	m := extractMember(t, decode(t, resp))
	assert.Equal(t, "member", m.Role)
}

func TestAddMemberHandler_Forbidden(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	carolToken := registerUser(t, app, "Carol", "carol@example.com", "password123")
	carolID := extractUserID(t, carolToken)
	p := createProject(t, app, aliceToken, "Alpha")

	resp := authPost(app, "/api/v1/projects/"+p.ID+"/members",
		map[string]string{"user_id": carolID, "role": "member"}, bobToken)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAddMemberHandler_UserNotFound(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	p := createProject(t, app, aliceToken, "Alpha")

	resp := authPost(app, "/api/v1/projects/"+p.ID+"/members",
		map[string]string{"user_id": "00000000-0000-0000-0000-000000000000", "role": "member"}, aliceToken)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAddMemberHandler_AlreadyMember(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	bobID := extractUserID(t, bobToken)
	p := createProject(t, app, aliceToken, "Alpha")

	// add bob once
	authPost(app, "/api/v1/projects/"+p.ID+"/members",
		map[string]string{"user_id": bobID, "role": "member"}, aliceToken)

	// attempt to add bob again
	resp := authPost(app, "/api/v1/projects/"+p.ID+"/members",
		map[string]string{"user_id": bobID, "role": "member"}, aliceToken)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestAddMemberHandler_InvalidRole(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	bobID := extractUserID(t, bobToken)
	p := createProject(t, app, aliceToken, "Alpha")

	resp := authPost(app, "/api/v1/projects/"+p.ID+"/members",
		map[string]string{"user_id": bobID, "role": "owner"}, aliceToken)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAddMemberHandler_NoAuth(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/some-id/members",
		strings.NewReader(`{"user_id":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// RemoveMember
// ---------------------------------------------------------------------------

func TestRemoveMemberHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	bobID := extractUserID(t, bobToken)
	p := createProject(t, app, aliceToken, "Alpha")
	authPost(app, "/api/v1/projects/"+p.ID+"/members",
		map[string]string{"user_id": bobID, "role": "member"}, aliceToken)

	resp := authDel(app, "/api/v1/projects/"+p.ID+"/members/"+bobID, aliceToken)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveMemberHandler_Forbidden(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	carolToken := registerUser(t, app, "Carol", "carol@example.com", "password123")
	bobID := extractUserID(t, bobToken)
	carolID := extractUserID(t, carolToken)
	p := createProject(t, app, aliceToken, "Alpha")
	authPost(app, "/api/v1/projects/"+p.ID+"/members",
		map[string]string{"user_id": bobID, "role": "member"}, aliceToken)

	// bob tries to remove carol (bob is not owner)
	resp := authDel(app, "/api/v1/projects/"+p.ID+"/members/"+carolID, bobToken)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRemoveMemberHandler_OwnerCannotRemoveSelf(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	aliceID := extractUserID(t, aliceToken)
	p := createProject(t, app, aliceToken, "Alpha")

	resp := authDel(app, "/api/v1/projects/"+p.ID+"/members/"+aliceID, aliceToken)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRemoveMemberHandler_MemberNotFound(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	p := createProject(t, app, aliceToken, "Alpha")

	resp := authDel(app, "/api/v1/projects/"+p.ID+"/members/00000000-0000-0000-0000-000000000000", aliceToken)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRemoveMemberHandler_NoAuth(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/some-id/members/some-user", nil)
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetMembers
// ---------------------------------------------------------------------------

func TestGetMembersHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	bobID := extractUserID(t, bobToken)
	p := createProject(t, app, aliceToken, "Alpha")
	authPost(app, "/api/v1/projects/"+p.ID+"/members",
		map[string]string{"user_id": bobID, "role": "admin"}, aliceToken)

	resp := authGet(app, "/api/v1/projects/"+p.ID+"/members", aliceToken)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	members := extractMembersList(t, decode(t, resp))
	assert.Len(t, members, 2) // owner + bob
}

func TestGetMembersHandler_MemberCanList(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	bobID := extractUserID(t, bobToken)
	p := createProject(t, app, aliceToken, "Alpha")
	authPost(app, "/api/v1/projects/"+p.ID+"/members",
		map[string]string{"user_id": bobID, "role": "member"}, aliceToken)

	// bob (non-owner member) can list members
	resp := authGet(app, "/api/v1/projects/"+p.ID+"/members", bobToken)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetMembersHandler_NonMemberForbidden(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	p := createProject(t, app, aliceToken, "Alpha")

	resp := authGet(app, "/api/v1/projects/"+p.ID+"/members", bobToken)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetMembersHandler_NoAuth(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/some-id/members", nil)
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// BulkDelete
// ---------------------------------------------------------------------------

type bulkDeleteResp struct {
	DeletedCount int `json:"deleted_count"`
}

func extractBulkDelete(t *testing.T, r apiResp) bulkDeleteResp {
	t.Helper()
	var bd bulkDeleteResp
	require.NoError(t, json.Unmarshal(r.Data, &bd))
	return bd
}

func TestBulkDeleteHandler_Success(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")
	p1 := createProject(t, app, token, "Alpha")
	p2 := createProject(t, app, token, "Beta")
	p3 := createProject(t, app, token, "Gamma")

	resp := authPost(app, "/api/v1/projects/bulk-delete",
		map[string]any{"ids": []string{p1.ID, p2.ID}}, token)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bd := extractBulkDelete(t, decode(t, resp))
	assert.Equal(t, 2, bd.DeletedCount)

	// remaining project should still be reachable
	resp2 := authGet(app, "/api/v1/projects/"+p3.ID, token)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	// deleted ones should 404
	resp3 := authGet(app, "/api/v1/projects/"+p1.ID, token)
	assert.Equal(t, http.StatusNotFound, resp3.StatusCode)
}

func TestBulkDeleteHandler_SkipsNonOwnedAndMissing(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	aliceToken := registerUser(t, app, "Alice", "alice@example.com", "password123")
	bobToken := registerUser(t, app, "Bob", "bob@example.com", "password123")
	aliceProject := createProject(t, app, aliceToken, "Alpha")
	bobProject := createProject(t, app, bobToken, "Bob's Project")

	// Bob attempts to bulk-delete his own + Alice's project + a non-existent ID
	resp := authPost(app, "/api/v1/projects/bulk-delete", map[string]any{
		"ids": []string{bobProject.ID, aliceProject.ID, "00000000-0000-0000-0000-000000000000"},
	}, bobToken)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bd := extractBulkDelete(t, decode(t, resp))
	assert.Equal(t, 1, bd.DeletedCount)

	// Alice's project survives
	resp2 := authGet(app, "/api/v1/projects/"+aliceProject.ID, aliceToken)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

func TestBulkDeleteHandler_EmptyIDs(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)
	token := registerUser(t, app, "Alice", "alice@example.com", "password123")

	resp := authPost(app, "/api/v1/projects/bulk-delete",
		map[string]any{"ids": []string{}}, token)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestBulkDeleteHandler_NoAuth(t *testing.T) {
	integrationDB.Truncate(t)
	app := buildProjectApp(integrationDB.DB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/bulk-delete",
		strings.NewReader(`{"ids":["some-id"]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
