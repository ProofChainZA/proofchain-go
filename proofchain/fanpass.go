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

// FanpassLeaderboardEntry is a single entry in the fanpass leaderboard.
type FanpassLeaderboardEntry struct {
	Rank            int                     `json:"rank"`
	UserID          string                  `json:"user_id"`
	FanScore        float64                 `json:"fan_score"`
	NormalizedScore float64                 `json:"normalized_score"`
	RawScore        float64                 `json:"raw_score"`
	Percentile      float64                 `json:"percentile"`
	ComputedAt      *string                 `json:"computed_at,omitempty"`
	User            *LeaderboardUserProfile `json:"user,omitempty"`
}

// FanpassGroupStats contains aggregate statistics for the fanpass group.
type FanpassGroupStats struct {
	AvgFanScore       float64 `json:"avg_fan_score"`
	TopNAvgFanScore   float64 `json:"top_n_avg_fan_score"`
	GlobalCount       int     `json:"global_count"`
	GlobalAvgFanScore float64 `json:"global_avg_fan_score"`
}

// FanpassLeaderboardResponse is the full fanpass leaderboard response.
type FanpassLeaderboardResponse struct {
	AggregationRuleID        *string                   `json:"aggregation_rule_id,omitempty"`
	AggregationRuleName      *string                   `json:"aggregation_rule_name,omitempty"`
	Filter                   map[string]interface{}     `json:"filter"`
	TotalUsers               int                        `json:"total_users"`
	GroupStats               FanpassGroupStats          `json:"group_stats"`
	Leaderboard              []FanpassLeaderboardEntry  `json:"leaderboard"`
	CurrentUser              *FanpassLeaderboardEntry   `json:"current_user,omitempty"`
	CurrentUserInLeaderboard bool                       `json:"current_user_in_leaderboard"`
}

// FanpassUserComparisonResponse contains a user's comparison across all cohorts.
type FanpassUserComparisonResponse struct {
	UserID  string                     `json:"user_id"`
	Filter  map[string]interface{}     `json:"filter"`
	Cohorts []UserCohortBreakdownEntry `json:"cohorts"`
}

// =============================================================================
// Options
// =============================================================================

// FanpassLeaderboardOptions configures the GetLeaderboard query.
type FanpassLeaderboardOptions struct {
	AggregationRuleID string
	Filters           map[string]string
	Country           string
	Limit             int
	TopN              int
	Fresh             bool
	UserID            string
}

// =============================================================================
// Client
// =============================================================================

// FanpassLeaderboardClient provides fanpass leaderboard operations.
type FanpassLeaderboardClient struct {
	http *HTTPClient
}

// NewFanpassLeaderboardClient creates a new fanpass leaderboard client.
func NewFanpassLeaderboardClient(http *HTTPClient) *FanpassLeaderboardClient {
	return &FanpassLeaderboardClient{http: http}
}

// GetLeaderboard returns the fanpass leaderboard with composite scores.
func (f *FanpassLeaderboardClient) GetLeaderboard(ctx context.Context, opts *FanpassLeaderboardOptions) (*FanpassLeaderboardResponse, error) {
	params := url.Values{}
	if opts != nil {
		if opts.AggregationRuleID != "" {
			params.Set("aggregation_rule_id", opts.AggregationRuleID)
		}
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

	var response FanpassLeaderboardResponse
	err := f.http.Get(ctx, "/passport-v2/fanpass/leaderboard", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetUserComparison returns a user's comparison across all cohorts (spider chart data).
func (f *FanpassLeaderboardClient) GetUserComparison(ctx context.Context, userID string, filters map[string]string, country string) (*FanpassUserComparisonResponse, error) {
	params := url.Values{}
	if len(filters) > 0 {
		filtersJSON, _ := json.Marshal(filters)
		params.Set("filters", string(filtersJSON))
	}
	if country != "" {
		params.Set("country", country)
	}

	var response FanpassUserComparisonResponse
	err := f.http.Get(ctx, "/passport-v2/fanpass/"+url.PathEscape(userID)+"/comparison", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
