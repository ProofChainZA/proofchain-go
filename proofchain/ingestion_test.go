package proofchain

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleHTTPError(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		assertErr func(t *testing.T, err error)
	}{
		{
			name:   "401 uses detail message",
			status: 401,
			body:   `{"detail":"key expired"}`,
			assertErr: func(t *testing.T, err error) {
				ae, ok := err.(*AuthenticationError)
				if !ok {
					t.Fatalf("want *AuthenticationError, got %T", err)
				}
				if ae.Message != "key expired" {
					t.Errorf("message = %q, want key expired", ae.Message)
				}
			},
		},
		{
			name:   "403 falls back to message field",
			status: 403,
			body:   `{"message":"forbidden zone"}`,
			assertErr: func(t *testing.T, err error) {
				ae, ok := err.(*AuthorizationError)
				if !ok {
					t.Fatalf("want *AuthorizationError, got %T", err)
				}
				if ae.Message != "forbidden zone" {
					t.Errorf("message = %q, want forbidden zone", ae.Message)
				}
			},
		},
		{
			name:   "404 not found",
			status: 404,
			body:   `{}`,
			assertErr: func(t *testing.T, err error) {
				if _, ok := err.(*NotFoundError); !ok {
					t.Errorf("want *NotFoundError, got %T", err)
				}
			},
		},
		{
			name:   "422 validation",
			status: 422,
			body:   `{"detail":"invalid"}`,
			assertErr: func(t *testing.T, err error) {
				if _, ok := err.(*ValidationError); !ok {
					t.Errorf("want *ValidationError, got %T", err)
				}
			},
		},
		{
			name:   "429 rate limit",
			status: 429,
			body:   `{}`,
			assertErr: func(t *testing.T, err error) {
				if _, ok := err.(*RateLimitError); !ok {
					t.Errorf("want *RateLimitError, got %T", err)
				}
			},
		},
		{
			name:   "503 server error",
			status: 503,
			body:   `{"detail":"unavailable"}`,
			assertErr: func(t *testing.T, err error) {
				se, ok := err.(*ServerError)
				if !ok {
					t.Fatalf("want *ServerError, got %T", err)
				}
				if se.StatusCode != 503 {
					t.Errorf("status = %d, want 503", se.StatusCode)
				}
			},
		},
		{
			name:   "418 default APIError with fallback message",
			status: 418,
			body:   `{}`,
			assertErr: func(t *testing.T, err error) {
				ae, ok := err.(*APIError)
				if !ok {
					t.Fatalf("want *APIError, got %T", err)
				}
				if ae.Message != "HTTP 418" {
					t.Errorf("message = %q, want HTTP 418", ae.Message)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertErr(t, handleHTTPError(tt.status, []byte(tt.body)))
		})
	}
}

// mustIngestionClient fails the test on constructor error.
func mustIngestionClient(t *testing.T, apiKey string, opts ...IngestionClientOption) *IngestionClient {
	t.Helper()
	c, err := NewIngestionClient(apiKey, opts...)
	if err != nil {
		t.Fatalf("NewIngestionClient: %v", err)
	}
	return c
}

func TestIngestionClientOptions(t *testing.T) {
	c := mustIngestionClient(t, "k",
		WithIngestURL("https://ingest.example.com"),
		WithIngestTimeout(12*time.Second),
	)
	if c.ingestURL != "https://ingest.example.com" {
		t.Errorf("ingestURL = %q, want the override", c.ingestURL)
	}
	if c.timeout != 12*time.Second {
		t.Errorf("timeout = %v, want 12s", c.timeout)
	}
	if c.httpClient.Timeout != 12*time.Second {
		t.Errorf("httpClient.Timeout = %v, want 12s", c.httpClient.Timeout)
	}
}

func TestIngest(t *testing.T) {
	var gotPath, gotSchemas, gotAPIKey string
	var gotPayload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotSchemas = r.Header.Get("X-Schemas")
		gotAPIKey = r.Header.Get("X-API-Key")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotPayload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"event_id":"ev1","certificate_id":"cert1","status":"queued","queue_position":3}`))
	}))
	defer srv.Close()

	client := mustIngestionClient(t, "api-key", WithIngestURL(srv.URL))
	resp, err := client.Ingest(context.Background(), &IngestEventRequest{
		UserID:    "user-1",
		EventType: "purchase",
		Data:      map[string]any{"amount": 9.99},
		SchemaIDs: []string{"s1", "s2"},
		Hot:       true,
	})
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}

	if gotPath != "/events/ingest" {
		t.Errorf("path = %q, want /events/ingest", gotPath)
	}
	if gotAPIKey != "api-key" {
		t.Errorf("X-API-Key = %q, want api-key", gotAPIKey)
	}
	if gotSchemas != "s1,s2" {
		t.Errorf("X-Schemas = %q, want s1,s2", gotSchemas)
	}
	if gotPayload["event_source"] != "sdk" {
		t.Errorf("event_source = %v, want sdk (default)", gotPayload["event_source"])
	}
	if gotPayload["hot"] != true {
		t.Errorf("hot = %v, want true", gotPayload["hot"])
	}
	if resp.EventID != "ev1" || resp.CertificateID != "cert1" || resp.Status != "queued" || resp.QueuePosition != 3 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestIngestDefaultsEmptyData(t *testing.T) {
	var gotPayload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotPayload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"event_id":"e","status":"queued"}`))
	}))
	defer srv.Close()

	client := mustIngestionClient(t, "k", WithIngestURL(srv.URL))
	_, err := client.Ingest(context.Background(), &IngestEventRequest{UserID: "u", EventType: "t"})
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}
	// nil Data should serialize to an empty object, not null.
	if _, ok := gotPayload["data"].(map[string]any); !ok {
		t.Errorf("data = %v, want empty object", gotPayload["data"])
	}
}

func TestIngestErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"detail":"bad key"}`))
	}))
	defer srv.Close()

	client := mustIngestionClient(t, "k", WithIngestURL(srv.URL))
	_, err := client.Ingest(context.Background(), &IngestEventRequest{UserID: "u", EventType: "t"})
	if _, ok := err.(*AuthenticationError); !ok {
		t.Fatalf("want *AuthenticationError, got %T (%v)", err, err)
	}
}

func TestIngestBatchTooLarge(t *testing.T) {
	client := mustIngestionClient(t, "k")
	events := make([]IngestEventRequest, 1001)
	_, err := client.IngestBatch(context.Background(), &BatchIngestRequest{Events: events})
	if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("want *ValidationError for oversized batch, got %T (%v)", err, err)
	}
}

func TestIngestBatch(t *testing.T) {
	var sentEvents []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events/ingest/batch" {
			t.Errorf("path = %q, want /events/ingest/batch", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		// Batch endpoint expects a bare array, not a wrapped object.
		_ = json.Unmarshal(body, &sentEvents)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total_events":2,"queued":2,"failed":0,"results":[{"event_id":"a","status":"queued"},{"event_id":"b","status":"queued"}]}`))
	}))
	defer srv.Close()

	client := mustIngestionClient(t, "k", WithIngestURL(srv.URL))
	resp, err := client.IngestBatch(context.Background(), &BatchIngestRequest{
		Events: []IngestEventRequest{
			{UserID: "u1", EventType: "t"},
			{UserID: "u2", EventType: "t"},
		},
	})
	if err != nil {
		t.Fatalf("IngestBatch failed: %v", err)
	}
	if len(sentEvents) != 2 {
		t.Fatalf("server received %d events, want 2", len(sentEvents))
	}
	if resp.TotalEvents != 2 || resp.Queued != 2 || len(resp.Results) != 2 {
		t.Errorf("unexpected response: %+v", resp)
	}
	if resp.Results[0].EventID != "a" || resp.Results[1].EventID != "b" {
		t.Errorf("result event ids = %+v", resp.Results)
	}
}

func TestIngestBatchAlternateResponseShape(t *testing.T) {
	// Server uses the "total" + "responses" field names instead of
	// "total_events" + "results"; the client must fall back to them.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total":1,"queued":1,"failed":0,"responses":[{"event_id":"z","certificate_id":"c","status":"confirmed"}]}`))
	}))
	defer srv.Close()

	client := mustIngestionClient(t, "k", WithIngestURL(srv.URL))
	resp, err := client.IngestBatch(context.Background(), &BatchIngestRequest{
		Events: []IngestEventRequest{{UserID: "u", EventType: "t"}},
	})
	if err != nil {
		t.Fatalf("IngestBatch failed: %v", err)
	}
	if resp.TotalEvents != 1 {
		t.Errorf("TotalEvents = %d, want 1 (from 'total' fallback)", resp.TotalEvents)
	}
	if len(resp.Results) != 1 || resp.Results[0].EventID != "z" {
		t.Errorf("results not populated from 'responses' fallback: %+v", resp.Results)
	}
}

func TestGetEventStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events/ev-123/status" {
			t.Errorf("path = %q, want /events/ev-123/status", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"settled"}`))
	}))
	defer srv.Close()

	client := mustIngestionClient(t, "k", WithIngestURL(srv.URL))
	status, err := client.GetEventStatus(context.Background(), "ev-123")
	if err != nil {
		t.Fatalf("GetEventStatus failed: %v", err)
	}
	if status != "settled" {
		t.Errorf("status = %q, want settled", status)
	}
}

func TestGetEventStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"no such event"}`))
	}))
	defer srv.Close()

	client := mustIngestionClient(t, "k", WithIngestURL(srv.URL))
	_, err := client.GetEventStatus(context.Background(), "nope")
	if _, ok := err.(*NotFoundError); !ok {
		t.Fatalf("want *NotFoundError, got %T (%v)", err, err)
	}
}
