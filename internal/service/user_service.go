package service

import (
	"banking-system/internal/model"
	"banking-system/internal/repository"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewUserService(repo repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{repo: repo, jwtSecret: jwtSecret}
}

func (s *UserService) Register(ctx context.Context, req *model.RegisterRequest) error {
	existing, _ := s.repo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return errors.New("email already taken")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
	}
	return s.repo.Create(ctx, user)
}
