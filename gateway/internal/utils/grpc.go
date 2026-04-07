package utils

import (
	"context"
	"fmt"
	"logtheus/shared/pkg/consts"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func FromGrpcToDTO[T any](src any, dest *T) *T {
	if src == nil || dest == nil {
		return dest
	}

	srcVal := reflect.ValueOf(src)
	destVal := reflect.ValueOf(dest).Elem()

	if srcVal.Kind() == reflect.Pointer {
		srcVal = srcVal.Elem()
	}

	if srcVal.Kind() != reflect.Struct || destVal.Kind() != reflect.Struct {
		return dest
	}

	for i := 0; i < destVal.NumField(); i++ {
		destField := destVal.Field(i)
		destFieldType := destVal.Type().Field(i)

		if !destField.CanSet() {
			continue
		}

		srcField := srcVal.FieldByName(destFieldType.Name)
		if !srcField.IsValid() {
			continue
		}

		if srcField.Type().String() == "*timestamppb.Timestamp" && destField.Type().Kind() == reflect.Int64 {
			if !srcField.IsNil() {
				timestamp := srcField.Interface().(*timestamppb.Timestamp)
				destField.SetInt(timestamp.AsTime().Unix())
			}
			continue
		}

		if srcField.Type().AssignableTo(destField.Type()) {
			destField.Set(srcField)
		}
	}

	return dest
}

func GrpcCodeToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusRequestTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
