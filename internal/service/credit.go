package service

import (
	"context"
	"errors"
	"time"

	"banking-system/internal/repository/postgres"
	"banking-system/internal/service/external"
)

type CreditService struct {
	creditRepo  *postgres.CreditRepository
	accountRepo *postgres.AccountRepository
	cbrClient   *external.CBRClient
}

func NewCreditService(cr *postgres.CreditRepository, ac *postgres.AccountRepository, cbr *external.CBRClient) *CreditService {
	return &CreditService{creditRepo: cr, accountRepo: ac, cbrClient: cbr}
}

func (s *CreditService) ApplyForCredit(ctx context.Context, userID, accountID string, principal float64, termMonths int) (string, error) {
	if principal <= 0 || termMonths <= 0 {
		return "", errors.New("invalid credit parameters")
	}

	if err := s.accountRepo.VerifyOwnership(ctx, accountID, userID); err != nil {
		return "", err
	}

	rate, err := s.cbrClient.GetKeyRate(ctx)
	if err != nil {
		rate = 16.0
	}

	monthlyPayment := CalculateAnnuityPayment(principal, rate, termMonths)
	schedules := make([]postgres.DBPaymentSchedule, termMonths)
	now := time.Now()

	principalPart := principal / float64(termMonths)
	interestPart := monthlyPayment - principalPart

	for i := 1; i <= termMonths; i++ {
		schedules[i-1] = postgres.DBPaymentSchedule{
			PaymentDate:   now.AddDate(0, i, 0),
			TotalPayment:  monthlyPayment,
			PrincipalPart: principalPart,
			InterestPart:  interestPart,
		}
	}

	dbCredit := postgres.DBCredit{
		UserID:       userID,
		AccountID:    accountID,
		Principal:    principal,
		InterestRate: rate,
		TermMonths:   termMonths,
	}

	return s.creditRepo.CreateCreditWithSchedule(ctx, dbCredit, schedules)
}

func (s *CreditService) GetPaymentSchedule(ctx context.Context, creditID, userID string) ([]postgres.DBPaymentSchedule, error) {
	ownerID, err := s.creditRepo.GetCreditOwner(ctx, creditID)
	if err != nil {
		return nil, err
	}
	if ownerID != userID {
		return nil, postgres.ErrAccessDenied
	}
	return s.creditRepo.GetScheduleByCreditID(ctx, creditID)
}
