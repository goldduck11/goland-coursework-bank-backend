package model

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusOverdue PaymentStatus = "over"
)

type PaymentSchedule struct {
	ID uuid.UUID `json:"id"`

	CreditID uuid.UUID `json:"credit_id"`

	PaymentDate time.Time `json:"payment_date"`

	PaymentAmount float64 `json:"payment_amount"`

	PrincipalAmount float64 `json:"principal_amount"`

	InterestAmount float64 `json:"interest_amount"`

	Paid bool `json:"paid"`

	PenaltyApplied bool `json:"penalty_applied"`
}
