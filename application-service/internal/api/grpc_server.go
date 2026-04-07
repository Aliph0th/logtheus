package api

import (
	"fmt"
	"log/slog"
	"net"

	"logtheus/shared/pkg/interceptors"
	applicationProto "logtheus/shared/pkg/pb/v1/application"
	"logtheus/shared/pkg/utils"

	"go.uber.org/dig"
	"google.golang.org/grpc"
)

func StartGRPCServer(port int, container *dig.Container) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	handler := utils.MustResolve[*ApplicationHandler](container)
	authInterceptor := utils.MustResolve[*interceptors.AuthInterceptor](container)
	errorInterceptor := utils.MustResolve[*interceptors.ErrorInterceptor](container)

	chainedInterceptor := interceptors.ChainUnaryInterceptor(
		authInterceptor.Unary(),
		errorInterceptor.Unary(),
	)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(chainedInterceptor),
	)
	applicationProto.RegisterApplicationServiceServer(grpcServer, handler)

	slog.Info("gRPC server starting", "port", port)
	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}
