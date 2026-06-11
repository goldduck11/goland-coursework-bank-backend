package postgres

import (
	"context"
	"database/sql"
	"time"
)

type CreditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{db: db}
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

// CreateCreditWithSchedule создает кредит и весь его график платежей в одной транзакции
func (r *CreditRepository) CreateCreditWithSchedule(ctx context.Context, credit DBCredit, schedules []DBPaymentSchedule) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// 1. Создаем запись о кредите
	var creditID string
	creditQuery := `INSERT INTO credits (user_id, account_id, principal, interest_rate, term_months) 
	                VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = tx.QueryRowContext(ctx, creditQuery, credit.UserID, credit.AccountID, credit.Principal, credit.InterestRate, credit.TermMonths).Scan(&creditID)
	if err != nil {
		return "", err
	}

	// 2. Зачисляем сумму кредита на счет клиента
	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", credit.Principal, credit.AccountID)
	if err != nil {
		return "", err
	}

	// 3. Создаем записи в истории транзакций (пополнение)
	_, err = tx.ExecContext(ctx, "INSERT INTO transactions (receiver_account_id, amount, transaction_type) VALUES ($1, $2, 'deposit')", credit.AccountID, credit.Principal)
	if err != nil {
		return "", err
	}

	// 4. Множественная вставка (Bulk Insert) графика платежей
	scheduleQuery := `INSERT INTO payment_schedules (credit_id, payment_date, total_payment, principal_part, interest_part, status) 
	                  VALUES ($1, $2, $3, $4, $5, 'pending')`

	// Подготавливаем стейтмент для оптимизации вставки в цикле
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

	// Комит транзакции
	if err := tx.Commit(); err != nil {
		return "", err
	}

	return creditID, nil
}

// GetScheduleByCreditID возвращает график платежей для эндпоинта /credits/{id}/schedule
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
