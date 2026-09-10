package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
	ErrTokenRevoked = errors.New("token revoked or replaced")
)

// TokenPair возвращается клиенту при Login / Refresh
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshSession сохраняется в БД (PostgreSQL / Redis)
type RefreshSession struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	Fingerprint  string    `json:"fingerprint"` // IP или хэш устройства
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}