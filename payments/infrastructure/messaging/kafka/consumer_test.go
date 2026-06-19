package kafkaadapter

import (
	"context"
	"testing"

	"Microservice-Payments-Service/payments/application/outboundservices"
)

func TestConsumerConsumedTopicsDeduplicatesGroupedTopics(t *testing.T) {
	consumer := &Consumer{
		topics: Topics{
			SubscriptionsEvents: "platform.events",
			BillingEvents:       "platform.events",
		},
		handler: &stubSubscriptionEventHandler{},
	}

	got := consumer.consumedTopics()
	if len(got) != 1 || got[0] != "platform.events" {
		t.Fatalf("consumedTopics() = %v, want [platform.events]", got)
	}
}

func TestConsumerHandleMessageDispatchesBillingEventFromGroupedTopic(t *testing.T) {
	handler := &stubSubscriptionEventHandler{}
	consumer := &Consumer{handler: handler}

	payload := []byte(`{"eventType":"billing.payment.requested","data":{"subscription_id":"sub-1","user_id":"user-1","payment_method_id":"pm-1","amount":42.5,"currency":"pen","source":"platform.events"}}`)
	if err := consumer.handleMessage(context.Background(), "platform.events", payload); err != nil {
		t.Fatalf("handleMessage() error = %v", err)
	}
	if handler.billingRequested == nil {
		t.Fatal("billingRequested = nil, want dispatched event")
	}
	if handler.billingRequested.PaymentMethodID != "pm-1" {
		t.Fatalf("payment method id = %s, want pm-1", handler.billingRequested.PaymentMethodID)
	}
}

type stubSubscriptionEventHandler struct {
	subscriptionCreated   *outboundservices.SubscriptionCreatedEvent
	subscriptionRenewal   *outboundservices.SubscriptionRenewalRequestedEvent
	subscriptionCancelled *outboundservices.SubscriptionCancelledEvent
	billingRequested      *outboundservices.BillingPaymentRequestedEvent
}

func (h *stubSubscriptionEventHandler) HandleSubscriptionCreated(_ context.Context, event outboundservices.SubscriptionCreatedEvent) error {
	h.subscriptionCreated = &event
	return nil
}

func (h *stubSubscriptionEventHandler) HandleSubscriptionRenewalRequested(_ context.Context, event outboundservices.SubscriptionRenewalRequestedEvent) error {
	h.subscriptionRenewal = &event
	return nil
}

func (h *stubSubscriptionEventHandler) HandleSubscriptionCancelled(_ context.Context, event outboundservices.SubscriptionCancelledEvent) error {
	h.subscriptionCancelled = &event
	return nil
}

func (h *stubSubscriptionEventHandler) HandleBillingPaymentRequested(_ context.Context, event outboundservices.BillingPaymentRequestedEvent) error {
	h.billingRequested = &event
	return nil
}
