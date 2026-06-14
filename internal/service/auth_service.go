package service

import (
	"context"
	"time"

	"banking-system/internal/repository/postgres"
	"banking-system/internal/service/external"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo         *postgres.UserRepository
	jwtSecret    string
	emailService *external.EmailService
}

func NewAuthService(repo *postgres.UserRepository, jwtSecret string, emailService *external.EmailService) *AuthService {
	return &AuthService{repo: repo, jwtSecret: jwtSecret, emailService: emailService}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) (string, error) {
	if len(password) < 6 {
		return "", postgres.ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	id, err := s.repo.CreateUser(ctx, username, email, string(hashedPassword))
	if err != nil {
		return "", err
	}

	if s.emailService != nil {
		s.emailService.SendRegistrationWelcome(email, username)
	}

	return id, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	userID, hashedPassword, err := s.repo.GetPasswordHashByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return "", postgres.ErrUserNotFound
	}

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
