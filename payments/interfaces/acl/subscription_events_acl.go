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
	SubscriptionID        string  `json:"subscription_id"`
	SubscriptionIDLegacy  string  `json:"SubscriptionID"`
	UserID                string  `json:"user_id"`
	UserIDLegacy          string  `json:"UserID"`
	PaymentMethodID       string  `json:"payment_method_id"`
	PaymentMethodIDLegacy string  `json:"PaymentMethodID"`
	Amount                float64 `json:"amount"`
	AmountLegacy          float64 `json:"Amount"`
	Currency              string  `json:"currency"`
	CurrencyLegacy        string  `json:"Currency"`
	Reason                string  `json:"reason"`
	ReasonLegacy          string  `json:"Reason"`
	Source                string  `json:"source"`
	SourceLegacy          string  `json:"Source"`
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
	var raw subscriptionEventPayload
	if err := json.Unmarshal(payload, &raw); err != nil {
		return subscriptionEventPayload{}, err
	}

	var envelope eventEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return subscriptionEventPayload{}, err
	}
	if envelope.Data == nil {
		normalizeSubscriptionEvent(&raw)
		return raw, nil
	}
	event := *envelope.Data
	mergeSubscriptionEvent(&event, raw)
	normalizeSubscriptionEvent(&event)
	return event, nil
}

func mergeSubscriptionEvent(target *subscriptionEventPayload, fallback subscriptionEventPayload) {
	if target.SubscriptionID == "" {
		target.SubscriptionID = fallback.SubscriptionID
	}
	if target.SubscriptionIDLegacy == "" {
		target.SubscriptionIDLegacy = fallback.SubscriptionIDLegacy
	}
	if target.UserID == "" {
		target.UserID = fallback.UserID
	}
	if target.UserIDLegacy == "" {
		target.UserIDLegacy = fallback.UserIDLegacy
	}
	if target.PaymentMethodID == "" {
		target.PaymentMethodID = fallback.PaymentMethodID
	}
	if target.PaymentMethodIDLegacy == "" {
		target.PaymentMethodIDLegacy = fallback.PaymentMethodIDLegacy
	}
	if target.Amount == 0 {
		target.Amount = fallback.Amount
	}
	if target.AmountLegacy == 0 {
		target.AmountLegacy = fallback.AmountLegacy
	}
	if target.Currency == "" {
		target.Currency = fallback.Currency
	}
	if target.CurrencyLegacy == "" {
		target.CurrencyLegacy = fallback.CurrencyLegacy
	}
	if target.Reason == "" {
		target.Reason = fallback.Reason
	}
	if target.ReasonLegacy == "" {
		target.ReasonLegacy = fallback.ReasonLegacy
	}
	if target.Source == "" {
		target.Source = fallback.Source
	}
	if target.SourceLegacy == "" {
		target.SourceLegacy = fallback.SourceLegacy
	}
}

func normalizeSubscriptionEvent(event *subscriptionEventPayload) {
	if event.SubscriptionID == "" {
		event.SubscriptionID = event.SubscriptionIDLegacy
	}
	if event.UserID == "" {
		event.UserID = event.UserIDLegacy
	}
	if event.PaymentMethodID == "" {
		event.PaymentMethodID = event.PaymentMethodIDLegacy
	}
	if event.Amount == 0 {
		event.Amount = event.AmountLegacy
	}
	if event.Currency == "" {
		event.Currency = event.CurrencyLegacy
	}
	if event.Reason == "" {
		event.Reason = event.ReasonLegacy
	}
	if event.Source == "" {
		event.Source = event.SourceLegacy
	}
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
