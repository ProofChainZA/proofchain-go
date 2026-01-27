// Package proofchain provides a Go client for the ProofChain API.
package proofchain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultIngestURL     = "https://ingest.proofchain.co.za"
	defaultIngestTimeout = 30 * time.Second
)

// handleHTTPError parses an HTTP error response and returns the appropriate error type.
func handleHTTPError(statusCode int, body []byte) error {
	var errResp struct {
		Detail  string `json:"detail"`
		Message string `json:"message"`
	}
	json.Unmarshal(body, &errResp)

	message := errResp.Detail
	if message == "" {
		message = errResp.Message
	}
	if message == "" {
		message = fmt.Sprintf("HTTP %d", statusCode)
	}

	switch statusCode {
	case 401:
		return NewAuthenticationError(message)
	case 403:
		return NewAuthorizationError(message)
	case 404:
		return NewNotFoundError(message)
	case 400, 422:
		return NewValidationError(message, nil)
	case 429:
		return NewRateLimitError(0)
	default:
		if statusCode >= 500 {
			return NewServerError(message, statusCode)
		}
		return &APIError{Message: message, StatusCode: statusCode}
	}
}

// IngestEventRequest is the request for ingesting a single event.
type IngestEventRequest struct {
	UserID      string                 `json:"user_id"`
	EventType   string                 `json:"event_type"`
	Data        map[string]interface{} `json:"data,omitempty"`
	EventSource string                 `json:"event_source,omitempty"`
	SchemaIDs   []string               `json:"-"` // Sent via header
}

// IngestEventResponse is the response from ingesting an event.
type IngestEventResponse struct {
	EventID               string `json:"event_id"`
	CertificateID         string `json:"certificate_id"`
	Status                string `json:"status"`
	QueuePosition         int    `json:"queue_position,omitempty"`
	EstimatedConfirmation string `json:"estimated_confirmation,omitempty"`
}

// BatchIngestRequest is the request for ingesting multiple events.
type BatchIngestRequest struct {
	Events []IngestEventRequest `json:"events"`
}

// BatchIngestResponse is the response from batch ingestion.
type BatchIngestResponse struct {
	TotalEvents int                   `json:"total_events"`
	Queued      int                   `json:"queued"`
	Failed      int                   `json:"failed"`
	Results     []IngestEventResponse `json:"results"`
}

// IngestionClientOption is a function that configures the ingestion client.
type IngestionClientOption func(*IngestionClient)

// WithIngestURL sets a custom ingestion URL.
func WithIngestURL(url string) IngestionClientOption {
	return func(c *IngestionClient) {
		c.ingestURL = url
	}
}

// WithIngestTimeout sets a custom timeout for ingestion requests.
func WithIngestTimeout(timeout time.Duration) IngestionClientOption {
	return func(c *IngestionClient) {
		c.timeout = timeout
	}
}

// IngestionClient is a high-performance client for the Rust ingestion API.
// Use this for maximum throughput when ingesting events.
type IngestionClient struct {
	apiKey     string
	ingestURL  string
	timeout    time.Duration
	httpClient *http.Client
}

// NewIngestionClient creates a new high-performance ingestion client.
//
// Example:
//
//	client := proofchain.NewIngestionClient("your-api-key")
//	result, err := client.Ingest(ctx, &proofchain.IngestEventRequest{
//	    UserID:    "user-123",
//	    EventType: "purchase",
//	    Data:      map[string]interface{}{"amount": 99.99},
//	})
func NewIngestionClient(apiKey string, opts ...IngestionClientOption) *IngestionClient {
	c := &IngestionClient{
		apiKey:    apiKey,
		ingestURL: defaultIngestURL,
		timeout:   defaultIngestTimeout,
		httpClient: &http.Client{
			Timeout: defaultIngestTimeout,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	c.httpClient.Timeout = c.timeout
	return c
}

// Ingest sends a single event to the high-performance Rust ingestion API.
// Events are attested immediately upon ingestion.
func (c *IngestionClient) Ingest(ctx context.Context, req *IngestEventRequest) (*IngestEventResponse, error) {
	source := req.EventSource
	if source == "" {
		source = "sdk"
	}

	data := req.Data
	if data == nil {
		data = map[string]interface{}{}
	}

	payload := map[string]interface{}{
		"user_id":      req.UserID,
		"event_type":   req.EventType,
		"data":         data,
		"event_source": source,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.ingestURL+"/events/ingest", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("User-Agent", userAgent)

	if len(req.SchemaIDs) > 0 {
		schemas := ""
		for i, id := range req.SchemaIDs {
			if i > 0 {
				schemas += ","
			}
			schemas += id
		}
		httpReq.Header.Set("X-Schemas", schemas)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, handleHTTPError(resp.StatusCode, respBody)
	}

	var result struct {
		EventID               string `json:"event_id"`
		CertificateID         string `json:"certificate_id"`
		Status                string `json:"status"`
		QueuePosition         int    `json:"queue_position"`
		EstimatedConfirmation string `json:"estimated_confirmation"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &IngestEventResponse{
		EventID:               result.EventID,
		CertificateID:         result.CertificateID,
		Status:                result.Status,
		QueuePosition:         result.QueuePosition,
		EstimatedConfirmation: result.EstimatedConfirmation,
	}, nil
}

// IngestBatch sends multiple events in a single request (up to 1000 events).
// More efficient than individual calls for bulk data.
func (c *IngestionClient) IngestBatch(ctx context.Context, req *BatchIngestRequest) (*BatchIngestResponse, error) {
	if len(req.Events) > 1000 {
		return nil, NewValidationError("batch size cannot exceed 1000 events", nil)
	}

	events := make([]map[string]interface{}, len(req.Events))
	for i, e := range req.Events {
		source := e.EventSource
		if source == "" {
			source = "sdk"
		}
		data := e.Data
		if data == nil {
			data = map[string]interface{}{}
		}
		events[i] = map[string]interface{}{
			"user_id":      e.UserID,
			"event_type":   e.EventType,
			"data":         data,
			"event_source": source,
		}
	}

	// Batch endpoint expects array directly, not wrapped in {"events": [...]}
	body, err := json.Marshal(events)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.ingestURL+"/events/ingest/batch", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, handleHTTPError(resp.StatusCode, respBody)
	}

	var result struct {
		TotalEvents int `json:"total_events"`
		Total       int `json:"total"`
		Queued      int `json:"queued"`
		Failed      int `json:"failed"`
		Results     []struct {
			EventID       string `json:"event_id"`
			CertificateID string `json:"certificate_id"`
			Status        string `json:"status"`
		} `json:"results"`
		Responses []struct {
			EventID       string `json:"event_id"`
			CertificateID string `json:"certificate_id"`
			Status        string `json:"status"`
		} `json:"responses"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	totalEvents := result.TotalEvents
	if totalEvents == 0 {
		totalEvents = result.Total
	}

	results := result.Results
	if len(results) == 0 {
		for _, r := range result.Responses {
			results = append(results, struct {
				EventID       string `json:"event_id"`
				CertificateID string `json:"certificate_id"`
				Status        string `json:"status"`
			}{
				EventID:       r.EventID,
				CertificateID: r.CertificateID,
				Status:        r.Status,
			})
		}
	}

	response := &BatchIngestResponse{
		TotalEvents: totalEvents,
		Queued:      result.Queued,
		Failed:      result.Failed,
		Results:     make([]IngestEventResponse, len(results)),
	}

	for i, r := range results {
		response.Results[i] = IngestEventResponse{
			EventID:       r.EventID,
			CertificateID: r.CertificateID,
			Status:        r.Status,
		}
	}

	return response, nil
}

// GetEventStatus retrieves the status of an event by ID.
func (c *IngestionClient) GetEventStatus(ctx context.Context, eventID string) (string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.ingestURL+"/events/"+eventID+"/status", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", handleHTTPError(resp.StatusCode, respBody)
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Status, nil
}
