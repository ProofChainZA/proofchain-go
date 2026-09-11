package proofchain

import (
	"context"
	"encoding/json"
	"net/url"
)

// FanScoreProfileMeta is the envelope header present on every profile response.
type FanScoreProfileMeta struct {
	Slug        string `json:"slug"`
	Version     int    `json:"version"`
	Audience    string `json:"audience"` // public | partner | self | tenant
	GeneratedAt string `json:"generated_at"`
	CacheTTLs   int    `json:"cache_ttl_s"`
}

// fanScoreTypedBlocks are the envelope keys FanScoreProfile decodes into a
// typed field. Every other key is kept verbatim in FanScoreProfile.Extra.
var fanScoreTypedBlocks = [...]string{"profile", "identity", "score", "cohorts", "rank", "points"}

// FanScoreProfile is a fan's FanScore object as shaped by a tenant profile.
//
// A block is nil when the profile does not expose it to the caller's audience
// or the fan has no data for it. Fields inside a block are optional because
// the tenant chooses which to expose, and a null field means "unknown" rather
// than zero. Blocks this SDK does not model (season, form, activity, tier,
// proof, rewards, quests, league, wallet, and any added later) stay available
// as raw JSON in Extra.
type FanScoreProfile struct {
	Profile  FanScoreProfileMeta `json:"profile"`
	Identity *FanScoreIdentity   `json:"identity,omitempty"`
	Score    *FanScoreScore      `json:"score,omitempty"`
	Cohorts  *FanScoreCohorts    `json:"cohorts,omitempty"`
	Rank     *FanScoreRank       `json:"rank,omitempty"`
	Points   *FanScorePoints     `json:"points,omitempty"`

	// Extra holds the envelope blocks without a typed field above, keyed by
	// block name. Nil when the profile enabled none of them.
	Extra map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON fills the typed blocks and keeps every other block in Extra,
// so an envelope carrying a block this SDK predates still decodes cleanly.
func (p *FanScoreProfile) UnmarshalJSON(data []byte) error {
	type plain FanScoreProfile
	var typed plain
	if err := json.Unmarshal(data, &typed); err != nil {
		return err
	}
	var extra map[string]json.RawMessage
	if err := json.Unmarshal(data, &extra); err != nil {
		return err
	}
	for _, key := range fanScoreTypedBlocks {
		delete(extra, key)
	}
	if len(extra) == 0 {
		extra = nil
	}
	typed.Extra = extra
	*p = FanScoreProfile(typed)
	return nil
}

// MarshalJSON writes the typed blocks back out alongside the blocks held in
// Extra, so re-encoding a decoded envelope does not drop them.
func (p FanScoreProfile) MarshalJSON() ([]byte, error) {
	type plain FanScoreProfile
	data, err := json.Marshal(plain(p))
	if err != nil {
		return nil, err
	}
	if len(p.Extra) == 0 {
		return data, nil
	}
	var merged map[string]json.RawMessage
	if err := json.Unmarshal(data, &merged); err != nil {
		return nil, err
	}
	for key, value := range p.Extra {
		if _, typed := merged[key]; !typed {
			merged[key] = value
		}
	}
	return json.Marshal(merged)
}

type FanScoreIdentity struct {
	FanID         *string                `json:"fan_id,omitempty"`
	ExternalID    *string                `json:"external_id,omitempty"`
	DisplayName   *string                `json:"display_name,omitempty"`
	AvatarURL     *string                `json:"avatar_url,omitempty"`
	Country       *string                `json:"country,omitempty"`
	City          *string                `json:"city,omitempty"`
	Language      *string                `json:"language,omitempty"`
	Email         *string                `json:"email,omitempty"`
	WalletAddress *string                `json:"wallet_address,omitempty"`
	Segments      []string               `json:"segments,omitempty"`
	Tags          map[string]interface{} `json:"tags,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
	Status        *string                `json:"status,omitempty"`
}

type FanScoreScore struct {
	Composite  *float64 `json:"composite,omitempty"`
	FanPass    *int     `json:"fanpass,omitempty"`
	Percentile *float64 `json:"percentile,omitempty"`
	EventCount *int     `json:"event_count,omitempty"`
	UpdatedAt  *string  `json:"updated_at,omitempty"`
	ReadTrack  *string  `json:"read_track,omitempty"`
}

type FanScoreCohortEntry struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Score                 float64  `json:"score"`
	Percentile            *float64 `json:"percentile"`
	SharePct              *float64 `json:"share_pct"`
	EventCount            int      `json:"event_count"`
	FirstEventAt          *string  `json:"first_event_at"`
	LastEventAt           *string  `json:"last_event_at"`
	FormIndex             *float64 `json:"form_index"`
	ContributesToFanscore bool     `json:"contributes_to_fanscore"`
	Scored                bool     `json:"scored"`
}

type FanScoreCohorts struct {
	Cohorts   []FanScoreCohortEntry `json:"cohorts,omitempty"`
	TopCohort *string               `json:"top_cohort,omitempty"`
}

type FanScoreRank struct {
	GlobalPosition *int `json:"global_position,omitempty"`
	TotalFans      *int `json:"total_fans,omitempty"`
}

type FanScorePoints struct {
	PointsBalance  *int `json:"points_balance,omitempty"`
	LifetimePoints *int `json:"lifetime_points,omitempty"`
}

// FanScoreBlockConfig is one block's slot in a profile definition: which
// fields it emits, how they are labelled and who may see each one.
type FanScoreBlockConfig struct {
	Key        string                 `json:"key"`
	Enabled    bool                   `json:"enabled"`
	Fields     []string               `json:"fields,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty"`
	Visibility map[string]string      `json:"visibility,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// FanScoreProfileDefinition is a profile as listed by GET /fanscore/profiles.
// Built-ins have no ID and no timestamps.
type FanScoreProfileDefinition struct {
	ID           *string                `json:"id"`
	Slug         string                 `json:"slug"`
	DisplayName  string                 `json:"display_name"`
	Description  *string                `json:"description"`
	Audience     string                 `json:"audience"`
	CacheTTLs    int                    `json:"cache_ttl_s"`
	IsDefault    bool                   `json:"is_default"`
	Version      int                    `json:"version"`
	Builtin      bool                   `json:"builtin"`
	Blocks       []FanScoreBlockConfig  `json:"blocks"`
	OutputSchema map[string]interface{} `json:"output_schema,omitempty"`
	CreatedAt    *string                `json:"created_at,omitempty"`
	UpdatedAt    *string                `json:"updated_at,omitempty"`
}

// FanScoreProfilesClient reads FanScore Profiles — the tenant-configurable
// fan object. A fan reference is a fan_id UUID, the tenant's external_id,
// "wallet:0x…" or "email:…" (email is never accepted on the public endpoint).
type FanScoreProfilesClient struct {
	http *HTTPClient
}

func NewFanScoreProfilesClient(http *HTTPClient) *FanScoreProfilesClient {
	return &FanScoreProfilesClient{http: http}
}

// GetProfile returns a fan's FanScore object. profile "" means the tenant's
// default. A tenant API key reads at audience tenant; an X-Partner-Key at
// audience partner.
func (f *FanScoreProfilesClient) GetProfile(ctx context.Context, fanRef, profile string) (*FanScoreProfile, error) {
	if profile == "" {
		profile = "default"
	}
	var out FanScoreProfile
	if err := f.http.Get(ctx, "/fanscore/profiles/"+url.PathEscape(profile)+"/fans/"+url.PathEscape(fanRef), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMyProfile returns the calling fan's own object (end-user JWT). Audience self.
func (f *FanScoreProfilesClient) GetMyProfile(ctx context.Context, profile string) (*FanScoreProfile, error) {
	params := url.Values{}
	if profile != "" {
		params.Set("profile", profile)
	}
	var out FanScoreProfile
	if err := f.http.Get(ctx, "/fanscore/me", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPublicProfile returns a public profile for a fan; no credential required.
func (f *FanScoreProfilesClient) GetPublicProfile(ctx context.Context, profile, fanRef string) (*FanScoreProfile, error) {
	var out FanScoreProfile
	if err := f.http.Get(ctx, "/fanscore/public/"+url.PathEscape(profile)+"/fans/"+url.PathEscape(fanRef), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListProfiles returns the tenant's profiles (own first, then built-ins).
func (f *FanScoreProfilesClient) ListProfiles(ctx context.Context) ([]FanScoreProfileDefinition, error) {
	var out struct {
		Profiles []FanScoreProfileDefinition `json:"profiles"`
	}
	if err := f.http.Get(ctx, "/fanscore/profiles", nil, &out); err != nil {
		return nil, err
	}
	return out.Profiles, nil
}

// GetProfileSchema returns the JSON Schema of a profile's envelope, the same
// document the builder validates against. profile "" means the tenant's
// default. Tenant credentials required.
func (f *FanScoreProfilesClient) GetProfileSchema(ctx context.Context, profile string) (map[string]interface{}, error) {
	if profile == "" {
		profile = "default"
	}
	var out map[string]interface{}
	if err := f.http.Get(ctx, "/fanscore/profiles/"+url.PathEscape(profile)+"/schema", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
