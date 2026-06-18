package acl

import (
	"encoding/json"
	"errors"
	"strings"

	"Microservice-Payments-Service/payments/application/outboundservices"
)

type eventEnvelope struct {
	EventType string                    `json:"eventType"`
	Data      *subscriptionEventPayload `json:"data"`
}

type subscriptionEventPayload struct {
	SubscriptionID  string  `json:"subscription_id"`
	UserID          string  `json:"user_id"`
	PaymentMethodID string  `json:"payment_method_id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Reason          string  `json:"reason"`
	Source          string  `json:"source"`
}

func EventType(payload []byte) (string, error) {
	var envelope eventEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return "", err
	}
	return strings.TrimSpace(envelope.EventType), nil
}

func TranslateSubscriptionCreated(payload []byte) (outboundservices.SubscriptionCreatedEvent, error) {
	event, err := decodeSubscriptionEvent(payload)
	if err != nil {
		return outboundservices.SubscriptionCreatedEvent{}, err
	}
	if event.SubscriptionID == "" || event.UserID == "" {
		return outboundservices.SubscriptionCreatedEvent{}, errors.New("subscription.created payload requires subscription_id and user_id")
	}
	return outboundservices.SubscriptionCreatedEvent{
		SubscriptionID: event.SubscriptionID,
		UserID:         event.UserID,
	}, nil
}

func TranslateSubscriptionRenewalRequested(payload []byte) (outboundservices.SubscriptionRenewalRequestedEvent, error) {
	event, err := decodeSubscriptionEvent(payload)
	if err != nil {
		return outboundservices.SubscriptionRenewalRequestedEvent{}, err
	}
	if event.SubscriptionID == "" || event.UserID == "" || event.PaymentMethodID == "" || event.Amount <= 0 {
		return outboundservices.SubscriptionRenewalRequestedEvent{}, errors.New("subscription.renewal.requested payload requires subscription_id, user_id, payment_method_id and amount")
	}
	return outboundservices.SubscriptionRenewalRequestedEvent{
		SubscriptionID:  event.SubscriptionID,
		UserID:          event.UserID,
		PaymentMethodID: event.PaymentMethodID,
		Amount:          event.Amount,
		Currency:        event.Currency,
	}, nil
}

func TranslateSubscriptionCancelled(payload []byte) (outboundservices.SubscriptionCancelledEvent, error) {
	event, err := decodeSubscriptionEvent(payload)
	if err != nil {
		return outboundservices.SubscriptionCancelledEvent{}, err
	}
	if event.SubscriptionID == "" || event.UserID == "" {
		return outboundservices.SubscriptionCancelledEvent{}, errors.New("subscription.cancelled payload requires subscription_id and user_id")
	}
	return outboundservices.SubscriptionCancelledEvent{
		SubscriptionID: event.SubscriptionID,
		UserID:         event.UserID,
		Reason:         event.Reason,
	}, nil
}

func TranslateBillingPaymentRequested(payload []byte) (outboundservices.BillingPaymentRequestedEvent, error) {
	event, err := decodeSubscriptionEvent(payload)
	if err != nil {
		return outboundservices.BillingPaymentRequestedEvent{}, err
	}
	if event.SubscriptionID == "" || event.UserID == "" || event.PaymentMethodID == "" || event.Amount <= 0 {
		return outboundservices.BillingPaymentRequestedEvent{}, errors.New("billing payment payload requires subscription_id, user_id, payment_method_id and amount")
	}
	return outboundservices.BillingPaymentRequestedEvent{
		SubscriptionID:  event.SubscriptionID,
		UserID:          event.UserID,
		PaymentMethodID: event.PaymentMethodID,
		Amount:          event.Amount,
		Currency:        event.Currency,
		Source:          firstNonEmptyString(event.Source, "billing.events"),
	}, nil
}

func decodeSubscriptionEvent(payload []byte) (subscriptionEventPayload, error) {
	var envelope eventEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return subscriptionEventPayload{}, err
	}
	if envelope.Data == nil {
		return subscriptionEventPayload{}, errors.New("event payload requires data")
	}
	event := *envelope.Data
	normalizeSubscriptionEvent(&event)
	return event, nil
}

func normalizeSubscriptionEvent(event *subscriptionEventPayload) {
	event.SubscriptionID = strings.TrimSpace(event.SubscriptionID)
	event.UserID = strings.TrimSpace(event.UserID)
	event.PaymentMethodID = strings.TrimSpace(event.PaymentMethodID)
	event.Currency = strings.ToLower(strings.TrimSpace(event.Currency))
	event.Reason = strings.TrimSpace(event.Reason)
	event.Source = strings.TrimSpace(event.Source)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
