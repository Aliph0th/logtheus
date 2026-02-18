package main

import (
	"log/slog"

	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/utils"
	sl "logtheus/shared/pkg/utils/logger"
	"logtheus/user/internal/api"
	"logtheus/user/internal/config"
	"logtheus/user/internal/models"
	"logtheus/user/internal/repository"
	service "logtheus/user/internal/services"
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

	userRepository := repository.NewUserRepository(db.DB)
	userService := service.NewUserService(userRepository, cfg)
	handler := api.NewUserHandler(userService)

	if err := api.StartGRPCServer(cfg.Server.Port, handler); err != nil {
		slog.Error("Failed to start gRPC server", sl.Error(err))
		os.Exit(1)
	}
}
