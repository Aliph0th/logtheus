package clients

import (
	"log/slog"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewLogEngineClient(address string) logEngineProto.LogEngineServiceClient {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to log-engine service", "address", address, "error", err)
		panic(err)
	}
	return logEngineProto.NewLogEngineServiceClient(conn)
}
