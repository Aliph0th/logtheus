package services

import (
	"context"
	"log/slog"
	"time"

	"logtheus/logengine/internal/config"
	internalUtils "logtheus/logengine/internal/utils"
	"logtheus/shared/pkg/consts"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	"logtheus/shared/pkg/types"
	sharedUtils "logtheus/shared/pkg/utils"
	sl "logtheus/shared/pkg/utils/logger"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type LogsConsumer struct {
	service *LogEngineService
	reader  *kafka.Reader

	batchMaxMessages int
	batchMaxBytes    int
	batchMaxWait     time.Duration

	dedupe *internalUtils.MessageDedupeCache
}

func NewLogsConsumer(cfg *config.AppConfig, service *LogEngineService) *LogsConsumer {
	brokers := sharedUtils.SplitBrokers(cfg.Kafka.Brokers)
	dialer := sharedUtils.NewKafkaDialer(&types.KafkaAuthOptions{
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

	batchMaxMessages := cfg.Kafka.LogsConsumerBatchMaxMessages
	batchMaxBytes := cfg.Kafka.LogsConsumerBatchMaxBytes
	batchMaxWait := time.Duration(cfg.Kafka.LogsConsumerBatchMaxWaitMs) * time.Millisecond
	dedupeTTL := time.Duration(cfg.Persistence.LogsDedupeTTLSeconds) * time.Second
	dedupeSize := cfg.Persistence.LogsDedupeCacheSize

	return &LogsConsumer{
		service:          service,
		reader:           reader,
		batchMaxMessages: batchMaxMessages,
		batchMaxBytes:    batchMaxBytes,
		batchMaxWait:     batchMaxWait,
		dedupe:           internalUtils.NewMessageDedupeCache(dedupeTTL, dedupeSize),
	}
}

func (c *LogsConsumer) Start(ctx context.Context) {
	slog.Info("Starting logs Kafka consumer")
	go c.consume(ctx)
}

func (c *LogsConsumer) consume(ctx context.Context) {
	ticker := time.NewTicker(c.batchMaxWait)
	defer ticker.Stop()

	messagesBatch := make([]kafka.Message, 0, c.batchMaxMessages)
	requestsBatch := make([]logEngineProto.SaveLogsRequest, 0, c.batchMaxMessages)
	hashesBatch := make([]string, 0, c.batchMaxMessages)
	batchBytes := 0

	flush := func() bool {
		if len(messagesBatch) == 0 {
			return true
		}

		for index := range requestsBatch {
			if err := c.service.SaveLogs(ctx, &requestsBatch[index]); err != nil {
				slog.Error("Failed to persist logs batch", sl.Error(err))
				time.Sleep(500 * time.Millisecond)
				return false
			}
		}

		if err := c.reader.CommitMessages(ctx, messagesBatch...); err != nil {
			slog.Error("Failed to commit logs batch events", sl.Error(err))
			time.Sleep(500 * time.Millisecond)
			return false
		}

		for _, hash := range hashesBatch {
			c.dedupe.MarkSeen(hash)
		}

		messagesBatch = messagesBatch[:0]
		requestsBatch = requestsBatch[:0]
		hashesBatch = hashesBatch[:0]
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
			slog.Error("Failed to read logs batch event", sl.Error(err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		hash := internalUtils.HashMessageValue(message.Value)
		if c.dedupe.Seen(hash) {
			if commitErr := c.reader.CommitMessages(ctx, message); commitErr != nil {
				slog.Error("Failed to commit deduped logs batch event", sl.Error(commitErr))
			}
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

		messagesBatch = append(messagesBatch, message)
		requestsBatch = append(requestsBatch, req)
		hashesBatch = append(hashesBatch, hash)
		batchBytes += len(message.Value)

		if len(messagesBatch) >= c.batchMaxMessages || batchBytes >= c.batchMaxBytes {
			_ = flush()
		}
	}
}

func (c *LogsConsumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}
