package commandservices

import (
	"context"

	"Microservice-Payments-Service/payments/application/outboundservices"
	paymentdomain "Microservice-Payments-Service/payments/domain"
	"Microservice-Payments-Service/payments/domain/model/commands"
	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/model/valueobjects"
	"Microservice-Payments-Service/payments/domain/repositories"
)

// WebhookCommandService handles incoming webhook notifications from the payment
// provider. A webhook is a callback: instead of us polling Stripe, Stripe calls
// our URL when something happens (a charge succeeded, failed, etc.). It reuses
// PaymentCommandService to apply the resulting status change.
type WebhookCommandService struct {
	events   repositories.WebhookEventRepository
	provider outboundservices.PaymentProvider
	payments *PaymentCommandService
}

func NewWebhookCommandService(events repositories.WebhookEventRepository, provider outboundservices.PaymentProvider, payments *PaymentCommandService) *WebhookCommandService {
	return &WebhookCommandService{events: events, provider: provider, payments: payments}
}

// HandleStripe processes one Stripe webhook. It returns (event, duplicate, err):
// the "duplicate" flag lets the controller answer 200 OK even when we ignore a
// repeat, so Stripe stops retrying it.
func (s *WebhookCommandService) HandleStripe(ctx context.Context, command commands.HandleStripeWebhookCommand) (*entities.PaymentWebhookEvent, bool, error) {
	// First, verify the signature and parse the event. This proves the request
	// really came from Stripe and was not forged.
	providerEvent, err := s.provider.ParseWebhookEvent(ctx, command.Payload, command.Signature)
	if err != nil {
		return nil, false, err
	}
	// Deduplication: Stripe may deliver the same event more than once, so we
	// check whether we have already seen this provider event id. Handling each
	// event exactly once keeps the operation idempotent.
	exists, err := s.events.ExistsByProviderEventID(ctx, valueobjects.WebhookProviderStripe, providerEvent.ProviderEventID)
	if err != nil {
		return nil, false, err
	}
	if exists {
		return nil, true, paymentdomain.ErrDuplicateWebhook
	}

	// Record the event before acting on it, so we have an audit trail.
	event := entities.NewPaymentWebhookEvent(valueobjects.WebhookProviderStripe, providerEvent.ProviderEventID, providerEvent.EventType, providerEvent.Payload)
	if err := s.events.Save(ctx, &event); err != nil {
		return nil, false, err
	}

	// Only the payment-intent events actually change a payment. The switch acts
	// as a whitelist; any other event type is stored but otherwise ignored.
	switch providerEvent.EventType {
	case "payment_intent.succeeded", "payment_intent.payment_failed", "payment_intent.canceled":
		if providerEvent.StripePaymentIntentID != "" {
			if _, _, err := s.payments.MarkFromProvider(ctx, providerEvent.StripePaymentIntentID, providerEvent.PaymentStatus); err != nil {
				return &event, false, err
			}
		}
	}

	// Mark the event as processed both in storage and on the in-memory object.
	if err := s.events.MarkProcessed(ctx, event.EventID); err != nil {
		return &event, false, err
	}
	event.MarkProcessed()
	return &event, false, nil
}
