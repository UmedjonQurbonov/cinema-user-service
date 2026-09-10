package security

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher { 
	return &BcryptHasher{} }


func (h *BcryptHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}

	return string(bytes), err
}

func (h *BcryptHasher) CheckPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("invalid password")
		} 
		return fmt.Errorf("bcrypt.CompareHashAndPassword: %w", err)
		
	}

	return nil
}