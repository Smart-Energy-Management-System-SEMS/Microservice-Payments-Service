package configuration

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const serviceName = "payments-service"

type Config struct {
	AppEnv                                 string
	ServerPort                             string
	APIBasePath                            string
	AutoMigrate                            bool
	ConfigServiceURL                       string
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
	cfg := Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		ServerPort:          getEnv("SERVER_PORT", "8085"),
		APIBasePath:         getEnv("API_BASE_PATH", "/api/v1"),
		AutoMigrate:         strings.EqualFold(getEnv("DB_AUTO_MIGRATE", "true"), "true"),
		ConfigServiceURL:    getEnv("CONFIG_SERVICE_URL", ""),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeCurrency:      getEnv("STRIPE_CURRENCY", "pen"),
		// Legacy/fallback config from env for backward compatibility.
		KafkaBrokers:                           splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaClientID:                          getEnv("KAFKA_CLIENT_ID", serviceName),
		KafkaPaymentProcessedTopic:             getEnv("KAFKA_PAYMENT_PROCESSED_TOPIC", "payment.processed"),
		KafkaPaymentFailedTopic:                getEnv("KAFKA_PAYMENT_FAILED_TOPIC", "payment.failed"),
		KafkaInvoiceGeneratedTopic:             getEnv("KAFKA_INVOICE_GENERATED_TOPIC", "invoice.generated"),
		KafkaPaymentMethodAddedTopic:           getEnv("KAFKA_PAYMENT_METHOD_ADDED_TOPIC", "payment.method.added"),
		KafkaSubscriptionCreatedTopic:          getEnv("KAFKA_SUBSCRIPTION_CREATED_TOPIC", "subscription.created"),
		KafkaSubscriptionRenewalRequestedTopic: getEnv("KAFKA_SUBSCRIPTION_RENEWAL_REQUESTED_TOPIC", "subscription.renewal.requested"),
		KafkaSubscriptionCancelledTopic:        getEnv("KAFKA_SUBSCRIPTION_CANCELLED_TOPIC", "subscription.cancelled"),
	}
	applyConfigServiceOverrides(context.Background(), &cfg)
	return cfg
}

func applyConfigServiceOverrides(ctx context.Context, cfg *Config) {
	if strings.TrimSpace(cfg.ConfigServiceURL) == "" {
		log.Println("config service disabled: CONFIG_SERVICE_URL not provided")
		return
	}
	client := NewConfigServiceClient(cfg.ConfigServiceURL)

	serviceCfg, err := client.GetServiceConfig(ctx, serviceName)
	if err != nil {
		log.Printf("config service service-config fetch failed: %v", err)
	} else {
		if serviceCfg.Port != "" {
			cfg.ServerPort = serviceCfg.Port
		}
		if serviceCfg.APIBasePath != "" {
			cfg.APIBasePath = serviceCfg.APIBasePath
		}
		if serviceCfg.StripeCurrency != "" {
			cfg.StripeCurrency = serviceCfg.StripeCurrency
		}
	}

	kafkaCfg, err := client.GetKafkaConfig(ctx)
	if err != nil {
		log.Printf("config service kafka-config fetch failed: %v", err)
		return
	}
	if len(kafkaCfg.BootstrapServers) > 0 {
		cfg.KafkaBrokers = kafkaCfg.BootstrapServers
	}
	if kafkaCfg.ClientID != "" {
		cfg.KafkaClientID = kafkaCfg.ClientID
	}
	cfg.KafkaPaymentProcessedTopic = firstNonEmpty(kafkaCfg.Topics.PaymentProcessed, cfg.KafkaPaymentProcessedTopic)
	cfg.KafkaPaymentFailedTopic = firstNonEmpty(kafkaCfg.Topics.PaymentFailed, cfg.KafkaPaymentFailedTopic)
	cfg.KafkaInvoiceGeneratedTopic = firstNonEmpty(kafkaCfg.Topics.InvoiceGenerated, cfg.KafkaInvoiceGeneratedTopic)
	cfg.KafkaPaymentMethodAddedTopic = firstNonEmpty(kafkaCfg.Topics.PaymentMethodAdded, cfg.KafkaPaymentMethodAddedTopic)
	cfg.KafkaSubscriptionCreatedTopic = firstNonEmpty(kafkaCfg.Topics.SubscriptionCreated, cfg.KafkaSubscriptionCreatedTopic)
	cfg.KafkaSubscriptionRenewalRequestedTopic = firstNonEmpty(kafkaCfg.Topics.SubscriptionRenewalRequested, cfg.KafkaSubscriptionRenewalRequestedTopic)
	cfg.KafkaSubscriptionCancelledTopic = firstNonEmpty(kafkaCfg.Topics.SubscriptionCancelled, cfg.KafkaSubscriptionCancelledTopic)
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

func firstNonEmpty(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
