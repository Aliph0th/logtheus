package interceptors

import (
	"context"

	grpcLib "google.golang.org/grpc"
)

func ChainUnaryInterceptor(interceptors ...grpcLib.UnaryServerInterceptor) grpcLib.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpcLib.UnaryServerInfo,
		handler grpcLib.UnaryHandler,
	) (any, error) {
		wrapped := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			wrapped = wrapHandler(interceptors[i], wrapped, info)
		}
		return wrapped(ctx, req)
	}
}

func wrapHandler(
	interceptor grpcLib.UnaryServerInterceptor,
	handler grpcLib.UnaryHandler,
	info *grpcLib.UnaryServerInfo,
) grpcLib.UnaryHandler {
	return func(ctx context.Context, req any) (any, error) {
		return interceptor(ctx, req, info, handler)
	}
}
