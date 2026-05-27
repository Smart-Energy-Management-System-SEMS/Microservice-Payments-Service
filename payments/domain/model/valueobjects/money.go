package valueobjects

import (
	"strings"

	paymentdomain "Microservice-Payments-Service/payments/domain"
)

type Money struct {
	Amount   float64
	Currency string
}

func NewMoney(amount float64, currency string) (Money, error) {
	if amount <= 0 {
		return Money{}, paymentdomain.ErrInvalidAmount
	}
	currency = strings.ToLower(strings.TrimSpace(currency))
	if currency == "" {
		return Money{}, paymentdomain.ErrInvalidCurrency
	}
	return Money{Amount: amount, Currency: currency}, nil
}
