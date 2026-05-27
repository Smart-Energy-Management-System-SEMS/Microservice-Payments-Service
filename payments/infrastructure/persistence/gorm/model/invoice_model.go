package model

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceModel struct {
	InvoiceID     uuid.UUID     `gorm:"type:uuid;primaryKey;column:invoice_id"`
	PaymentID     uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex;column:payment_id"`
	PaymentRef    *PaymentModel `gorm:"foreignKey:PaymentID;references:PaymentID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	InvoiceNumber string        `gorm:"type:varchar(80);not null;uniqueIndex;column:invoice_number"`
	IssuedAt      time.Time     `gorm:"type:timestamp;not null;column:issued_at"`
	TotalAmount   float64       `gorm:"type:numeric(10,2);not null;column:total_amount"`
	PDFURL        string        `gorm:"type:text;column:pdf_url"`
}

func (InvoiceModel) TableName() string {
	return "invoices"
}
