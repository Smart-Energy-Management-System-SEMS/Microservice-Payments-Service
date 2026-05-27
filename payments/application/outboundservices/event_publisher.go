package outboundservices

import (
	"context"

	"Microservice-Payments-Service/payments/domain/model/entities"
)

type EventPublisher interface {
	PublishPaymentProcessed(ctx context.Context, payment entities.Payment, invoice entities.Invoice) error
	PublishPaymentFailed(ctx context.Context, payment entities.Payment) error
	PublishInvoiceGenerated(ctx context.Context, invoice entities.Invoice) error
	PublishPaymentMethodAdded(ctx context.Context, method entities.PaymentMethod) error
	Close() error
}
