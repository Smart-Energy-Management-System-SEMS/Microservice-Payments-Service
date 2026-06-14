package kafkaadapter

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"

	segmentio "github.com/segmentio/kafka-go"

	"Microservice-Payments-Service/payments/application/eventhandlers"
	"Microservice-Payments-Service/payments/application/outboundservices"
	"Microservice-Payments-Service/payments/interfaces/acl"
)

type Consumer struct {
	config  ConnectionConfig
	topics  Topics
	handler subscriptionEventHandler
	readers []*segmentio.Reader
	wg      sync.WaitGroup
}

type subscriptionEventHandler interface {
	HandleSubscriptionCreated(ctx context.Context, event outboundservices.SubscriptionCreatedEvent) error
	HandleSubscriptionRenewalRequested(ctx context.Context, event outboundservices.SubscriptionRenewalRequestedEvent) error
	HandleSubscriptionCancelled(ctx context.Context, event outboundservices.SubscriptionCancelledEvent) error
	HandleBillingPaymentRequested(ctx context.Context, event outboundservices.BillingPaymentRequestedEvent) error
}

func NewConsumer(config ConnectionConfig, topics Topics, handler *eventhandlers.SubscriptionEventsHandler) *Consumer {
	return &Consumer{config: config, topics: topics, handler: handler}
}

func (c *Consumer) Start(ctx context.Context) error {
	if len(c.config.Brokers) == 0 {
		log.Println("kafka consumer disabled: no brokers configured")
		return nil
	}
	for _, topic := range c.consumedTopics() {
		c.consume(ctx, topic, func(ctx context.Context, topic string, value []byte) error {
			return c.handleMessage(ctx, topic, value)
		})
	}
	return nil
}

func (c *Consumer) Close() error {
	var lastErr error
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			lastErr = err
		}
	}
	c.wg.Wait()
	return lastErr
}

func (c *Consumer) consume(ctx context.Context, topic string, handle func(context.Context, string, []byte) error) {
	if topic == "" {
		return
	}
	reader := segmentio.NewReader(segmentio.ReaderConfig{
		Brokers:  c.config.Brokers,
		Topic:    topic,
		GroupID:  c.config.ConsumerGroup,
		MinBytes: 1,
		MaxBytes: 10e6,
		Dialer:   c.config.dialer(),
	})
	c.readers = append(c.readers, reader)
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			message, err := reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				log.Printf("kafka fetch error topic=%s: %v", topic, err)
				continue
			}
			if err := handle(ctx, topic, message.Value); err != nil {
				log.Printf("kafka handler error topic=%s partition=%d offset=%d: %v", topic, message.Partition, message.Offset, err)
				continue
			}
			if err := reader.CommitMessages(ctx, message); err != nil {
				log.Printf("kafka commit error topic=%s: %v", topic, err)
			}
		}
	}()
}

func (c *Consumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	eventType, err := acl.EventType(value)
	if err != nil {
		return err
	}
	switch strings.TrimSpace(eventType) {
	case "subscription.created":
		event, err := acl.TranslateSubscriptionCreated(value)
		if err != nil {
			return err
		}
		return c.handler.HandleSubscriptionCreated(ctx, event)
	case "subscription.renewal.requested":
		event, err := acl.TranslateSubscriptionRenewalRequested(value)
		if err != nil {
			return err
		}
		return c.handler.HandleSubscriptionRenewalRequested(ctx, event)
	case "subscription.cancelled":
		event, err := acl.TranslateSubscriptionCancelled(value)
		if err != nil {
			return err
		}
		return c.handler.HandleSubscriptionCancelled(ctx, event)
	case "billing.payment.requested", "payment.requested":
		event, err := acl.TranslateBillingPaymentRequested(value)
		if err != nil {
			return err
		}
		return c.handler.HandleBillingPaymentRequested(ctx, event)
	case "invoice.generated":
		log.Printf("kafka event ignored topic=%s eventType=%s reason=downstream-output", topic, eventType)
		return nil
	default:
		log.Printf("kafka event ignored topic=%s eventType=%s", topic, eventType)
		return nil
	}
}

func (c *Consumer) consumedTopics() []string {
	return uniqueTopics(c.topics.SubscriptionsEvents, c.topics.BillingEvents)
}
