package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mnabil1718/taskflow/internal/config"
	"github.com/mnabil1718/taskflow/internal/model"
	"github.com/mnabil1718/taskflow/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenInvalid       = errors.New("invalid or expired token")
	ErrValidation         = errors.New("validation error")
)

type AuthService interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.TokenPair, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.TokenPair, error)
	Logout(ctx context.Context, rawRefreshToken string) error
	RefreshToken(ctx context.Context, rawRefreshToken string) (*model.TokenPair, error)
	ValidateAccessToken(tokenStr string) (*model.Claims, error)
}

type authService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	cfg       *config.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	cfg *config.Config,
) AuthService {
	return &authService{userRepo: userRepo, tokenRepo: tokenRepo, cfg: cfg}
}

func (s *authService) Register(ctx context.Context, req *model.RegisterRequest) (*model.TokenPair, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{Name: req.Name, Email: req.Email, PasswordHash: string(hash)}
	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}

	return s.generateTokenPair(ctx, user)
}

func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.TokenPair, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("%w: email and password are required", ErrValidation)
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(ctx, user)
}

func (s *authService) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := hashToken(rawRefreshToken)
	return s.tokenRepo.DeleteByHash(ctx, hash)
}

func (s *authService) RefreshToken(ctx context.Context, rawRefreshToken string) (*model.TokenPair, error) {
	hash := hashToken(rawRefreshToken)

	stored, err := s.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTokenInvalid
		}
		return nil, err
	}

	if time.Now().After(stored.ExpiresAt) {
		_ = s.tokenRepo.DeleteByHash(ctx, hash)
		return nil, ErrTokenInvalid
	}

	user, err := s.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}

	// rotate: delete old token, issue new pair
	if err := s.tokenRepo.DeleteByHash(ctx, hash); err != nil {
		return nil, err
	}

	return s.generateTokenPair(ctx, user)
}

func (s *authService) ValidateAccessToken(tokenStr string) (*model.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWT.AccessSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return &model.Claims{
		UserID: claims["user_id"].(string),
		Email:  claims["email"].(string),
	}, nil
}

func (s *authService) generateTokenPair(ctx context.Context, user *model.User) (*model.TokenPair, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	rawRefresh, refreshHash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}

	rt := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshExpiry),
	}
	if err := s.tokenRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &model.TokenPair{AccessToken: accessToken, RefreshToken: rawRefresh}, nil
}

func (s *authService) generateAccessToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(s.cfg.JWT.AccessExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWT.AccessSecret))
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

func newRefreshToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	raw = base64.URLEncoding.EncodeToString(b)
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func validateRegister(req *model.RegisterRequest) error {
	if req.Name == "" || len(req.Name) < 2 {
		return fmt.Errorf("%w: name must be at least 2 characters", ErrValidation)
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return fmt.Errorf("%w: invalid email address", ErrValidation)
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrValidation)
	}
	return nil
}
