package model

import (
	"errors"
	"strings"
	"time"
)

// User соответствует таблице users в БД.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // не попадает в JSON
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest — тело POST /register.
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() error {
	if len(r.Username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if !strings.Contains(r.Email, "@") || len(r.Email) < 5 {
		return errors.New("invalid email format")
	}
	if len(r.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}

// LoginRequest — тело POST /login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse — ответ на успешный логин.
type LoginResponse struct {
	Token string `json:"token"`
}
