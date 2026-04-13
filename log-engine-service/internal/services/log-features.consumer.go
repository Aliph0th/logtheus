package services

import (
	"context"
	"log/slog"
	"time"

	"logtheus/logengine/internal/config"
	"logtheus/shared/pkg/consts"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	"logtheus/shared/pkg/types"
	"logtheus/shared/pkg/utils"
	sl "logtheus/shared/pkg/utils/logger"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type LogFeaturesConsumer struct {
	service *LogFeatureService
	reader  *kafka.Reader
}

func NewLogFeaturesConsumer(cfg *config.AppConfig, service *LogFeatureService) *LogFeaturesConsumer {
	brokers := utils.SplitBrokers(cfg.Kafka.Brokers)
	dialer := utils.NewKafkaDialer(&types.KafkaAuthOptions{
		Username:  cfg.Kafka.Username,
		Password:  cfg.Kafka.Password,
		Mechanism: cfg.Kafka.Mechanism,
	})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    consts.KAFKA_TOPIC_LOGS_FEATURE,
		GroupID:  consts.KAFKA_GROUP_LOG_FEATURES,
		MinBytes: 1,
		MaxBytes: 10e6,
		Dialer:   dialer,
	})

	return &LogFeaturesConsumer{service: service, reader: reader}
}

func (c *LogFeaturesConsumer) Start(ctx context.Context) {
	slog.Info("Starting log features Kafka consumer")
	go c.consume(ctx)
}

func (c *LogFeaturesConsumer) consume(ctx context.Context) {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("Failed to read log features event", sl.Error(err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var req logEngineProto.SaveLogFeaturesRequest
		if err := proto.Unmarshal(message.Value, &req); err != nil {
			slog.Error("Failed to decode log features event", sl.Error(err))
			if commitErr := c.reader.CommitMessages(ctx, message); commitErr != nil {
				slog.Error("Failed to commit malformed log features event", sl.Error(commitErr))
			}
			continue
		}

		if err := c.service.SaveFeatures(ctx, &req); err != nil {
			slog.Error("Failed to persist log features batch", sl.Error(err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			slog.Error("Failed to commit log features event", sl.Error(err))
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (c *LogFeaturesConsumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}
