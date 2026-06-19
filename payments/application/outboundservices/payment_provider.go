// Package outboundservices declares the "ports" the application uses to reach
// the outside world. In hexagonal (ports & adapters) architecture, a port is an
// interface owned by the inner layers, while the concrete "adapter" lives in
// infrastructure. PaymentProvider below is the port for any card processor; the
// Stripe adapter implements it.
package outboundservices

import (
	"context"
	"encoding/json"
)

// The structs below are plain, provider-neutral data carriers (DTOs). They let
// the application talk about payment concepts without importing Stripe types, so
// the domain/application stay independent of the specific provider SDK.

// PaymentMethodDetails describes a saved card (or other method) in generic terms.
type PaymentMethodDetails struct {
	Type     string
	Brand    string
	Last4    string
	ExpMonth int
	ExpYear  int
}

// CreatePaymentIntentRequest is the input needed to start a charge. The id
// fields are also forwarded as metadata so we can correlate the provider's
// records with ours later (e.g. when a webhook arrives).
type CreatePaymentIntentRequest struct {
	Amount                float64
	Currency              string
	StripePaymentMethodID string
	UserID                string
	SubscriptionID        string
	PaymentID             string
}

// PaymentIntentResult is the minimal response we care about from the provider.
type PaymentIntentResult struct {
	ID     string
	Status string
}

// ProviderWebhookEvent is the normalised form of an incoming webhook, after the
// adapter has verified and parsed it. json.RawMessage keeps the original payload
// bytes untouched for storage/auditing.
type ProviderWebhookEvent struct {
	ProviderEventID       string
	EventType             string
	Payload               json.RawMessage
	StripePaymentIntentID string
	PaymentStatus         string
}

// PaymentProvider is the port: the contract every payment provider adapter must
// fulfil. Because the application depends only on this interface, we could swap
// Stripe for another processor, or use a fake in tests, without touching the
// business logic.
type PaymentProvider interface {
	GetPaymentMethodDetails(ctx context.Context, stripePaymentMethodID string) (*PaymentMethodDetails, error)
	AttachPaymentMethod(ctx context.Context, stripePaymentMethodID string, customerID string) error
	CreatePaymentIntent(ctx context.Context, request CreatePaymentIntentRequest) (*PaymentIntentResult, error)
	ConfirmPaymentIntent(ctx context.Context, paymentIntentID string) (*PaymentIntentResult, error)
	ParseWebhookEvent(ctx context.Context, payload []byte, signature string) (*ProviderWebhookEvent, error)
}
