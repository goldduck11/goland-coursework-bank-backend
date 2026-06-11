package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/sirupsen/logrus"
)

type CreditScheduler struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewCreditScheduler(db *sql.DB, logger *logrus.Logger) *CreditScheduler {
	return &CreditScheduler{db: db, logger: logger}
}

func (s *CreditScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(12 * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.processOverduePayments(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *CreditScheduler) processOverduePayments(ctx context.Context) {
	s.logger.Info("Starting background credit payment iteration...")

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to start scheduler transaction")
		return
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT ps.id, ps.credit_id, ps.total_payment, c.account_id, a.balance 
		FROM payment_schedules ps
		JOIN credits c ON ps.credit_id = c.id
		JOIN accounts a ON c.account_id = a.id
		WHERE ps.status IN ('pending', 'overdue') AND ps.payment_date <= CURRENT_TIMESTAMP
	`)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch due payments")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var psID, creditID, accountID string
		var totalPayment, balance float64

		if err := rows.Scan(&psID, &creditID, &totalPayment, &accountID, &balance); err != nil {
			continue
		}

		if balance >= totalPayment {
			_, _ = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", totalPayment, accountID)
			_, _ = tx.ExecContext(ctx, "UPDATE payment_schedules SET status = 'paid', updated_at = CURRENT_TIMESTAMP WHERE id = $1", psID)

			_, _ = tx.ExecContext(ctx, "INSERT INTO transactions (sender_account_id, amount, transaction_type) VALUES ($1, $2, 'withdraw')", accountID, totalPayment)
		} else {
			fineAmount := totalPayment * 0.10
			_, _ = tx.ExecContext(ctx, `
				UPDATE payment_schedules 
				SET total_payment = total_payment + $1, status = 'overdue', updated_at = CURRENT_TIMESTAMP 
				WHERE id = $2`, fineAmount, psID)

			_, _ = tx.ExecContext(ctx, "UPDATE credits SET status = 'overdue' WHERE id = $1", creditID)
			s.logger.Warnf("Credit payment %s is overdue. Fine of %f applied.", psID, fineAmount)

		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.WithError(err).Error("Failed to commit scheduler transaction")
	}
}
