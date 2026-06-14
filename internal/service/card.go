package service

import (
	"context"
	"errors"
	"fmt"

	"banking-system/internal/repository/postgres"
	"banking-system/internal/utils"
)

type CardView struct {
	ID           string `json:"id"`
	AccountID    string `json:"account_id"`
	MaskedNumber string `json:"number"`
	Expiry       string `json:"expiry"`
}

type CardService struct {
	repo        *postgres.CardRepository
	accountRepo *postgres.AccountRepository
	hmacSecret  string
}

func NewCardService(repo *postgres.CardRepository, accountRepo *postgres.AccountRepository, hmacSecret string) *CardService {
	return &CardService{repo: repo, accountRepo: accountRepo, hmacSecret: hmacSecret}
}

func (s *CardService) IssueCard(ctx context.Context, userID, accountID string) error {
	if err := s.accountRepo.VerifyOwnership(ctx, accountID, userID); err != nil {
		return err
	}

	cardNumber, err := utils.GenerateLuhnNumber("400000")
	if err != nil {
		return err
	}
	expiration := utils.GenerateExpiry(3)
	cvv, err := utils.GenerateCVV()
	if err != nil {
		return err
	}

	cvvHash, err := utils.HashCVV(cvv)
	if err != nil {
		return err
	}

	encNumber, err := utils.EncryptPGP(cardNumber)
	if err != nil {
		return fmt.Errorf("encrypt card number: %w", err)
	}
	encExpiration, err := utils.EncryptPGP(expiration)
	if err != nil {
		return fmt.Errorf("encrypt expiry: %w", err)
	}

	hmacSig := utils.GenerateHMAC(cardNumber+expiration, s.hmacSecret)

	return s.repo.SaveCard(ctx, postgres.DBKeyCard{
		AccountID:           accountID,
		UserID:              userID,
		EncryptedNumber:     encNumber,
		EncryptedExpiration: encExpiration,
		CVVHash:             cvvHash,
		HMACSignature:       hmacSig,
	})
}

func (s *CardService) GetCards(ctx context.Context, userID string) ([]CardView, error) {
	cards, err := s.repo.GetCardsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]CardView, 0, len(cards))
	for _, c := range cards {
		view, err := s.decryptCard(c)
		if err != nil {
			return nil, err
		}
		result = append(result, view)
	}
	return result, nil
}

func (s *CardService) PayWithCard(ctx context.Context, userID, cardID string, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	card, err := s.repo.GetCardByID(ctx, cardID)
	if err != nil {
		return err
	}
	if card.UserID != userID {
		return postgres.ErrAccessDenied
	}

	return s.accountRepo.PayFromAccount(ctx, card.AccountID, amount, userID)
}

func (s *CardService) decryptCard(c postgres.DBKeyCard) (CardView, error) {
	number, err := utils.DecryptPGP(c.EncryptedNumber)
	if err != nil {
		return CardView{}, fmt.Errorf("decrypt number: %w", err)
	}
	expiry, err := utils.DecryptPGP(c.EncryptedExpiration)
	if err != nil {
		return CardView{}, fmt.Errorf("decrypt expiry: %w", err)
	}

	if !utils.VerifyHMAC(number+expiry, c.HMACSignature, s.hmacSecret) {
		return CardView{}, errors.New("card data integrity check failed")
	}

	return CardView{
		ID:           c.ID,
		AccountID:    c.AccountID,
		MaskedNumber: utils.MaskCardNumber(number),
		Expiry:       expiry,
	}, nil
}
