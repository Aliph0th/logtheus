package clients

import (
	"log/slog"
	ingestionProto "logtheus/shared/pkg/pb/v1/ingestion"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewIngestionClient(address string) ingestionProto.IngestionServiceClient {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to ingestion service", "address", address, "error", err)
		panic(err)
	}
	return ingestionProto.NewIngestionServiceClient(conn)
}
