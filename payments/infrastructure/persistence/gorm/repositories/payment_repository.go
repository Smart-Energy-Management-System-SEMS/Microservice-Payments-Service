package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/model/valueobjects"
	gormmodel "Microservice-Payments-Service/payments/infrastructure/persistence/gorm/model"
)

// GormPaymentRepository is the database implementation of the domain's
// PaymentRepository interface, built on GORM (a Go ORM). It is an adapter: the
// domain declares WHAT persistence operations exist, this struct says HOW they
// run against SQL.
type GormPaymentRepository struct {
	db *gorm.DB
}

func NewGormPaymentRepository(db *gorm.DB) *GormPaymentRepository {
	return &GormPaymentRepository{db: db}
}

// Save inserts a new payment row. paymentToModel converts the domain entity into
// the DB-shaped struct, and WithContext(ctx) ties the query to the request
// context so it can be cancelled.
func (r *GormPaymentRepository) Save(ctx context.Context, payment *entities.Payment) error {
	model := paymentToModel(payment)
	return r.db.WithContext(ctx).Create(&model).Error
}

// Update writes changes to an existing payment row (GORM's Save upserts by
// primary key).
func (r *GormPaymentRepository) Update(ctx context.Context, payment *entities.Payment) error {
	model := paymentToModel(payment)
	return r.db.WithContext(ctx).Save(&model).Error
}

// FindByID loads one payment. The "?" is a parameterised query placeholder,
// which prevents SQL injection. On success we map the DB model back to a domain
// entity so callers only ever deal with domain types.
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

// FindByStripePaymentIntentID looks a payment up by the Stripe intent id. This
// is the lookup used when a webhook arrives: Stripe gives us the intent id, and
// we find the matching local payment to update.
func (r *GormPaymentRepository) FindByStripePaymentIntentID(ctx context.Context, stripePaymentIntentID string) (*entities.Payment, error) {
	var model gormmodel.PaymentModel
	if err := r.db.WithContext(ctx).First(&model, "stripe_payment_intent_id = ?", stripePaymentIntentID).Error; err != nil {
		return nil, err
	}
	entity := paymentToEntity(model)
	return &entity, nil
}

// paymentsToEntities maps a slice of DB models to a slice of domain entities.
func paymentsToEntities(models []gormmodel.PaymentModel) []entities.Payment {
	items := make([]entities.Payment, 0, len(models))
	for _, model := range models {
		items = append(items, paymentToEntity(model))
	}
	return items
}

// paymentToModel and paymentToEntity are the "mappers" that translate between
// the two representations of a payment: the domain entity (with its value-object
// status) and the persistence model (plain DB columns). Keeping them separate
// means the database schema and the domain can evolve independently.

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
