package proofchain

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// fanScoreEnvelope is a representative tenant-audience envelope. It carries
// the blocks the SDK types (identity, score, cohorts, points), a block the
// profile enabled but the fan has no data for (rank: null) and three blocks
// the SDK does not model (season, form, tier).
const fanScoreEnvelope = `{
  "profile": {"slug":"vip","version":3,"audience":"tenant","generated_at":"2026-09-11T08:30:00+00:00","cache_ttl_s":60},
  "identity": {"fan_id":"5c1f9c2e-0a1d-4f3b-9a77-7b1f0c4d8e11","external_id":"acme-42","display_name":"Thandi",
               "avatar_url":null,"country":"ZA","city":"Cape Town","language":"en","email":"thandi@example.com",
               "wallet_address":"0xAbC0000000000000000000000000000000000001","segments":["season-ticket"],
               "tags":{"tier":"gold","vip":true},"attributes":{"seat":"A12"},"status":"active"},
  "score": {"composite":1284.5,"fanpass":871,"percentile":92.4,"event_count":318,
            "updated_at":"2026-09-11T07:59:12+00:00","read_track":"eventscore"},
  "cohorts": {"cohorts":[{"id":"9b2a1f4c-3d5e-4a6b-8c7d-0e1f2a3b4c5d","name":"Matchday","score":902.25,
                          "percentile":95.1,"share_pct":70.2,"event_count":210,
                          "first_event_at":"2025-08-02T18:00:00+00:00","last_event_at":"2026-09-10T20:15:00+00:00",
                          "form_index":1.34,"contributes_to_fanscore":true,"scored":true},
                         {"id":"1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5e","name":"Merch","score":0.0,
                          "percentile":null,"share_pct":null,"event_count":0,
                          "first_event_at":null,"last_event_at":null,
                          "form_index":null,"contributes_to_fanscore":false,"scored":false}],
              "top_cohort":"Matchday"},
  "rank": null,
  "points": {"points_balance":4200,"lifetime_points":18350},
  "season": {"active_season":{"id":"7f8e9d0c-1b2a-4c3d-9e8f-7a6b5c4d3e2f","name":"2026/27",
                              "starts_at":"2026-07-01T00:00:00+00:00","ends_at":null},
             "season_score":214.5,"cohort_season_scores":[]},
  "form": {"form_index":1.12,"window_days":30,"recent_event_count":27},
  "tier": {"tiers":[{"set":"global","name":"Platinum","position":2}]}
}`

// newFanScoreClient starts a test server and returns a profiles client aimed
// at it. apiKey "" plus no options means an unauthenticated caller.
func newFanScoreClient(t *testing.T, apiKey string, handler http.HandlerFunc, opts ...HTTPClientOption) *FanScoreProfilesClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewFanScoreProfilesClient(NewHTTPClient(apiKey, append(opts, WithBaseURL(srv.URL))...))
}

func TestFanScoreProfilesGetProfile(t *testing.T) {
	tests := []struct {
		name        string
		profile     string
		fanRef      string
		wantPath    string
		wantEscaped string
	}{
		{
			name:        "empty profile falls back to default",
			fanRef:      "5c1f9c2e-0a1d-4f3b-9a77-7b1f0c4d8e11",
			wantPath:    "/fanscore/profiles/default/fans/5c1f9c2e-0a1d-4f3b-9a77-7b1f0c4d8e11",
			wantEscaped: "/fanscore/profiles/default/fans/5c1f9c2e-0a1d-4f3b-9a77-7b1f0c4d8e11",
		},
		{
			name:        "wallet ref keeps its colon",
			profile:     "vip",
			fanRef:      "wallet:0xAbC0000000000000000000000000000000000001",
			wantPath:    "/fanscore/profiles/vip/fans/wallet:0xAbC0000000000000000000000000000000000001",
			wantEscaped: "/fanscore/profiles/vip/fans/wallet:0xAbC0000000000000000000000000000000000001",
		},
		{
			name:        "email ref keeps its colon and at-sign",
			fanRef:      "email:fan+tag@example.com",
			wantPath:    "/fanscore/profiles/default/fans/email:fan+tag@example.com",
			wantEscaped: "/fanscore/profiles/default/fans/email:fan+tag@example.com",
		},
		{
			name:        "external id with a slash and a space is escaped",
			fanRef:      "acme/fan 42",
			wantPath:    "/fanscore/profiles/default/fans/acme/fan 42",
			wantEscaped: "/fanscore/profiles/default/fans/acme%2Ffan%2042",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath, gotEscaped, gotQuery, gotAPIKey string
			client := newFanScoreClient(t, "tenant-key", func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotEscaped = r.URL.Path, r.URL.EscapedPath()
				gotQuery, gotAPIKey = r.URL.RawQuery, r.Header.Get("X-API-Key")
				w.Header().Set("X-FanScore-Profile", "vip;v=3")
				_, _ = w.Write([]byte(fanScoreEnvelope))
			})

			profile, err := client.GetProfile(context.Background(), tt.fanRef, tt.profile)
			if err != nil {
				t.Fatalf("GetProfile failed: %v", err)
			}
			if gotPath != tt.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tt.wantPath)
			}
			if gotEscaped != tt.wantEscaped {
				t.Errorf("escaped path = %q, want %q", gotEscaped, tt.wantEscaped)
			}
			if gotQuery != "" {
				t.Errorf("query = %q, want empty", gotQuery)
			}
			if gotAPIKey != "tenant-key" {
				t.Errorf("X-API-Key = %q, want %q", gotAPIKey, "tenant-key")
			}
			if profile.Profile.Slug != "vip" {
				t.Errorf("profile.slug = %q, want %q", profile.Profile.Slug, "vip")
			}
		})
	}
}

// fanScoreBatchReply pairs a resolved reference with one that resolved to no
// fan: null in profiles and listed in not_found.
const fanScoreMissingRef = "wallet:0xAbC0000000000000000000000000000000000009"
const fanScoreBatchReply = `{
  "profiles": {
    "acme-42": ` + fanScoreEnvelope + `,
    "` + fanScoreMissingRef + `": null
  },
  "not_found": ["` + fanScoreMissingRef + `"]
}`

func TestFanScoreProfilesGetProfiles(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		wantPath string
	}{
		{name: "empty profile falls back to default", wantPath: "/fanscore/profiles/default/fans"},
		{name: "named profile", profile: "vip", wantPath: "/fanscore/profiles/vip/fans"},
		{name: "profile name with a space is escaped", profile: "match day", wantPath: "/fanscore/profiles/match%20day/fans"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotEscaped, gotQuery, gotAPIKey, gotContentType string
			var gotBody struct {
				FanRefs []string `json:"fan_refs"`
			}
			client := newFanScoreClient(t, "tenant-key", func(w http.ResponseWriter, r *http.Request) {
				gotMethod, gotEscaped = r.Method, r.URL.EscapedPath()
				gotQuery, gotAPIKey = r.URL.RawQuery, r.Header.Get("X-API-Key")
				gotContentType = r.Header.Get("Content-Type")
				if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
					t.Errorf("decoding the request body failed: %v", err)
				}
				_, _ = w.Write([]byte(fanScoreBatchReply))
			})

			refs := []string{"acme-42", fanScoreMissingRef}
			batch, err := client.GetProfiles(context.Background(), refs, tt.profile)
			if err != nil {
				t.Fatalf("GetProfiles failed: %v", err)
			}
			if gotMethod != http.MethodPost {
				t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
			}
			if gotEscaped != tt.wantPath {
				t.Errorf("escaped path = %q, want %q", gotEscaped, tt.wantPath)
			}
			if gotQuery != "" {
				t.Errorf("query = %q, want empty", gotQuery)
			}
			if gotAPIKey != "tenant-key" {
				t.Errorf("X-API-Key = %q, want %q", gotAPIKey, "tenant-key")
			}
			if gotContentType != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", gotContentType)
			}
			if len(gotBody.FanRefs) != 2 || gotBody.FanRefs[0] != "acme-42" || gotBody.FanRefs[1] != fanScoreMissingRef {
				t.Errorf("body fan_refs = %v, want %v sent verbatim and in order", gotBody.FanRefs, refs)
			}

			if len(batch.Profiles) != 2 {
				t.Fatalf("profiles = %v, want one entry per reference sent", batch.Profiles)
			}
			found := batch.Profiles["acme-42"]
			if found == nil {
				t.Fatal("profiles[acme-42] is nil, want the decoded envelope")
			}
			if found.Profile.Slug != "vip" || found.Score == nil || found.Score.Composite == nil || *found.Score.Composite != 1284.5 {
				t.Errorf("profiles[acme-42] = %+v, want the vip envelope scoring 1284.5", found)
			}
			if found.Rank != nil {
				t.Errorf("profiles[acme-42].rank = %v, want nil for a null block", found.Rank)
			}
			missing, ok := batch.Profiles[fanScoreMissingRef]
			if !ok {
				t.Fatalf("profiles is missing the %q key; a null entry must still be keyed", fanScoreMissingRef)
			}
			if missing != nil {
				t.Errorf("profiles[%q] = %+v, want nil for a reference that resolved to no fan", fanScoreMissingRef, missing)
			}
			if len(batch.NotFound) != 1 || batch.NotFound[0] != fanScoreMissingRef {
				t.Errorf("not_found = %v, want [%q]", batch.NotFound, fanScoreMissingRef)
			}
		})
	}
}

func TestFanScoreProfilesGetProfilesRejectsOversizeBatch(t *testing.T) {
	called := false
	client := newFanScoreClient(t, "tenant-key", func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, _ = w.Write([]byte(fanScoreBatchReply))
	})

	refs := make([]string, MaxFanScoreProfileBatch+1)
	for i := range refs {
		refs[i] = "acme-" + strconv.Itoa(i)
	}
	batch, err := client.GetProfiles(context.Background(), refs, "vip")
	if batch != nil {
		t.Errorf("batch = %v, want nil when the cap is exceeded", batch)
	}
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("err = %v (%T), want a *ValidationError", err, err)
	}
	if called {
		t.Error("an oversize batch reached the server; it must be refused before sending")
	}

	refs = refs[:MaxFanScoreProfileBatch]
	if _, err := client.GetProfiles(context.Background(), refs, "vip"); err != nil {
		t.Fatalf("a batch exactly at the cap failed: %v", err)
	}
	if !called {
		t.Error("a batch exactly at the cap was not sent")
	}
}

func TestFanScoreProfilesGetMyProfile(t *testing.T) {
	tests := []struct {
		name      string
		profile   string
		wantQuery string
	}{
		{name: "no profile sends no query", wantQuery: ""},
		{name: "named profile", profile: "vip", wantQuery: "profile=vip"},
		{name: "profile name is query-escaped", profile: "match day", wantQuery: "profile=match+day"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath, gotQuery, gotAuth, gotTenant, gotAPIKey string
			client := newFanScoreClient(t, "", func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
				gotAuth = r.Header.Get("Authorization")
				gotTenant = r.Header.Get("X-Tenant-ID")
				gotAPIKey = r.Header.Get("X-API-Key")
				_, _ = w.Write([]byte(fanScoreEnvelope))
			}, WithUserToken("jwt-abc", "tenant-1"))

			if _, err := client.GetMyProfile(context.Background(), tt.profile); err != nil {
				t.Fatalf("GetMyProfile failed: %v", err)
			}
			if gotPath != "/fanscore/me" {
				t.Errorf("path = %q, want %q", gotPath, "/fanscore/me")
			}
			if gotQuery != tt.wantQuery {
				t.Errorf("query = %q, want %q", gotQuery, tt.wantQuery)
			}
			if gotAuth != "Bearer jwt-abc" {
				t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer jwt-abc")
			}
			if gotTenant != "tenant-1" {
				t.Errorf("X-Tenant-ID = %q, want %q", gotTenant, "tenant-1")
			}
			if gotAPIKey != "" {
				t.Errorf("X-API-Key = %q, want it unset for an end-user token", gotAPIKey)
			}
		})
	}
}

func TestFanScoreProfilesGetPublicProfile(t *testing.T) {
	var gotPath, gotEscaped, gotQuery, gotAPIKey, gotAuth string
	client := newFanScoreClient(t, "", func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotEscaped = r.URL.Path, r.URL.EscapedPath()
		gotQuery = r.URL.RawQuery
		gotAPIKey, gotAuth = r.Header.Get("X-API-Key"), r.Header.Get("Authorization")
		_, _ = w.Write([]byte(fanScoreEnvelope))
	})

	const fanRef = "wallet:0xAbC0000000000000000000000000000000000001"
	if _, err := client.GetPublicProfile(context.Background(), "showcase", fanRef); err != nil {
		t.Fatalf("GetPublicProfile failed: %v", err)
	}
	wantPath := "/fanscore/public/showcase/fans/" + fanRef
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}
	if gotEscaped != wantPath {
		t.Errorf("escaped path = %q, want %q", gotEscaped, wantPath)
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty", gotQuery)
	}
	if gotAPIKey != "" || gotAuth != "" {
		t.Errorf("public read sent credentials: X-API-Key=%q Authorization=%q", gotAPIKey, gotAuth)
	}
}

func TestFanScoreProfilesListProfiles(t *testing.T) {
	const catalog = `{"profiles":[
      {"id":"a0b1c2d3-4e5f-4a6b-8c9d-0e1f2a3b4c5d","slug":"vip","display_name":"VIP","description":"Box holders",
       "audience":"tenant","cache_ttl_s":30,"is_default":true,"version":3,"builtin":false,
       "blocks":[{"key":"score","enabled":true,"fields":["composite","fanpass"],"labels":{"fanpass":"FanPass #"},
                  "visibility":{"composite":"partner"},"options":{"percentile_method":"exact"}}],
       "output_schema":{"type":"object"},"created_at":"2026-08-01T10:00:00+00:00","updated_at":"2026-09-01T10:00:00+00:00"},
      {"id":null,"slug":"public-card","display_name":"Public card","description":null,
       "audience":"public","cache_ttl_s":60,"is_default":false,"version":1,"builtin":true,
       "blocks":[{"key":"identity","enabled":true,"fields":[],"labels":{},"visibility":{},"options":{}}],
       "output_schema":null,"created_at":null,"updated_at":null}]}`

	var gotPath, gotQuery string
	client := newFanScoreClient(t, "tenant-key", func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		_, _ = w.Write([]byte(catalog))
	})

	profiles, err := client.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}
	if gotPath != "/fanscore/profiles" {
		t.Errorf("path = %q, want %q", gotPath, "/fanscore/profiles")
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty", gotQuery)
	}
	if len(profiles) != 2 {
		t.Fatalf("got %d profiles, want 2", len(profiles))
	}

	own := profiles[0]
	if own.ID == nil || *own.ID != "a0b1c2d3-4e5f-4a6b-8c9d-0e1f2a3b4c5d" {
		t.Errorf("own.ID = %v, want the catalog uuid", own.ID)
	}
	if own.Slug != "vip" || own.DisplayName != "VIP" || own.Audience != "tenant" {
		t.Errorf("own identity = %q/%q/%q, want vip/VIP/tenant", own.Slug, own.DisplayName, own.Audience)
	}
	if own.CacheTTLs != 30 || own.Version != 3 || !own.IsDefault || own.Builtin {
		t.Errorf("own flags = ttl %d, v%d, default %v, builtin %v; want 30, v3, true, false",
			own.CacheTTLs, own.Version, own.IsDefault, own.Builtin)
	}
	if len(own.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(own.Blocks))
	}
	block := own.Blocks[0]
	if block.Key != "score" || !block.Enabled {
		t.Errorf("block = %q enabled=%v, want score enabled", block.Key, block.Enabled)
	}
	if len(block.Fields) != 2 || block.Fields[0] != "composite" {
		t.Errorf("block.Fields = %v, want [composite fanpass]", block.Fields)
	}
	if block.Labels["fanpass"] != "FanPass #" {
		t.Errorf("block.Labels[fanpass] = %q, want %q", block.Labels["fanpass"], "FanPass #")
	}
	if block.Visibility["composite"] != "partner" {
		t.Errorf("block.Visibility[composite] = %q, want %q", block.Visibility["composite"], "partner")
	}
	if block.Options["percentile_method"] != "exact" {
		t.Errorf("block.Options[percentile_method] = %v, want exact", block.Options["percentile_method"])
	}
	if own.OutputSchema["type"] != "object" {
		t.Errorf("own.OutputSchema = %v, want type object", own.OutputSchema)
	}
	if own.CreatedAt == nil || own.UpdatedAt == nil {
		t.Errorf("own timestamps = %v/%v, want both set", own.CreatedAt, own.UpdatedAt)
	}

	builtin := profiles[1]
	if !builtin.Builtin {
		t.Error("builtin.Builtin = false, want true")
	}
	if builtin.ID != nil || builtin.Description != nil || builtin.CreatedAt != nil {
		t.Errorf("builtin nullables = %v/%v/%v, want all nil", builtin.ID, builtin.Description, builtin.CreatedAt)
	}
}

func TestFanScoreProfilesGetProfileSchema(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		wantPath string
	}{
		{name: "named profile", profile: "vip", wantPath: "/fanscore/profiles/vip/schema"},
		{name: "empty profile falls back to default", wantPath: "/fanscore/profiles/default/schema"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath, gotQuery string
			client := newFanScoreClient(t, "tenant-key", func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
				_, _ = w.Write([]byte(`{"type":"object","required":["profile"],
                    "properties":{"profile":{"type":"object"},"score":{"type":["object","null"]}}}`))
			})

			schema, err := client.GetProfileSchema(context.Background(), tt.profile)
			if err != nil {
				t.Fatalf("GetProfileSchema failed: %v", err)
			}
			if gotPath != tt.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tt.wantPath)
			}
			if gotQuery != "" {
				t.Errorf("query = %q, want empty", gotQuery)
			}
			if schema["type"] != "object" {
				t.Errorf("schema[type] = %v, want object", schema["type"])
			}
			props, ok := schema["properties"].(map[string]interface{})
			if !ok {
				t.Fatalf("schema[properties] = %T, want an object", schema["properties"])
			}
			if _, ok := props["score"]; !ok {
				t.Errorf("schema properties = %v, want a score block", props)
			}
		})
	}
}

func TestFanScoreProfileEnvelopeDecodesUnknownBlocks(t *testing.T) {
	client := newFanScoreClient(t, "tenant-key", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(fanScoreEnvelope))
	})

	profile, err := client.GetProfile(context.Background(), "acme-42", "vip")
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	if profile.Profile.Version != 3 || profile.Profile.Audience != "tenant" || profile.Profile.CacheTTLs != 60 {
		t.Errorf("header = v%d/%s/ttl %d, want v3/tenant/ttl 60",
			profile.Profile.Version, profile.Profile.Audience, profile.Profile.CacheTTLs)
	}
	if profile.Profile.GeneratedAt != "2026-09-11T08:30:00+00:00" {
		t.Errorf("generated_at = %q, want the envelope value", profile.Profile.GeneratedAt)
	}

	if profile.Identity == nil {
		t.Fatal("identity block is nil, want decoded")
	}
	if profile.Identity.City == nil || *profile.Identity.City != "Cape Town" {
		t.Errorf("identity.city = %v, want Cape Town", profile.Identity.City)
	}
	if profile.Identity.Status == nil || *profile.Identity.Status != "active" {
		t.Errorf("identity.status = %v, want active", profile.Identity.Status)
	}
	if profile.Identity.Tags["tier"] != "gold" {
		t.Errorf("identity.tags[tier] = %v, want gold", profile.Identity.Tags["tier"])
	}
	if profile.Identity.Attributes["seat"] != "A12" {
		t.Errorf("identity.attributes[seat] = %v, want A12", profile.Identity.Attributes["seat"])
	}
	if profile.Identity.AvatarURL != nil {
		t.Errorf("identity.avatar_url = %v, want nil for an unknown value", profile.Identity.AvatarURL)
	}

	if profile.Score == nil || profile.Score.FanPass == nil || *profile.Score.FanPass != 871 {
		t.Fatalf("score.fanpass = %v, want 871", profile.Score)
	}
	if profile.Cohorts == nil || len(profile.Cohorts.Cohorts) != 2 {
		t.Fatalf("cohorts = %v, want 2 entries", profile.Cohorts)
	}
	scored, placeholder := profile.Cohorts.Cohorts[0], profile.Cohorts.Cohorts[1]
	if scored.FormIndex == nil || *scored.FormIndex != 1.34 {
		t.Errorf("cohorts[0].form_index = %v, want 1.34", scored.FormIndex)
	}
	if placeholder.FormIndex != nil || placeholder.Percentile != nil || placeholder.Scored {
		t.Errorf("cohorts[1] = %+v, want an unscored placeholder with null metrics", placeholder)
	}
	if profile.Points == nil || profile.Points.PointsBalance == nil || *profile.Points.PointsBalance != 4200 {
		t.Errorf("points = %v, want a 4200 balance", profile.Points)
	}
	if profile.Rank != nil {
		t.Errorf("rank = %v, want nil for a null block", profile.Rank)
	}

	for _, key := range []string{"season", "form", "tier"} {
		if _, ok := profile.Extra[key]; !ok {
			t.Errorf("Extra is missing the %q block; have %v", key, extraKeys(profile.Extra))
		}
	}
	if len(profile.Extra) != 3 {
		t.Errorf("Extra = %v, want exactly season, form and tier", extraKeys(profile.Extra))
	}
	var form struct {
		FormIndex        *float64 `json:"form_index"`
		WindowDays       int      `json:"window_days"`
		RecentEventCount int      `json:"recent_event_count"`
	}
	if err := json.Unmarshal(profile.Extra["form"], &form); err != nil {
		t.Fatalf("decoding the raw form block failed: %v", err)
	}
	if form.WindowDays != 30 || form.RecentEventCount != 27 {
		t.Errorf("form block = %+v, want window 30 and 27 recent events", form)
	}
}

func TestFanScoreProfileReencodeKeepsUnknownBlocks(t *testing.T) {
	var profile FanScoreProfile
	if err := json.Unmarshal([]byte(fanScoreEnvelope), &profile); err != nil {
		t.Fatalf("decoding the envelope failed: %v", err)
	}

	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("re-encoding the envelope failed: %v", err)
	}
	var round map[string]json.RawMessage
	if err := json.Unmarshal(data, &round); err != nil {
		t.Fatalf("decoding the re-encoded envelope failed: %v", err)
	}
	for _, key := range []string{"profile", "identity", "score", "cohorts", "points", "season", "form", "tier"} {
		if _, ok := round[key]; !ok {
			t.Errorf("re-encoded envelope dropped the %q block", key)
		}
	}

	var again FanScoreProfile
	if err := json.Unmarshal(data, &again); err != nil {
		t.Fatalf("re-decoding the envelope failed: %v", err)
	}
	if again.Profile != profile.Profile {
		t.Errorf("header round-trip = %+v, want %+v", again.Profile, profile.Profile)
	}
	if len(again.Extra) != len(profile.Extra) {
		t.Errorf("Extra round-trip = %v, want %v", extraKeys(again.Extra), extraKeys(profile.Extra))
	}
}

func extraKeys(extra map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(extra))
	for key := range extra {
		keys = append(keys, key)
	}
	return keys
}
