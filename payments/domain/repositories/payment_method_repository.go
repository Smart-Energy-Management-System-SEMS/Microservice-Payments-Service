package repositories

import (
	"context"

	"github.com/google/uuid"

	"Microservice-Payments-Service/payments/domain/model/entities"
)

type PaymentMethodRepository interface {
	Save(ctx context.Context, method *entities.PaymentMethod) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.PaymentMethod, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]entities.PaymentMethod, error)
	ClearDefaultForUser(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}
