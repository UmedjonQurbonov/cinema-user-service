// internal/infrastructure/security/token.go
package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/UmedjonQurbonov/cinema-user-service/internal/domain"
	gjwt "github.com/golang-jwt/jwt/v5"
)

type JWTTokenManager struct {
	secretKey []byte
	accessTTL time.Duration
}

func NewJWTTokenManager(secretKey string, accessTTL time.Duration) *JWTTokenManager {
	return &JWTTokenManager{
		secretKey: []byte(secretKey),
		accessTTL: accessTTL,
	}
}

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	gjwt.RegisteredClaims
}

func (m *JWTTokenManager) GenerateAccessToken(userID int64, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: gjwt.RegisteredClaims{
			IssuedAt:  gjwt.NewNumericDate(now),
			ExpiresAt: gjwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	token := gjwt.NewWithClaims(gjwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *JWTTokenManager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (m *JWTTokenManager) ValidateAccessToken(tokenString string) (int64, string, error) {
	token, err := gjwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *gjwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*gjwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.secretKey, nil
		},
	)
	if err != nil {
		return 0, "", domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, "", domain.ErrInvalidToken
	}

	return claims.UserID, claims.Role, nil
}