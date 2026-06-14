package repository

import (
	"banking-system/internal/model"
	"context"
	"time"

	"github.com/google/uuid"
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

type TransactionRepository interface {
	Create(transaction *model.Transaction) error

	GetByAccountID(accountID uuid.UUID) ([]model.Transaction, error)

	GetMonthlyIncome(accountID uuid.UUID, month time.Time) (float64, error)

	GetMonthlyExpense(accountID uuid.UUID, month time.Time) (float64, error)
}

type PaymentScheduleRepository interface {
	Create(schedule *model.PaymentSchedule) error

	GetByCreditID(creditID uuid.UUID) ([]model.PaymentSchedule, error)

	GetUnpaidSchedules() ([]model.PaymentSchedule, error)

	MarkPaid(id uuid.UUID) error

	ApplyPenalty(id uuid.UUID) error
}
