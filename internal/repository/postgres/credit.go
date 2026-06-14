package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type CreditRepository struct {
	db *sql.DB
}

type DBCredit struct {
	ID           string
	UserID       string
	AccountID    string
	Principal    float64
	InterestRate float64
	TermMonths   int
}

type DBPaymentSchedule struct {
	CreditID      string
	PaymentDate   time.Time
	TotalPayment  float64
	PrincipalPart float64
	InterestPart  float64
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

func (r *CreditRepository) CreateCreditWithSchedule(ctx context.Context, credit DBCredit, schedules []DBPaymentSchedule) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var creditID string
	creditQuery := `INSERT INTO credits (user_id, account_id, principal, interest_rate, term_months) 
	                VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = tx.QueryRowContext(ctx, creditQuery, credit.UserID, credit.AccountID, credit.Principal, credit.InterestRate, credit.TermMonths).Scan(&creditID)
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", credit.Principal, credit.AccountID)
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO transactions (receiver_account_id, amount, transaction_type) VALUES ($1, $2, 'deposit')",
		credit.AccountID, credit.Principal)
	if err != nil {
		return "", err
	}

	scheduleQuery := `INSERT INTO payment_schedules (credit_id, payment_date, total_payment, principal_part, interest_part, status) 
	                  VALUES ($1, $2, $3, $4, $5, 'pending')`
	stmt, err := tx.PrepareContext(ctx, scheduleQuery)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	for _, s := range schedules {
		_, err = stmt.ExecContext(ctx, creditID, s.PaymentDate, s.TotalPayment, s.PrincipalPart, s.InterestPart)
		if err != nil {
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return creditID, nil
}

func (r *CreditRepository) GetScheduleByCreditID(ctx context.Context, creditID string) ([]DBPaymentSchedule, error) {
	query := `SELECT payment_date, total_payment, principal_part, interest_part 
	          FROM payment_schedules WHERE credit_id = $1 ORDER BY payment_date ASC`
	rows, err := r.db.QueryContext(ctx, query, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []DBPaymentSchedule
	for rows.Next() {
		var s DBPaymentSchedule
		if err := rows.Scan(&s.PaymentDate, &s.TotalPayment, &s.PrincipalPart, &s.InterestPart); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

func (r *CreditRepository) GetCreditOwner(ctx context.Context, creditID string) (string, error) {
	var userID string
	err := r.db.QueryRowContext(ctx, "SELECT user_id FROM credits WHERE id = $1", creditID).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("credit not found")
		}
		return "", err
	}
	return userID, nil
}
