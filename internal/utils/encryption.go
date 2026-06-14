package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

var (
	pgpPublicKey  string
	pgpPrivateKey string
	pgpPassphrase string
)

// InitPGPKeys инициализирует PGP-ключи из конфигурации.
// Если ключи не заданы, генерируется временная пара для разработки.
func InitPGPKeys(publicKey, privateKey, passphrase string) error {
	if publicKey != "" && privateKey != "" {
		pgpPublicKey = publicKey
		pgpPrivateKey = privateKey
		pgpPassphrase = passphrase
		return nil
	}

	entity, err := openpgp.NewEntity("Bank", "Service", "bank@local", nil)
	if err != nil {
		return fmt.Errorf("failed to generate dev PGP keys: %w", err)
	}

	pubBuf := new(bytes.Buffer)
	if err := entity.SerializePublic(pubBuf); err != nil {
		return err
	}
	pubArmor := new(bytes.Buffer)
	if err := armor.Encode(pubArmor, openpgp.PublicKeyType, pubBuf.Bytes()); err != nil {
		return err
	}
	pgpPublicKey = pubArmor.String()

	privBuf := new(bytes.Buffer)
	if err := entity.SerializePrivate(privBuf, nil); err != nil {
		return err
	}
	privArmor := new(bytes.Buffer)
	if err := armor.Encode(privArmor, openpgp.PrivateKeyType, privBuf.Bytes()); err != nil {
		return err
	}
	pgpPrivateKey = privArmor.String()
	pgpPassphrase = ""

	return nil
}

// HashCVV хеширует CVV код с помощью bcrypt.
func HashCVV(cvv string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	return string(hash), err
}

// GenerateHMAC создает подпись целостности строки данных.
func GenerateHMAC(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC проверяет валидность данных по подписи.
func VerifyHMAC(data, signature, secret string) bool {
	expectedSign := GenerateHMAC(data, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSign))
}

// EncryptPGP шифрует данные публичным PGP-ключом.
func EncryptPGP(secretData string) (string, error) {
	entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(pgpPublicKey))
	if err != nil {
		return "", fmt.Errorf("read public key: %w", err)
	}

	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, entityList, nil, nil, nil)
	if err != nil {
		return "", fmt.Errorf("encrypt: %w", err)
	}

	if _, err = w.Write([]byte(secretData)); err != nil {
		return "", err
	}
	if err = w.Close(); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf.Bytes()), nil
}

// DecryptPGP расшифровывает данные приватным PGP-ключом.
func DecryptPGP(encryptedHex string) (string, error) {
	ciphertext, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("decode hex: %w", err)
	}

	entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(pgpPrivateKey))
	if err != nil {
		return "", fmt.Errorf("read private key: %w", err)
	}

	var entity *openpgp.Entity
	if len(entityList) > 0 {
		entity = entityList[0]
	}
	if entity == nil {
		return "", fmt.Errorf("no private key entity found")
	}

	if entity.PrivateKey != nil && entity.PrivateKey.Encrypted && pgpPassphrase != "" {
		if err := entity.PrivateKey.Decrypt([]byte(pgpPassphrase)); err != nil {
			return "", fmt.Errorf("decrypt private key: %w", err)
		}
	}

	md, err := openpgp.ReadMessage(bytes.NewReader(ciphertext), entityList, nil, nil)
	if err != nil {
		return "", fmt.Errorf("read message: %w", err)
	}

	plaintext, err := io.ReadAll(md.UnverifiedBody)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// MaskCardNumber маскирует номер карты: **** **** **** 1234.
func MaskCardNumber(number string) string {
	if len(number) < 4 {
		return "****"
	}
	return "**** **** **** " + number[len(number)-4:]
}

// RandomBytes генерирует n случайных байт.
func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}
