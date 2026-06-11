package service

import "math"

func CalculateAnnuityPayment(principal float64, annualRate float64, termMonths int) (total float64) {
	monthlyRate := annualRate / 12.0 / 100.0
	if monthlyRate == 0 {
		return principal / float64(termMonths)
	}
	// Расчет коэффициента аннуитета
	coeff := (monthlyRate * math.Pow(1+monthlyRate, float64(termMonths))) / (math.Pow(1+monthlyRate, float64(termMonths)) - 1)
	return principal * coeff
}
