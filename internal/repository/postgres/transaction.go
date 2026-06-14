package postgres

import (
	"context"
	"database/sql"
	"time"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) GetMonthlyIncome(ctx context.Context, userID string, month time.Time) (float64, error) {
	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	end := start.AddDate(0, 1, 0)

	query := `
		SELECT COALESCE(SUM(t.amount), 0)
		FROM transactions t
		JOIN accounts a ON t.receiver_account_id = a.id
		WHERE a.user_id = $1
		  AND t.transaction_type IN ('deposit', 'transfer')
		  AND t.created_at >= $2 AND t.created_at < $3
	`
	var total float64
	err := r.db.QueryRowContext(ctx, query, userID, start, end).Scan(&total)
	return total, err
}

func (r *TransactionRepository) GetMonthlyExpense(ctx context.Context, userID string, month time.Time) (float64, error) {
	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	end := start.AddDate(0, 1, 0)

	query := `
		SELECT COALESCE(SUM(t.amount), 0)
		FROM transactions t
		JOIN accounts a ON t.sender_account_id = a.id
		WHERE a.user_id = $1
		  AND t.transaction_type IN ('withdraw', 'transfer', 'payment')
		  AND t.created_at >= $2 AND t.created_at < $3
	`
	var total float64
	err := r.db.QueryRowContext(ctx, query, userID, start, end).Scan(&total)
	return total, err
}

func (r *TransactionRepository) GetCreditLoad(ctx context.Context, userID string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(ps.total_payment), 0)
		FROM payment_schedules ps
		JOIN credits c ON ps.credit_id = c.id
		WHERE c.user_id = $1 AND ps.status IN ('pending', 'overdue')
	`
	var total float64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	return total, err
}

func (r *TransactionRepository) GetTotalBalance(ctx context.Context, userID string) (float64, error) {
	query := `SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE user_id = $1`
	var total float64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return total, nil
}
