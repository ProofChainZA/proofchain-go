package proofchain

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// mustHTTPClient / mustClient fail the test on constructor error.
func mustHTTPClient(t *testing.T, apiKey string, opts ...HTTPClientOption) *HTTPClient {
	t.Helper()
	c, err := NewHTTPClient(apiKey, opts...)
	if err != nil {
		t.Fatalf("NewHTTPClient: %v", err)
	}
	return c
}

func mustClient(t *testing.T, apiKey string, opts ...HTTPClientOption) *Client {
	t.Helper()
	c, err := NewClient(apiKey, opts...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// captured records the last request a capture server received.
type captured struct {
	method  string
	path    string
	rawPath string // escaped path as sent on the wire
	rawQ    string
	header  http.Header
	body    []byte
}

// newCaptureServer starts a server that records the last request it received
// and replies with status/reply.
func newCaptureServer(t *testing.T, status int, reply string) (*httptest.Server, *captured) {
	t.Helper()
	rec := &captured{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.method = r.Method
		rec.path = r.URL.Path
		rec.rawPath = r.URL.EscapedPath()
		rec.rawQ = r.URL.RawQuery
		rec.header = r.Header.Clone()
		rec.body, _ = io.ReadAll(r.Body)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, reply)
	}))
	t.Cleanup(srv.Close)
	return srv, rec
}

// ---------------------------------------------------------------------------
// Scheme validation
// ---------------------------------------------------------------------------

func TestBaseURLSchemeValidation(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		opts    []HTTPClientOption
		wantErr bool
	}{
		{"https ok", "https://api.proofchain.co.za", nil, false},
		{"http loopback ok", "http://localhost:8080", nil, false},
		{"http 127.0.0.1 ok", "http://127.0.0.1:8080", nil, false},
		{"http ::1 ok", "http://[::1]:8080", nil, false},
		{"http remote rejected", "http://api.proofchain.co.za", nil, true},
		{"http nip.io rejected", "http://api.127.0.0.1.nip.io:8080", nil, true},
		{"http remote with opt-in", "http://api.internal:8080", []HTTPClientOption{WithInsecureHTTP()}, false},
		{"ftp rejected", "ftp://api.proofchain.co.za", nil, true},
		{"no scheme rejected", "api.proofchain.co.za", nil, true},
		{"empty rejected", "", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := append([]HTTPClientOption{WithBaseURL(tt.url)}, tt.opts...)
			_, err := NewClient("atst_x", opts...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewClient(%q) err = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestBaseURLFromEnvValidated(t *testing.T) {
	t.Setenv("PROOFCHAIN_API_KEY", "atst_x")
	t.Setenv("PROOFCHAIN_BASE_URL", "http://api.proofchain.co.za")
	if _, err := NewClientFromEnv(); err == nil {
		t.Fatal("expected error for plaintext PROOFCHAIN_BASE_URL")
	}
	t.Setenv("PROOFCHAIN_BASE_URL", "https://api.proofchain.co.za/")
	c, err := NewClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if c.http.baseURL != "https://api.proofchain.co.za" {
		t.Errorf("trailing slash not trimmed: %q", c.http.baseURL)
	}
}

func TestIngestURLSchemeValidation(t *testing.T) {
	if _, err := NewIngestionClient("atst_x", WithIngestURL("http://ingest.proofchain.co.za")); err == nil {
		t.Fatal("expected error for plaintext ingest URL")
	}
	if _, err := NewIngestionClient("atst_x", WithIngestURL("http://ingest.internal"), WithIngestInsecureHTTP()); err != nil {
		t.Fatalf("opt-in insecure should be accepted: %v", err)
	}
	if _, err := NewIngestionClient("atst_x", WithIngestURL("http://127.0.0.1:9000")); err != nil {
		t.Fatalf("loopback should be accepted: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Redirects are never followed
// ---------------------------------------------------------------------------

func TestRedirectNotFollowedWithCredentials(t *testing.T) {
	var leaked atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "" || r.Header.Get("Authorization") != "" {
			leaked.Add(1)
		}
		_, _ = w.Write([]byte(`{"id":"e1"}`))
	}))
	t.Cleanup(target.Close)

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+r.URL.Path, http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	client := mustClient(t, "atst_secret", WithBaseURL(origin.URL), WithRetries(0))
	_, err := client.Events.Get(context.Background(), "e1")
	if !errors.Is(err, errRedirectNotFollowed) {
		t.Fatalf("expected errRedirectNotFollowed, got %v", err)
	}
	if _, err := client.Vault.Download(context.Background(), "f1"); !errors.Is(err, errRedirectNotFollowed) {
		t.Fatalf("GetRaw: expected errRedirectNotFollowed, got %v", err)
	}

	ingest, err := NewIngestionClient("atst_secret", WithIngestURL(origin.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ingest.Ingest(context.Background(), &IngestEventRequest{UserID: "u", EventType: "e"}); !errors.Is(err, errRedirectNotFollowed) {
		t.Fatalf("ingest: expected errRedirectNotFollowed, got %v", err)
	}

	if n := leaked.Load(); n != 0 {
		t.Fatalf("credentials were forwarded to the redirect target %d time(s)", n)
	}
}

func TestCustomHTTPClientNeverFollowsRedirects(t *testing.T) {
	custom := &http.Client{} // default policy follows redirects
	c := mustHTTPClient(t, "atst_x", WithHTTPClient(custom))
	if c.httpClient.CheckRedirect == nil {
		t.Fatal("CheckRedirect not set on custom client")
	}
	if err := c.httpClient.CheckRedirect(nil, nil); !errors.Is(err, http.ErrUseLastResponse) {
		t.Fatalf("CheckRedirect = %v, want ErrUseLastResponse", err)
	}
	if custom.CheckRedirect != nil {
		t.Fatal("caller's http.Client must not be mutated")
	}
}

// ---------------------------------------------------------------------------
// Path escaping
// ---------------------------------------------------------------------------

func TestSegEscapesPathSegment(t *testing.T) {
	tests := map[string]string{
		"plain":        "plain",
		"a/b":          "a%2Fb",
		"../admin":     "..%2Fadmin",
		"x?y=1":        "x%3Fy=1",
		"x#frag":       "x%23frag",
		"user@x.com":   "user@x.com",
		"with space":   "with%20space",
		"ü":            "%C3%BC",
		"a%2Fb":        "a%252Fb",
		"a;b":          "a%3Bb",
		"semi;colon/1": "semi%3Bcolon%2F1",
	}
	for in, want := range tests {
		if got := seg(in); got != want {
			t.Errorf("seg(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCallerIdentifiersAreEscapedInPaths(t *testing.T) {
	srv, rec := newCaptureServer(t, http.StatusOK, `{}`)
	client := mustClient(t, "atst_key", WithBaseURL(srv.URL), WithRetries(0))
	ctx := context.Background()
	evil := "../../tenant/api-keys?x=1#f"
	wantSeg := seg(evil)

	cases := []struct {
		name string
		call func() error
		want string
	}{
		{"Events.Get", func() error { _, err := client.Events.Get(ctx, evil); return err }, "/tenant/events/" + wantSeg},
		{"Webhooks.Delete", func() error { return client.Webhooks.Delete(ctx, evil) }, "/webhooks/" + wantSeg},
		{"Users.GDPRDelete", func() error {
			_, err := client.Users.GDPRDelete(ctx, evil, &GDPRDeletionRequest{})
			return err
		}, "/end-users/" + wantSeg + "/gdpr"},
		{"Quests.DeleteStep", func() error { return client.Quests.DeleteStep(ctx, evil, evil) }, "/quests/" + wantSeg + "/steps/" + wantSeg},
		{"Quests.StartStep", func() error { _, err := client.Quests.StartStep(ctx, evil, evil, 2); return err }, "/quests/" + wantSeg + "/progress/" + wantSeg + "/step/2/start"},
		{"Wallets.GetInfo", func() error { _, err := client.Wallets.GetInfo(ctx, evil, nil); return err }, "/wallets/" + wantSeg + "/info"},
		{"Wallets.GetTokenByContract", func() error { _, err := client.Wallets.GetTokenByContract(ctx, evil, "a&b"); return err }, "/tokens/by-contract/" + wantSeg},
		{"Tenant.DeleteAPIKey", func() error { return client.Tenant.DeleteAPIKey(ctx, evil) }, "/tenant/api-keys/" + wantSeg},
		{"Vault.Delete", func() error { return client.Vault.Delete(ctx, evil) }, "/tenant/vault/files/" + wantSeg},
		{"Passports.SetFieldValue", func() error { return client.Passports.SetFieldValue(ctx, evil, evil, 1) }, "/passports/" + wantSeg + "/fields/" + wantSeg},
		{"Rewards.GetUserRewards", func() error { _, err := client.Rewards.GetUserRewards(ctx, evil, nil); return err }, "/rewards/users/" + wantSeg + "/rewards"},
		{"Search.ByUser", func() error { _, err := client.Search.ByUser(ctx, evil, 0, 0); return err }, "/search/by-user/" + wantSeg},
		{"PartnerKeys.RequestOTT", func() error { _, err := client.PartnerKeys.RequestOTT(ctx, evil); return err }, "/partner-keys/" + wantSeg + "/ott/request"},
		{"Credentials.GetType", func() error { _, err := client.Credentials.GetType(ctx, evil); return err }, "/credentials/types/" + wantSeg},
		{"Credentials.Revoke", func() error { _, err := client.Credentials.Revoke(ctx, evil, "r"); return err }, "/credentials/issued/" + wantSeg + "/revoke"},
		{"Rewards.ClaimReward", func() error { _, err := client.Rewards.ClaimReward(ctx, evil, "0xabc"); return err }, "/rewards/earned/" + wantSeg + "/claim"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = tc.call()
			if rec.rawPath != tc.want {
				t.Errorf("raw path = %q, want %q", rec.rawPath, tc.want)
			}
			if strings.Contains(rec.rawPath, "?") || strings.Contains(rec.rawPath, "#") {
				t.Errorf("identifier leaked query/fragment delimiters into path: %q", rec.rawPath)
			}
		})
	}
	// query values are encoded too
	_, _ = client.Wallets.GetTokenByContract(ctx, "c", "a&b=c")
	if rec.rawQ != "network=a%26b%3Dc" {
		t.Errorf("query not encoded: %q", rec.rawQ)
	}
}

// ---------------------------------------------------------------------------
// Retry rebuilds the request body
// ---------------------------------------------------------------------------

func TestRetryResendsRequestBody(t *testing.T) {
	var calls atomic.Int32
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 1024)
		n, _ := r.Body.Read(b)
		bodies = append(bodies, string(b[:n]))
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"detail":"slow down"}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"e1"}`))
	}))
	t.Cleanup(srv.Close)

	client := mustClient(t, "atst_key", WithBaseURL(srv.URL), WithRetries(1))
	if _, err := client.Events.Create(context.Background(), &CreateEventRequest{EventType: "login", UserID: "u1"}); err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if len(bodies) != 2 || bodies[0] == "" || bodies[0] != bodies[1] {
		t.Fatalf("retry did not resend the body: %q", bodies)
	}
}

// ---------------------------------------------------------------------------
// CreateAPIKey validation
// ---------------------------------------------------------------------------

func TestCreateAPIKeyScopeValidation(t *testing.T) {
	srv, rec := newCaptureServer(t, http.StatusOK, `{"id":"k1","key":"atst_secretsecretsecret","key_prefix":"atst_sec","scopes":["read"],"expires_at":"2026-01-01T00:00:00Z","allowed_cidrs":["10.0.0.0/8"],"notify_new_ip":false,"is_legacy_admin":false}`)
	client := mustClient(t, "atst_key", WithBaseURL(srv.URL))
	ctx := context.Background()

	bad := []*CreateAPIKeyRequest{
		{Name: "k"},
		{Name: "k", Scopes: []string{}},
		{Name: "k", Scopes: []string{"admin"}},
		{Name: "k", Scopes: []string{"read", "*"}},
		{Name: "k", Scopes: []string{"read", " ADMIN "}},
		{Name: "", Scopes: []string{"read"}},
	}
	for i, req := range bad {
		rec.path = ""
		_, err := client.Tenant.CreateAPIKey(ctx, req)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("case %d: expected *ValidationError, got %v", i, err)
		}
		if rec.path != "" {
			t.Errorf("case %d: invalid request must not reach the server", i)
		}
	}

	days := 30
	notify := false
	res, err := client.Tenant.CreateAPIKey(ctx, &CreateAPIKeyRequest{
		Name: "k", Scopes: []string{"read", "write"}, ExpiresInDays: &days,
		AllowedCIDRs: []string{"10.0.0.0/8"}, NotifyNewIP: &notify,
	})
	if err != nil {
		t.Fatal(err)
	}
	var sent map[string]interface{}
	if err := json.Unmarshal(rec.body, &sent); err != nil {
		t.Fatal(err)
	}
	if sent["expires_in_days"] != float64(30) || sent["notify_new_ip"] != false || sent["scopes"] == nil || sent["allowed_cidrs"] == nil {
		t.Errorf("request payload = %v", sent)
	}
	if res.Key != "atst_secretsecretsecret" || res.ExpiresAt == nil || len(res.AllowedCIDRs) != 1 || res.NotifyNewIP || res.IsLegacyAdmin {
		t.Errorf("response not decoded: %+v", res)
	}
	for _, s := range []string{fmt.Sprint(*res), fmt.Sprintf("%v", res), fmt.Sprintf("%#v", *res), fmt.Sprintf("%s", res)} {
		if strings.Contains(s, "secretsecret") {
			t.Errorf("secret key leaked through formatting: %s", s)
		}
		if !strings.Contains(s, "atst_sec…") {
			t.Errorf("expected redacted prefix in %s", s)
		}
	}
}

// ---------------------------------------------------------------------------
// Debug output redaction
// ---------------------------------------------------------------------------

func TestClientsRedactAPIKeyInFormatting(t *testing.T) {
	const key = "atst_0123456789abcdef_SECRET"
	h := mustHTTPClient(t, key)
	g, err := NewGRPCClient(key)
	if err != nil {
		t.Fatal(err)
	}
	in, err := NewIngestionClient(key)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []interface{}{h, g, in} {
		for _, s := range []string{fmt.Sprintf("%v", v), fmt.Sprintf("%s", v), fmt.Sprintf("%#v", v), fmt.Sprintf("%+v", v)} {
			if strings.Contains(s, "SECRET") {
				t.Errorf("%T leaked key: %s", v, s)
			}
			if !strings.Contains(s, "atst_012…") {
				t.Errorf("%T missing redacted prefix: %s", v, s)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Webhook signature
// ---------------------------------------------------------------------------

func signWebhook(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookSignature(t *testing.T) {
	secret := "whsec_test"
	payload := []byte(`{"event_type":"document.attested","id":"e1"}`)
	good := signWebhook(secret, payload)

	if err := VerifyWebhookSignature(secret, payload, good); err != nil {
		t.Fatalf("good signature rejected: %v", err)
	}
	if err := VerifyWebhookSignature(secret, payload, "  "+good+"\n"); err != nil {
		t.Fatalf("whitespace-padded header rejected: %v", err)
	}
	bad := []string{
		"",
		"sha256=",
		"sha256=zz",
		strings.TrimPrefix(good, "sha256="),
		"sha1=" + strings.TrimPrefix(good, "sha256="),
		signWebhook("wrong-secret", payload),
		signWebhook(secret, []byte(`{"event_type":"document.attested","id":"e2"}`)),
	}
	for _, h := range bad {
		if err := VerifyWebhookSignature(secret, payload, h); !errors.Is(err, ErrInvalidWebhookSignature) {
			t.Errorf("header %q: expected ErrInvalidWebhookSignature, got %v", h, err)
		}
	}
	if err := VerifyWebhookSignature("", payload, good); err == nil {
		t.Error("empty secret must be rejected")
	}
}

// ---------------------------------------------------------------------------
// gRPC plaintext only for loopback
// ---------------------------------------------------------------------------

func TestGRPCInsecurePlaintextOnlyLoopback(t *testing.T) {
	ok := []string{"localhost:50051", "127.0.0.1:50051", "[::1]:50051", "localhost"}
	for _, ep := range ok {
		c, err := NewGRPCClient("atst_x", WithGRPCEndpoint(ep), WithInsecurePlaintext())
		if err != nil {
			t.Errorf("%s: expected loopback plaintext to be allowed: %v", ep, err)
			continue
		}
		if !c.plaintext {
			t.Errorf("%s: plaintext not set", ep)
		}
	}
	notOK := []string{"grpc.proofchain.co.za:50051", "grpc.proofchain.co.za:443", "10.0.0.5:50051", "grpc.127.0.0.1.nip.io:50051"}
	for _, ep := range notOK {
		if _, err := NewGRPCClient("atst_x", WithGRPCEndpoint(ep), WithInsecurePlaintext()); err == nil {
			t.Errorf("%s: plaintext must be rejected for non-loopback host", ep)
		}
	}
	// TLS is the default regardless of port.
	c, err := NewGRPCClient("atst_x", WithGRPCEndpoint("grpc.proofchain.co.za:50051"))
	if err != nil {
		t.Fatal(err)
	}
	if c.plaintext {
		t.Fatal("non-443 port must still use TLS")
	}
	if _, err := NewMultiStreamClient("atst_x", WithGRPCEndpoint("10.0.0.5:50051"), WithInsecurePlaintext()); err == nil {
		t.Error("NewMultiStreamClient must propagate the plaintext rejection")
	}
}
