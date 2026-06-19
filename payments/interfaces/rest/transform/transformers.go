// Package transform converts domain entities into REST "resources" (DTOs) for
// the API responses. This layer decides exactly which fields the outside world
// sees and in which format, so internal refactors of the domain do not
// accidentally change our public JSON contract. Note how ids and timestamps are
// turned into strings, the friendliest shape for HTTP clients.
package transform

import (
	"time"

	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/interfaces/rest/resources"
)

// ToPaymentMethodResponse maps a single payment method entity to its API DTO.
func ToPaymentMethodResponse(method entities.PaymentMethod) resources.PaymentMethodResponse {
	return resources.PaymentMethodResponse{
		PaymentMethodID:       method.PaymentMethodID.String(),
		UserID:                method.UserID.String(),
		Type:                  method.Type,
		Brand:                 method.Brand,
		Last4:                 method.Last4,
		ExpMonth:              method.ExpMonth,
		ExpYear:               method.ExpYear,
		StripePaymentMethodID: method.StripePaymentMethodID,
		IsDefault:             method.IsDefault,
		CreatedAt:             formatTime(method.CreatedAt),
	}
}

func ToPaymentMethodResponses(methods []entities.PaymentMethod) []resources.PaymentMethodResponse {
	items := make([]resources.PaymentMethodResponse, 0, len(methods))
	for _, method := range methods {
		items = append(items, ToPaymentMethodResponse(method))
	}
	return items
}

// ToPaymentResponse maps a payment entity to its API DTO. PaidAt is handled
// carefully: it is a pointer because a payment may not have been paid yet, so we
// only format a date string when a value is present, leaving it null otherwise.
func ToPaymentResponse(payment entities.Payment) resources.PaymentResponse {
	var paidAt *string
	if payment.PaidAt != nil {
		value := formatTime(*payment.PaidAt)
		paidAt = &value
	}
	return resources.PaymentResponse{
		PaymentID:             payment.PaymentID.String(),
		SubscriptionID:        payment.SubscriptionID.String(),
		UserID:                payment.UserID.String(),
		PaymentMethodID:       payment.PaymentMethodID.String(),
		Amount:                payment.Amount,
		Currency:              payment.Currency,
		Status:                payment.Status.String(),
		PaymentMethod:         payment.PaymentMethod,
		StripePaymentIntentID: payment.StripePaymentIntentID,
		PaidAt:                paidAt,
		CreatedAt:             formatTime(payment.CreatedAt),
	}
}

// ToPaymentResponses maps a whole slice by reusing the single-item function.
func ToPaymentResponses(payments []entities.Payment) []resources.PaymentResponse {
	items := make([]resources.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		items = append(items, ToPaymentResponse(payment))
	}
	return items
}

func ToInvoiceResponse(invoice entities.Invoice) resources.InvoiceResponse {
	return resources.InvoiceResponse{
		InvoiceID:     invoice.InvoiceID.String(),
		PaymentID:     invoice.PaymentID.String(),
		InvoiceNumber: invoice.InvoiceNumber,
		IssuedAt:      formatTime(invoice.IssuedAt),
		TotalAmount:   invoice.TotalAmount,
		PDFURL:        invoice.PDFURL,
	}
}

// formatTime renders a timestamp as a UTC RFC3339 string (e.g.
// "2026-06-03T10:00:00Z"), a standard, unambiguous format that clients in any
// timezone can parse reliably.
func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
