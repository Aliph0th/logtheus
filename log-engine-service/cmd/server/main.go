package main

import (
	"context"
	"log/slog"
	"logtheus/logengine/internal/services"

	"logtheus/logengine/internal/api"
	"logtheus/logengine/internal/config"
	"logtheus/logengine/internal/di"
	"logtheus/logengine/internal/models"
	"logtheus/logengine/internal/storages"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/utils"
	sl "logtheus/shared/pkg/utils/logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := utils.LoadConfig[config.AppConfig](".env")
	if err != nil {
		slog.Error("Failed to load config", sl.Error(err))
		os.Exit(1)
	}

	logger := sl.SetupLogger(cfg.Env)
	slog.SetDefault(logger)

	container := di.Build(cfg)
	clickHouse := utils.MustResolve[*storages.ClickHouse](container)
	defer clickHouse.Close()
	logsConsumer := utils.MustResolve[*services.LogsConsumer](container)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	defer func() {
		if err := logsConsumer.Close(); err != nil {
			slog.Error("Failed to close logs consumer", sl.Error(err))
		}
	}()

	if cfg.Env == consts.DEVELOPMENT {
		clickHouse.Migrate(
			&models.LogRecord{},
			"ENGINE=MergeTree() PARTITION BY toDate(received_at) ORDER BY (project_id, application_id, received_at)",
		)
	}

	logsConsumer.Start(ctx)

	if err := api.StartGRPCServer(cfg.Server.Port, container); err != nil {
		slog.Error("Failed to start gRPC server", sl.Error(err))
		os.Exit(1)
	}
}
