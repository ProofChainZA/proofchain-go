package proofchain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// =============================================================================
// Types
// =============================================================================

// CohortDefinition represents a cohort scoring definition.
type CohortDefinition struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenant_id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description *string  `json:"description,omitempty"`
	ScoringType string   `json:"scoring_type"`
	Icon        *string  `json:"icon,omitempty"`
	Color       *string  `json:"color,omitempty"`
	Status      string   `json:"status"`
	AvgScore    *float64 `json:"avg_score,omitempty"`
	TotalUsers  *int     `json:"total_users,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// LeaderboardUserProfile contains user profile data in leaderboard entries.
type LeaderboardUserProfile struct {
	ExternalID  string                 `json:"external_id"`
	DisplayName *string                `json:"display_name,omitempty"`
	FirstName   *string                `json:"first_name,omitempty"`
	LastName    *string                `json:"last_name,omitempty"`
	Email       *string                `json:"email,omitempty"`
	AvatarURL   *string                `json:"avatar_url,omitempty"`
	Country     *string                `json:"country,omitempty"`
	City        *string                `json:"city,omitempty"`
	Attributes  map[string]interface{} `json:"attributes"`
}

// CohortLeaderboardEntry is a single entry in a cohort leaderboard.
type CohortLeaderboardEntry struct {
	Rank               int                     `json:"rank"`
	UserID             string                  `json:"user_id"`
	Score              float64                 `json:"score"`
	PercentileGlobal   float64                 `json:"percentile_global"`
	PercentileFiltered *float64                `json:"percentile_filtered,omitempty"`
	ComputedAt         *string                 `json:"computed_at,omitempty"`
	User               *LeaderboardUserProfile `json:"user,omitempty"`
}

// CohortGroupStats contains aggregate statistics for a cohort group.
type CohortGroupStats struct {
	FilteredAvgPercentile     *float64 `json:"filtered_avg_percentile,omitempty"`
	FilteredTopNAvgPercentile *float64 `json:"filtered_top_n_avg_percentile,omitempty"`
	GlobalCount               int      `json:"global_count"`
	GlobalAvgPercentile       float64  `json:"global_avg_percentile"`
}

// CohortLeaderboardResponse is the full leaderboard response.
type CohortLeaderboardResponse struct {
	CohortID                 string                   `json:"cohort_id"`
	CohortName               string                   `json:"cohort_name"`
	Filter                   map[string]interface{}   `json:"filter"`
	TotalUsers               int                      `json:"total_users"`
	GroupStats               CohortGroupStats         `json:"group_stats"`
	Leaderboard              []CohortLeaderboardEntry `json:"leaderboard"`
	CurrentUser              *CohortLeaderboardEntry  `json:"current_user,omitempty"`
	CurrentUserInLeaderboard bool                     `json:"current_user_in_leaderboard"`
}

// UserCohortBreakdownEntry is a single cohort entry in a user breakdown.
type UserCohortBreakdownEntry struct {
	CohortID                   string   `json:"cohort_id"`
	CohortSlug                 string   `json:"cohort_slug"`
	CohortName                 string   `json:"cohort_name"`
	Icon                       *string  `json:"icon,omitempty"`
	Color                      *string  `json:"color,omitempty"`
	UserPercentile             float64  `json:"user_percentile"`
	FilteredGroupAvgPercentile *float64 `json:"filtered_group_avg_percentile,omitempty"`
	GlobalGroupAvgPercentile   float64  `json:"global_group_avg_percentile"`
}

// UserBreakdownResponse contains a user's breakdown across all cohorts.
type UserBreakdownResponse struct {
	UserID  string                     `json:"user_id"`
	Filter  map[string]interface{}     `json:"filter"`
	Cohorts []UserCohortBreakdownEntry `json:"cohorts"`
}

// =============================================================================
// Options
// =============================================================================

// ListCohortsOptions configures the List query.
type ListCohortsOptions struct {
	Status string // "active", "inactive", "draft"
	Limit  int
	Offset int
}

// CohortLeaderboardOptions configures the GetLeaderboard query.
type CohortLeaderboardOptions struct {
	Filters map[string]string
	Country string
	Limit   int
	TopN    int
	Fresh   bool
	UserID  string
}

// =============================================================================
// Client
// =============================================================================

// CohortLeaderboardClient provides cohort leaderboard operations.
type CohortLeaderboardClient struct {
	http *HTTPClient
}

// NewCohortLeaderboardClient creates a new cohort leaderboard client.
func NewCohortLeaderboardClient(http *HTTPClient) *CohortLeaderboardClient {
	return &CohortLeaderboardClient{http: http}
}

// List returns all cohort definitions.
func (c *CohortLeaderboardClient) List(ctx context.Context, opts *ListCohortsOptions) ([]CohortDefinition, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	}

	var definitions []CohortDefinition
	err := c.http.Get(ctx, "/cohorts/definitions", params, &definitions)
	if err != nil {
		return nil, err
	}
	return definitions, nil
}

// Get returns a cohort definition by ID.
func (c *CohortLeaderboardClient) Get(ctx context.Context, cohortID string) (*CohortDefinition, error) {
	var definition CohortDefinition
	err := c.http.Get(ctx, "/cohorts/definitions/"+url.PathEscape(cohortID), nil, &definition)
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

// GetLeaderboard returns the filtered cohort leaderboard with global and filtered percentiles.
func (c *CohortLeaderboardClient) GetLeaderboard(ctx context.Context, cohortID string, opts *CohortLeaderboardOptions) (*CohortLeaderboardResponse, error) {
	params := url.Values{}
	if opts != nil {
		if len(opts.Filters) > 0 {
			filtersJSON, _ := json.Marshal(opts.Filters)
			params.Set("filters", string(filtersJSON))
		}
		if opts.Country != "" {
			params.Set("country", opts.Country)
		}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.TopN > 0 {
			params.Set("top_n", fmt.Sprintf("%d", opts.TopN))
		}
		if opts.Fresh {
			params.Set("fresh", "true")
		}
		if opts.UserID != "" {
			params.Set("user_id", opts.UserID)
		}
	}

	var response CohortLeaderboardResponse
	err := c.http.Get(ctx, "/cohorts/definitions/"+url.PathEscape(cohortID)+"/leaderboard", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetUserBreakdown returns a user's breakdown across all cohorts (for spider charts).
func (c *CohortLeaderboardClient) GetUserBreakdown(ctx context.Context, userID string, filters map[string]string, country string) (*UserBreakdownResponse, error) {
	params := url.Values{}
	if len(filters) > 0 {
		filtersJSON, _ := json.Marshal(filters)
		params.Set("filters", string(filtersJSON))
	}
	if country != "" {
		params.Set("country", country)
	}

	var response UserBreakdownResponse
	err := c.http.Get(ctx, "/cohorts/users/"+url.PathEscape(userID)+"/breakdown", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
