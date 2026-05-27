package entities

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	shareddomain "Microservice-Payments-Service/payments/shared/domain"
)

type Invoice struct {
	InvoiceID     uuid.UUID
	PaymentID     uuid.UUID
	InvoiceNumber string
	IssuedAt      time.Time
	TotalAmount   float64
	PDFURL        string
}

func NewInvoice(paymentID uuid.UUID, totalAmount float64, pdfURL string) Invoice {
	invoiceID := shareddomain.NewID()
	shortID := strings.Split(invoiceID.String(), "-")[0]
	return Invoice{
		InvoiceID:     invoiceID,
		PaymentID:     paymentID,
		InvoiceNumber: fmt.Sprintf("INV-%s-%s", time.Now().UTC().Format("20060102"), strings.ToUpper(shortID)),
		IssuedAt:      time.Now().UTC(),
		TotalAmount:   totalAmount,
		PDFURL:        pdfURL,
	}
}
