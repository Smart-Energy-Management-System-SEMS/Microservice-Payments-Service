package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type PaymentWebhookEventModel struct {
	EventID         uuid.UUID      `gorm:"type:uuid;primaryKey;column:event_id"`
	Provider        string         `gorm:"type:varchar(30);not null;uniqueIndex:idx_provider_event;column:provider"`
	ProviderEventID string         `gorm:"type:varchar(150);not null;uniqueIndex:idx_provider_event;column:provider_event_id"`
	EventType       string         `gorm:"type:varchar(80);not null;column:event_type"`
	Payload         datatypes.JSON `gorm:"type:jsonb;not null;column:payload"`
	Processed       bool           `gorm:"not null;default:false;column:processed"`
	ReceivedAt      time.Time      `gorm:"type:timestamp;not null;column:received_at"`
	ProcessedAt     *time.Time     `gorm:"type:timestamp;column:processed_at"`
}

func (PaymentWebhookEventModel) TableName() string {
	return "payment_webhook_events"
}
