package kafkaadapter

import (
	"context"
	"encoding/json"
	"log"
	"time"

	segmentio "github.com/segmentio/kafka-go"

	"Microservice-Payments-Service/payments/domain/model/entities"
)

type Topics struct {
	PaymentProcessed              string
	PaymentFailed                 string
	InvoiceGenerated              string
	PaymentMethodAdded            string
	SubscriptionCreated           string
	SubscriptionRenewalRequested  string
	SubscriptionCancelled         string
}

type Producer struct {
	brokers []string
	topics  Topics
	writers map[string]*segmentio.Writer
}

func NewProducer(brokers []string, topics Topics) *Producer {
	return &Producer{brokers: brokers, topics: topics, writers: map[string]*segmentio.Writer{}}
}

func (p *Producer) PublishPaymentProcessed(ctx context.Context, payment entities.Payment, invoice entities.Invoice) error {
	return p.publish(ctx, p.topics.PaymentProcessed, payment.PaymentID.String(), map[string]interface{}{
		"event_type": "payment.processed",
		"occurred_at": time.Now().UTC(),
		"payment_id": payment.PaymentID,
		"subscription_id": payment.SubscriptionID,
		"user_id": payment.UserID,
		"amount": payment.Amount,
		"currency": payment.Currency,
		"status": payment.Status.String(),
		"invoice_id": invoice.InvoiceID,
	})
}

func (p *Producer) PublishPaymentFailed(ctx context.Context, payment entities.Payment) error {
	return p.publish(ctx, p.topics.PaymentFailed, payment.PaymentID.String(), map[string]interface{}{
		"event_type": "payment.failed",
		"occurred_at": time.Now().UTC(),
		"payment_id": payment.PaymentID,
		"subscription_id": payment.SubscriptionID,
		"user_id": payment.UserID,
		"amount": payment.Amount,
		"currency": payment.Currency,
		"status": payment.Status.String(),
	})
}

func (p *Producer) PublishInvoiceGenerated(ctx context.Context, invoice entities.Invoice) error {
	return p.publish(ctx, p.topics.InvoiceGenerated, invoice.InvoiceID.String(), map[string]interface{}{
		"event_type": "invoice.generated",
		"occurred_at": time.Now().UTC(),
		"invoice_id": invoice.InvoiceID,
		"payment_id": invoice.PaymentID,
		"invoice_number": invoice.InvoiceNumber,
		"issued_at": invoice.IssuedAt,
		"total_amount": invoice.TotalAmount,
		"pdf_url": invoice.PDFURL,
	})
}

func (p *Producer) PublishPaymentMethodAdded(ctx context.Context, method entities.PaymentMethod) error {
	return p.publish(ctx, p.topics.PaymentMethodAdded, method.PaymentMethodID.String(), map[string]interface{}{
		"event_type": "payment.method.added",
		"occurred_at": time.Now().UTC(),
		"payment_method_id": method.PaymentMethodID,
		"user_id": method.UserID,
		"type": method.Type,
		"brand": method.Brand,
		"last4": method.Last4,
		"is_default": method.IsDefault,
	})
}

func (p *Producer) Close() error {
	var lastErr error
	for _, writer := range p.writers {
		if err := writer.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

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
	return writer.WriteMessages(ctx, segmentio.Message{Key: []byte(key), Value: value})
}

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
