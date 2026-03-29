package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"logtheus/mail/internal/config"
	sharedConsts "logtheus/shared/pkg/consts"
	sharedTypes "logtheus/shared/pkg/types"
	"logtheus/shared/pkg/utils"
	sl "logtheus/shared/pkg/utils/logger"
	"time"

	"github.com/segmentio/kafka-go"
)

type MailConsumer struct {
	mailService  *MailService
	verifyReader *kafka.Reader
	inviteReader *kafka.Reader
}

func NewMailConsumer(cfg *config.AppConfig, mailService *MailService) *MailConsumer {
	brokers := utils.SplitBrokers(cfg.Kafka.Brokers)
	dialer := utils.NewKafkaDialer(&sharedTypes.KafkaAuthOptions{
		Username:  cfg.Kafka.Username,
		Password:  cfg.Kafka.Password,
		Mechanism: cfg.Kafka.Mechanism,
	})

	return &MailConsumer{
		mailService:  mailService,
		verifyReader: newReader(brokers, sharedConsts.KAFKA_TOPIC_MAIL_VERIFY, dialer),
		inviteReader: newReader(brokers, sharedConsts.KAFKA_TOPIC_MAIL_INVITE, dialer),
	}
}

func (c *MailConsumer) Start(ctx context.Context) {
	slog.Info("Starting mail Kafka consumers")
	go c.consumeVerify(ctx)
	go c.consumeInvite(ctx)
}

func (c *MailConsumer) consumeVerify(ctx context.Context) {
	for {
		message, err := c.verifyReader.ReadMessage(ctx)
		if err != nil {
			slog.Error("Failed to read verify email event", sl.Error(err))
			if ctx.Err() != nil {
				return
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var event sharedTypes.VerifyEmailEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			slog.Error("Failed to decode verify email event", sl.Error(err))
			continue
		}

		if err := c.mailService.SendVerifyEmail(event.Email, event.Username, event.Code, event.ExpirationMinutes); err != nil {
			slog.Error("Failed to send verify email", "email", event.Email, sl.Error(err))
		}
	}
}

func (c *MailConsumer) consumeInvite(ctx context.Context) {
	for {
		message, err := c.inviteReader.ReadMessage(ctx)
		if err != nil {
			slog.Error("Failed to read invite email event", sl.Error(err))
			if ctx.Err() != nil {
				return
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var event sharedTypes.InviteEmailEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			slog.Error("Failed to decode invite email event", sl.Error(err))
			continue
		}

		minutes := int(time.Until(event.ExpiresAt).Minutes())
		if minutes < 0 {
			minutes = 0
		}
		expiresMinutes := uint16(minutes)
		if err := c.mailService.SendInviteEmail(event.Email, event.InviteeName, event.Referrer, event.ProjectName, event.Code, expiresMinutes); err != nil {
			slog.Error("Failed to send invite email", "email", event.Email, sl.Error(err))
		}
	}
}

func newReader(brokers []string, topic string, dialer *kafka.Dialer) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  sharedConsts.KAFKA_GROUP_MAIL_SERVICE,
		MinBytes: 1,
		MaxBytes: 10e6,
		Dialer:   dialer,
	})
}
