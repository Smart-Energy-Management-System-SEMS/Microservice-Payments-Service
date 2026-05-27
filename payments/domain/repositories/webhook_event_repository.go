package repositories

import (
	"context"

	"github.com/google/uuid"

	"Microservice-Payments-Service/payments/domain/model/entities"
)

type WebhookEventRepository interface {
	Save(ctx context.Context, event *entities.PaymentWebhookEvent) error
	ExistsByProviderEventID(ctx context.Context, provider string, providerEventID string) (bool, error)
	MarkProcessed(ctx context.Context, id uuid.UUID) error
}
