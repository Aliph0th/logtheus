package clients

import (
	"log/slog"
	mailProto "logtheus/shared/pkg/pb/v1/mail"
	"logtheus/user/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewMailClient(cfg *config.AppConfig) mailProto.MailServiceClient {
	conn, err := grpc.NewClient(cfg.Services.Mail, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to mail service", "address", cfg.Services.Mail, "error", err)
		panic(err)
	}
	return mailProto.NewMailServiceClient(conn)
}
