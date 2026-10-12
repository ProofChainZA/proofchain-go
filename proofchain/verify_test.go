package proofchain

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCertificateVerifyResultIsValid(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{status: "VALID", want: true},
		{status: "INVALID", want: false},
		{status: "REVOKED", want: false},
		{status: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			r := &CertificateVerifyResult{Status: tt.status}
			if got := r.IsValid(); got != tt.want {
				t.Errorf("IsValid() with status %q = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestVerifyResourceCertificate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/verify/cert/ABC123" {
			t.Errorf("path = %q, want /verify/cert/ABC123", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"certificate_id":"ABC123","status":"VALID","type":"event"}`))
	}))
	defer srv.Close()

	client := mustClient(t, "k", WithBaseURL(srv.URL))
	cert, err := client.VerifyResource.Certificate(context.Background(), "ABC123")
	if err != nil {
		t.Fatalf("Certificate verify failed: %v", err)
	}
	if !cert.IsValid() {
		t.Error("expected certificate to be valid")
	}
	if cert.CertificateID != "ABC123" {
		t.Errorf("certificate id = %q, want ABC123", cert.CertificateID)
	}
}

func TestVerifyResourceProof(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/verify/proof" {
			t.Errorf("path = %q, want /verify/proof", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"valid":true,"message":"proof ok","verified_at":"2026-07-21"}`))
	}))
	defer srv.Close()

	client := mustClient(t, "k", WithBaseURL(srv.URL))
	res, err := client.VerifyResource.Proof(context.Background(), &ProofVerifyRequest{
		Leaf:  "leaf",
		Proof: []string{"a", "b"},
		Root:  "root",
	})
	if err != nil {
		t.Fatalf("Proof verify failed: %v", err)
	}
	if !res.Valid {
		t.Error("expected proof to be valid")
	}
}

func TestVerifyResourceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := mustClient(t, "k", WithBaseURL(srv.URL))
	_, err := client.VerifyResource.Certificate(context.Background(), "missing")
	if _, ok := err.(*NotFoundError); !ok {
		t.Fatalf("want *NotFoundError, got %T (%v)", err, err)
	}
}
