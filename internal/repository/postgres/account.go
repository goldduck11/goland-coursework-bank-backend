package postgres

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds on account")
	ErrAccountNotFound   = errors.New("account not found")
	ErrAccessDenied      = errors.New("access denied: you do not own this account")
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) CreateAccount(ctx context.Context, userID, accountNumber string) (string, error) {
	query := `INSERT INTO accounts (user_id, account_number, currency) VALUES ($1, $2, 'RUB') RETURNING id`
	var id string
	err := r.db.QueryRowContext(ctx, query, userID, accountNumber).Scan(&id)
	return id, err
}

// TransferFunds выполняет атомарный перевод денег со счета на счет с проверкой баланса и прав
func (r *AccountRepository) TransferFunds(ctx context.Context, senderID, receiverID string, amount float64, ownerUserID string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback() // Откатит изменения, если вызовем return до tx.Commit()

	// 1. Проверяем и блокируем счет отправителя (FOR UPDATE)
	var senderBalance float64
	var accountOwnerID string
	err = tx.QueryRowContext(ctx, "SELECT balance, user_id FROM accounts WHERE id = $1 FOR UPDATE", senderID).Scan(&senderBalance, &accountOwnerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAccountNotFound
		}
		return err
	}

	// Защита прав доступа
	if accountOwnerID != ownerUserID {
		return ErrAccessDenied
	}

	// Проверка баланса
	if senderBalance < amount {
		return ErrInsufficientFunds
	}

	// 2. Списываем средства
	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, senderID)
	if err != nil {
		return err
	}

	// 3. Зачисляем средства получателю
	res, err := tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, receiverID)
	if err != nil {
		return err
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("receiver account not found")
	}

	// 4. Записываем операцию в историю транзакций
	logQuery := `INSERT INTO transactions (sender_account_id, receiver_account_id, amount, transaction_type) 
	             VALUES ($1, $2, $3, 'transfer')`
	_, err = tx.ExecContext(ctx, logQuery, senderID, receiverID, amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}
