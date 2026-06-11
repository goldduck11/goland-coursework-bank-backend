package repository

import (
	"awesomeProject/internal/model"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
}

type AccountRepository interface {
	Create(ctx context.Context, account *model.Account) error
	FindByID(ctx context.Context, id int64) (*model.Account, error)
	FindByUserID(ctx context.Context, userID int64) ([]*model.Account, error)
	UpdateBalance(ctx context.Context, id int64, amount float64) error
}
