package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

// GenerateLuhnNumber генерирует 16-значный номер карты, валидный по алгоритму Луна.
// iin — первые 6 цифр (Bank Identification Number), например "400000".
func GenerateLuhnNumber(iin string) (string, error) {
	// 16 символов итого: 6 (IIN) + 9 (случайных) + 1 (контрольная)
	number := iin
	for i := 0; i < 9; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		number += n.String()
	}

	checksum := 0
	isSecond := true
	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))
		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		checksum += digit
		isSecond = !isSecond
	}

	controlDigit := (10 - (checksum % 10)) % 10
	return number + strconv.Itoa(controlDigit), nil
}

// GenerateExpiry генерирует срок действия карты через N лет от сегодня.
func GenerateExpiry(yearsFromNow int) string {
	t := time.Now().AddDate(yearsFromNow, 0, 0)
	return fmt.Sprintf("%02d/%02d", t.Month(), t.Year()%100)
}

// GenerateCVV генерирует случайный 3-значный CVV.
func GenerateCVV() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900))
	if err != nil {
		return "", err
	}
	cvv := int(n.Int64()) + 100
	return strconv.Itoa(cvv), nil
}
