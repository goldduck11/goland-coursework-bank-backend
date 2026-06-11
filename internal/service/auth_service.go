package service

import (
	"context"
	"time"

	"awesomeProject/internal/repository/postgres"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      *postgres.UserRepository
	jwtSecret string
}

func NewAuthService(repo *postgres.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{repo: repo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) (string, error) {
	// Хеширование пароля через bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return s.repo.CreateUser(ctx, username, email, string(hashedPassword))
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	userID, hashedPassword, err := s.repo.GetPasswordHashByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	// Проверка соответствия пароля
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return "", postgres.ErrUserNotFound // Скрываем детали для безопасности
	}

	// Генерация JWT токена на 24 часа
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
