package configuration

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort                             string
	APIBasePath                            string
	AutoMigrate                            bool
	DatabaseURL                            string
	StripeSecretKey                        string
	StripeWebhookSecret                    string
	StripeCurrency                         string
	KafkaBrokers                           []string
	KafkaClientID                          string
	KafkaPaymentProcessedTopic             string
	KafkaPaymentFailedTopic                string
	KafkaInvoiceGeneratedTopic             string
	KafkaPaymentMethodAddedTopic           string
	KafkaSubscriptionCreatedTopic          string
	KafkaSubscriptionRenewalRequestedTopic string
	KafkaSubscriptionCancelledTopic        string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded, using environment variables: %v", err)
	}
	return Config{
		ServerPort:                             getEnv("SERVER_PORT", "8085"),
		APIBasePath:                            getEnv("API_BASE_PATH", "/api/v1"),
		AutoMigrate:                            strings.EqualFold(getEnv("DB_AUTO_MIGRATE", "true"), "true"),
		DatabaseURL:                            getEnv("DATABASE_URL", ""),
		StripeSecretKey:                        getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:                    getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeCurrency:                         getEnv("STRIPE_CURRENCY", "pen"),
		KafkaBrokers:                           splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaClientID:                          getEnv("KAFKA_CLIENT_ID", "payments-service"),
		KafkaPaymentProcessedTopic:             getEnv("KAFKA_PAYMENT_PROCESSED_TOPIC", "payment.processed"),
		KafkaPaymentFailedTopic:                getEnv("KAFKA_PAYMENT_FAILED_TOPIC", "payment.failed"),
		KafkaInvoiceGeneratedTopic:             getEnv("KAFKA_INVOICE_GENERATED_TOPIC", "invoice.generated"),
		KafkaPaymentMethodAddedTopic:           getEnv("KAFKA_PAYMENT_METHOD_ADDED_TOPIC", "payment.method.added"),
		KafkaSubscriptionCreatedTopic:          getEnv("KAFKA_SUBSCRIPTION_CREATED_TOPIC", "subscription.created"),
		KafkaSubscriptionRenewalRequestedTopic: getEnv("KAFKA_SUBSCRIPTION_RENEWAL_REQUESTED_TOPIC", "subscription.renewal.requested"),
		KafkaSubscriptionCancelledTopic:        getEnv("KAFKA_SUBSCRIPTION_CANCELLED_TOPIC", "subscription.cancelled"),
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}
