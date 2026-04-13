package main

import (
	"log/slog"

	"logtheus/application/internal/api"
	"logtheus/application/internal/config"
	"logtheus/application/internal/di"
	"logtheus/application/internal/models"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/storages"
	"logtheus/shared/pkg/utils"
	sl "logtheus/shared/pkg/utils/logger"
	"os"
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
	db := utils.MustResolve[*storages.Database](container)
	defer db.Close()

	if cfg.Env == consts.DEVELOPMENT {
		if err := db.Migrate(&models.Application{}, &models.ApplicationKey{}); err != nil {
			slog.Error("Failed to migrate database", sl.Error(err))
			os.Exit(1)
		}
	}

	if err := api.StartGRPCServer(cfg.Server.Port, container); err != nil {
		slog.Error("Failed to start gRPC server", sl.Error(err))
		os.Exit(1)
	}
}
