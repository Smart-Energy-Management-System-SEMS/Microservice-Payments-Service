package model

import (
	"time"

	"github.com/google/uuid"
)

type PaymentModel struct {
	PaymentID             uuid.UUID           `gorm:"type:uuid;primaryKey;column:payment_id"`
	SubscriptionID        uuid.UUID           `gorm:"type:uuid;not null;index;column:subscription_id"`
	UserID                uuid.UUID           `gorm:"type:uuid;not null;index;column:user_id"`
	PaymentMethodID       uuid.UUID           `gorm:"type:uuid;not null;index;column:payment_method_id"`
	PaymentMethodRef      *PaymentMethodModel `gorm:"foreignKey:PaymentMethodID;references:PaymentMethodID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Amount                float64             `gorm:"type:numeric(10,2);not null;column:amount"`
	Currency              string              `gorm:"type:varchar(10);not null;column:currency"`
	Status                string              `gorm:"type:varchar(30);not null;index;column:status"`
	PaymentMethod         string              `gorm:"type:varchar(50);not null;column:payment_method"`
	StripePaymentIntentID string              `gorm:"type:varchar(150);index;column:stripe_payment_intent_id"`
	PaidAt                *time.Time          `gorm:"type:timestamp;column:paid_at"`
	CreatedAt             time.Time           `gorm:"type:timestamp;not null;column:created_at"`
}

func (PaymentModel) TableName() string {
	return "payments"
}
