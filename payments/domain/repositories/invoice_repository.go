package repositories

import (
	"context"

	"github.com/google/uuid"

	"Microservice-Payments-Service/payments/domain/model/entities"
)

type InvoiceRepository interface {
	Save(ctx context.Context, invoice *entities.Invoice) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Invoice, error)
	FindByPaymentID(ctx context.Context, paymentID uuid.UUID) (*entities.Invoice, error)
}
