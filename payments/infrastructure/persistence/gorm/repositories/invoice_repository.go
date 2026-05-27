package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"Microservice-Payments-Service/payments/domain/model/entities"
	gormmodel "Microservice-Payments-Service/payments/infrastructure/persistence/gorm/model"
)

type GormInvoiceRepository struct {
	db *gorm.DB
}

func NewGormInvoiceRepository(db *gorm.DB) *GormInvoiceRepository {
	return &GormInvoiceRepository{db: db}
}

func (r *GormInvoiceRepository) Save(ctx context.Context, invoice *entities.Invoice) error {
	model := invoiceToModel(invoice)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormInvoiceRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Invoice, error) {
	var model gormmodel.InvoiceModel
	if err := r.db.WithContext(ctx).First(&model, "invoice_id = ?", id).Error; err != nil {
		return nil, err
	}
	entity := invoiceToEntity(model)
	return &entity, nil
}

func (r *GormInvoiceRepository) FindByPaymentID(ctx context.Context, paymentID uuid.UUID) (*entities.Invoice, error) {
	var model gormmodel.InvoiceModel
	if err := r.db.WithContext(ctx).First(&model, "payment_id = ?", paymentID).Error; err != nil {
		return nil, err
	}
	entity := invoiceToEntity(model)
	return &entity, nil
}

func invoiceToModel(invoice *entities.Invoice) gormmodel.InvoiceModel {
	return gormmodel.InvoiceModel{
		InvoiceID:     invoice.InvoiceID,
		PaymentID:     invoice.PaymentID,
		InvoiceNumber: invoice.InvoiceNumber,
		IssuedAt:      invoice.IssuedAt,
		TotalAmount:   invoice.TotalAmount,
		PDFURL:        invoice.PDFURL,
	}
}

func invoiceToEntity(model gormmodel.InvoiceModel) entities.Invoice {
	return entities.Invoice{
		InvoiceID:     model.InvoiceID,
		PaymentID:     model.PaymentID,
		InvoiceNumber: model.InvoiceNumber,
		IssuedAt:      model.IssuedAt,
		TotalAmount:   model.TotalAmount,
		PDFURL:        model.PDFURL,
	}
}
