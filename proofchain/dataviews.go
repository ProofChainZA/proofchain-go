package proofchain

import (
	"context"
	"fmt"
	"net/url"
)

// =============================================================================
// Types
// =============================================================================

// DataViewSummary represents a summary of a data view.
type DataViewSummary struct {
	Name             string   `json:"name"`
	DisplayName      *string  `json:"display_name,omitempty"`
	Description      string   `json:"description"`
	ViewType         *string  `json:"view_type,omitempty"`
	SourceCategories []string `json:"source_categories,omitempty"`
}

// DataViewListResponse contains categorised data views.
type DataViewListResponse struct {
	OwnViews     []DataViewSummary `json:"own_views"`
	PublicViews  []DataViewSummary `json:"public_views"`
	BuiltinViews []DataViewSummary `json:"builtin_views"`
}

// DataViewComputation defines a computation within a data view.
type DataViewComputation struct {
	Type           string             `json:"type"`
	Name           *string            `json:"name,omitempty"`
	EventTypes     []string           `json:"event_types,omitempty"`
	TimeWindowDays *int               `json:"time_window_days,omitempty"`
	EventWeights   map[string]float64 `json:"event_weights,omitempty"`
	MaxScore       *float64           `json:"max_score,omitempty"`
	DecayRate      *float64           `json:"decay_rate,omitempty"`
	Field          *string            `json:"field,omitempty"`
	Operation      *string            `json:"operation,omitempty"`
	GroupBy        *string            `json:"group_by,omitempty"`
	Limit          *int               `json:"limit,omitempty"`
	Fields         []string           `json:"fields,omitempty"`
	Tiers          []TierDefinition   `json:"tiers,omitempty"`
	ScoreSource    *string            `json:"score_source,omitempty"`
}

// TierDefinition defines a tier within a computation.
type TierDefinition struct {
	Name string  `json:"name"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

// DataViewDetail contains full details of a data view.
type DataViewDetail struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	DisplayName      string      `json:"display_name"`
	Description      string      `json:"description"`
	ViewType         string      `json:"view_type"`
	Computation      interface{} `json:"computation"` // single DataViewComputation or []DataViewComputation
	SourceCategories []string    `json:"source_categories"`
	IsPublic         bool        `json:"is_public"`
	CreatedAt        string      `json:"created_at"`
	UpdatedAt        string      `json:"updated_at"`
}

// DataViewExecuteResult is the result of executing a data view.
type DataViewExecuteResult struct {
	ViewName       string                 `json:"view_name"`
	DisplayName    string                 `json:"display_name"`
	Identifier     string                 `json:"identifier"`
	IdentifierType string                 `json:"identifier_type"`
	Data           map[string]interface{} `json:"data"`
	ComputedAt     string                 `json:"computed_at"`
	TotalEvents    int                    `json:"total_events"`
}

// DataViewPreviewResult is the result of previewing a computation.
type DataViewPreviewResult struct {
	Preview         bool                   `json:"preview"`
	WalletAddress   string                 `json:"wallet_address"`
	ComputationType string                 `json:"computation_type"`
	Result          map[string]interface{} `json:"result"`
	EventsProcessed int                    `json:"events_processed"`
	TimeWindowDays  int                    `json:"time_window_days"`
}

// FanProfileView is the builtin fan profile view result.
type FanProfileView struct {
	WalletAddress  string         `json:"wallet_address"`
	FanScore       float64        `json:"fan_score"`
	TotalEvents    int            `json:"total_events"`
	EventBreakdown map[string]int `json:"event_breakdown"`
	FirstSeen      string         `json:"first_seen"`
	LastSeen       string         `json:"last_seen"`
	LoyaltyTier    string         `json:"loyalty_tier"`
	ComputedAt     string         `json:"computed_at"`
}

// ActivitySummaryView is the builtin activity summary view result.
type ActivitySummaryView struct {
	WalletAddress    string         `json:"wallet_address"`
	TotalEvents      int            `json:"total_events"`
	EventCountByType map[string]int `json:"event_count_by_type"`
	LastActivity     string         `json:"last_activity"`
	ActiveDays       int            `json:"active_days"`
	PeriodDays       int            `json:"period_days"`
	ComputedAt       string         `json:"computed_at"`
}

// EventMetadata contains available event types and counts.
type EventMetadata struct {
	EventTypes  []EventTypeInfo `json:"event_types"`
	TotalEvents int             `json:"total_events"`
	Categories  []string        `json:"categories"`
}

// EventTypeInfo describes a single event type.
type EventTypeInfo struct {
	EventType string `json:"event_type"`
	Count     int    `json:"count"`
	LastSeen  string `json:"last_seen"`
}

// ViewTemplate is a pre-configured computation pattern.
type ViewTemplate struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Computation DataViewComputation `json:"computation"`
}

// =============================================================================
// Request types
// =============================================================================

// CreateDataViewRequest creates a new custom data view.
type CreateDataViewRequest struct {
	Name             string      `json:"name"`
	DisplayName      string      `json:"display_name"`
	Description      string      `json:"description"`
	ViewType         string      `json:"view_type,omitempty"`
	Computation      interface{} `json:"computation"` // single or []DataViewComputation
	SourceCategories []string    `json:"source_categories,omitempty"`
	IsPublic         *bool       `json:"is_public,omitempty"`
}

// UpdateDataViewRequest updates an existing data view.
type UpdateDataViewRequest struct {
	DisplayName      *string     `json:"display_name,omitempty"`
	Description      *string     `json:"description,omitempty"`
	Computation      interface{} `json:"computation,omitempty"`
	SourceCategories []string    `json:"source_categories,omitempty"`
	IsPublic         *bool       `json:"is_public,omitempty"`
}

// DataViewPreviewRequest previews a computation without saving.
type DataViewPreviewRequest struct {
	WalletAddress  string      `json:"wallet_address"`
	Computation    interface{} `json:"computation"` // single or []DataViewComputation
	TimeWindowDays *int        `json:"time_window_days,omitempty"`
	Limit          *int        `json:"limit,omitempty"`
}

// =============================================================================
// Client
// =============================================================================

// DataViewsClient provides data view operations.
type DataViewsClient struct {
	http *HTTPClient
}

// NewDataViewsClient creates a new data views client.
func NewDataViewsClient(http *HTTPClient) *DataViewsClient {
	return &DataViewsClient{http: http}
}

// List returns all available data views (own, public, builtin).
func (d *DataViewsClient) List(ctx context.Context) (*DataViewListResponse, error) {
	var response DataViewListResponse
	err := d.http.Get(ctx, "/data-mesh/views", nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Get returns detailed information about a specific data view.
func (d *DataViewsClient) Get(ctx context.Context, viewName string) (*DataViewDetail, error) {
	var detail DataViewDetail
	err := d.http.Get(ctx, "/data-mesh/views/custom/"+url.PathEscape(viewName), nil, &detail)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// Create creates a new custom data view.
func (d *DataViewsClient) Create(ctx context.Context, req *CreateDataViewRequest) (*DataViewDetail, error) {
	var detail DataViewDetail
	err := d.http.Post(ctx, "/data-mesh/views/custom", req, &detail)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// Update updates an existing data view.
func (d *DataViewsClient) Update(ctx context.Context, viewName string, req *UpdateDataViewRequest) (*DataViewDetail, error) {
	var detail DataViewDetail
	err := d.http.Patch(ctx, "/data-mesh/views/custom/"+url.PathEscape(viewName), req, &detail)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// Delete deletes a data view.
func (d *DataViewsClient) Delete(ctx context.Context, viewName string) error {
	return d.http.Delete(ctx, "/data-mesh/views/custom/" + url.PathEscape(viewName))
}

// Execute executes a data view for a specific identifier (user ID or wallet address).
func (d *DataViewsClient) Execute(ctx context.Context, identifier, viewName string) (*DataViewExecuteResult, error) {
	var result DataViewExecuteResult
	err := d.http.Get(ctx, "/data-mesh/views/"+url.PathEscape(identifier)+"/custom/"+url.PathEscape(viewName), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Preview previews a computation without saving it.
func (d *DataViewsClient) Preview(ctx context.Context, req *DataViewPreviewRequest) (*DataViewPreviewResult, error) {
	var result DataViewPreviewResult
	err := d.http.Post(ctx, "/data-mesh/views/preview", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFanProfile returns the builtin fan profile view for a wallet.
func (d *DataViewsClient) GetFanProfile(ctx context.Context, walletAddress string) (*FanProfileView, error) {
	var result FanProfileView
	err := d.http.Get(ctx, "/data-mesh/views/"+url.PathEscape(walletAddress)+"/fan-profile", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetActivitySummary returns the builtin activity summary view for a wallet.
func (d *DataViewsClient) GetActivitySummary(ctx context.Context, walletAddress string, days int) (*ActivitySummaryView, error) {
	params := url.Values{}
	if days > 0 {
		params.Set("days", fmt.Sprintf("%d", days))
	}

	var result ActivitySummaryView
	err := d.http.Get(ctx, "/data-mesh/views/"+url.PathEscape(walletAddress)+"/activity-summary", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetEventMetadata returns available event types and their counts.
func (d *DataViewsClient) GetEventMetadata(ctx context.Context) (*EventMetadata, error) {
	var result EventMetadata
	err := d.http.Get(ctx, "/data-mesh/event-metadata", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTemplates returns available view templates.
func (d *DataViewsClient) GetTemplates(ctx context.Context) ([]ViewTemplate, error) {
	var result struct {
		Templates []ViewTemplate `json:"templates"`
	}
	err := d.http.Get(ctx, "/data-mesh/views/templates", nil, &result)
	if err != nil {
		return nil, err
	}
	return result.Templates, nil
}
