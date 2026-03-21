package internal

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
