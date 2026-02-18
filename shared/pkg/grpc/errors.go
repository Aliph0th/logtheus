package grpc

import (
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCError struct {
	Code    codes.Code
	Message string
	Slug    string
	Err     error
}

func NewGRPCError(code codes.Code, message string) *GRPCError {
	return &GRPCError{Code: code, Message: message}
}

func NewGRPCErrorWithMessage(code codes.Code, message string, err error) *GRPCError {
	return &GRPCError{Code: code, Message: message, Err: err}
}

func (e *GRPCError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code.String()
}

func (e *GRPCError) ToGRPCStatus() error {
	st := status.New(e.Code, e.Message)
	// if e.Slug != "" {
	// 	newSt, err := st.WithDetails(protoadapt.MessageV1Of(e.Slug))
	// 	if err != nil {
	// 		slog.Warn("failed to add details to gRPC status", "error", err)
	// 	} else {
	// 		st = newSt
	// 	}
	// }

	return st.Err()
}

// func (e *GRPCError) WithSlug(slug string) *GRPCError {
// 	e.Slug = slug
// 	return e
// }

func WithInvalidArgument(msg string) *GRPCError {
	return NewGRPCError(codes.InvalidArgument, msg)
}

func WithNotFound(msg string) *GRPCError {
	return NewGRPCError(codes.NotFound, msg)
}

func WithAlreadyExists(msg string) *GRPCError {
	return NewGRPCError(codes.AlreadyExists, msg)
}

func WithPermissionDenied(msg string) *GRPCError {
	return NewGRPCError(codes.PermissionDenied, msg)
}

func WithUnauthenticated(msg string) *GRPCError {
	return NewGRPCError(codes.Unauthenticated, msg)
}

func WithInternal(msg string) *GRPCError {
	return NewGRPCError(codes.Internal, msg)
}

func HandleError(err error, op string) error {
	if err == nil {
		return nil
	}

	var grpcErr *GRPCError
	if errors.As(err, &grpcErr) {
		if grpcErr != nil {
			slog.Warn("gRPC operation failed", "operation", op, "code", grpcErr.Code.String(), "message", grpcErr.Message)
			return grpcErr.ToGRPCStatus()
		}
	}

	slog.Error("Unexpected error", "operation", op, "error", err)
	return status.Error(codes.Internal, "Internal Server Error")
}
