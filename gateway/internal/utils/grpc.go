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
