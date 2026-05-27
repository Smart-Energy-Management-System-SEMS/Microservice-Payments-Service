package queryservices

import (
	"context"

	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/model/queries"
	"Microservice-Payments-Service/payments/domain/repositories"
	shareddomain "Microservice-Payments-Service/payments/shared/domain"
)

type PaymentMethodQueryService struct {
	repository repositories.PaymentMethodRepository
}

func NewPaymentMethodQueryService(repository repositories.PaymentMethodRepository) *PaymentMethodQueryService {
	return &PaymentMethodQueryService{repository: repository}
}

func (s *PaymentMethodQueryService) FindByUser(ctx context.Context, query queries.GetPaymentMethodsByUserQuery) ([]entities.PaymentMethod, error) {
	userID, err := shareddomain.ParseID(query.UserID)
	if err != nil {
		return nil, shareddomain.ErrInvalidUUID
	}
	return s.repository.FindByUserID(ctx, userID)
}
