package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"Microservice-Payments-Service/payments/domain/model/entities"
	gormmodel "Microservice-Payments-Service/payments/infrastructure/persistence/gorm/model"
)

type GormWebhookEventRepository struct {
	db *gorm.DB
}

func NewGormWebhookEventRepository(db *gorm.DB) *GormWebhookEventRepository {
	return &GormWebhookEventRepository{db: db}
}

func (r *GormWebhookEventRepository) Save(ctx context.Context, event *entities.PaymentWebhookEvent) error {
	model := gormmodel.PaymentWebhookEventModel{
		EventID:         event.EventID,
		Provider:        event.Provider,
		ProviderEventID: event.ProviderEventID,
		EventType:       event.EventType,
		Payload:         datatypes.JSON(event.Payload),
		Processed:       event.Processed,
		ReceivedAt:      event.ReceivedAt,
		ProcessedAt:     event.ProcessedAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormWebhookEventRepository) ExistsByProviderEventID(ctx context.Context, provider string, providerEventID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&gormmodel.PaymentWebhookEventModel{}).Where("provider = ? AND provider_event_id = ?", provider, providerEventID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormWebhookEventRepository) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&gormmodel.PaymentWebhookEventModel{}).Where("event_id = ?", id).Updates(map[string]interface{}{
		"processed": true,
		"processed_at": now,
	}).Error
}
