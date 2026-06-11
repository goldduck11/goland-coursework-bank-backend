package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// HashCVV хеширует CVV код с помощью bcrypt
func HashCVV(cvv string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	return string(bytes), err
}

// GenerateHMAC создает подпись целостности строки данных на основе секретного ключа
func GenerateHMAC(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC проверяет валидность данных по подписи
func VerifyHMAC(data, signature, secret string) bool {
	expectedSign := GenerateHMAC(data, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSign))
}
