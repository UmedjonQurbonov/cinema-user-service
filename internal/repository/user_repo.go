package repository

import (
	"context"
	"fmt"

	"github.com/UmedjonQurbonov/cinema-user-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, user *domain.User) (bool, error) {
	const query = `SELECT EXISTS (
		SELECT 1 
		FROM users 
		WHERE email = $1
	);`

	var exists bool 

	err := r.db.QueryRow(ctx, query, user.Email).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("r.db.QueryRow: %w", err)
	}

	return exists, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (int, error) {
	const query = `INSERT INTO users (email, full_name, password_hash) VALUES ($1, $2, $3) returning id`
	var id int
	err := r.db.QueryRow(ctx, query, user.Email, user.Full_name, user.Password_hash).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("r.db.QueryRow: %w", err)
	}

	return id, nil
}