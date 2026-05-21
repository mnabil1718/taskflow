package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/mnabil1718/taskflow/internal/config"
	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/repository"
	"github.com/mnabil1718/taskflow/internal/service"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockUserRepo struct {
	byEmail   map[string]*model.User
	byID      map[string]*model.User
	createErr error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		byEmail: make(map[string]*model.User),
		byID:    make(map[string]*model.User),
	}
}

func (m *mockUserRepo) Create(_ context.Context, user *model.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	user.ID = "user-001"
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.byEmail[user.Email] = user
	m.byID[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (*model.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) Search(_ context.Context, query, excludeID string, limit int) ([]*model.User, error) {
	out := make([]*model.User, 0)
	for _, u := range m.byID {
		if u.ID == excludeID {
			continue
		}
		if !strings.Contains(strings.ToLower(u.Name), strings.ToLower(query)) &&
			!strings.Contains(strings.ToLower(u.Email), strings.ToLower(query)) {
			continue
		}
		out = append(out, u)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

type mockTokenRepo struct {
	tokens    map[string]*model.RefreshToken
	createErr error
}

func newMockTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{tokens: make(map[string]*model.RefreshToken)}
}

func (m *mockTokenRepo) Create(_ context.Context, token *model.RefreshToken) error {
	if m.createErr != nil {
		return m.createErr
	}
	token.ID = "token-001"
	token.CreatedAt = time.Now()
	m.tokens[token.TokenHash] = token
	return nil
}

func (m *mockTokenRepo) GetByHash(_ context.Context, hash string) (*model.RefreshToken, error) {
	t, ok := m.tokens[hash]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return t, nil
}

func (m *mockTokenRepo) DeleteByHash(_ context.Context, hash string) error {
	delete(m.tokens, hash)
	return nil
}

func (m *mockTokenRepo) DeleteByUserID(_ context.Context, userID string) error {
	for k, v := range m.tokens {
		if v.UserID == userID {
			delete(m.tokens, k)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func testCfg() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:  "test-access-secret-at-least-32-chars",
			RefreshSecret: "test-refresh-secret-at-least-32-ch",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
	}
}

// seedUser adds a user with a known bcrypt password to userRepo.
// bcrypt.MinCost keeps tests fast.
func seedUser(r *mockUserRepo, id, name, email, plainPassword string) *model.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.MinCost)
	u := &model.User{
		ID:           id,
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	r.byEmail[email] = u
	r.byID[id] = u
	return u
}

func expiredToken(secret string) string {
	claims := jwt.MapClaims{
		"user_id": "u1",
		"email":   "x@x.com",
		"exp":     time.Now().Add(-time.Hour).Unix(),
	}
	t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return t
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func TestRegister_Success(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepo(), newMockTokenRepo(), testCfg())

	pair, err := svc.Register(context.Background(), &model.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestRegister_ValidationErrors(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepo(), newMockTokenRepo(), testCfg())
	ctx := context.Background()

	cases := []struct {
		name string
		req  model.RegisterRequest
	}{
		{"short name", model.RegisterRequest{Name: "A", Email: "a@b.com", Password: "password123"}},
		{"invalid email", model.RegisterRequest{Name: "Alice", Email: "not-an-email", Password: "password123"}},
		{"short password", model.RegisterRequest{Name: "Alice", Email: "a@b.com", Password: "short"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Register(ctx, &tc.req)
			require.Error(t, err)
			assert.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	userRepo := newMockUserRepo()
	userRepo.createErr = repository.ErrDuplicateEmail
	svc := service.NewAuthService(userRepo, newMockTokenRepo(), testCfg())

	_, err := svc.Register(context.Background(), &model.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	})

	assert.ErrorIs(t, err, service.ErrEmailTaken)
}

// ---------------------------------------------------------------------------
// Login
// ---------------------------------------------------------------------------

func TestLogin_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	seedUser(userRepo, "u1", "Alice", "alice@example.com", "password123")
	svc := service.NewAuthService(userRepo, newMockTokenRepo(), testCfg())

	pair, err := svc.Login(context.Background(), &model.LoginRequest{
		Email:    "alice@example.com",
		Password: "password123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestLogin_UserNotFound(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepo(), newMockTokenRepo(), testCfg())

	_, err := svc.Login(context.Background(), &model.LoginRequest{
		Email:    "nobody@example.com",
		Password: "password123",
	})

	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := newMockUserRepo()
	seedUser(userRepo, "u1", "Alice", "alice@example.com", "password123")
	svc := service.NewAuthService(userRepo, newMockTokenRepo(), testCfg())

	_, err := svc.Login(context.Background(), &model.LoginRequest{
		Email:    "alice@example.com",
		Password: "wrongpassword",
	})

	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestLogin_MissingFields(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepo(), newMockTokenRepo(), testCfg())
	ctx := context.Background()

	_, err := svc.Login(ctx, &model.LoginRequest{Email: "", Password: "pass"})
	assert.ErrorIs(t, err, service.ErrValidation)

	_, err = svc.Login(ctx, &model.LoginRequest{Email: "a@b.com", Password: ""})
	assert.ErrorIs(t, err, service.ErrValidation)
}

// ---------------------------------------------------------------------------
// Logout
// ---------------------------------------------------------------------------

func TestLogout_RemovesToken(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	seedUser(userRepo, "u1", "Alice", "alice@example.com", "password123")
	svc := service.NewAuthService(userRepo, tokenRepo, testCfg())
	ctx := context.Background()

	pair, err := svc.Login(ctx, &model.LoginRequest{
		Email:    "alice@example.com",
		Password: "password123",
	})
	require.NoError(t, err)
	assert.Len(t, tokenRepo.tokens, 1)

	err = svc.Logout(ctx, pair.RefreshToken)
	require.NoError(t, err)
	assert.Empty(t, tokenRepo.tokens)
}

// ---------------------------------------------------------------------------
// RefreshToken
// ---------------------------------------------------------------------------

func TestRefreshToken_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	seedUser(userRepo, "u1", "Alice", "alice@example.com", "password123")
	svc := service.NewAuthService(userRepo, tokenRepo, testCfg())
	ctx := context.Background()

	first, err := svc.Login(ctx, &model.LoginRequest{
		Email:    "alice@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	second, err := svc.RefreshToken(ctx, first.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, second.AccessToken)
	assert.NotEmpty(t, second.RefreshToken)
	// old token should be gone, new one stored
	assert.Len(t, tokenRepo.tokens, 1)
	assert.NotEqual(t, first.RefreshToken, second.RefreshToken)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepo(), newMockTokenRepo(), testCfg())

	_, err := svc.RefreshToken(context.Background(), "not-a-real-token")

	assert.ErrorIs(t, err, service.ErrTokenInvalid)
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	seedUser(userRepo, "u1", "Alice", "alice@example.com", "password123")
	svc := service.NewAuthService(userRepo, tokenRepo, testCfg())
	ctx := context.Background()

	pair, _ := svc.Login(ctx, &model.LoginRequest{
		Email:    "alice@example.com",
		Password: "password123",
	})

	// manually expire the stored token
	for _, v := range tokenRepo.tokens {
		v.ExpiresAt = time.Now().Add(-time.Hour)
	}

	_, err := svc.RefreshToken(ctx, pair.RefreshToken)

	assert.ErrorIs(t, err, service.ErrTokenInvalid)
	assert.Empty(t, tokenRepo.tokens) // expired token cleaned up
}

// ---------------------------------------------------------------------------
// ValidateAccessToken
// ---------------------------------------------------------------------------

func TestValidateAccessToken_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	seedUser(userRepo, "u1", "Alice", "alice@example.com", "password123")
	svc := service.NewAuthService(userRepo, newMockTokenRepo(), testCfg())
	ctx := context.Background()

	pair, err := svc.Login(ctx, &model.LoginRequest{
		Email:    "alice@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(pair.AccessToken)

	require.NoError(t, err)
	assert.Equal(t, "u1", claims.UserID)
	assert.Equal(t, "alice@example.com", claims.Email)
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepo(), newMockTokenRepo(), testCfg())

	_, err := svc.ValidateAccessToken("garbage.token.value")

	assert.ErrorIs(t, err, service.ErrTokenInvalid)
}

func TestValidateAccessToken_ExpiredToken(t *testing.T) {
	cfg := testCfg()
	svc := service.NewAuthService(newMockUserRepo(), newMockTokenRepo(), cfg)

	_, err := svc.ValidateAccessToken(expiredToken(cfg.JWT.AccessSecret))

	assert.ErrorIs(t, err, service.ErrTokenInvalid)
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepo(), newMockTokenRepo(), testCfg())

	token := expiredToken("completely-different-secret-value!")

	_, err := svc.ValidateAccessToken(token)

	assert.ErrorIs(t, err, service.ErrTokenInvalid)
}
