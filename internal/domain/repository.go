package domain

import "context"


type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
}

type TokenProvider interface {
	GenerateToken(userID int64, role string) (string, error)
	ValidateToken(token string) (userID int64, role string, err error)
}


type SessionRepository interface {
	Create(ctx context.Context, session *RefreshSession) error
	GetByToken(ctx context.Context, refreshToken string) (*RefreshSession, error)
	DeleteByToken(ctx context.Context, refreshToken string) error
	DeleteAllByUserID(ctx context.Context, userID int64) error // Для Logout со всех устройств
}

type TokenManager interface {
	GenerateAccessToken(userID int64, email, role string) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(token string) (userID int64, role string, err error)
}