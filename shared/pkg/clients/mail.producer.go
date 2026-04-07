package clients

import (
	"context"
	"encoding/json"
	"log/slog"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/types"
	"logtheus/shared/pkg/utils"
	sl "logtheus/shared/pkg/utils/logger"
	"time"

	"github.com/segmentio/kafka-go"
)

type MailEventProducer struct {
	verifyWriter *kafka.Writer
	inviteWriter *kafka.Writer
}

func NewMailEventProducer(brokers string, auth *types.KafkaAuthOptions) *MailEventProducer {
	brokerList := utils.SplitBrokers(brokers)
	transport := utils.NewKafkaTransport(auth)
	return &MailEventProducer{
		verifyWriter: newKafkaWriter(brokerList, consts.KAFKA_TOPIC_MAIL_VERIFY, transport),
		inviteWriter: newKafkaWriter(brokerList, consts.KAFKA_TOPIC_MAIL_INVITE, transport),
	}
}

func (p *MailEventProducer) PublishVerifyEmail(ctx context.Context, event *types.VerifyEmailEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.verifyWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.Email),
		Value: data,
	})
}

func (p *MailEventProducer) PublishInviteEmail(ctx context.Context, event *types.InviteEmailEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.inviteWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.Email),
		Value: data,
	})
}

func (p *MailEventProducer) Close() error {
	if err := p.verifyWriter.Close(); err != nil {
		return err
	}
	if err := p.inviteWriter.Close(); err != nil {
		return err
	}
	return nil
}

func newKafkaWriter(brokers []string, topic string, transport *kafka.Transport) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 20 * time.Millisecond,
		Transport:    transport,
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				slog.Error("[KAFKA_PRODUCER] delivery failed", "topic", topic, sl.Error(err))
			}
		},
	}
}
