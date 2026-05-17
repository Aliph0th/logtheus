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
	sharedStorages "logtheus/shared/pkg/storages"
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
	featuresDB := utils.MustResolve[*sharedStorages.Database](container)
	defer clickHouse.Close()
	defer featuresDB.Close()

	logsConsumer := utils.MustResolve[*services.LogsConsumer](container)
	logFeaturesConsumer := utils.MustResolve[*services.LogFeaturesConsumer](container)
	clusteringCleanupService := utils.MustResolve[*services.ClusteringCleanupService](container)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	defer func() {
		if err := logsConsumer.Close(); err != nil {
			slog.Error("Failed to close logs consumer", sl.Error(err))
		}
		if err := logFeaturesConsumer.Close(); err != nil {
			slog.Error("Failed to close log features consumer", sl.Error(err))
		}
	}()

	if cfg.Env == consts.DEVELOPMENT {
		if err := clickHouse.Migrate(
			&models.LogRecord{},
			"ENGINE=MergeTree() PARTITION BY toDate(received_at) ORDER BY (project_id, application_id, received_at, log_id)",
		); err != nil {
			slog.Error("Failed to migrate ClickHouse", sl.Error(err))
			os.Exit(1)
		}
		featuresDB.DB.Config.DisableForeignKeyConstraintWhenMigrating = true
		if err := featuresDB.Migrate(
			&models.LogFeature{},
			&models.ClusteringJob{},
			&models.ClusteringAssignment{},
			&models.ClusteringClusterSummary{},
		); err != nil {
			slog.Error("Failed to migrate Postgres", sl.Error(err))
			os.Exit(1)
		}
		featuresDB.DB.Config.DisableForeignKeyConstraintWhenMigrating = false
		if !featuresDB.DB.Migrator().HasConstraint(&models.ClusteringJob{}, "Assignments") {
			if err := featuresDB.DB.Migrator().CreateConstraint(&models.ClusteringJob{}, "Assignments"); err != nil {
				slog.Error("Failed to create clustering assignment constraint", sl.Error(err))
				os.Exit(1)
			}
		}
		if !featuresDB.DB.Migrator().HasConstraint(&models.ClusteringJob{}, "Summaries") {
			if err := featuresDB.DB.Migrator().CreateConstraint(&models.ClusteringJob{}, "Summaries"); err != nil {
				slog.Error("Failed to create clustering summary constraint", sl.Error(err))
				os.Exit(1)
			}
		}
	}

	logsConsumer.Start(ctx)
	logFeaturesConsumer.Start(ctx)
	clusteringCleanupService.Start(ctx)

	if err := api.StartGRPCServer(cfg.Server.Port, container); err != nil {
		slog.Error("Failed to start gRPC server", sl.Error(err))
		os.Exit(1)
	}
}
