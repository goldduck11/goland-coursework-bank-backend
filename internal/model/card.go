package model

import "time"

type Card struct {
	ID              int64     `json:"id"`
	AccountID       int64     `json:"account_id"`
	NumberEncrypted []byte    `json:"-"`
	ExpiryEncrypted []byte    `json:"-"`
	CVVHash         string    `json:"-"`
	HMACSignature   string    `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
}

// CardResponse — то что видит клиент.
// Число карты показываем частично: **** **** **** 1234
type CardResponse struct {
	ID           int64     `json:"id"`
	AccountID    int64     `json:"account_id"`
	MaskedNumber string    `json:"number"`
	Expiry       string    `json:"expiry"`
	CreatedAt    time.Time `json:"created_at"`
}
