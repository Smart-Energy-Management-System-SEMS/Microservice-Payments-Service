package configuration

import (
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	gormmodel "Microservice-Payments-Service/payments/infrastructure/persistence/gorm/model"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&gormmodel.PaymentMethodModel{},
		&gormmodel.PaymentModel{},
		&gormmodel.InvoiceModel{},
		&gormmodel.PaymentWebhookEventModel{},
	)
}
