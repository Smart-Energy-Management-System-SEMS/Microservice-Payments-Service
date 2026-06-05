package kafkaadapter

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	segmentio "github.com/segmentio/kafka-go"
)

const (
	defaultTopicPartitions        = 1
	defaultTopicReplicationFactor = 1
)

// EnsureTopics makes sure every configured topic exists in the broker before
// producers and consumers start using them. This keeps local/dev setups
// convenient and makes first-run startup on a new machine more predictable.
func EnsureTopics(brokers []string, topics Topics) error {
	if len(brokers) == 0 {
		log.Println("kafka topic ensure skipped: no brokers configured")
		return nil
	}

	topicNames := uniqueTopics(
		topics.PaymentProcessed,
		topics.PaymentFailed,
		topics.InvoiceGenerated,
		topics.PaymentMethodAdded,
		topics.SubscriptionCreated,
		topics.SubscriptionRenewalRequested,
		topics.SubscriptionCancelled,
	)
	if len(topicNames) == 0 {
		log.Println("kafka topic ensure skipped: no topics configured")
		return nil
	}

	conn, err := segmentio.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("dial broker %s: %w", brokers[0], err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("resolve kafka controller: %w", err)
	}

	controllerAddr := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	controllerConn, err := segmentio.Dial("tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("dial kafka controller %s: %w", controllerAddr, err)
	}
	defer controllerConn.Close()

	configs := make([]segmentio.TopicConfig, 0, len(topicNames))
	for _, topic := range topicNames {
		configs = append(configs, segmentio.TopicConfig{
			Topic:             topic,
			NumPartitions:     defaultTopicPartitions,
			ReplicationFactor: defaultTopicReplicationFactor,
		})
	}

	if err := controllerConn.CreateTopics(configs...); err != nil {
		if isTopicAlreadyExistsError(err) {
			log.Printf("kafka topics already existed: %v", topicNames)
			return nil
		}
		if isTopicEnsureNonFatalError(err) {
			log.Printf("kafka topic ensure skipped due to broker limitation: %v", err)
			return nil
		}
		return fmt.Errorf("create kafka topics %v: %w", topicNames, err)
	}

	log.Printf("kafka topics ensured: %v", topicNames)
	return nil
}

func uniqueTopics(values ...string) []string {
	seen := make(map[string]struct{}, len(values))
	topics := make([]string, 0, len(values))
	for _, value := range values {
		topic := strings.TrimSpace(value)
		if topic == "" {
			continue
		}
		if _, ok := seen[topic]; ok {
			continue
		}
		seen[topic] = struct{}{}
		topics = append(topics, topic)
	}
	return topics
}

func isTopicAlreadyExistsError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "already exists") || strings.Contains(message, "topic with this name already exists")
}

func isTopicEnsureNonFatalError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unsupported version") || strings.Contains(message, "eof")
}
