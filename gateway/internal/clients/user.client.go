package clients

import (
	"log/slog"
	"logtheus/gateway/internal/config"
	userProto "logtheus/shared/pkg/pb/v1/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewUserClient(cfg *config.AppConfig) userProto.UserServiceClient {
	conn, err := grpc.NewClient(cfg.Services.User, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to user service", "address", cfg.Services.User, "error", err)
		panic(err)
	}
	return userProto.NewUserServiceClient(conn)
}
