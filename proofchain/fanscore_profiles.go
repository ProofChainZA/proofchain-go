package proofchain

import (
	"context"
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

// FanScoreProfile is a fan's FanScore object as shaped by a tenant profile.
//
// Blocks are present only when the profile enables them and the caller's
// audience may see at least one field; a present block that is nil means the
// fan has no data for it. Fields inside a block are optional because the
// tenant chooses which to expose. Blocks this SDK does not model (season,
// form, activity, and any added later) are available raw via Extra.
type FanScoreProfile struct {
	Profile  FanScoreProfileMeta `json:"profile"`
	Identity *FanScoreIdentity   `json:"identity,omitempty"`
	Score    *FanScoreScore      `json:"score,omitempty"`
	Cohorts  *FanScoreCohorts    `json:"cohorts,omitempty"`
	Rank     *FanScoreRank       `json:"rank,omitempty"`
	Points   *FanScorePoints     `json:"points,omitempty"`
}

type FanScoreIdentity struct {
	FanID         *string  `json:"fan_id,omitempty"`
	ExternalID    *string  `json:"external_id,omitempty"`
	DisplayName   *string  `json:"display_name,omitempty"`
	AvatarURL     *string  `json:"avatar_url,omitempty"`
	Country       *string  `json:"country,omitempty"`
	Language      *string  `json:"language,omitempty"`
	Email         *string  `json:"email,omitempty"`
	WalletAddress *string  `json:"wallet_address,omitempty"`
	Segments      []string `json:"segments,omitempty"`
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

// FanScoreProfileDefinition is a profile as listed by GET /fanscore/profiles.
type FanScoreProfileDefinition struct {
	ID          *string                  `json:"id"`
	Slug        string                   `json:"slug"`
	DisplayName string                   `json:"display_name"`
	Description *string                  `json:"description"`
	Audience    string                   `json:"audience"`
	CacheTTLs   int                      `json:"cache_ttl_s"`
	IsDefault   bool                     `json:"is_default"`
	Version     int                      `json:"version"`
	Builtin     bool                     `json:"builtin"`
	Blocks      []map[string]interface{} `json:"blocks"`
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
