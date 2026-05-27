package configuration

import (
	"errors"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	gormmodel "Microservice-Payments-Service/payments/infrastructure/persistence/gorm/model"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Warn,
			Colorful:      true,
		},
	)
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{Logger: gormLogger})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&gormmodel.PaymentMethodModel{},
		&gormmodel.PaymentModel{},
		&gormmodel.InvoiceModel{},
		&gormmodel.PaymentWebhookEventModel{},
	)
}
