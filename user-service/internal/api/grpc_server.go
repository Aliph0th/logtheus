package api

import (
	"fmt"
	"log/slog"
	"net"

	userProto "logtheus/shared/pkg/pb/v1/user"

	"google.golang.org/grpc"
)

func StartGRPCServer(port int, handler *UserHandler) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	grpcServer := grpc.NewServer()
	userProto.RegisterUserServiceServer(grpcServer, handler)

	slog.Info("[USER_SERVICE] gRPC server starting", "port", port)
	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}
