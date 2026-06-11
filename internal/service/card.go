package service

import (
	"awesomeProject/internal/repository/postgres"
	"awesomeProject/internal/utils"
	"context"
)

type CardService struct {
	repo *postgres.CardRepository
}

func NewCardService(repo *postgres.CardRepository) *CardService {
	return &CardService{repo: repo}
}

func (s *CardService) IssueCard(ctx context.Context, userID, accountID string) error {
	// 1. Генерация номера карты по алгоритму Луна (IIN: 400000)
	cardNumber, err := utils.GenerateLuhnNumber("400000")
	if err != nil {
		return err
	}
	expiration := "12/29" // Срок действия (захардкожено для примера)
	cvv := "123"          // Имитация генерации CVV

	// 2. Хеширование CVV через bcrypt
	cvvHash, err := utils.HashCVV(cvv)
	if err != nil {
		return err
	}

	// 3. Шифрование номера и срока PGP (для примера имитируем строку-результат шифрования)
	// В продакшене здесь вызов пакета github.com/ProtonMail/gopenpgp/v2
	encNumber := "[PGP_ENCRYPTED_DATA:" + cardNumber + "]"
	encExpiration := "[PGP_ENCRYPTED_DATA:" + expiration + "]"

	// 4. Расчет HMAC для проверки целостности записи в БД
	secretKey := "card_integrity_secret"
	combinedData := cardNumber + expiration + cvv
	hmacSig := utils.GenerateHMAC(combinedData, secretKey)

	// Сохранение в БД через репозиторий
	dbCard := postgres.DBKeyCard{
		AccountID:           accountID,
		UserID:              userID,
		EncryptedNumber:     encNumber,
		EncryptedExpiration: encExpiration,
		CVVHash:             cvvHash,
		HMACSignature:       hmacSig,
	}

	return s.repo.SaveCard(ctx, dbCard)
}

func (s *CardService) GetCards(ctx context.Context, userID string) ([]postgres.DBKeyCard, error) {
	// Возвращает список карт (в хендлере их можно расшифровать обратно)
	return s.repo.GetCardsByUserID(ctx, userID)
}
