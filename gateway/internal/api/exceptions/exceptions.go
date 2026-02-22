package excepts

import (
	"errors"
	"logtheus/gateway/internal/utils"
	"logtheus/shared/pkg/consts"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type AppError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Err     error  `json:"-"`
	Details any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return http.StatusText(e.Status)
}

func New(status int, msg string) *AppError {
	return &AppError{Status: status, Message: msg}
}

func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}
func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

func RespondError(ctx *gin.Context, err error) {
	var appError *AppError
	if errors.As(err, &appError) && appError != nil {
		status := appError.Status
		if status == 0 {
			status = http.StatusInternalServerError
		}
		ctx.AbortWithStatusJSON(status, gin.H{
			"error": appError,
		})
		return
	}

	// Handle gRPC errors
	st, ok := status.FromError(err)
	if ok {
		appErr := &AppError{
			Status:  utils.GrpcCodeToHTTPStatus(st.Code()),
			Message: st.Message(),
		}

		detail := st.Details()[0]
		if errInfo, ok := detail.(*errdetails.ErrorInfo); ok {
			appErr.Code = errInfo.Reason
		}

		ctx.AbortWithStatusJSON(appErr.Status, gin.H{
			"error": appErr,
		})
		return
	}

	ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"message": "Internal Server Error",
			"status":  http.StatusInternalServerError,
			"code":    consts.ERROR_CODE_INTERNAL,
		},
	})
}

func WithBadRequest(msg string) *AppError   { return New(http.StatusBadRequest, msg) }
func WithUnauthorized(msg string) *AppError { return New(http.StatusUnauthorized, msg) }
func WithForbidden(msg string) *AppError    { return New(http.StatusForbidden, msg) }
func WithNotFound(msg string) *AppError     { return New(http.StatusNotFound, msg) }
func WithConflict(msg string) *AppError     { return New(http.StatusConflict, msg) }
func WithInternal(msg string) *AppError     { return New(http.StatusInternalServerError, msg) }
