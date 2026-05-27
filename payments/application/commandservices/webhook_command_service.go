package commandservices

import (
	"context"

	"Microservice-Payments-Service/payments/application/outboundservices"
	"Microservice-Payments-Service/payments/domain/model/commands"
	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/model/valueobjects"
	"Microservice-Payments-Service/payments/domain/repositories"
	shareddomain "Microservice-Payments-Service/payments/shared/domain"
)

type WebhookCommandService struct {
	events   repositories.WebhookEventRepository
	provider outboundservices.PaymentProvider
	payments *PaymentCommandService
}

func NewWebhookCommandService(events repositories.WebhookEventRepository, provider outboundservices.PaymentProvider, payments *PaymentCommandService) *WebhookCommandService {
	return &WebhookCommandService{events: events, provider: provider, payments: payments}
}

func (s *WebhookCommandService) HandleStripe(ctx context.Context, command commands.HandleStripeWebhookCommand) (*entities.PaymentWebhookEvent, bool, error) {
	providerEvent, err := s.provider.ParseWebhookEvent(ctx, command.Payload, command.Signature)
	if err != nil {
		return nil, false, err
	}
	exists, err := s.events.ExistsByProviderEventID(ctx, valueobjects.WebhookProviderStripe, providerEvent.ProviderEventID)
	if err != nil {
		return nil, false, err
	}
	if exists {
		return nil, true, shareddomain.ErrDuplicateWebhook
	}

	event := entities.NewPaymentWebhookEvent(valueobjects.WebhookProviderStripe, providerEvent.ProviderEventID, providerEvent.EventType, providerEvent.Payload)
	if err := s.events.Save(ctx, &event); err != nil {
		return nil, false, err
	}

	switch providerEvent.EventType {
	case "payment_intent.succeeded", "payment_intent.payment_failed", "payment_intent.canceled":
		if providerEvent.StripePaymentIntentID != "" {
			if _, _, err := s.payments.MarkFromProvider(ctx, providerEvent.StripePaymentIntentID, providerEvent.PaymentStatus); err != nil {
				return &event, false, err
			}
		}
	}

	if err := s.events.MarkProcessed(ctx, event.EventID); err != nil {
		return &event, false, err
	}
	event.MarkProcessed()
	return &event, false, nil
}
