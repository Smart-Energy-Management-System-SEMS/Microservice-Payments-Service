package queryservices

import (
	"context"

	paymentdomain "Microservice-Payments-Service/payments/domain"
	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/model/queries"
	"Microservice-Payments-Service/payments/domain/repositories"
)

type PaymentMethodQueryService struct {
	repository repositories.PaymentMethodRepository
}

func NewPaymentMethodQueryService(repository repositories.PaymentMethodRepository) *PaymentMethodQueryService {
	return &PaymentMethodQueryService{repository: repository}
}

func (s *PaymentMethodQueryService) FindByUser(ctx context.Context, query queries.GetPaymentMethodsByUserQuery) ([]entities.PaymentMethod, error) {
	userID, err := paymentdomain.ParseID(query.UserID)
	if err != nil {
		return nil, paymentdomain.ErrInvalidUUID
	}
	return s.repository.FindByUserID(ctx, userID)
}
