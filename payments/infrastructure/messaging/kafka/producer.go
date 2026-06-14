// Package kafkaadapter is the messaging adapter: it implements the EventPublisher
// port by sending domain events to Apache Kafka. Publishing events lets other
// microservices react to what happens here (a payment processed, an invoice
// generated, ...) without being directly coupled to this service.
package kafkaadapter

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	segmentio "github.com/segmentio/kafka-go"

	"Microservice-Payments-Service/payments/domain/model/entities"
)

type Topics struct {
	PaymentsEvents      string
	BillingEvents       string
	SubscriptionsEvents string
}

// Producer sends messages to Kafka. It caches one writer per topic in the
// "writers" map and reuses them, since creating a writer is comparatively
// expensive.
type Producer struct {
	config  ConnectionConfig
	topics  Topics
	mu      sync.Mutex
	writers map[string]*segmentio.Writer
}

// NewProducer builds a Producer with an initialised (non-nil) writers map.
func NewProducer(config ConnectionConfig, topics Topics) *Producer {
	return &Producer{config: config, topics: topics, writers: map[string]*segmentio.Writer{}}
}

// The Publish* methods each build the JSON payload for one event type and hand
// it to the shared publish() helper. They use the payment/invoice id as the
// Kafka message key so all events about the same entity go to the same partition
// and therefore preserve their order.

func (p *Producer) PublishPaymentProcessed(ctx context.Context, payment entities.Payment, invoice entities.Invoice) error {
	return p.publish(ctx, p.topics.PaymentsEvents, payment.PaymentID.String(), "payment.processed", map[string]interface{}{
		"payment_id":      payment.PaymentID,
		"subscription_id": payment.SubscriptionID,
		"user_id":         payment.UserID,
		"amount":          payment.Amount,
		"currency":        payment.Currency,
		"status":          payment.Status.String(),
		"invoice_id":      invoice.InvoiceID,
	})
}

func (p *Producer) PublishPaymentFailed(ctx context.Context, payment entities.Payment) error {
	return p.publish(ctx, p.topics.PaymentsEvents, payment.PaymentID.String(), "payment.failed", map[string]interface{}{
		"payment_id":      payment.PaymentID,
		"subscription_id": payment.SubscriptionID,
		"user_id":         payment.UserID,
		"amount":          payment.Amount,
		"currency":        payment.Currency,
		"status":          payment.Status.String(),
	})
}

func (p *Producer) PublishInvoiceGenerated(ctx context.Context, invoice entities.Invoice) error {
	return p.publish(ctx, p.topics.BillingEvents, invoice.InvoiceID.String(), "invoice.generated", map[string]interface{}{
		"invoice_id":     invoice.InvoiceID,
		"payment_id":     invoice.PaymentID,
		"invoice_number": invoice.InvoiceNumber,
		"issued_at":      invoice.IssuedAt,
		"total_amount":   invoice.TotalAmount,
		"pdf_url":        invoice.PDFURL,
	})
}

func (p *Producer) PublishPaymentMethodAdded(ctx context.Context, method entities.PaymentMethod) error {
	return p.publish(ctx, p.topics.PaymentsEvents, method.PaymentMethodID.String(), "payment.method.added", map[string]interface{}{
		"payment_method_id": method.PaymentMethodID,
		"user_id":           method.UserID,
		"type":              method.Type,
		"brand":             method.Brand,
		"last4":             method.Last4,
		"is_default":        method.IsDefault,
	})
}

// Close shuts down all cached writers, e.g. during graceful shutdown. It keeps
// the last error but still tries to close every writer.
func (p *Producer) Close() error {
	var lastErr error
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, writer := range p.writers {
		if err := writer.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// publish is the shared low-level send. If messaging is not configured (no topic
// or no brokers) it logs and returns nil instead of failing — this lets the
// service run locally without a Kafka cluster. Otherwise it serialises the
// payload to JSON and writes the message.
func (p *Producer) publish(ctx context.Context, topic string, key string, eventType string, data interface{}) error {
	if topic == "" || len(p.config.Brokers) == 0 {
		log.Printf("kafka publish skipped eventType=%s topic=%s brokers=%v", eventType, topic, p.config.Brokers)
		return nil
	}
	log.Printf("kafka publish started eventType=%s topic=%s key=%s", eventType, topic, key)
	value, err := json.Marshal(map[string]interface{}{
		"eventType":  eventType,
		"occurredAt": time.Now().UTC(),
		"data":       data,
	})
	if err != nil {
		return err
	}
	writer := p.writer(topic)
	if err := writer.WriteMessages(ctx, segmentio.Message{Key: []byte(key), Value: value}); err != nil {
		log.Printf("kafka publish failed eventType=%s topic=%s key=%s brokers=%v err=%v", eventType, topic, key, p.config.Brokers, err)
		return err
	}
	log.Printf("kafka publish succeeded eventType=%s topic=%s key=%s", eventType, topic, key)
	return nil
}

// writer returns the cached writer for a topic, creating it on first use
// ("lazy initialisation"). LeastBytes balances messages toward the least-loaded
// partition.
func (p *Producer) writer(topic string) *segmentio.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()
	if writer, ok := p.writers[topic]; ok {
		return writer
	}
	writer := &segmentio.Writer{
		Addr:      segmentio.TCP(p.config.Brokers...),
		Topic:     topic,
		Balancer:  &segmentio.LeastBytes{},
		Transport: p.config.transport(),
	}
	p.writers[topic] = writer
	return writer
}
