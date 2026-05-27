package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/model/valueobjects"
	gormmodel "Microservice-Payments-Service/payments/infrastructure/persistence/gorm/model"
)

type GormPaymentRepository struct {
	db *gorm.DB
}

func NewGormPaymentRepository(db *gorm.DB) *GormPaymentRepository {
	return &GormPaymentRepository{db: db}
}

func (r *GormPaymentRepository) Save(ctx context.Context, payment *entities.Payment) error {
	model := paymentToModel(payment)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormPaymentRepository) Update(ctx context.Context, payment *entities.Payment) error {
	model := paymentToModel(payment)
	return r.db.WithContext(ctx).Save(&model).Error
}

func (r *GormPaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error) {
	var model gormmodel.PaymentModel
	if err := r.db.WithContext(ctx).First(&model, "payment_id = ?", id).Error; err != nil {
		return nil, err
	}
	entity := paymentToEntity(model)
	return &entity, nil
}

func (r *GormPaymentRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Payment, error) {
	var models []gormmodel.PaymentModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return paymentsToEntities(models), nil
}

func (r *GormPaymentRepository) FindBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]entities.Payment, error) {
	var models []gormmodel.PaymentModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models, "subscription_id = ?", subscriptionID).Error; err != nil {
		return nil, err
	}
	return paymentsToEntities(models), nil
}

func (r *GormPaymentRepository) FindByStripePaymentIntentID(ctx context.Context, stripePaymentIntentID string) (*entities.Payment, error) {
	var model gormmodel.PaymentModel
	if err := r.db.WithContext(ctx).First(&model, "stripe_payment_intent_id = ?", stripePaymentIntentID).Error; err != nil {
		return nil, err
	}
	entity := paymentToEntity(model)
	return &entity, nil
}

func paymentsToEntities(models []gormmodel.PaymentModel) []entities.Payment {
	items := make([]entities.Payment, 0, len(models))
	for _, model := range models {
		items = append(items, paymentToEntity(model))
	}
	return items
}

func paymentToModel(payment *entities.Payment) gormmodel.PaymentModel {
	return gormmodel.PaymentModel{
		PaymentID:             payment.PaymentID,
		SubscriptionID:        payment.SubscriptionID,
		UserID:                payment.UserID,
		PaymentMethodID:       payment.PaymentMethodID,
		Amount:                payment.Amount,
		Currency:              payment.Currency,
		Status:                payment.Status.String(),
		PaymentMethod:         payment.PaymentMethod,
		StripePaymentIntentID: payment.StripePaymentIntentID,
		PaidAt:                payment.PaidAt,
		CreatedAt:             payment.CreatedAt,
	}
}

func paymentToEntity(model gormmodel.PaymentModel) entities.Payment {
	return entities.Payment{
		PaymentID:             model.PaymentID,
		SubscriptionID:        model.SubscriptionID,
		UserID:                model.UserID,
		PaymentMethodID:       model.PaymentMethodID,
		Amount:                model.Amount,
		Currency:              model.Currency,
		Status:                valueobjects.PaymentStatus(model.Status),
		PaymentMethod:         model.PaymentMethod,
		StripePaymentIntentID: model.StripePaymentIntentID,
		PaidAt:                model.PaidAt,
		CreatedAt:             model.CreatedAt,
	}
}
