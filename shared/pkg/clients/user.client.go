package clients

import (
	"log/slog"
	userProto "logtheus/shared/pkg/pb/v1/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewUserClient(address string) userProto.UserServiceClient {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to user service", "address", address, "error", err)
		panic(err)
	}
	return userProto.NewUserServiceClient(conn)
}
