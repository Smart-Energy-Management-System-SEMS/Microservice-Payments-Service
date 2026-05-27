package model

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethodModel struct {
	PaymentMethodID       uuid.UUID `gorm:"type:uuid;primaryKey;column:payment_method_id"`
	UserID                uuid.UUID `gorm:"type:uuid;not null;index;column:user_id"`
	Type                  string    `gorm:"type:varchar(30);not null;column:type"`
	Brand                 string    `gorm:"type:varchar(30);column:brand"`
	Last4                 string    `gorm:"type:char(4);column:last4"`
	ExpMonth              int       `gorm:"type:smallint;column:exp_month"`
	ExpYear               int       `gorm:"type:smallint;column:exp_year"`
	StripePaymentMethodID string    `gorm:"type:varchar(150);not null;uniqueIndex;column:stripe_payment_method_id"`
	IsDefault             bool      `gorm:"not null;default:false;column:is_default"`
	CreatedAt             time.Time `gorm:"type:timestamp;not null;column:created_at"`
}

func (PaymentMethodModel) TableName() string {
	return "payment_methods"
}
