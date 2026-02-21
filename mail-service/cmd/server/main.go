package main

import (
	"log/slog"

	"logtheus/mail/internal/api"
	"logtheus/mail/internal/config"
	"logtheus/mail/internal/di"
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

	if err := api.StartGRPCServer(cfg.Server.Port, container); err != nil {
		slog.Error("Failed to start gRPC server", sl.Error(err))
		os.Exit(1)
	}
}
