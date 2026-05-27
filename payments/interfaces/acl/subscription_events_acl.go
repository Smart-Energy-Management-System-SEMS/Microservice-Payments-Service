package acl

import (
	"encoding/json"
	"errors"
	"strings"

	"Microservice-Payments-Service/payments/application/outboundservices"
)

type subscriptionEventPayload struct {
	SubscriptionID  string                    `json:"subscription_id"`
	UserID          string                    `json:"user_id"`
	PaymentMethodID string                    `json:"payment_method_id"`
	Amount          float64                   `json:"amount"`
	Currency        string                    `json:"currency"`
	Reason          string                    `json:"reason"`
	Data            *subscriptionEventPayload `json:"data"`
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

func decodeSubscriptionEvent(payload []byte) (subscriptionEventPayload, error) {
	var envelope subscriptionEventPayload
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return subscriptionEventPayload{}, err
	}
	if envelope.Data != nil {
		mergeSubscriptionEvent(&envelope, envelope.Data)
	}
	normalizeSubscriptionEvent(&envelope)
	return envelope, nil
}

func mergeSubscriptionEvent(target *subscriptionEventPayload, data *subscriptionEventPayload) {
	if target.SubscriptionID == "" {
		target.SubscriptionID = data.SubscriptionID
	}
	if target.UserID == "" {
		target.UserID = data.UserID
	}
	if target.PaymentMethodID == "" {
		target.PaymentMethodID = data.PaymentMethodID
	}
	if target.Amount == 0 {
		target.Amount = data.Amount
	}
	if target.Currency == "" {
		target.Currency = data.Currency
	}
	if target.Reason == "" {
		target.Reason = data.Reason
	}
}

func normalizeSubscriptionEvent(event *subscriptionEventPayload) {
	event.SubscriptionID = strings.TrimSpace(event.SubscriptionID)
	event.UserID = strings.TrimSpace(event.UserID)
	event.PaymentMethodID = strings.TrimSpace(event.PaymentMethodID)
	event.Currency = strings.ToLower(strings.TrimSpace(event.Currency))
	event.Reason = strings.TrimSpace(event.Reason)
}
