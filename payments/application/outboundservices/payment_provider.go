package outboundservices

import (
	"context"
	"encoding/json"
)

type PaymentMethodDetails struct {
	Type     string
	Brand    string
	Last4    string
	ExpMonth int
	ExpYear  int
}

type CreatePaymentIntentRequest struct {
	Amount                float64
	Currency              string
	StripePaymentMethodID string
	UserID                string
	SubscriptionID        string
	PaymentID             string
}

type PaymentIntentResult struct {
	ID     string
	Status string
}

type ProviderWebhookEvent struct {
	ProviderEventID       string
	EventType             string
	Payload               json.RawMessage
	StripePaymentIntentID string
	PaymentStatus         string
}

type PaymentProvider interface {
	GetPaymentMethodDetails(ctx context.Context, stripePaymentMethodID string) (*PaymentMethodDetails, error)
	AttachPaymentMethod(ctx context.Context, stripePaymentMethodID string, customerID string) error
	CreatePaymentIntent(ctx context.Context, request CreatePaymentIntentRequest) (*PaymentIntentResult, error)
	ConfirmPaymentIntent(ctx context.Context, paymentIntentID string) (*PaymentIntentResult, error)
	ParseWebhookEvent(ctx context.Context, payload []byte, signature string) (*ProviderWebhookEvent, error)
}
