package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/UmedjonQurbonov/cinema-user-service/internal/domain"
	"github.com/UmedjonQurbonov/cinema-user-service/pkg/jwt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidData       = errors.New("invalid user data")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrCodeNotFound      = errors.New("verification code expired or not found")
	ErrInvalidCode       = errors.New("invalid verification code")
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (int, error)
	ExistsByEmail(ctx context.Context, user *domain.User) (bool, error)
}

type TempUserStore interface {
	SavePendingUser(ctx context.Context, email, code string, user *domain.User, ttl time.Duration) error
	GetPendingUser(ctx context.Context, email string) (*domain.User, string, error)
	DeletePendingUser(ctx context.Context, email string) error
}

type Mailer interface {
	SendVerificationCode(ctx context.Context, toEmail, code string) error
}

type UserService struct {
	repo      UserRepository
	tempStore TempUserStore
	mailer    Mailer
}

func NewServiceUser(repo UserRepository, tempStore TempUserStore, mailer Mailer) *UserService {
	return &UserService{
		repo:      repo,
		tempStore: tempStore,
		mailer:    mailer,
	}
}

func generateCode(length int) string {
	digits := "0123456789"
	bytes := make([]byte, length)
	_, _ = io.ReadFull(rand.Reader, bytes)
	for i := range bytes {
		bytes[i] = digits[bytes[i]%byte(len(digits))]
	}
	return string(bytes)
}

func (s *UserService) SendVerificationCode(ctx context.Context, user *domain.User) error {
	if strings.TrimSpace(user.Full_name) == "" {
		return fmt.Errorf("validation error: %w", ErrInvalidData)
	}
	if strings.TrimSpace(user.Email) == "" || !strings.Contains(user.Email, "@") {
		return fmt.Errorf("validation error: %w", ErrInvalidData)
	}
	if strings.TrimSpace(user.Password_hash) == "" {
		return fmt.Errorf("validation error: %w", ErrInvalidData)
	}

	exists, err := s.repo.ExistsByEmail(ctx, user)
	if err != nil {
		return fmt.Errorf("s.repo.ExistsByEmail: %w", err)
	}
	if exists {
		return ErrUserAlreadyExists
	}

	hashPassword, err := jwt.HashPassword(user.Password_hash)
	if err != nil {
		return fmt.Errorf("jwt.HashPassword: %w", err)
	}

	reqUser := &domain.User{
		ID:            user.ID,
		Full_name:     user.Full_name,
		Email:         user.Email,
		Password_hash: hashPassword,
		Role:          "USER",
	}

	code := generateCode(6)

	if err = s.tempStore.SavePendingUser(ctx, user.Email, code, reqUser, 15*time.Minute); err != nil {
		return fmt.Errorf("s.tempStore.SavePendingUser: %w", err)
	}

	if err = s.mailer.SendVerificationCode(ctx, user.Email, code); err != nil {
		// Очищаем Redis, чтобы не оставлять невалидную сессию
		_ = s.tempStore.DeletePendingUser(ctx, user.Email)
		return fmt.Errorf("s.mailer.SendVerificationCode: %w", err)
	}

	return nil
}

func (s *UserService) VerifyAndRegister(ctx context.Context, email, inputCode string) (int, error) {
	pendingUser, expectedCode, err := s.tempStore.GetPendingUser(ctx, email)
	if err != nil {
		return 0, fmt.Errorf("verification error: %w", ErrCodeNotFound)
	}

	if expectedCode != inputCode {
		return 0, fmt.Errorf("verification error: %w", ErrInvalidCode)
	}

	id, err := s.repo.Create(ctx, pendingUser)
	if err != nil {
		return 0, fmt.Errorf("s.repo.Create: %w", err)
	}

	// Запись создана успешно, ошибку удаления из Redis можно не возвращать клиенту
	_ = s.tempStore.DeletePendingUser(ctx, email)

	return id, nil
}