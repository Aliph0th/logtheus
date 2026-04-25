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

	batchMaxMessages int
	batchMaxBytes    int
	batchMaxWait     time.Duration
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

	batchMaxMessages := cfg.Kafka.FeaturesConsumerBatchMaxMessages
	batchMaxBytes := cfg.Kafka.FeaturesConsumerBatchMaxBytes
	batchMaxWait := time.Duration(cfg.Kafka.FeaturesConsumerBatchMaxWaitMs) * time.Millisecond

	return &LogFeaturesConsumer{
		service:          service,
		reader:           reader,
		batchMaxMessages: batchMaxMessages,
		batchMaxBytes:    batchMaxBytes,
		batchMaxWait:     batchMaxWait,
	}
}

func (c *LogFeaturesConsumer) Start(ctx context.Context) {
	slog.Info("Starting log features Kafka consumer")
	go c.consume(ctx)
}

func (c *LogFeaturesConsumer) consume(ctx context.Context) {
	ticker := time.NewTicker(c.batchMaxWait)
	defer ticker.Stop()

	messagesBatch := make([]kafka.Message, 0, c.batchMaxMessages)
	requestsBatch := make([]logEngineProto.SaveLogFeaturesRequest, 0, c.batchMaxMessages)
	batchBytes := 0

	flush := func() bool {
		if len(messagesBatch) == 0 {
			return true
		}

		for index := range requestsBatch {
			if err := c.service.SaveFeatures(ctx, &requestsBatch[index]); err != nil {
				slog.Error("Failed to persist log features batch", sl.Error(err))
				time.Sleep(500 * time.Millisecond)
				return false
			}
		}

		if err := c.reader.CommitMessages(ctx, messagesBatch...); err != nil {
			slog.Error("Failed to commit log features batch events", sl.Error(err))
			time.Sleep(500 * time.Millisecond)
			return false
		}

		messagesBatch = messagesBatch[:0]
		requestsBatch = requestsBatch[:0]
		batchBytes = 0
		return true
	}

	for {
		select {
		case <-ctx.Done():
			_ = flush()
			return
		case <-ticker.C:
			_ = flush()
		default:
		}

		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				_ = flush()
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

		messagesBatch = append(messagesBatch, message)
		requestsBatch = append(requestsBatch, req)
		batchBytes += len(message.Value)

		if len(messagesBatch) >= c.batchMaxMessages || batchBytes >= c.batchMaxBytes {
			_ = flush()
		}
	}
}

func (c *LogFeaturesConsumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}
