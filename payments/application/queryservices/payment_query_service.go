package queryservices

import (
	"context"

	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/model/queries"
	"Microservice-Payments-Service/payments/domain/repositories"
	shareddomain "Microservice-Payments-Service/payments/shared/domain"
)

type PaymentQueryService struct {
	repository repositories.PaymentRepository
}

func NewPaymentQueryService(repository repositories.PaymentRepository) *PaymentQueryService {
	return &PaymentQueryService{repository: repository}
}

func (s *PaymentQueryService) FindByID(ctx context.Context, id string) (*entities.Payment, error) {
	paymentID, err := shareddomain.ParseID(id)
	if err != nil {
		return nil, shareddomain.ErrInvalidUUID
	}
	return s.repository.FindByID(ctx, paymentID)
}

func (s *PaymentQueryService) FindByUser(ctx context.Context, query queries.GetPaymentsByUserQuery) ([]entities.Payment, error) {
	userID, err := shareddomain.ParseID(query.UserID)
	if err != nil {
		return nil, shareddomain.ErrInvalidUUID
	}
	return s.repository.FindByUserID(ctx, userID)
}

func (s *PaymentQueryService) FindBySubscription(ctx context.Context, query queries.GetPaymentsBySubscriptionQuery) ([]entities.Payment, error) {
	subscriptionID, err := shareddomain.ParseID(query.SubscriptionID)
	if err != nil {
		return nil, shareddomain.ErrInvalidUUID
	}
	return s.repository.FindBySubscriptionID(ctx, subscriptionID)
}
