package main

import (
	"fmt"
	"log/slog"
	"logtheus/internal/api"
	"logtheus/internal/config"
	"logtheus/internal/consts"
	"logtheus/internal/models"
	"logtheus/internal/storage"
	sl "logtheus/internal/utils/logger"
	"os"
)

func main() {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		slog.Error("Failed to load config", sl.Error(err))
		os.Exit(1)
	}

	logger := sl.SetupLogger(cfg.Env)
	slog.SetDefault(logger)

	db, err := storage.NewPostgres(cfg)
	if err != nil {
		slog.Error("Failed to setup database", sl.Error(err))
		os.Exit(1)
	}

	if cfg.Env == consts.DEVELOPMENT {
		db.Migrate(&models.User{}, &models.Token{})
	}

	defer db.Close()

	router := api.NewRouter(db.DB)

	slog.Info("Server starting", "mode", cfg.Env, "port", cfg.Server.Port)
	if err := router.Run(fmt.Sprintf("localhost:%d", cfg.Server.Port)); err != nil {
		slog.Error("Failed to start server", sl.Error(err))
		os.Exit(1)
	}
}
