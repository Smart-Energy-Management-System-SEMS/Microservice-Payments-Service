package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ConfigServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

type ServiceConfigResponse struct {
	Name           string `json:"name"`
	Port           string `json:"port"`
	APIBasePath    string `json:"api_base_path"`
	StripeCurrency string `json:"stripe_currency"`
}

type KafkaTopicsConfig struct {
	PaymentsEvents               string `json:"payments_events"`
	BillingEvents                string `json:"billing_events"`
	SubscriptionsEvents          string `json:"subscriptions_events"`
	PaymentProcessed             string `json:"payment_processed"`
	PaymentFailed                string `json:"payment_failed"`
	InvoiceGenerated             string `json:"invoice_generated"`
	PaymentMethodAdded           string `json:"payment_method_added"`
	SubscriptionCreated          string `json:"subscription_created"`
	SubscriptionRenewalRequested string `json:"subscription_renewal_requested"`
	SubscriptionCancelled        string `json:"subscription_cancelled"`
}

type KafkaConfigResponse struct {
	BootstrapServers     []string            `json:"bootstrap_servers"`
	BootstrapServersText string              `json:"bootstrapServers"`
	BrokersText          string              `json:"brokers"`
	BrokerList           []string            `json:"brokerList"`
	ClientID             string              `json:"client_id"`
	ClientIDAlt          string              `json:"clientId"`
	Topics               KafkaTopicsConfig   `json:"topics"`
	PublishTopics        map[string][]string `json:"publishTopics"`
	ProducedTopics       map[string][]string `json:"producedTopics"`
	ConsumeTopics        map[string][]string `json:"consumeTopics"`
	ConsumedTopics       map[string][]string `json:"consumedTopics"`
}

type ServicesEnvelope struct {
	Services []ServiceConfigResponse `json:"services"`
}

type ServiceEnvelope struct {
	Service ServiceConfigResponse `json:"service"`
	Kafka   KafkaConfigResponse   `json:"kafka"`
}

type KafkaEnvelope struct {
	Kafka KafkaConfigResponse `json:"kafka"`
}

func NewConfigServiceClient(baseURL string) *ConfigServiceClient {
	return &ConfigServiceClient{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *ConfigServiceClient) GetServicesConfig(ctx context.Context) ([]ServiceConfigResponse, error) {
	var envelope ServicesEnvelope
	if err := c.getJSON(ctx, "/api/v1/config/services", &envelope); err != nil {
		return nil, err
	}
	return envelope.Services, nil
}

func (c *ConfigServiceClient) GetKafkaConfig(ctx context.Context) (KafkaConfigResponse, error) {
	var envelope KafkaEnvelope
	if err := c.getJSON(ctx, "/api/v1/config/kafka", &envelope); err == nil {
		return envelope.Kafka, nil
	}

	var direct KafkaConfigResponse
	if err := c.getJSON(ctx, "/api/v1/config/kafka", &direct); err != nil {
		return KafkaConfigResponse{}, err
	}
	return direct, nil
}

func (c *ConfigServiceClient) GetServiceConfig(ctx context.Context, serviceName string) (ServiceConfigResponse, error) {
	path := fmt.Sprintf("/api/v1/config/%s", url.PathEscape(strings.TrimSpace(serviceName)))

	var envelope ServiceEnvelope
	if err := c.getJSON(ctx, path, &envelope); err == nil {
		if envelope.Service.Name == "" {
			envelope.Service.Name = strings.TrimSpace(serviceName)
		}
		return envelope.Service, nil
	}

	var direct ServiceConfigResponse
	if err := c.getJSON(ctx, path, &direct); err != nil {
		return ServiceConfigResponse{}, err
	}
	if direct.Name == "" {
		direct.Name = strings.TrimSpace(serviceName)
	}
	return direct, nil
}

func (c *ConfigServiceClient) getJSON(ctx context.Context, path string, target interface{}) error {
	if c.baseURL == "" {
		return fmt.Errorf("config service base url is empty")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("config service response %d for %s", resp.StatusCode, path)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
