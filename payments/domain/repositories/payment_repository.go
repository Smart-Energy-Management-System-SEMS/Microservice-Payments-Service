package repositories

import (
	"context"

	"github.com/google/uuid"

	"Microservice-Payments-Service/payments/domain/model/entities"
)

type PaymentRepository interface {
	Save(ctx context.Context, payment *entities.Payment) error
	Update(ctx context.Context, payment *entities.Payment) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Payment, error)
	FindBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]entities.Payment, error)
	FindByStripePaymentIntentID(ctx context.Context, stripePaymentIntentID string) (*entities.Payment, error)
}
