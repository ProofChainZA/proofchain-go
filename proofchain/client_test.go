package proofchain

import (
	"context"
	"os"
	"testing"
)

func getTestClient(t *testing.T) *Client {
	apiKey := os.Getenv("PROOFCHAIN_API_KEY")
	if apiKey == "" {
		apiKey = "atst_d68b397e80587a87d5a5bd11160f400d9dfdd62e913315ec7b2b440a73609be0"
	}
	baseURL := os.Getenv("PROOFCHAIN_BASE_URL")
	if baseURL == "" {
		baseURL = "http://api.127.0.0.1.nip.io:8080"
	}
	return NewClient(apiKey, WithBaseURL(baseURL))
}

func TestTenantInfo(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	info, err := client.TenantInfo(ctx)
	if err != nil {
		t.Fatalf("TenantInfo failed: %v", err)
	}

	if info.Name == "" {
		t.Error("Expected tenant name to be non-empty")
	}
	t.Logf("Tenant: %s, Tier: %s", info.Name, info.Tier)
}

func TestUsage(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	usage, err := client.Usage(ctx, "month")
	if err != nil {
		t.Fatalf("Usage failed: %v", err)
	}

	t.Logf("Events this month: %d/%d", usage.EventsThisMonth, usage.MaxEventsPerMonth)
}

func TestEventsList(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	events, err := client.Events.List(ctx, &ListEventsRequest{Limit: 5})
	if err != nil {
		t.Fatalf("Events.List failed: %v", err)
	}

	t.Logf("Found %d events", len(events))
}

func TestCreateEvent(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	// Create a new test event
	event, err := client.Events.Create(ctx, &CreateEventRequest{
		UserID:    "sdk-test@acme.com",
		EventType: "go_sdk_test",
		Data: map[string]interface{}{
			"test_run":  "integration_test",
			"timestamp": "2025-12-17T12:26:00Z",
			"sdk":       "go",
		},
	})
	if err != nil {
		t.Fatalf("Events.Create failed: %v", err)
	}

	t.Logf("Created event: ID=%s, CertificateID=%s, IPFS=%s",
		event.ID, event.CertificateID, event.IPFSHash)
}

func TestChannelsList(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	channels, err := client.Channels.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("Channels.List failed: %v", err)
	}

	t.Logf("Found %d channels", len(channels))
}

func TestCertificatesList(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	certs, err := client.Certificates.List(ctx, &ListCertificatesRequest{Limit: 5})
	if err != nil {
		t.Fatalf("Certificates.List failed: %v", err)
	}

	t.Logf("Found %d certificates", len(certs))
}

func TestWebhooksList(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	webhooks, err := client.Webhooks.List(ctx)
	if err != nil {
		t.Fatalf("Webhooks.List failed: %v", err)
	}

	t.Logf("Found %d webhooks", len(webhooks))
}

func TestVaultList(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	vault, err := client.Vault.List(ctx, "")
	if err != nil {
		t.Fatalf("Vault.List failed: %v", err)
	}

	t.Logf("Vault: %d files, %d bytes", vault.TotalFiles, vault.TotalSize)
}

func TestSearchQuery(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	results, err := client.Search.Query(ctx, &SearchQueryRequest{Limit: 5})
	if err != nil {
		t.Fatalf("Search.Query failed: %v", err)
	}

	t.Logf("Search: %d total results", results.Total)
}

func TestSearchFacets(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	facets, err := client.Search.Facets(ctx, nil, nil)
	if err != nil {
		t.Fatalf("Search.Facets failed: %v", err)
	}

	t.Logf("Facets: %d event types, %d statuses", len(facets.EventTypes), len(facets.Statuses))
}

func TestVerifyCertificate(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	// Use a known certificate ID from the test tenant
	cert, err := client.VerifyResource.Certificate(ctx, "5282DC4D5342AA2E")
	if err != nil {
		t.Fatalf("Verify.Certificate failed: %v", err)
	}

	if !cert.IsValid() {
		t.Error("Expected certificate to be valid")
	}
	t.Logf("Certificate %s: %s", cert.CertificateID, cert.Status)
}

func TestTenantAPIKeys(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	keys, err := client.Tenant.ListAPIKeys(ctx)
	if err != nil {
		t.Fatalf("Tenant.ListAPIKeys failed: %v", err)
	}

	t.Logf("Found %d API keys", len(keys))
}

func TestTenantBlockchainStats(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	stats, err := client.Tenant.BlockchainStats(ctx)
	if err != nil {
		t.Fatalf("Tenant.BlockchainStats failed: %v", err)
	}

	t.Logf("Blockchain: %s, TXs: %d", stats.ChainName, stats.TotalTransactions)
}
