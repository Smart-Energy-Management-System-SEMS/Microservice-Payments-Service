package kafkaadapter

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"

	segmentio "github.com/segmentio/kafka-go"

	"Microservice-Payments-Service/payments/application/eventhandlers"
	"Microservice-Payments-Service/payments/application/outboundservices"
)

type Consumer struct {
	brokers  []string
	clientID string
	topics   Topics
	handler  *eventhandlers.SubscriptionEventsHandler
	readers  []*segmentio.Reader
	wg       sync.WaitGroup
}

func NewConsumer(brokers []string, clientID string, topics Topics, handler *eventhandlers.SubscriptionEventsHandler) *Consumer {
	return &Consumer{brokers: brokers, clientID: clientID, topics: topics, handler: handler}
}

func (c *Consumer) Start(ctx context.Context) error {
	if len(c.brokers) == 0 {
		log.Println("kafka consumer disabled: no brokers configured")
		return nil
	}
	c.consume(ctx, c.topics.SubscriptionCreated, func(ctx context.Context, value []byte) error {
		var event outboundservices.SubscriptionCreatedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		return c.handler.HandleSubscriptionCreated(ctx, event)
	})
	c.consume(ctx, c.topics.SubscriptionRenewalRequested, func(ctx context.Context, value []byte) error {
		var event outboundservices.SubscriptionRenewalRequestedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		return c.handler.HandleSubscriptionRenewalRequested(ctx, event)
	})
	c.consume(ctx, c.topics.SubscriptionCancelled, func(ctx context.Context, value []byte) error {
		var event outboundservices.SubscriptionCancelledEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		return c.handler.HandleSubscriptionCancelled(ctx, event)
	})
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

func (c *Consumer) consume(ctx context.Context, topic string, handle func(context.Context, []byte) error) {
	if topic == "" {
		return
	}
	reader := segmentio.NewReader(segmentio.ReaderConfig{
		Brokers:  c.brokers,
		Topic:    topic,
		GroupID:  c.clientID + "-group",
		MinBytes: 1,
		MaxBytes: 10e6,
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
			if err := handle(ctx, message.Value); err != nil {
				log.Printf("kafka handler error topic=%s partition=%d offset=%d: %v", topic, message.Partition, message.Offset, err)
				continue
			}
			if err := reader.CommitMessages(ctx, message); err != nil {
				log.Printf("kafka commit error topic=%s: %v", topic, err)
			}
		}
	}()
}
