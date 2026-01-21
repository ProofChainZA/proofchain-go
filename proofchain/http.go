package proofchain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	defaultBaseURL = "https://api.proofchain.co.za"
	defaultTimeout = 30 * time.Second
	userAgent      = "proofchain-go/0.1.0"
)

// HTTPClient handles HTTP requests to the ProofChain API.
type HTTPClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

// HTTPClientOption is a function that configures the HTTP client.
type HTTPClientOption func(*HTTPClient)

// WithBaseURL sets a custom base URL.
func WithBaseURL(baseURL string) HTTPClientOption {
	return func(c *HTTPClient) {
		c.baseURL = baseURL
	}
}

// WithTimeout sets a custom timeout.
func WithTimeout(timeout time.Duration) HTTPClientOption {
	return func(c *HTTPClient) {
		c.httpClient.Timeout = timeout
	}
}

// WithRetries sets the maximum number of retries.
func WithRetries(maxRetries int) HTTPClientOption {
	return func(c *HTTPClient) {
		c.maxRetries = maxRetries
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) HTTPClientOption {
	return func(c *HTTPClient) {
		c.httpClient = httpClient
	}
}

// NewHTTPClient creates a new HTTP client.
func NewHTTPClient(apiKey string, opts ...HTTPClientOption) *HTTPClient {
	c := &HTTPClient{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		maxRetries: 3,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// NewHTTPClientFromEnv creates a client using the PROOFCHAIN_API_KEY environment variable.
func NewHTTPClientFromEnv(opts ...HTTPClientOption) (*HTTPClient, error) {
	apiKey := os.Getenv("PROOFCHAIN_API_KEY")
	if apiKey == "" {
		return nil, NewAuthenticationError("PROOFCHAIN_API_KEY environment variable not set")
	}

	baseURL := os.Getenv("PROOFCHAIN_BASE_URL")
	if baseURL != "" {
		opts = append(opts, WithBaseURL(baseURL))
	}

	return NewHTTPClient(apiKey, opts...), nil
}

// Request makes an HTTP request to the API.
func (c *HTTPClient) Request(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, method, path, body, nil, result)
}

// RequestWithParams makes an HTTP request with query parameters.
func (c *HTTPClient) RequestWithParams(ctx context.Context, method, path string, params url.Values, result interface{}) error {
	return c.doRequest(ctx, method, path, nil, params, result)
}

// RequestMultipart makes a multipart form request.
func (c *HTTPClient) RequestMultipart(ctx context.Context, path string, fields map[string]string, fileField, filename string, fileContent []byte, result interface{}) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file
	part, err := writer.CreateFormFile(fileField, filename)
	if err != nil {
		return NewNetworkError(err)
	}
	if _, err := part.Write(fileContent); err != nil {
		return NewNetworkError(err)
	}

	// Add other fields
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return NewNetworkError(err)
		}
	}

	if err := writer.Close(); err != nil {
		return NewNetworkError(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, &buf)
	if err != nil {
		return NewNetworkError(err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", userAgent)

	return c.executeRequest(req, result)
}

func (c *HTTPClient) doRequest(ctx context.Context, method, path string, body interface{}, params url.Values, result interface{}) error {
	fullURL := c.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return NewNetworkError(err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return NewNetworkError(err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	return c.executeRequest(req, result)
}

func (c *HTTPClient) executeRequest(req *http.Request, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			if ctx := req.Context(); ctx.Err() != nil {
				return NewTimeoutError()
			}
			lastErr = NewNetworkError(err)
			continue
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = NewNetworkError(err)
			continue
		}

		if err := c.handleResponse(resp.StatusCode, respBody, result); err != nil {
			// Retry on rate limit
			if rateLimitErr, ok := err.(*RateLimitError); ok && attempt < c.maxRetries {
				sleepDuration := time.Duration(rateLimitErr.RetryAfter) * time.Second
				if sleepDuration > 60*time.Second {
					sleepDuration = 60 * time.Second
				}
				if sleepDuration > 0 {
					time.Sleep(sleepDuration)
				}
				lastErr = err
				continue
			}
			return err
		}

		return nil
	}

	if lastErr != nil {
		return lastErr
	}
	return NewNetworkError(fmt.Errorf("request failed after %d retries", c.maxRetries))
}

func (c *HTTPClient) handleResponse(statusCode int, body []byte, result interface{}) error {
	switch statusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		if result != nil && len(body) > 0 {
			if err := json.Unmarshal(body, result); err != nil {
				return NewNetworkError(fmt.Errorf("failed to parse response: %w", err))
			}
		}
		return nil

	case http.StatusNoContent:
		return nil

	case http.StatusUnauthorized:
		return NewAuthenticationError("")

	case http.StatusForbidden:
		return NewAuthorizationError("")

	case http.StatusNotFound:
		return NewNotFoundError("")

	case http.StatusUnprocessableEntity, http.StatusBadRequest:
		var errResp struct {
			Detail string                  `json:"detail"`
			Errors []ValidationErrorDetail `json:"errors"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil {
			return NewValidationError(errResp.Detail, errResp.Errors)
		}
		return NewValidationError(string(body), nil)

	case http.StatusTooManyRequests:
		retryAfter := 60
		// Try to parse Retry-After header if available
		return NewRateLimitError(retryAfter)

	default:
		if statusCode >= 500 {
			return NewServerError(string(body), statusCode)
		}
		return &APIError{
			Message:    fmt.Sprintf("HTTP %d: %s", statusCode, string(body)),
			StatusCode: statusCode,
		}
	}
}

// Get makes a GET request.
func (c *HTTPClient) Get(ctx context.Context, path string, params url.Values, result interface{}) error {
	return c.RequestWithParams(ctx, http.MethodGet, path, params, result)
}

// Post makes a POST request.
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.Request(ctx, http.MethodPost, path, body, result)
}

// Put makes a PUT request.
func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.Request(ctx, http.MethodPut, path, body, result)
}

// Patch makes a PATCH request.
func (c *HTTPClient) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.Request(ctx, http.MethodPatch, path, body, result)
}

// Delete makes a DELETE request.
func (c *HTTPClient) Delete(ctx context.Context, path string) error {
	return c.Request(ctx, http.MethodDelete, path, nil, nil)
}

// GetRaw makes a GET request and returns raw bytes (for file downloads).
func (c *HTTPClient) GetRaw(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, NewNetworkError(err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewNetworkError(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewNetworkError(err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleResponse(resp.StatusCode, body, nil)
	}

	return body, nil
}

// Helper to convert int to string for query params
func intToString(i int) string {
	return strconv.Itoa(i)
}
