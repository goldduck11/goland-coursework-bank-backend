package service

import (
	"context"
	"crypto/rand"
	"math/big"

	"banking-system/internal/repository/postgres"
)

type AccountService struct {
	repo *postgres.AccountRepository
}

func NewAccountService(repo *postgres.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) CreateAccount(ctx context.Context, userID string) (string, error) {
	accNum := "40817810"
	for i := 0; i < 12; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		accNum += n.String()
	}
	return s.repo.CreateAccount(ctx, userID, accNum)
}

func (s *AccountService) GetAccounts(ctx context.Context, userID string) ([]postgres.DBAccount, error) {
	return s.repo.GetAccountsByUserID(ctx, userID)
}

func (s *AccountService) Deposit(ctx context.Context, accountID string, amount float64, ownerUserID string) error {
	return s.repo.DepositFunds(ctx, accountID, amount, ownerUserID)
}

func (s *AccountService) Withdraw(ctx context.Context, accountID string, amount float64, ownerUserID string) error {
	return s.repo.WithdrawFunds(ctx, accountID, amount, ownerUserID)
}

func (s *AccountService) Transfer(ctx context.Context, senderID, receiverID string, amount float64, ownerUserID string) error {
	return s.repo.TransferFunds(ctx, senderID, receiverID, amount, ownerUserID)
}
