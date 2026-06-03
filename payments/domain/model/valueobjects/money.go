// Package valueobjects contains "value objects": small immutable types defined
// by their value rather than an identity. Two Money objects with the same amount
// and currency are considered the same.
package valueobjects

import (
	"strings"

	paymentdomain "Microservice-Payments-Service/payments/domain"
)

// Money represents an amount of money in a specific currency. Bundling the
// amount and currency together avoids a classic bug: passing a bare number
// around with no idea whether it is dollars, soles or euros.
type Money struct {
	Amount   float64
	Currency string
}

// NewMoney is a factory function: the safe, validated way to build a Money
// value. It enforces two business rules and returns an error (instead of an
// invalid object) when they are broken, so a Money that exists is always valid.
func NewMoney(amount float64, currency string) (Money, error) {
	// Rule 1: a payment amount must be positive.
	if amount <= 0 {
		return Money{}, paymentdomain.ErrInvalidAmount
	}
	// Normalise the currency to lowercase and trim spaces so "USD ", "usd" and
	// "Usd" are all stored the same way (Stripe also expects lowercase codes).
	currency = strings.ToLower(strings.TrimSpace(currency))
	// Rule 2: a currency is required.
	if currency == "" {
		return Money{}, paymentdomain.ErrInvalidCurrency
	}
	return Money{Amount: amount, Currency: currency}, nil
}
