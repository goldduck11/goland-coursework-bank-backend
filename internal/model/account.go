package model

import "time"

type Currency string

const (
	CurrencyRUB Currency = "RUB"
)

type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  Currency  `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAccountRequest struct {
	Currency Currency `json:"currency"`
}

type DepositRequest struct {
	Amount float64 `json:"amount"`
}

type TransferRequest struct {
	FromAccountID int64   `json:"from_account_id"`
	ToAccountID   int64   `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}
