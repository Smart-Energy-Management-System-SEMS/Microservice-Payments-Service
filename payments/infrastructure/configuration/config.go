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
	AppEnv                        string
	ServerPort                    string
	APIBasePath                   string
	SwaggerEnabled                bool
	SwaggerServerURL              string
	AutoMigrate                   bool
	KafkaEnsureTopics             bool
	ConfigServiceURL              string
	CORSAllowedOrigins            []string
	CORSAllowCredentials          bool
	DatabaseURL                   string
	StripeSecretKey               string
	StripeWebhookSecret           string
	StripeCurrency                string
	KafkaBrokers                  []string
	KafkaSecurityProtocol         string
	KafkaSASLMechanism            string
	KafkaUsername                 string
	KafkaPassword                 string
	KafkaClientID                 string
	KafkaConsumerGroup            string
	KafkaPaymentsEventsTopic      string
	KafkaBillingEventsTopic       string
	KafkaSubscriptionsEventsTopic string
}

// Load builds the configuration. It first tries to read a local .env file (handy
// for development; harmless in production where it just won't exist), then reads
// each value from the environment with a sensible default, and finally lets a
// central config service override selected values.
func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded, using environment variables: %v", err)
	}
	kafkaClientID := getEnv("KAFKA_CLIENT_ID", serviceName)
	cfg := Config{
		AppEnv:               getEnv("APP_ENV", "development"),
		ServerPort:           firstNonEmpty(getEnv("PORT", ""), getEnv("SERVER_PORT", "8085")),
		APIBasePath:          getEnv("API_BASE_PATH", "/api/v1"),
		SwaggerEnabled:       strings.EqualFold(getEnv("SWAGGER_ENABLED", "true"), "true"),
		SwaggerServerURL:     getEnv("SWAGGER_SERVER_URL", ""),
		AutoMigrate:          strings.EqualFold(getEnv("DB_AUTO_MIGRATE", "true"), "true"),
		KafkaEnsureTopics:    strings.EqualFold(getEnv("KAFKA_ENSURE_TOPICS", "false"), "true"),
		ConfigServiceURL:     getEnv("CONFIG_SERVICE_URL", ""),
		CORSAllowedOrigins:   splitCSV(getEnvWithFallback("CORS_ALLOWED_ORIGINS", "ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")),
		CORSAllowCredentials: strings.EqualFold(getEnv("CORS_ALLOW_CREDENTIALS", "false"), "true"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		StripeSecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeCurrency:       getEnv("STRIPE_CURRENCY", "pen"),
		// Read the official KAFKA_TOPIC_* names first, but keep the legacy
		// KAFKA_*_TOPIC variants as fallbacks so existing deployments keep working.
		KafkaBrokers:                  splitCSV(getEnv("KAFKA_BROKERS", "")),
		KafkaSecurityProtocol:         getEnv("KAFKA_SECURITY_PROTOCOL", ""),
		KafkaSASLMechanism:            getEnv("KAFKA_SASL_MECHANISM", ""),
		KafkaUsername:                 getEnvWithFallback("KAFKA_USERNAME", "KAFKA_SASL_USERNAME", ""),
		KafkaPassword:                 getEnvWithFallback("KAFKA_PASSWORD", "KAFKA_SASL_PASSWORD", ""),
		KafkaClientID:                 kafkaClientID,
		KafkaConsumerGroup:            firstNonEmpty(getEnvWithFallback("KAFKA_CONSUMER_GROUP", "KAFKA_GROUP_ID", ""), kafkaClientID+"-group"),
		KafkaPaymentsEventsTopic:      firstNonEmpty(getEnv("KAFKA_TOPIC_PAYMENTS_EVENTS", ""), "payments.events"),
		KafkaBillingEventsTopic:       firstNonEmpty(getEnv("KAFKA_TOPIC_BILLING_EVENTS", ""), "billing.events"),
		KafkaSubscriptionsEventsTopic: firstNonEmpty(getEnv("KAFKA_TOPIC_SUBSCRIPTIONS_EVENTS", ""), "subscriptions.events"),
	}
	cfg.KafkaUsername = normalizeKafkaUsername(cfg.KafkaUsername, cfg.KafkaBrokers)
	applyConfigServiceOverrides(context.Background(), &cfg)
	cfg.KafkaUsername = normalizeKafkaUsername(cfg.KafkaUsername, cfg.KafkaBrokers)
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
	} else if len(kafkaCfg.BrokerList) > 0 {
		cfg.KafkaBrokers = kafkaCfg.BrokerList
	} else if brokers := splitCSV(firstNonEmpty(kafkaCfg.BootstrapServersText, kafkaCfg.BrokersText)); len(brokers) > 0 {
		cfg.KafkaBrokers = brokers
	}
	if clientID := firstNonEmpty(kafkaCfg.ClientID, kafkaCfg.ClientIDAlt); clientID != "" {
		cfg.KafkaClientID = clientID
	}
	cfg.KafkaPaymentsEventsTopic = firstNonEmpty(
		kafkaCfg.Topics.PaymentsEvents,
		topicForService(serviceName, kafkaCfg.PublishTopics, "payments.events"),
		topicForService(serviceName, kafkaCfg.ProducedTopics, "payments.events"),
		cfg.KafkaPaymentsEventsTopic,
	)
	cfg.KafkaBillingEventsTopic = firstNonEmpty(
		kafkaCfg.Topics.BillingEvents,
		topicForService(serviceName, kafkaCfg.PublishTopics, "billing.events"),
		topicForService(serviceName, kafkaCfg.ProducedTopics, "billing.events"),
		topicForService(serviceName, kafkaCfg.ConsumeTopics, "billing.events"),
		topicForService(serviceName, kafkaCfg.ConsumedTopics, "billing.events"),
		cfg.KafkaBillingEventsTopic,
	)
	cfg.KafkaSubscriptionsEventsTopic = firstNonEmpty(
		kafkaCfg.Topics.SubscriptionsEvents,
		topicForService(serviceName, kafkaCfg.ConsumeTopics, "subscriptions.events"),
		topicForService(serviceName, kafkaCfg.ConsumedTopics, "subscriptions.events"),
		cfg.KafkaSubscriptionsEventsTopic,
	)
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

// getEnvWithFallback prefers the primary env var, then a compatibility alias,
// and finally the provided default.
func getEnvWithFallback(primaryKey string, legacyKey string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(primaryKey))
	if value != "" {
		return value
	}
	return getEnv(legacyKey, fallback)
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
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func topicForService(service string, topicsByService map[string][]string, expected string) string {
	if len(topicsByService) == 0 {
		return ""
	}
	expected = strings.TrimSpace(expected)
	for _, topic := range topicsByService[strings.TrimSpace(service)] {
		if strings.EqualFold(strings.TrimSpace(topic), expected) {
			return expected
		}
	}
	return ""
}

func normalizeKafkaUsername(username string, brokers []string) string {
	trimmed := strings.TrimSpace(username)
	if trimmed == "" {
		return trimmed
	}
	if !looksLikeAzureEventHubsBroker(brokers) {
		return trimmed
	}
	if strings.EqualFold(trimmed, "$ConnectionString") {
		return "$ConnectionString"
	}
	if strings.EqualFold(trimmed, "ConnectionString") || strings.EqualFold(trimmed, "onnectionString") {
		return "$ConnectionString"
	}
	return trimmed
}

func looksLikeAzureEventHubsBroker(brokers []string) bool {
	for _, broker := range brokers {
		if strings.Contains(strings.ToLower(strings.TrimSpace(broker)), ".servicebus.windows.net") {
			return true
		}
	}
	return false
}
