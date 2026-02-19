package grpc

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCError struct {
	Code    codes.Code
	Message string
	Slug    string
}

func NewGRPCError(code codes.Code, message string) *GRPCError {
	return &GRPCError{Code: code, Message: message}
}

func (e *GRPCError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code.String()
}

func (e *GRPCError) WithSlug(slug string) *GRPCError {
	e.Slug = slug
	return e
}

func (e *GRPCError) ToGRPCStatus() error {
	st := status.New(e.Code, e.Message)
	if e.Slug != "" {
		errdetails := &errdetails.ErrorInfo{
			Reason: e.Slug,
		}
		st, _ = st.WithDetails(errdetails)
	}
	return st.Err()
}

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

func WithInternalError(slug, msg string) *GRPCError {
	return NewGRPCError(codes.Internal, msg).WithSlug(slug)
}

func GetErrorSlug(err error) string {
	if grpcErr, ok := err.(*GRPCError); ok {
		return grpcErr.Slug
	}
	return ""
}

func FormatErrorResponse(err error) map[string]interface{} {
	if grpcErr, ok := err.(*GRPCError); ok {
		return map[string]interface{}{
			"code":    grpcErr.Code.String(),
			"message": grpcErr.Message,
			"slug":    grpcErr.Slug,
		}
	}
	return map[string]interface{}{
		"code":    "UNKNOWN",
		"message": err.Error(),
		"slug":    "",
	}
}
