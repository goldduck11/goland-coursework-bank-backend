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

type DBAccount struct {
	ID            string  `json:"id"`
	AccountNumber string  `json:"account_number"`
	Currency      string  `json:"currency"`
	Balance       float64 `json:"balance"`
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

func (r *AccountRepository) VerifyOwnership(ctx context.Context, accountID, userID string) error {
	var ownerID string
	err := r.db.QueryRowContext(ctx, "SELECT user_id FROM accounts WHERE id = $1", accountID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAccountNotFound
		}
		return err
	}
	if ownerID != userID {
		return ErrAccessDenied
	}
	return nil
}

func (r *AccountRepository) GetAccountsByUserID(ctx context.Context, userID string) ([]DBAccount, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, account_number, currency, balance FROM accounts WHERE user_id = $1 ORDER BY created_at`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []DBAccount
	for rows.Next() {
		var a DBAccount
		if err := rows.Scan(&a.ID, &a.AccountNumber, &a.Currency, &a.Balance); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (r *AccountRepository) DepositFunds(ctx context.Context, accountID string, amount float64, ownerUserID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := r.lockAndVerifyOwner(ctx, tx, accountID, ownerUserID); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, accountID); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO transactions (receiver_account_id, amount, transaction_type) VALUES ($1, $2, 'deposit')`,
		accountID, amount); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AccountRepository) WithdrawFunds(ctx context.Context, accountID string, amount float64, ownerUserID string) error {
	return r.deductFunds(ctx, accountID, amount, ownerUserID, "withdraw")
}

func (r *AccountRepository) PayFromAccount(ctx context.Context, accountID string, amount float64, ownerUserID string) error {
	return r.deductFunds(ctx, accountID, amount, ownerUserID, "payment")
}

func (r *AccountRepository) deductFunds(ctx context.Context, accountID string, amount float64, ownerUserID, txType string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	balance, err := r.lockAndVerifyOwnerWithBalance(ctx, tx, accountID, ownerUserID)
	if err != nil {
		return err
	}
	if balance < amount {
		return ErrInsufficientFunds
	}

	if _, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, accountID); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO transactions (sender_account_id, amount, transaction_type) VALUES ($1, $2, $3)`,
		accountID, amount, txType); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AccountRepository) TransferFunds(ctx context.Context, senderID, receiverID string, amount float64, ownerUserID string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	senderBalance, err := r.lockAndVerifyOwnerWithBalance(ctx, tx, senderID, ownerUserID)
	if err != nil {
		return err
	}
	if senderBalance < amount {
		return ErrInsufficientFunds
	}

	if _, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, senderID); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, receiverID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("receiver account not found")
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO transactions (sender_account_id, receiver_account_id, amount, transaction_type) 
		 VALUES ($1, $2, $3, 'transfer')`, senderID, receiverID, amount); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AccountRepository) lockAndVerifyOwner(ctx context.Context, tx *sql.Tx, accountID, ownerUserID string) error {
	_, err := r.lockAndVerifyOwnerWithBalance(ctx, tx, accountID, ownerUserID)
	return err
}

func (r *AccountRepository) lockAndVerifyOwnerWithBalance(ctx context.Context, tx *sql.Tx, accountID, ownerUserID string) (float64, error) {
	var balance float64
	var accountOwnerID string
	err := tx.QueryRowContext(ctx,
		"SELECT balance, user_id FROM accounts WHERE id = $1 FOR UPDATE", accountID).
		Scan(&balance, &accountOwnerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrAccountNotFound
		}
		return 0, err
	}
	if accountOwnerID != ownerUserID {
		return 0, ErrAccessDenied
	}
	return balance, nil
}
