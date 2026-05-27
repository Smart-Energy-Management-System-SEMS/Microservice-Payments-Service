package transform

import (
	"time"

	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/interfaces/rest/resources"
)

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

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
