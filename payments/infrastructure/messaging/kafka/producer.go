// Package kafkaadapter is the messaging adapter: it implements the EventPublisher
// port by sending domain events to Apache Kafka. Publishing events lets other
// microservices react to what happens here (a payment processed, an invoice
// generated, ...) without being directly coupled to this service.
package kafkaadapter

import (
	"context"
	"encoding/json"
	"log"
	"time"

	segmentio "github.com/segmentio/kafka-go"

	"Microservice-Payments-Service/payments/domain/model/entities"
)

// Topics maps each kind of event to the Kafka topic name it is published on.
// Keeping the names in config (not hard-coded) lets each environment use its own.
type Topics struct {
	PaymentProcessed             string
	PaymentFailed                string
	InvoiceGenerated             string
	PaymentMethodAdded           string
	SubscriptionCreated          string
	SubscriptionRenewalRequested string
	SubscriptionCancelled        string
}

// Producer sends messages to Kafka. It caches one writer per topic in the
// "writers" map and reuses them, since creating a writer is comparatively
// expensive.
type Producer struct {
	brokers []string
	topics  Topics
	writers map[string]*segmentio.Writer
}

// NewProducer builds a Producer with an initialised (non-nil) writers map.
func NewProducer(brokers []string, topics Topics) *Producer {
	return &Producer{brokers: brokers, topics: topics, writers: map[string]*segmentio.Writer{}}
}

// The Publish* methods each build the JSON payload for one event type and hand
// it to the shared publish() helper. They use the payment/invoice id as the
// Kafka message key so all events about the same entity go to the same partition
// and therefore preserve their order.

func (p *Producer) PublishPaymentProcessed(ctx context.Context, payment entities.Payment, invoice entities.Invoice) error {
	return p.publish(ctx, p.topics.PaymentProcessed, payment.PaymentID.String(), map[string]interface{}{
		"event_type":      "payment.processed",
		"occurred_at":     time.Now().UTC(),
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
	return p.publish(ctx, p.topics.PaymentFailed, payment.PaymentID.String(), map[string]interface{}{
		"event_type":      "payment.failed",
		"occurred_at":     time.Now().UTC(),
		"payment_id":      payment.PaymentID,
		"subscription_id": payment.SubscriptionID,
		"user_id":         payment.UserID,
		"amount":          payment.Amount,
		"currency":        payment.Currency,
		"status":          payment.Status.String(),
	})
}

func (p *Producer) PublishInvoiceGenerated(ctx context.Context, invoice entities.Invoice) error {
	return p.publish(ctx, p.topics.InvoiceGenerated, invoice.InvoiceID.String(), map[string]interface{}{
		"event_type":     "invoice.generated",
		"occurred_at":    time.Now().UTC(),
		"invoice_id":     invoice.InvoiceID,
		"payment_id":     invoice.PaymentID,
		"invoice_number": invoice.InvoiceNumber,
		"issued_at":      invoice.IssuedAt,
		"total_amount":   invoice.TotalAmount,
		"pdf_url":        invoice.PDFURL,
	})
}

func (p *Producer) PublishPaymentMethodAdded(ctx context.Context, method entities.PaymentMethod) error {
	return p.publish(ctx, p.topics.PaymentMethodAdded, method.PaymentMethodID.String(), map[string]interface{}{
		"event_type":        "payment.method.added",
		"occurred_at":       time.Now().UTC(),
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
func (p *Producer) publish(ctx context.Context, topic string, key string, payload interface{}) error {
	if topic == "" || len(p.brokers) == 0 {
		log.Printf("kafka publish skipped for topic=%s", topic)
		return nil
	}
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	writer := p.writer(topic)
	if err := writer.WriteMessages(ctx, segmentio.Message{Key: []byte(key), Value: value}); err != nil {
		log.Printf("kafka publish failed topic=%s key=%s brokers=%v err=%v", topic, key, p.brokers, err)
		return err
	}
	log.Printf("kafka publish succeeded topic=%s key=%s", topic, key)
	return nil
}

// writer returns the cached writer for a topic, creating it on first use
// ("lazy initialisation"). LeastBytes balances messages toward the least-loaded
// partition.
func (p *Producer) writer(topic string) *segmentio.Writer {
	if writer, ok := p.writers[topic]; ok {
		return writer
	}
	writer := &segmentio.Writer{
		Addr:     segmentio.TCP(p.brokers...),
		Topic:    topic,
		Balancer: &segmentio.LeastBytes{},
	}
	p.writers[topic] = writer
	return writer
}
