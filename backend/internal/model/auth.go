package model

import "time"

type RegisterRequest struct {
	Name     string `json:"name" example:"Alice"`
	Email    string `json:"email" example:"alice@example.com"`
	Password string `json:"password" example:"password123"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"alice@example.com"`
	Password string `json:"password" example:"password123"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" example:"k9F2pQ7sR1tU4vWxYz0AbCdEfGhIjKlMnOpQrStUvWx="`
}

type TokenPair struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYzMwMzAxMmEtNjI3NS00YWEzLWFkZWMtZWJmYjEyM2Y0NTY3IiwiZW1haWwiOiJhbGljZUBleGFtcGxlLmNvbSIsImV4cCI6MTczMzc4NzIwMH0.signature"`
	RefreshToken string `json:"refresh_token" example:"k9F2pQ7sR1tU4vWxYz0AbCdEfGhIjKlMnOpQrStUvWx="`
}

type Claims struct {
	UserID string
	Email  string
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}
