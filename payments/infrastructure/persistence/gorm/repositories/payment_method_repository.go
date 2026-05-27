package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"Microservice-Payments-Service/payments/domain/model/entities"
	gormmodel "Microservice-Payments-Service/payments/infrastructure/persistence/gorm/model"
)

type GormPaymentMethodRepository struct {
	db *gorm.DB
}

func NewGormPaymentMethodRepository(db *gorm.DB) *GormPaymentMethodRepository {
	return &GormPaymentMethodRepository{db: db}
}

func (r *GormPaymentMethodRepository) Save(ctx context.Context, method *entities.PaymentMethod) error {
	model := paymentMethodToModel(method)
	return r.db.WithContext(ctx).Save(&model).Error
}

func (r *GormPaymentMethodRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.PaymentMethod, error) {
	var model gormmodel.PaymentMethodModel
	if err := r.db.WithContext(ctx).First(&model, "payment_method_id = ?", id).Error; err != nil {
		return nil, err
	}
	entity := paymentMethodToEntity(model)
	return &entity, nil
}

func (r *GormPaymentMethodRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entities.PaymentMethod, error) {
	var models []gormmodel.PaymentMethodModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	items := make([]entities.PaymentMethod, 0, len(models))
	for _, model := range models {
		items = append(items, paymentMethodToEntity(model))
	}
	return items, nil
}

func (r *GormPaymentMethodRepository) ClearDefaultForUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&gormmodel.PaymentMethodModel{}).Where("user_id = ?", userID).Update("is_default", false).Error
}

func (r *GormPaymentMethodRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&gormmodel.PaymentMethodModel{}, "payment_method_id = ?", id).Error
}

func paymentMethodToModel(method *entities.PaymentMethod) gormmodel.PaymentMethodModel {
	return gormmodel.PaymentMethodModel{
		PaymentMethodID:       method.PaymentMethodID,
		UserID:                method.UserID,
		Type:                  method.Type,
		Brand:                 method.Brand,
		Last4:                 method.Last4,
		ExpMonth:              method.ExpMonth,
		ExpYear:               method.ExpYear,
		StripePaymentMethodID: method.StripePaymentMethodID,
		IsDefault:             method.IsDefault,
		CreatedAt:             method.CreatedAt,
	}
}

func paymentMethodToEntity(model gormmodel.PaymentMethodModel) entities.PaymentMethod {
	return entities.PaymentMethod{
		PaymentMethodID:       model.PaymentMethodID,
		UserID:                model.UserID,
		Type:                  model.Type,
		Brand:                 model.Brand,
		Last4:                 model.Last4,
		ExpMonth:              model.ExpMonth,
		ExpYear:               model.ExpYear,
		StripePaymentMethodID: model.StripePaymentMethodID,
		IsDefault:             model.IsDefault,
		CreatedAt:             model.CreatedAt,
	}
}
