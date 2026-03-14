package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	ErrValidation       = "VALIDATION_ERROR"
	ErrBadRequest       = "BAD_REQUEST"
	ErrUnauthorized     = "UNAUTHORIZED"
	ErrForbidden        = "FORBIDDEN"
	ErrNotFound         = "NOT_FOUND"
	ErrResourceConflict = "RESOURCE_CONFLICT"
	ErrUploadError      = "UPLOAD_ERROR"
	ErrInternalError    = "INTERNAL_ERROR"
)

type ApiError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	Errors     []any  `json:"errors,omitempty"`
}

func (e *ApiError) Error() string {
	return e.Message
}

func NewApiError(statusCode int, message, code string, errs []any) *ApiError {
	if message == "" {
		message = "Something went wrong"
	}
	if code == "" {
		code = ErrInternalError
	}
	return &ApiError{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
		Errors:     errs,
	}
}

type ValidationErrorDetail struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func HandleError(err error) (int, any) {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode, gin.H{
			"success": false,
			"code":    apiErr.Code,
			"message": apiErr.Message,
			"errors":  apiErr.Errors,
		}
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		var details []any
		for _, v := range validationErrs {
			details = append(details, ValidationErrorDetail{
				Path:    strings.ToLower(v.Field()),
				Message: fmt.Sprintf("failed on the '%s' tag", v.Tag()),
				Code:    v.Tag(),
			})
		}

		return 422, gin.H{
			"success": false,
			"code":    ErrValidation,
			"message": "Validation failed",
			"errors":  details,
		}
	}

	// Unhandled error fallback
	return 500, gin.H{
		"success": false,
		"code":    ErrInternalError,
		"message": err.Error(),
	}
}

type ApiResponse[T any] struct {
	StatusCode int    `json:"statusCode"`
	Data       T      `json:"data"`
	Message    string `json:"message"`
	Success    bool   `json:"success"`
	Code       string `json:"code"`
}

func NewApiResponse[T any](statusCode int, data T, message string) ApiResponse[T] {
	if message == "" {
		message = "Success"
	}
	return ApiResponse[T]{
		StatusCode: statusCode,
		Data:       data,
		Message:    message,
		Success:    statusCode < 400,
		Code:       "SUCCESS",
	}
}
