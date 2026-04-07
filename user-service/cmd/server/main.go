package main

import (
	"log/slog"

	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/storages"
	"logtheus/shared/pkg/utils"
	sl "logtheus/shared/pkg/utils/logger"
	"logtheus/user/internal/api"
	"logtheus/user/internal/config"
	"logtheus/user/internal/di"
	"logtheus/user/internal/models"
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
	redis := utils.MustResolve[*storages.RedisDatabase](container)
	defer redis.Close()

	if cfg.Env == consts.DEVELOPMENT {
		db.Migrate(&models.User{})
	}

	if err := api.StartGRPCServer(cfg.Server.Port, container); err != nil {
		slog.Error("Failed to start gRPC server", sl.Error(err))
		os.Exit(1)
	}
}
