package excepts

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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
func NewWithDetails(status int, msg string, details any) *AppError {
	return &AppError{Status: status, Message: msg, Details: details}
}

func Wrap(err error, status int, code string) *AppError {
	return &AppError{Status: status, Err: err, Code: code}
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
	ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"message": "Internal Server Error",
			"status":  http.StatusInternalServerError,
		},
	})
}

func WithBadRequest(msg string) *AppError   { return New(http.StatusBadRequest, msg) }
func WithUnauthorized(msg string) *AppError { return New(http.StatusUnauthorized, msg) }
func WithForbidden(msg string) *AppError    { return New(http.StatusForbidden, msg) }
func WithNotFound(msg string) *AppError     { return New(http.StatusNotFound, msg) }
func WithConflict(msg string) *AppError     { return New(http.StatusConflict, msg) }
func WithInternal(msg string) *AppError     { return New(http.StatusInternalServerError, msg) }
