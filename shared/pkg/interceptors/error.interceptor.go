package interceptors

import (
	"context"
	"errors"
	"log/slog"

	"logtheus/shared/pkg/grpc"
	sl "logtheus/shared/pkg/utils/logger"

	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorInterceptor struct{}

func NewErrorInterceptor() *ErrorInterceptor {
	return &ErrorInterceptor{}
}

func (e *ErrorInterceptor) Unary() grpcLib.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpcLib.UnaryServerInfo,
		handler grpcLib.UnaryHandler,
	) (any, error) {
		response, err := handler(ctx, req)

		if err == nil {
			return response, nil
		}

		var grpcErr *grpc.GRPCError
		if errors.As(err, &grpcErr) {
			if grpcErr.Code == codes.Internal {
				slog.Error(
					"gRPC error occurred",
					"method", info.FullMethod,
					"code", grpcErr.Code.String(),
					"message", grpcErr.Message,
					"slug", grpcErr.Slug,
					sl.Error(err),
				)
			}
			return nil, grpcErr.ToGRPCStatus()
		}

		if st, ok := status.FromError(err); ok {
			slog.Warn(
				"gRPC status error occurred",
				"method", info.FullMethod,
				"code", st.Code().String(),
				"message", st.Message(),
				sl.Error(err),
			)
			return nil, err
		}

		slog.Error(
			"unexpected error in gRPC handler",
			"method", info.FullMethod,
			sl.Error(err),
		)
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}
}
