package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"banking-system/internal/repository/postgres"
)

type MonthlyStats struct {
	Month        string  `json:"month"`
	Income       float64 `json:"income"`
	Expense      float64 `json:"expense"`
	NetFlow      float64 `json:"net_flow"`
	CreditLoad   float64 `json:"credit_load"`
	TotalBalance float64 `json:"total_balance"`
}

type AnalyticsService struct {
	db              *sql.DB
	transactionRepo *postgres.TransactionRepository
}

func NewAnalyticsService(db *sql.DB, transactionRepo *postgres.TransactionRepository) *AnalyticsService {
	return &AnalyticsService{db: db, transactionRepo: transactionRepo}
}

func (s *AnalyticsService) PredictBalance(ctx context.Context, accountID string, days int, userID string) (float64, error) {
	if days < 1 || days > 365 {
		return 0, errors.New("prediction period must be between 1 and 365 days")
	}

	var currentBalance float64
	var ownerID string
	err := s.db.QueryRowContext(ctx,
		"SELECT balance, user_id FROM accounts WHERE id = $1", accountID).
		Scan(&currentBalance, &ownerID)
	if err != nil {
		return 0, err
	}
	if ownerID != userID {
		return 0, errors.New("access denied")
	}

	targetDate := time.Now().AddDate(0, 0, days)

	var totalFutureDeductions float64
	query := `
		SELECT COALESCE(SUM(ps.total_payment), 0)
		FROM payment_schedules ps
		JOIN credits c ON ps.credit_id = c.id
		WHERE c.account_id = $1 
		  AND ps.status IN ('pending', 'overdue')
		  AND ps.payment_date > CURRENT_TIMESTAMP 
		  AND ps.payment_date <= $2
	`
	err = s.db.QueryRowContext(ctx, query, accountID, targetDate).Scan(&totalFutureDeductions)
	if err != nil {
		return 0, err
	}

	predictedBalance := currentBalance - totalFutureDeductions
	if predictedBalance < 0 {
		predictedBalance = 0
	}
	return predictedBalance, nil
}

func (s *AnalyticsService) GetMonthlyStats(ctx context.Context, userID string) (*MonthlyStats, error) {
	now := time.Now()
	income, err := s.transactionRepo.GetMonthlyIncome(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	expense, err := s.transactionRepo.GetMonthlyExpense(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	creditLoad, err := s.transactionRepo.GetCreditLoad(ctx, userID)
	if err != nil {
		return nil, err
	}
	totalBalance, err := s.transactionRepo.GetTotalBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &MonthlyStats{
		Month:        now.Format("2006-01"),
		Income:       income,
		Expense:      expense,
		NetFlow:      income - expense,
		CreditLoad:   creditLoad,
		TotalBalance: totalBalance,
	}, nil
}
