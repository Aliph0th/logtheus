package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"logtheus/mail/internal/config"
	"logtheus/mail/internal/di"
	"logtheus/mail/internal/services"
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
	consumer := utils.MustResolve[*services.MailConsumer](container)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("Mail worker starting")
	consumer.Start(ctx)
	slog.Info("Mail worker started, waiting for shutdown signal")

	<-ctx.Done()
	slog.Info("Mail worker stopped")
}
