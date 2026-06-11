package utils

import (
	"crypto/rand"
	"math/big"
	"strconv"
)

// GenerateLuhnNumber генерирует 16-значный номер карты, валидный по алгоритму Луна
func GenerateLuhnNumber(iin string) (string, error) {
	// Длина номера 16 знаков. iin занимает 6 знаков. 9 знаков генерируем случайно. 1 знак — контрольный.
	number := iin
	for i := 0; i < 9; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		number += n.String()
	}

	// Вычисляем контрольную цифру
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
