package clients

import (
	"log/slog"
	mailProto "logtheus/shared/pkg/pb/v1/mail"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewMailClient(address string) mailProto.MailServiceClient {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to mail service", "address", address, "error", err)
		panic(err)
	}
	return mailProto.NewMailServiceClient(conn)
}
