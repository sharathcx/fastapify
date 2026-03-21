package internal

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

func NotFound(message string) *ApiError {
	return NewApiError(404, message, ErrNotFound, nil)
}

func BadRequest(message string) *ApiError {
	return NewApiError(400, message, ErrBadRequest, nil)
}

func Unauthorized(message string) *ApiError {
	return NewApiError(401, message, ErrUnauthorized, nil)
}

func Forbidden(message string) *ApiError {
	return NewApiError(403, message, ErrForbidden, nil)
}

func Conflict(message string) *ApiError {
	return NewApiError(409, message, ErrResourceConflict, nil)
}

func InternalError(message string) *ApiError {
	return NewApiError(500, message, ErrInternalError, nil)
}
