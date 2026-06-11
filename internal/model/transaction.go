package model

import "time"

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypeTransfer TransactionType = "transfer"
	TransactionTypePayment  TransactionType = "payment"
)

type Transaction struct {
	ID          int64           `json:"id"`
	AccountID   int64           `json:"account_id"`
	Amount      float64         `json:"amount"`
	Type        TransactionType `json:"type"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}
