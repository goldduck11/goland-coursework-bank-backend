package service

import (
	"context"
	"time"

	"awesomeProject/internal/repository/postgres"
	"awesomeProject/internal/service/external"
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
	// 1. Получаем актуальную ключевую ставку ЦБ РФ через SOAP
	rate, err := s.cbrClient.GetKeyRate(ctx)
	if err != nil {
		rate = 16.0 // Фолбэк (запасной вариант), если ЦБ недоступен
	}

	// 2. Рассчитываем размер ежемесячного аннуитетного платежа
	monthlyPayment := CalculateAnnuityPayment(principal, rate, termMonths)

	// 3. Формируем массив графиков платежей на N месяцев вперед
	schedules := make([]postgres.DBPaymentSchedule, termMonths)
	now := time.Now()

	// Упрощенный расчет деления платежа на ОД и Проценты
	principalPart := principal / float64(termMonths)
	interestPart := monthlyPayment - principalPart

	for i := 1; i <= termMonths; i++ {
		schedules[i-1] = postgres.DBPaymentSchedule{
			PaymentDate:   now.AddDate(0, i, 0), // Каждый следующий месяц
			TotalPayment:  monthlyPayment,
			PrincipalPart: principalPart,
			InterestPart:  interestPart,
		}
	}

	// 4. Запись в БД
	dbCredit := postgres.DBCredit{
		UserID:       userID,
		AccountID:    accountID,
		Principal:    principal,
		InterestRate: rate,
		TermMonths:   termMonths,
	}

	return s.creditRepo.CreateCreditWithSchedule(ctx, dbCredit, schedules)
}

func (s *CreditService) GetPaymentSchedule(ctx context.Context, creditID string) ([]postgres.DBPaymentSchedule, error) {
	return s.creditRepo.GetScheduleByCreditID(ctx, creditID)
}
