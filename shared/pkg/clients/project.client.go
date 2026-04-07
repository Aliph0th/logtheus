package clients

import (
	"log/slog"
	projectProto "logtheus/shared/pkg/pb/v1/project"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewProjectClient(address string) projectProto.ProjectServiceClient {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to project service", "address", address, "error", err)
		panic(err)
	}
	return projectProto.NewProjectServiceClient(conn)
}
