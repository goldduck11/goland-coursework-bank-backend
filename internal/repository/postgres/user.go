package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

var (
	ErrUserAlreadyExists = errors.New("username or email already taken")
	ErrUserNotFound      = errors.New("user not found")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser создает нового пользователя и возвращает его UUID
func (r *UserRepository) CreateUser(ctx context.Context, username, email, passwordHash string) (string, error) {
	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id`

	var id string
	err := r.db.QueryRowContext(ctx, query, username, email, passwordHash).Scan(&id)
	if err != nil {
		// Проверяем ошибку уникальности PostgreSQL (код 23505)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return "", ErrUserAlreadyExists
		}
		return "", fmt.Errorf("failed to insert user: %w", err)
	}

	return id, nil
}

// GetPasswordHashByEmail находит хеш пароля и ID для аутентификации
func (r *UserRepository) GetPasswordHashByEmail(ctx context.Context, email string) (string, string, error) {
	query := `SELECT id, password_hash FROM users WHERE email = $1`

	var id, hash string
	err := r.db.QueryRowContext(ctx, query, email).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrUserNotFound
		}
		return "", "", err
	}

	return id, hash, nil
}
