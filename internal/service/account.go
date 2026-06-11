package service

import (
	"context"
	"crypto/rand"
	"math/big"

	"awesomeProject/internal/repository/postgres"
)

type AccountService struct {
	repo *postgres.AccountRepository
}

func NewAccountService(repo *postgres.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) CreateAccount(ctx context.Context, userID string) (string, error) {
	// Генерация 20-значного номера счета для примера (начинается с 408)
	accNum := "40817810"
	for i := 0; i < 12; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		accNum += n.String()
	}

	return s.repo.CreateAccount(ctx, userID, accNum)
}

func (s *AccountService) Transfer(ctx context.Context, senderID, receiverID string, amount float64, ownerUserID string) error {
	return s.repo.TransferFunds(ctx, senderID, receiverID, amount, ownerUserID)
}
