package internal

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

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
