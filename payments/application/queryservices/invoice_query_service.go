package queryservices

import (
	"context"

	paymentdomain "Microservice-Payments-Service/payments/domain"
	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/repositories"
)

type InvoiceQueryService struct {
	repository repositories.InvoiceRepository
}

func NewInvoiceQueryService(repository repositories.InvoiceRepository) *InvoiceQueryService {
	return &InvoiceQueryService{repository: repository}
}

func (s *InvoiceQueryService) FindByID(ctx context.Context, id string) (*entities.Invoice, error) {
	invoiceID, err := paymentdomain.ParseID(id)
	if err != nil {
		return nil, paymentdomain.ErrInvalidUUID
	}
	return s.repository.FindByID(ctx, invoiceID)
}

func (s *InvoiceQueryService) FindByPaymentID(ctx context.Context, id string) (*entities.Invoice, error) {
	paymentID, err := paymentdomain.ParseID(id)
	if err != nil {
		return nil, paymentdomain.ErrInvalidUUID
	}
	return s.repository.FindByPaymentID(ctx, paymentID)
}
