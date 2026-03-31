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

type LogsConsumer struct {
	service *LogEngineService
	reader  *kafka.Reader
}

func NewLogsConsumer(cfg *config.AppConfig, service *LogEngineService) *LogsConsumer {
	brokers := utils.SplitBrokers(cfg.Kafka.Brokers)
	dialer := utils.NewKafkaDialer(&types.KafkaAuthOptions{
		Username:  cfg.Kafka.Username,
		Password:  cfg.Kafka.Password,
		Mechanism: cfg.Kafka.Mechanism,
	})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    consts.KAFKA_TOPIC_LOGS_INGEST,
		GroupID:  consts.KAFKA_GROUP_LOG_ENGINE,
		MinBytes: 1,
		MaxBytes: 10e6,
		Dialer:   dialer,
	})

	return &LogsConsumer{service: service, reader: reader}
}

func (c *LogsConsumer) Start(ctx context.Context) {
	slog.Info("Starting logs Kafka consumer")
	go c.consume(ctx)
}

func (c *LogsConsumer) consume(ctx context.Context) {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("Failed to read logs batch event", sl.Error(err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var req logEngineProto.SaveLogsRequest
		if err := proto.Unmarshal(message.Value, &req); err != nil {
			slog.Error("Failed to decode logs batch event", sl.Error(err))
			if commitErr := c.reader.CommitMessages(ctx, message); commitErr != nil {
				slog.Error("Failed to commit malformed logs batch event", sl.Error(commitErr))
			}
			continue
		}

		if err := c.service.SaveLogs(ctx, &req); err != nil {
			slog.Error("Failed to persist logs batch", sl.Error(err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			slog.Error("Failed to commit logs batch event", sl.Error(err))
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (c *LogsConsumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}
