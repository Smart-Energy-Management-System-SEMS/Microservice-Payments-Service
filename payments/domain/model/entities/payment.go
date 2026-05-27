package entities

import (
	"time"

	"github.com/google/uuid"

	"Microservice-Payments-Service/payments/domain/model/valueobjects"
	shareddomain "Microservice-Payments-Service/payments/shared/domain"
)

type Payment struct {
	PaymentID             uuid.UUID
	SubscriptionID        uuid.UUID
	UserID                uuid.UUID
	PaymentMethodID       uuid.UUID
	Amount                float64
	Currency              string
	Status                valueobjects.PaymentStatus
	PaymentMethod         string
	StripePaymentIntentID string
	PaidAt                *time.Time
	CreatedAt             time.Time
}

func NewPayment(subscriptionID, userID, paymentMethodID uuid.UUID, amount float64, currency, paymentMethod string) (Payment, error) {
	money, err := valueobjects.NewMoney(amount, currency)
	if err != nil {
		return Payment{}, err
	}
	return Payment{
		PaymentID:       shareddomain.NewID(),
		SubscriptionID:  subscriptionID,
		UserID:          userID,
		PaymentMethodID: paymentMethodID,
		Amount:          money.Amount,
		Currency:        money.Currency,
		Status:          valueobjects.PaymentStatusPending,
		PaymentMethod:   paymentMethod,
		CreatedAt:       time.Now().UTC(),
	}, nil
}

func (p *Payment) MarkProcessing(stripePaymentIntentID string) {
	p.Status = valueobjects.PaymentStatusProcessing
	p.StripePaymentIntentID = stripePaymentIntentID
}

func (p *Payment) MarkProcessed(stripePaymentIntentID string) {
	p.Status = valueobjects.PaymentStatusProcessed
	p.StripePaymentIntentID = stripePaymentIntentID
	now := time.Now().UTC()
	p.PaidAt = &now
}

func (p *Payment) MarkFailed(stripePaymentIntentID string) {
	p.Status = valueobjects.PaymentStatusFailed
	p.StripePaymentIntentID = stripePaymentIntentID
}

func (p *Payment) MarkCancelled(stripePaymentIntentID string) {
	p.Status = valueobjects.PaymentStatusCancelled
	p.StripePaymentIntentID = stripePaymentIntentID
}
