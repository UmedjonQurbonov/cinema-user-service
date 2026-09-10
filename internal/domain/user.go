package domain

import (
	"errors"

	"github.com/golang/protobuf/ptypes/timestamp"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

type User struct {
	ID            int
	Email         string
	Password_hash string
	Full_name     string
	Role          Role
	Created_at    timestamp.Timestamp
}