package entities

import (
	"time"

	"github.com/google/uuid"

	shareddomain "Microservice-Payments-Service/payments/shared/domain"
)

type PaymentMethod struct {
	PaymentMethodID       uuid.UUID
	UserID                uuid.UUID
	Type                  string
	Brand                 string
	Last4                 string
	ExpMonth              int
	ExpYear               int
	StripePaymentMethodID string
	IsDefault             bool
	CreatedAt             time.Time
}

func NewPaymentMethod(userID uuid.UUID, methodType, brand, last4 string, expMonth, expYear int, stripePaymentMethodID string, isDefault bool) PaymentMethod {
	return PaymentMethod{
		PaymentMethodID:       shareddomain.NewID(),
		UserID:                userID,
		Type:                  methodType,
		Brand:                 brand,
		Last4:                 last4,
		ExpMonth:              expMonth,
		ExpYear:               expYear,
		StripePaymentMethodID: stripePaymentMethodID,
		IsDefault:             isDefault,
		CreatedAt:             time.Now().UTC(),
	}
}

func (m *PaymentMethod) MarkDefault() {
	m.IsDefault = true
}

func (m *PaymentMethod) RemoveDefault() {
	m.IsDefault = false
}
