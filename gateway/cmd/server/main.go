package main

import (
	"fmt"
	"log/slog"
	"logtheus/gateway/internal/api"
	"logtheus/gateway/internal/config"
	"logtheus/gateway/internal/di"
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
	router := api.NewRouter(container)

	slog.Info("Server starting", "mode", cfg.Env, "port", cfg.Server.Port)
	if err := router.Run(fmt.Sprintf("localhost:%d", cfg.Server.Port)); err != nil {
		slog.Error("Failed to start server", sl.Error(err))
		os.Exit(1)
	}
}
