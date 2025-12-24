package main

import (
	"log/slog"
	"logtheus/internal/config"
	"logtheus/internal/constants"
	"os"
	"time"

	"github.com/Marlliton/slogpretty"
)

func main() {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.Env)
	slog.SetDefault(logger)

	slog.Info("Server starting", "mode", cfg.Env, "port", cfg.Server.Port)
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger
	switch env {
	case constants.DEVELOPMENT:
		logger = slog.New(slogpretty.New(os.Stdout, &slogpretty.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
			Colorful:   true,
		}))
	case constants.PRODUCTION:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
	return logger
}
