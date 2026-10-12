package proofchain

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientWiresResources(t *testing.T) {
	c := mustClient(t, "k")
	checks := map[string]any{
		"http":         c.http,
		"Documents":    c.Documents,
		"Events":       c.Events,
		"Channels":     c.Channels,
		"Certificates": c.Certificates,
		"Webhooks":     c.Webhooks,
		"Vault":        c.Vault,
		"Search":       c.Search,
		"Verify":       c.VerifyResource,
		"Tenant":       c.Tenant,
		"Passports":    c.Passports,
		"Wallets":      c.Wallets,
		"Users":        c.Users,
		"Rewards":      c.Rewards,
		"Quests":       c.Quests,
		"Schemas":      c.Schemas,
		"DataViews":    c.DataViews,
		"Cohorts":      c.Cohorts,
		"Fanpass":      c.Fanpass,
		"Credentials":  c.Credentials,
		"PartnerKeys":  c.PartnerKeys,
	}
	for name, ptr := range checks {
		if ptr == nil {
			t.Errorf("resource %q was not initialized", name)
		}
	}
}

func TestNewClientFromEnv(t *testing.T) {
	t.Run("missing key errors", func(t *testing.T) {
		t.Setenv("PROOFCHAIN_API_KEY", "")
		if _, err := NewClientFromEnv(); err == nil {
			t.Error("expected error when key is unset")
		}
	})
	t.Run("present key succeeds", func(t *testing.T) {
		t.Setenv("PROOFCHAIN_API_KEY", "env-key")
		c, err := NewClientFromEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.http.apiKey != "env-key" {
			t.Errorf("apiKey = %q, want env-key", c.http.apiKey)
		}
	})
}

func TestNewEndUserClientUsesTokenAuth(t *testing.T) {
	c, err := NewEndUserClient("jwt-token", "tenant-slug")
	if err != nil {
		t.Fatalf("NewEndUserClient: %v", err)
	}
	if c.http.apiKey != "" {
		t.Errorf("apiKey = %q, want empty for end-user client", c.http.apiKey)
	}
	if c.http.userToken != "jwt-token" {
		t.Errorf("userToken = %q, want jwt-token", c.http.userToken)
	}
	if c.http.tenantID != "tenant-slug" {
		t.Errorf("tenantID = %q, want tenant-slug", c.http.tenantID)
	}
}

func TestClientVerify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/verify/Qm123" {
			t.Errorf("path = %q, want /verify/Qm123", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"valid":true,"ipfs_hash":"Qm123","proof_verified":true,"message":"ok"}`))
	}))
	defer srv.Close()

	c := mustClient(t, "k", WithBaseURL(srv.URL))
	res, err := c.Verify(context.Background(), "Qm123")
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !res.Valid || res.IPFSHash != "Qm123" || !res.ProofVerified {
		t.Errorf("unexpected verification result: %+v", res)
	}
}
