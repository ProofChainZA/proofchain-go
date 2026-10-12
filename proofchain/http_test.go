package proofchain

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

type sampleResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestHandleResponseStatusMapping(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantErr    bool
		assertErr  func(t *testing.T, err error)
		wantResult *sampleResult
	}{
		{
			name:       "200 decodes body",
			status:     http.StatusOK,
			body:       `{"id":"abc","name":"widget"}`,
			wantResult: &sampleResult{ID: "abc", Name: "widget"},
		},
		{
			name:   "204 no content",
			status: http.StatusNoContent,
			body:   "",
		},
		{
			name:    "401 unauthorized",
			status:  http.StatusUnauthorized,
			body:    "",
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				if _, ok := err.(*AuthenticationError); !ok {
					t.Errorf("want *AuthenticationError, got %T", err)
				}
			},
		},
		{
			name:    "403 forbidden",
			status:  http.StatusForbidden,
			body:    "",
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				if _, ok := err.(*AuthorizationError); !ok {
					t.Errorf("want *AuthorizationError, got %T", err)
				}
			},
		},
		{
			name:    "404 not found",
			status:  http.StatusNotFound,
			body:    "",
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				if _, ok := err.(*NotFoundError); !ok {
					t.Errorf("want *NotFoundError, got %T", err)
				}
			},
		},
		{
			name:    "422 validation with detail",
			status:  http.StatusUnprocessableEntity,
			body:    `{"detail":"bad input","errors":[{"field":"x","message":"required"}]}`,
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("want *ValidationError, got %T", err)
				}
				if ve.Message != "bad input" {
					t.Errorf("message = %q, want %q", ve.Message, "bad input")
				}
				if len(ve.Errors) != 1 || ve.Errors[0].Field != "x" {
					t.Errorf("validation details not parsed: %+v", ve.Errors)
				}
			},
		},
		{
			name:    "400 plain body validation",
			status:  http.StatusBadRequest,
			body:    "plain text error",
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				if _, ok := err.(*ValidationError); !ok {
					t.Errorf("want *ValidationError, got %T", err)
				}
			},
		},
		{
			name:    "500 server error",
			status:  http.StatusInternalServerError,
			body:    "kaboom",
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				se, ok := err.(*ServerError)
				if !ok {
					t.Fatalf("want *ServerError, got %T", err)
				}
				if se.StatusCode != 500 {
					t.Errorf("status = %d, want 500", se.StatusCode)
				}
			},
		},
		{
			name:    "418 falls through to APIError",
			status:  http.StatusTeapot,
			body:    "teapot",
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				ae, ok := err.(*APIError)
				if !ok {
					t.Fatalf("want *APIError, got %T", err)
				}
				if ae.StatusCode != http.StatusTeapot {
					t.Errorf("status = %d, want %d", ae.StatusCode, http.StatusTeapot)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				if tt.body != "" {
					_, _ = w.Write([]byte(tt.body))
				}
			}))
			defer srv.Close()

			client := mustHTTPClient(t, "test-key", WithBaseURL(srv.URL), WithRetries(0))
			var result sampleResult
			err := client.Get(context.Background(), "/thing", nil, &result)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.assertErr != nil {
					tt.assertErr(t, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantResult != nil && result != *tt.wantResult {
				t.Errorf("result = %+v, want %+v", result, *tt.wantResult)
			}
		})
	}
}

func TestSetAuthHeaders(t *testing.T) {
	tests := []struct {
		name       string
		opts       []HTTPClientOption
		wantAPIKey string
		wantBearer string
		wantTenant string
	}{
		{
			name:       "api key auth",
			opts:       nil,
			wantAPIKey: "secret-key",
		},
		{
			name:       "user token auth",
			opts:       []HTTPClientOption{WithUserToken("jwt-abc", "tenant-1")},
			wantBearer: "Bearer jwt-abc",
			wantTenant: "tenant-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotAPIKey, gotAuth, gotTenant string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAPIKey = r.Header.Get("X-API-Key")
				gotAuth = r.Header.Get("Authorization")
				gotTenant = r.Header.Get("X-Tenant-ID")
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			client := mustHTTPClient(t, "secret-key", append([]HTTPClientOption{WithBaseURL(srv.URL)}, tt.opts...)...)
			if err := client.Get(context.Background(), "/ping", nil, nil); err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if gotAPIKey != tt.wantAPIKey {
				t.Errorf("X-API-Key = %q, want %q", gotAPIKey, tt.wantAPIKey)
			}
			if gotAuth != tt.wantBearer {
				t.Errorf("Authorization = %q, want %q", gotAuth, tt.wantBearer)
			}
			if gotTenant != tt.wantTenant {
				t.Errorf("X-Tenant-ID = %q, want %q", gotTenant, tt.wantTenant)
			}
		})
	}
}

func TestRequestSendsBodyAndQueryParams(t *testing.T) {
	var gotBody string
	var gotQuery string
	var gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		gotBody = string(buf)
		gotQuery = r.URL.RawQuery
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	defer srv.Close()

	client := mustHTTPClient(t, "k", WithBaseURL(srv.URL))

	// POST with JSON body
	var res sampleResult
	err := client.Post(context.Background(), "/create", map[string]string{"name": "x"}, &res)
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	if gotBody != `{"name":"x"}` {
		t.Errorf("body = %q, want %q", gotBody, `{"name":"x"}`)
	}
	if gotContentType != "application/json" {
		t.Errorf("content-type = %q, want application/json", gotContentType)
	}

	// GET with query params
	params := url.Values{}
	params.Set("limit", "5")
	if err := client.Get(context.Background(), "/list", params, &res); err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if gotQuery != "limit=5" {
		t.Errorf("query = %q, want limit=5", gotQuery)
	}
}

func TestExecuteRequestRetriesRateLimitThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"done"}`))
	}))
	defer srv.Close()

	client := mustHTTPClient(t, "k", WithBaseURL(srv.URL), WithRetries(2))
	start := time.Now()
	var res sampleResult
	err := client.Get(context.Background(), "/x", nil, &res)
	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if res.ID != "done" {
		t.Errorf("id = %q, want done", res.ID)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("server calls = %d, want 2", calls)
	}
	// Retry-After was 1s, so the retry should not have been instant.
	if elapsed := time.Since(start); elapsed < 900*time.Millisecond {
		t.Errorf("retry happened too fast (%v); Retry-After was not honored", elapsed)
	}
}

func TestExecuteRequestRateLimitNoRetry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := mustHTTPClient(t, "k", WithBaseURL(srv.URL), WithRetries(0))
	err := client.Get(context.Background(), "/x", nil, nil)
	if _, ok := err.(*RateLimitError); !ok {
		t.Fatalf("want *RateLimitError, got %T (%v)", err, err)
	}
}

func TestExecuteRequestNetworkError(t *testing.T) {
	// Point at a closed server to force a dial failure.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := srv.URL
	srv.Close()

	client := mustHTTPClient(t, "k", WithBaseURL(closedURL), WithRetries(1))
	err := client.Get(context.Background(), "/x", nil, nil)
	if _, ok := err.(*NetworkError); !ok {
		t.Fatalf("want *NetworkError, got %T (%v)", err, err)
	}
}

func TestExecuteRequestContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := mustHTTPClient(t, "k", WithBaseURL(srv.URL), WithRetries(0))
	err := client.Get(ctx, "/x", nil, nil)
	if _, ok := err.(*TimeoutError); !ok {
		t.Fatalf("want *TimeoutError, got %T (%v)", err, err)
	}
}

func TestGetRaw(t *testing.T) {
	want := []byte("binary-file-content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "k" {
			t.Errorf("missing auth header on GetRaw")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(want)
	}))
	defer srv.Close()

	client := mustHTTPClient(t, "k", WithBaseURL(srv.URL))
	got, err := client.GetRaw(context.Background(), "/download")
	if err != nil {
		t.Fatalf("GetRaw failed: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("GetRaw = %q, want %q", got, want)
	}
}

func TestGetRawErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := mustHTTPClient(t, "k", WithBaseURL(srv.URL))
	_, err := client.GetRaw(context.Background(), "/missing")
	if _, ok := err.(*NotFoundError); !ok {
		t.Fatalf("want *NotFoundError, got %T (%v)", err, err)
	}
}

func TestRequestMultipart(t *testing.T) {
	var gotFilename, gotField string
	var gotFileBytes []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm failed: %v", err)
		}
		gotField = r.FormValue("ipfs_hash")
		file, hdr, err := r.FormFile("file")
		if err != nil {
			t.Errorf("FormFile failed: %v", err)
		} else {
			defer file.Close()
			gotFilename = hdr.Filename
			gotFileBytes = make([]byte, hdr.Size)
			_, _ = file.Read(gotFileBytes)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"uploaded"}`))
	}))
	defer srv.Close()

	client := mustHTTPClient(t, "k", WithBaseURL(srv.URL))
	var res sampleResult
	err := client.RequestMultipart(context.Background(), "/verify/document",
		map[string]string{"ipfs_hash": "Qm123"}, "file", "doc.pdf", []byte("PDFDATA"), &res)
	if err != nil {
		t.Fatalf("RequestMultipart failed: %v", err)
	}
	if res.ID != "uploaded" {
		t.Errorf("id = %q, want uploaded", res.ID)
	}
	if gotFilename != "doc.pdf" {
		t.Errorf("filename = %q, want doc.pdf", gotFilename)
	}
	if gotField != "Qm123" {
		t.Errorf("field = %q, want Qm123", gotField)
	}
	if string(gotFileBytes) != "PDFDATA" {
		t.Errorf("file bytes = %q, want PDFDATA", gotFileBytes)
	}
}

func TestNewHTTPClientFromEnv(t *testing.T) {
	t.Run("missing key errors", func(t *testing.T) {
		t.Setenv("PROOFCHAIN_API_KEY", "")
		if _, err := NewHTTPClientFromEnv(); err == nil {
			t.Error("expected error when PROOFCHAIN_API_KEY unset")
		}
	})
	t.Run("uses env key and base url", func(t *testing.T) {
		t.Setenv("PROOFCHAIN_API_KEY", "env-key")
		t.Setenv("PROOFCHAIN_BASE_URL", "https://custom.example.com")
		c, err := NewHTTPClientFromEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.apiKey != "env-key" {
			t.Errorf("apiKey = %q, want env-key", c.apiKey)
		}
		if c.baseURL != "https://custom.example.com" {
			t.Errorf("baseURL = %q, want the env override", c.baseURL)
		}
	})
}

func TestHTTPClientOptions(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{}}
	c := mustHTTPClient(t, "k",
		WithBaseURL("https://x"),
		WithHTTPClient(custom),
		WithTimeout(15*time.Second),
		WithRetries(7),
	)
	// WithHTTPClient replaces the client; WithTimeout/WithRetries set fields.
	if c.maxRetries != 7 {
		t.Errorf("maxRetries = %d, want 7", c.maxRetries)
	}
	if c.baseURL != "https://x" {
		t.Errorf("baseURL = %q, want https://x", c.baseURL)
	}
	// The constructor hardens a copy of the caller's client (no redirects)
	// rather than mutating it, so compare the transport and timeout.
	if c.httpClient.Transport != custom.Transport {
		t.Error("WithHTTPClient did not replace the underlying client")
	}
	if c.httpClient.Timeout != 15*time.Second {
		t.Errorf("timeout = %v, want 15s", c.httpClient.Timeout)
	}
	if c.httpClient.CheckRedirect == nil {
		t.Error("hardened client must not follow redirects")
	}
}
