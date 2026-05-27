package services

import "Microservice-Payments-Service/payments/domain/model/valueobjects"

type PaymentStatusMapper struct{}

func (PaymentStatusMapper) FromStripe(status string) valueobjects.PaymentStatus {
	switch status {
	case "succeeded":
		return valueobjects.PaymentStatusProcessed
	case "requires_payment_method", "payment_failed":
		return valueobjects.PaymentStatusFailed
	case "canceled", "cancelled":
		return valueobjects.PaymentStatusCancelled
	case "processing", "requires_confirmation", "requires_action":
		return valueobjects.PaymentStatusProcessing
	default:
		return valueobjects.PaymentStatusProcessing
	}
}
