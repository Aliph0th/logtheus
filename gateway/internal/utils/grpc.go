package utils

import (
	"context"
	"fmt"
	"logtheus/shared/pkg/consts"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func GetGRPCContextWithAuth(ginCtx *gin.Context) context.Context {
	grpcCtx := context.Background()
	payload := GetAuthPayload(ginCtx)
	if payload == nil {
		return grpcCtx
	}
	grpcCtx = metadata.AppendToOutgoingContext(
		grpcCtx,
		consts.X_USER_ID_METADATA_KEY, fmt.Sprintf("%d", payload.UserID),
		consts.X_EMAIL_VERIFIED_METADATA_KEY, fmt.Sprintf("%v", payload.IsEmailVerified),
	)

	return grpcCtx
}

func GrpcCodeToHTTPStatus(code codes.Code) int {
	switch code.String() {
	case "Ok":
		return http.StatusOK
	case "Canceled":
		return http.StatusRequestTimeout
	case "Unknown":
		return http.StatusInternalServerError
	case "InvalidArgument":
		return http.StatusBadRequest
	case "DeadlineExceeded":
		return http.StatusRequestTimeout
	case "NotFound":
		return http.StatusNotFound
	case "AlreadyExists":
		return http.StatusConflict
	case "PermissionDenied":
		return http.StatusForbidden
	case "ResourceExhausted":
		return http.StatusTooManyRequests
	case "FailedPrecondition":
		return http.StatusBadRequest
	case "Aborted":
		return http.StatusConflict
	case "OutOfRange":
		return http.StatusBadRequest
	case "Unimplemented":
		return http.StatusNotImplemented
	case "Internal":
		return http.StatusInternalServerError
	case "Unavailable":
		return http.StatusServiceUnavailable
	case "DataLoss":
		return http.StatusInternalServerError
	case "Unauthenticated":
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
