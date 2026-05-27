package valueobjects

import (
	"strings"

	shareddomain "Microservice-Payments-Service/payments/shared/domain"
)

type Money struct {
	Amount   float64
	Currency string
}

func NewMoney(amount float64, currency string) (Money, error) {
	if amount <= 0 {
		return Money{}, shareddomain.ErrInvalidAmount
	}
	currency = strings.ToLower(strings.TrimSpace(currency))
	if currency == "" {
		return Money{}, shareddomain.ErrInvalidCurrency
	}
	return Money{Amount: amount, Currency: currency}, nil
}
