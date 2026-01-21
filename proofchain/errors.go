package proofchain

import (
	"fmt"
)

// APIError is the base error type for ProofChain API errors.
type APIError struct {
	Message      string                 `json:"message"`
	StatusCode   int                    `json:"status_code,omitempty"`
	ResponseBody map[string]interface{} `json:"response_body,omitempty"`
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
	}
	return e.Message
}

// AuthenticationError is returned when authentication fails (401).
type AuthenticationError struct {
	APIError
}

// NewAuthenticationError creates a new AuthenticationError.
func NewAuthenticationError(message string) *AuthenticationError {
	if message == "" {
		message = "Invalid or missing API key"
	}
	return &AuthenticationError{
		APIError: APIError{Message: message, StatusCode: 401},
	}
}

// AuthorizationError is returned when access is forbidden (403).
type AuthorizationError struct {
	APIError
}

// NewAuthorizationError creates a new AuthorizationError.
func NewAuthorizationError(message string) *AuthorizationError {
	if message == "" {
		message = "Access denied"
	}
	return &AuthorizationError{
		APIError: APIError{Message: message, StatusCode: 403},
	}
}

// NotFoundError is returned when a resource is not found (404).
type NotFoundError struct {
	APIError
}

// NewNotFoundError creates a new NotFoundError.
func NewNotFoundError(message string) *NotFoundError {
	if message == "" {
		message = "Resource not found"
	}
	return &NotFoundError{
		APIError: APIError{Message: message, StatusCode: 404},
	}
}

// ValidationError is returned when request validation fails (400/422).
type ValidationError struct {
	APIError
	Errors []ValidationErrorDetail `json:"errors,omitempty"`
}

// ValidationErrorDetail contains details about a validation error.
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewValidationError creates a new ValidationError.
func NewValidationError(message string, errors []ValidationErrorDetail) *ValidationError {
	if message == "" {
		message = "Validation error"
	}
	return &ValidationError{
		APIError: APIError{Message: message, StatusCode: 422},
		Errors:   errors,
	}
}

// RateLimitError is returned when rate limit is exceeded (429).
type RateLimitError struct {
	APIError
	RetryAfter int `json:"retry_after,omitempty"`
}

// NewRateLimitError creates a new RateLimitError.
func NewRateLimitError(retryAfter int) *RateLimitError {
	return &RateLimitError{
		APIError:   APIError{Message: "Rate limit exceeded", StatusCode: 429},
		RetryAfter: retryAfter,
	}
}

// ServerError is returned when the server returns an error (5xx).
type ServerError struct {
	APIError
}

// NewServerError creates a new ServerError.
func NewServerError(message string, statusCode int) *ServerError {
	if message == "" {
		message = "Server error"
	}
	if statusCode == 0 {
		statusCode = 500
	}
	return &ServerError{
		APIError: APIError{Message: message, StatusCode: statusCode},
	}
}

// NetworkError is returned when a network error occurs.
type NetworkError struct {
	APIError
	Cause error
}

// NewNetworkError creates a new NetworkError.
func NewNetworkError(cause error) *NetworkError {
	message := "Network error"
	if cause != nil {
		message = fmt.Sprintf("Network error: %v", cause)
	}
	return &NetworkError{
		APIError: APIError{Message: message},
		Cause:    cause,
	}
}

// Unwrap returns the underlying error.
func (e *NetworkError) Unwrap() error {
	return e.Cause
}

// TimeoutError is returned when a request times out.
type TimeoutError struct {
	APIError
}

// NewTimeoutError creates a new TimeoutError.
func NewTimeoutError() *TimeoutError {
	return &TimeoutError{
		APIError: APIError{Message: "Request timed out"},
	}
}
