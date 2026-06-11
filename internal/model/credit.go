package model

import "time"

type CreditStatus string

const (
	CreditStatusActive CreditStatus = "active"
	CreditStatusClosed CreditStatus = "closed"
)

type Credit struct {
	ID         int64        `json:"id"`
	UserID     int64        `json:"user_id"`
	AccountID  int64        `json:"account_id"`
	Amount     float64      `json:"amount"`
	Rate       float64      `json:"rate"`
	TermMonths int          `json:"term_months"`
	Status     CreditStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
}

type CreateCreditRequest struct {
	AccountID  int64   `json:"account_id"`
	Amount     float64 `json:"amount"`
	TermMonths int     `json:"term_months"`
}
