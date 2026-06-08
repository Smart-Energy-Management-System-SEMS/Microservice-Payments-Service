package acl

import (
	"encoding/json"
	"errors"
	"strings"

	"Microservice-Payments-Service/payments/application/outboundservices"
)

type subscriptionEventPayload struct {
	SubscriptionID        string                    `json:"subscription_id"`
	SubscriptionIDLegacy  string                    `json:"SubscriptionID"`
	UserID                string                    `json:"user_id"`
	UserIDLegacy          string                    `json:"UserID"`
	PaymentMethodID       string                    `json:"payment_method_id"`
	PaymentMethodIDLegacy string                    `json:"PaymentMethodID"`
	Amount                float64                   `json:"amount"`
	AmountLegacy          float64                   `json:"Amount"`
	Currency              string                    `json:"currency"`
	CurrencyLegacy        string                    `json:"Currency"`
	Reason                string                    `json:"reason"`
	ReasonLegacy          string                    `json:"Reason"`
	Data                  *subscriptionEventPayload `json:"data"`
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
	if target.SubscriptionIDLegacy == "" {
		target.SubscriptionIDLegacy = data.SubscriptionIDLegacy
	}
	if target.UserID == "" {
		target.UserID = data.UserID
	}
	if target.UserIDLegacy == "" {
		target.UserIDLegacy = data.UserIDLegacy
	}
	if target.PaymentMethodID == "" {
		target.PaymentMethodID = data.PaymentMethodID
	}
	if target.PaymentMethodIDLegacy == "" {
		target.PaymentMethodIDLegacy = data.PaymentMethodIDLegacy
	}
	if target.Amount == 0 {
		target.Amount = data.Amount
	}
	if target.AmountLegacy == 0 {
		target.AmountLegacy = data.AmountLegacy
	}
	if target.Currency == "" {
		target.Currency = data.Currency
	}
	if target.CurrencyLegacy == "" {
		target.CurrencyLegacy = data.CurrencyLegacy
	}
	if target.Reason == "" {
		target.Reason = data.Reason
	}
	if target.ReasonLegacy == "" {
		target.ReasonLegacy = data.ReasonLegacy
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
	event.SubscriptionID = strings.TrimSpace(event.SubscriptionID)
	event.UserID = strings.TrimSpace(event.UserID)
	event.PaymentMethodID = strings.TrimSpace(event.PaymentMethodID)
	event.Currency = strings.ToLower(strings.TrimSpace(event.Currency))
	event.Reason = strings.TrimSpace(event.Reason)
}
