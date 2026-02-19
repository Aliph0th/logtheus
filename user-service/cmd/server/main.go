package main

import (
	"log/slog"

	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/utils"
	sl "logtheus/shared/pkg/utils/logger"
	"logtheus/user/internal/api"
	"logtheus/user/internal/config"
	"logtheus/user/internal/di"
	"logtheus/user/internal/models"
	"logtheus/user/internal/storages"
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

	db, err := storages.NewPostgres(cfg)
	if err != nil {
		slog.Error("Failed to setup database", sl.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	if cfg.Env == consts.DEVELOPMENT {
		db.Migrate(&models.User{})
	}

	container := di.Build(cfg, db.DB)

	if err := api.StartGRPCServer(cfg.Server.Port, container); err != nil {
		slog.Error("Failed to start gRPC server", sl.Error(err))
		os.Exit(1)
	}
}
