package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	shareddomain "Microservice-Payments-Service/payments/shared/domain"
)

type PaymentWebhookEvent struct {
	EventID         uuid.UUID
	Provider        string
	ProviderEventID string
	EventType       string
	Payload         json.RawMessage
	Processed       bool
	ReceivedAt      time.Time
	ProcessedAt     *time.Time
}

func NewPaymentWebhookEvent(provider, providerEventID, eventType string, payload json.RawMessage) PaymentWebhookEvent {
	return PaymentWebhookEvent{
		EventID:         shareddomain.NewID(),
		Provider:        provider,
		ProviderEventID: providerEventID,
		EventType:       eventType,
		Payload:         payload,
		Processed:       false,
		ReceivedAt:      time.Now().UTC(),
	}
}

func (e *PaymentWebhookEvent) MarkProcessed() {
	e.Processed = true
	now := time.Now().UTC()
	e.ProcessedAt = &now
}
