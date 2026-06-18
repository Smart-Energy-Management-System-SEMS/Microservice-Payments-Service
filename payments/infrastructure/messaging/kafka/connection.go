package kafkaadapter

import (
	"crypto/tls"
	"log"
	"strings"
	"time"

	segmentio "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type ConnectionConfig struct {
	Brokers          []string
	SecurityProtocol string
	SASLMechanism    string
	Username         string
	Password         string
	ClientID         string
	ConsumerGroup    string
}

func (c ConnectionConfig) dialer() *segmentio.Dialer {
	dialer := &segmentio.Dialer{
		ClientID:  c.ClientID,
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	if usesTLS(c.SecurityProtocol) {
		dialer.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	if mechanism := c.saslMechanism(); mechanism != nil {
		dialer.SASLMechanism = mechanism
	}
	return dialer
}

func (c ConnectionConfig) transport() *segmentio.Transport {
	dialer := c.dialer()
	transport := &segmentio.Transport{
		ClientID: c.ClientID,
	}
	transport.TLS = dialer.TLS
	transport.SASL = dialer.SASLMechanism
	return transport
}

func (c ConnectionConfig) saslMechanism() sasl.Mechanism {
	username := strings.TrimSpace(c.Username)
	password := strings.TrimSpace(c.Password)
	if username == "" || password == "" {
		return nil
	}

	switch strings.ToUpper(strings.TrimSpace(c.SASLMechanism)) {
	case "", "PLAIN":
		return plain.Mechanism{
			Username: username,
			Password: password,
		}
	case "SCRAM-SHA-256":
		mechanism, err := scram.Mechanism(scram.SHA256, username, password)
		if err != nil {
			log.Printf("kafka SCRAM-SHA-256 configuration ignored: %v", err)
			return nil
		}
		return mechanism
	case "SCRAM-SHA-512":
		mechanism, err := scram.Mechanism(scram.SHA512, username, password)
		if err != nil {
			log.Printf("kafka SCRAM-SHA-512 configuration ignored: %v", err)
			return nil
		}
		return mechanism
	default:
		log.Printf("kafka SASL mechanism %q is not supported; continuing without SASL", c.SASLMechanism)
		return nil
	}
}

func usesTLS(protocol string) bool {
	switch strings.ToUpper(strings.TrimSpace(protocol)) {
	case "SSL", "SASL_SSL":
		return true
	default:
		return false
	}
}
