// Package configuration loads every setting the service needs. It follows the
// Twelve-Factor "config in the environment" idea: hard-coded defaults are
// overridden by environment variables (and a local .env file), which in turn can
// be overridden by a central config service. This keeps secrets and per-env
// values out of the source code.
package configuration

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const serviceName = "payments-service"

// Config is the full, flat set of settings used across the service.
type Config struct {
	AppEnv                                 string
	ServerPort                             string
	APIBasePath                            string
	AutoMigrate                            bool
	KafkaEnsureTopics                      bool
	ConfigServiceURL                       string
	CORSAllowedOrigins                     []string
	CORSAllowCredentials                   bool
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

// Load builds the configuration. It first tries to read a local .env file (handy
// for development; harmless in production where it just won't exist), then reads
// each value from the environment with a sensible default, and finally lets a
// central config service override selected values.
func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded, using environment variables: %v", err)
	}
	cfg := Config{
		AppEnv:               getEnv("APP_ENV", "development"),
		ServerPort:           firstNonEmpty(getEnv("PORT", ""), getEnv("SERVER_PORT", "8085")),
		APIBasePath:          getEnv("API_BASE_PATH", "/api/v1"),
		AutoMigrate:          strings.EqualFold(getEnv("DB_AUTO_MIGRATE", "true"), "true"),
		KafkaEnsureTopics:    strings.EqualFold(getEnv("KAFKA_ENSURE_TOPICS", "false"), "true"),
		ConfigServiceURL:     getEnv("CONFIG_SERVICE_URL", ""),
		CORSAllowedOrigins:   splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")),
		CORSAllowCredentials: strings.EqualFold(getEnv("CORS_ALLOW_CREDENTIALS", "false"), "true"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		StripeSecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeCurrency:       getEnv("STRIPE_CURRENCY", "pen"),
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

// applyConfigServiceOverrides asks a central config service for values and
// overlays them on top of the local config. If no config service URL is set, it
// simply returns (local config is enough). The rule throughout is "only override
// when the remote actually provided a non-empty value", so a partial remote
// response never wipes a good local default. A failed fetch is logged but not
// fatal — the service keeps running with what it already has (resilience).
func applyConfigServiceOverrides(ctx context.Context, cfg *Config) {
	if strings.TrimSpace(cfg.ConfigServiceURL) == "" {
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

// The helpers below keep the Load function clean and free of repetition.

// getEnv reads an environment variable, returning the fallback when it is unset
// or blank.
func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

// splitCSV turns "a, b ,c" into a clean slice ["a","b","c"], dropping empties.
// Used for list settings like CORS origins and Kafka brokers.
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

// firstNonEmpty returns the value if it has content, otherwise the fallback.
// It is the building block for the "only override when provided" merge rule.
func firstNonEmpty(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
