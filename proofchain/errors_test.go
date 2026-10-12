package proofchain

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIErrorError(t *testing.T) {
	tests := []struct {
		name string
		err  *APIError
		want string
	}{
		{name: "with status code", err: &APIError{Message: "boom", StatusCode: 500}, want: "[500] boom"},
		{name: "without status code", err: &APIError{Message: "boom"}, want: "boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorConstructorDefaults(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{name: "auth default message", err: NewAuthenticationError(""), wantStatus: 401, wantMsg: "Invalid or missing API key"},
		{name: "auth custom message", err: NewAuthenticationError("nope"), wantStatus: 401, wantMsg: "nope"},
		{name: "authorization default", err: NewAuthorizationError(""), wantStatus: 403, wantMsg: "Access denied"},
		{name: "not found default", err: NewNotFoundError(""), wantStatus: 404, wantMsg: "Resource not found"},
		{name: "validation default", err: NewValidationError("", nil), wantStatus: 422, wantMsg: "Validation error"},
		{name: "rate limit", err: NewRateLimitError(30), wantStatus: 429, wantMsg: "Rate limit exceeded"},
		{name: "server default status", err: NewServerError("", 0), wantStatus: 500, wantMsg: "Server error"},
		{name: "server custom status", err: NewServerError("bad gateway", 502), wantStatus: 502, wantMsg: "bad gateway"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var api *APIError
			// All typed errors embed APIError; extract via a small type switch.
			switch e := tt.err.(type) {
			case *AuthenticationError:
				api = &e.APIError
			case *AuthorizationError:
				api = &e.APIError
			case *NotFoundError:
				api = &e.APIError
			case *ValidationError:
				api = &e.APIError
			case *RateLimitError:
				api = &e.APIError
			case *ServerError:
				api = &e.APIError
			default:
				t.Fatalf("unhandled error type %T", tt.err)
			}
			if api.StatusCode != tt.wantStatus {
				t.Errorf("StatusCode = %d, want %d", api.StatusCode, tt.wantStatus)
			}
			if api.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", api.Message, tt.wantMsg)
			}
		})
	}
}

func TestRateLimitErrorRetryAfter(t *testing.T) {
	err := NewRateLimitError(45)
	if err.RetryAfter != 45 {
		t.Errorf("RetryAfter = %d, want 45", err.RetryAfter)
	}
}

func TestValidationErrorDetails(t *testing.T) {
	details := []ValidationErrorDetail{
		{Field: "email", Message: "required"},
		{Field: "name", Message: "too short"},
	}
	err := NewValidationError("invalid", details)
	if len(err.Errors) != 2 {
		t.Fatalf("len(Errors) = %d, want 2", len(err.Errors))
	}
	if err.Errors[0].Field != "email" {
		t.Errorf("first field = %q, want email", err.Errors[0].Field)
	}
}

func TestNetworkErrorUnwrap(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := NewNetworkError(cause)

	if got := err.Unwrap(); !errors.Is(got, cause) {
		t.Errorf("Unwrap() did not return the wrapped cause")
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is could not match the wrapped cause through NetworkError")
	}
	if err.Message != "Network error: connection refused" {
		t.Errorf("Message = %q, want it to include the cause", err.Message)
	}
}

func TestNetworkErrorNilCause(t *testing.T) {
	err := NewNetworkError(nil)
	if err.Message != "Network error" {
		t.Errorf("Message = %q, want %q", err.Message, "Network error")
	}
	if err.Unwrap() != nil {
		t.Error("expected nil unwrap for nil cause")
	}
}

func TestTimeoutError(t *testing.T) {
	err := NewTimeoutError()
	if err.Message != "Request timed out" {
		t.Errorf("Message = %q, want %q", err.Message, "Request timed out")
	}
}
