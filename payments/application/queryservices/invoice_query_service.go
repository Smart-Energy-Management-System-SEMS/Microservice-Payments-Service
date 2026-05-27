package queryservices

import (
	"context"

	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/repositories"
	shareddomain "Microservice-Payments-Service/payments/shared/domain"
)

type InvoiceQueryService struct {
	repository repositories.InvoiceRepository
}

func NewInvoiceQueryService(repository repositories.InvoiceRepository) *InvoiceQueryService {
	return &InvoiceQueryService{repository: repository}
}

func (s *InvoiceQueryService) FindByID(ctx context.Context, id string) (*entities.Invoice, error) {
	invoiceID, err := shareddomain.ParseID(id)
	if err != nil {
		return nil, shareddomain.ErrInvalidUUID
	}
	return s.repository.FindByID(ctx, invoiceID)
}

func (s *InvoiceQueryService) FindByPaymentID(ctx context.Context, id string) (*entities.Invoice, error) {
	paymentID, err := shareddomain.ParseID(id)
	if err != nil {
		return nil, shareddomain.ErrInvalidUUID
	}
	return s.repository.FindByPaymentID(ctx, paymentID)
}
