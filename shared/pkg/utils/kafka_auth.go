package utils

import (
	"logtheus/shared/pkg/types"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
)

func NewKafkaDialer(auth *types.KafkaAuthOptions) *kafka.Dialer {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	dialer.SASLMechanism = newKafkaSASLMechanism(auth)

	return dialer
}

func NewKafkaTransport(auth *types.KafkaAuthOptions) *kafka.Transport {
	return &kafka.Transport{
		SASL: newKafkaSASLMechanism(auth),
	}
}

func newKafkaSASLMechanism(auth *types.KafkaAuthOptions) sasl.Mechanism {
	if auth == nil {
		return nil
	}

	if strings.TrimSpace(auth.Username) == "" || strings.TrimSpace(auth.Password) == "" {
		return nil
	}

	mechanism := strings.ToUpper(strings.TrimSpace(auth.Mechanism))
	if mechanism == "" {
		mechanism = "PLAIN"
	}

	if mechanism == "PLAIN" {
		return plain.Mechanism{
			Username: auth.Username,
			Password: auth.Password,
		}
	}

	return nil
}
