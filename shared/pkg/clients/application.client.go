package clients

import (
	"log/slog"
	applicationProto "logtheus/shared/pkg/pb/v1/application"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewApplicationClient(address string) applicationProto.ApplicationServiceClient {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to application service", "address", address, "error", err)
		panic(err)
	}
	return applicationProto.NewApplicationServiceClient(conn)
}
