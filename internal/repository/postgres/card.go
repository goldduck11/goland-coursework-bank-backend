package postgres

import (
	"context"
	"database/sql"
	"errors"
)

type CardRepository struct {
	db *sql.DB
}

type DBKeyCard struct {
	ID                  string
	AccountID           string
	UserID              string
	EncryptedNumber     string
	EncryptedExpiration string
	CVVHash             string
	HMACSignature       string
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) SaveCard(ctx context.Context, c DBKeyCard) error {
	query := `INSERT INTO cards (account_id, user_id, encrypted_number, encrypted_expiration, cvv_hash, hmac_signature) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, c.AccountID, c.UserID, c.EncryptedNumber, c.EncryptedExpiration, c.CVVHash, c.HMACSignature)
	return err
}

func (r *CardRepository) GetCardsByUserID(ctx context.Context, userID string) ([]DBKeyCard, error) {
	query := `SELECT id, account_id, user_id, encrypted_number, encrypted_expiration, cvv_hash, hmac_signature 
	          FROM cards WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []DBKeyCard
	for rows.Next() {
		var c DBKeyCard
		if err := rows.Scan(&c.ID, &c.AccountID, &c.UserID, &c.EncryptedNumber, &c.EncryptedExpiration, &c.CVVHash, &c.HMACSignature); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *CardRepository) GetCardByID(ctx context.Context, cardID string) (DBKeyCard, error) {
	query := `SELECT id, account_id, user_id, encrypted_number, encrypted_expiration, cvv_hash, hmac_signature 
	          FROM cards WHERE id = $1`
	var c DBKeyCard
	err := r.db.QueryRowContext(ctx, query, cardID).Scan(
		&c.ID, &c.AccountID, &c.UserID, &c.EncryptedNumber, &c.EncryptedExpiration, &c.CVVHash, &c.HMACSignature)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DBKeyCard{}, errors.New("card not found")
		}
		return DBKeyCard{}, err
	}
	return c, nil
}
