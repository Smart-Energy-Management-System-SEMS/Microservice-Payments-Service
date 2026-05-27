package eventhandlers

import (
	"context"
	"log"

	"Microservice-Payments-Service/payments/application/commandservices"
	"Microservice-Payments-Service/payments/application/outboundservices"
	"Microservice-Payments-Service/payments/domain/model/commands"
)

type SubscriptionEventsHandler struct {
	payments *commandservices.PaymentCommandService
}

func NewSubscriptionEventsHandler(payments *commandservices.PaymentCommandService) *SubscriptionEventsHandler {
	return &SubscriptionEventsHandler{payments: payments}
}

func (h *SubscriptionEventsHandler) HandleSubscriptionCreated(ctx context.Context, event outboundservices.SubscriptionCreatedEvent) error {
	log.Printf("subscription.created received for subscription_id=%s user_id=%s", event.SubscriptionID, event.UserID)
	return nil
}

func (h *SubscriptionEventsHandler) HandleSubscriptionRenewalRequested(ctx context.Context, event outboundservices.SubscriptionRenewalRequestedEvent) error {
	_, _, err := h.payments.Process(ctx, commands.ProcessPaymentCommand{
		SubscriptionID:  event.SubscriptionID,
		UserID:          event.UserID,
		PaymentMethodID: event.PaymentMethodID,
		Amount:          event.Amount,
		Currency:        event.Currency,
		PaymentMethod:   "card",
	})
	if err != nil {
		log.Printf("subscription.renewal.requested processing failed: %v", err)
	}
	return err
}

func (h *SubscriptionEventsHandler) HandleSubscriptionCancelled(ctx context.Context, event outboundservices.SubscriptionCancelledEvent) error {
	log.Printf("subscription.cancelled received for subscription_id=%s user_id=%s", event.SubscriptionID, event.UserID)
	return nil
}
