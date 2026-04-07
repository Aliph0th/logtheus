package grpc

import (
	"logtheus/shared/pkg/consts"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
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

func (e *GRPCError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	if e.Message != "" {
		return e.Message
	}
	return e.Code.String()
}

func (e *GRPCError) WithSlug(slug string) *GRPCError {
	e.Slug = slug
	return e
}

func (e *GRPCError) WithError(err error) *GRPCError {
	e.Err = err
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
	return NewGRPCError(codes.NotFound, msg).WithSlug(consts.ERROR_CODE_NOT_FOUND)
}

func WithAlreadyExists(msg string) *GRPCError {
	return NewGRPCError(codes.AlreadyExists, msg)
}

func WithResourceExhausted(msg string) *GRPCError {
	return NewGRPCError(codes.ResourceExhausted, msg)
}

func WithPermissionDenied(msg string) *GRPCError {
	return NewGRPCError(codes.PermissionDenied, msg)
}

func WithUnauthenticated(msg string) *GRPCError {
	return NewGRPCError(codes.Unauthenticated, msg)
}

func WithInternal(err error) *GRPCError {
	return NewGRPCError(codes.Internal, "Internal Server Error").WithError(err)
}
