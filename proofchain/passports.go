package proofchain

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// =============================================================================
// Types
// =============================================================================

// Passport represents a user's passport with points, level, and traits
type Passport struct {
	ID             string                 `json:"id"`
	TenantID       string                 `json:"tenant_id"`
	UserID         string                 `json:"user_id"`
	WalletAddress  *string                `json:"wallet_address,omitempty"`
	Level          int                    `json:"level"`
	Points         int                    `json:"points"`
	Experience     int                    `json:"experience"`
	Traits         map[string]interface{} `json:"traits"`
	CustomMetadata map[string]interface{} `json:"custom_metadata"`
	OnChainTokenID *string                `json:"on_chain_token_id,omitempty"`
	OnChainTxHash  *string                `json:"on_chain_tx_hash,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	LastUpdatedAt  time.Time              `json:"last_updated_at"`
}

// FieldValue represents a computed or manual field value on a passport
type FieldValue struct {
	FieldKey    string      `json:"field_key"`
	FieldName   string      `json:"field_name"`
	Value       interface{} `json:"value"`
	DataType    string      `json:"data_type"`
	ComputedAt  *time.Time  `json:"computed_at,omitempty"`
	ManuallySet bool        `json:"manually_set"`
}

// PassportWithFields includes field values
type PassportWithFields struct {
	Passport
	TemplateID   *string      `json:"template_id,omitempty"`
	TemplateName *string      `json:"template_name,omitempty"`
	FieldValues  []FieldValue `json:"field_values"`
}

// PassportTemplate defines the structure of a passport
type PassportTemplate struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Slug        string          `json:"slug"`
	IconURL     *string         `json:"icon_url,omitempty"`
	IsDefault   bool            `json:"is_default"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Fields      []TemplateField `json:"fields"`
}

// TemplateField defines a field within a template
type TemplateField struct {
	ID            string                 `json:"id"`
	TemplateID    string                 `json:"template_id"`
	Name          string                 `json:"name"`
	FieldKey      string                 `json:"field_key"`
	Description   *string                `json:"description,omitempty"`
	DataType      string                 `json:"data_type"`
	DefaultValue  interface{}            `json:"default_value,omitempty"`
	Formula       *string                `json:"formula,omitempty"`
	FormulaType   *string                `json:"formula_type,omitempty"`
	Aggregation   *string                `json:"aggregation,omitempty"`
	EventFilter   map[string]interface{} `json:"event_filter,omitempty"`
	DisplayFormat *string                `json:"display_format,omitempty"`
	Icon          *string                `json:"icon,omitempty"`
	Color         *string                `json:"color,omitempty"`
	SortOrder     int                    `json:"sort_order"`
	IsVisible     bool                   `json:"is_visible"`
	IsComputed    bool                   `json:"is_computed"`
	CreatedAt     time.Time              `json:"created_at"`
}

// Badge represents a badge that can be earned
type Badge struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	BadgeID      string                 `json:"badge_id"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description,omitempty"`
	IconURL      *string                `json:"icon_url,omitempty"`
	Rarity       string                 `json:"rarity"`
	Requirements map[string]interface{} `json:"requirements"`
	CreatedAt    time.Time              `json:"created_at"`
}

// Achievement represents an achievement that can be earned
type Achievement struct {
	ID            string                 `json:"id"`
	TenantID      string                 `json:"tenant_id"`
	AchievementID string                 `json:"achievement_id"`
	Name          string                 `json:"name"`
	Description   *string                `json:"description,omitempty"`
	Category      *string                `json:"category,omitempty"`
	PointsReward  int                    `json:"points_reward"`
	Requirements  map[string]interface{} `json:"requirements"`
	CreatedAt     time.Time              `json:"created_at"`
}

// UserBadge represents a badge earned by a user
type UserBadge struct {
	ID         string                 `json:"id"`
	PassportID string                 `json:"passport_id"`
	BadgeID    string                 `json:"badge_id"`
	Badge      *Badge                 `json:"badge,omitempty"`
	EarnedAt   time.Time              `json:"earned_at"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// UserAchievement represents an achievement earned by a user
type UserAchievement struct {
	ID            string       `json:"id"`
	PassportID    string       `json:"passport_id"`
	AchievementID string       `json:"achievement_id"`
	Achievement   *Achievement `json:"achievement,omitempty"`
	EarnedAt      time.Time    `json:"earned_at"`
	Progress      float64      `json:"progress"`
	Completed     bool         `json:"completed"`
}

// PassportHistory represents a history entry for a passport
type PassportHistory struct {
	ID           string      `json:"id"`
	PassportID   string      `json:"passport_id"`
	EventType    string      `json:"event_type"`
	OldValue     interface{} `json:"old_value,omitempty"`
	NewValue     interface{} `json:"new_value,omitempty"`
	ChangeReason *string     `json:"change_reason,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

// =============================================================================
// Request Types
// =============================================================================

// CreatePassportRequest is the request to create a new passport
type CreatePassportRequest struct {
	UserID         string                 `json:"user_id"`
	WalletAddress  *string                `json:"wallet_address,omitempty"`
	Level          int                    `json:"level,omitempty"`
	Points         int                    `json:"points,omitempty"`
	Experience     int                    `json:"experience,omitempty"`
	Traits         map[string]interface{} `json:"traits,omitempty"`
	CustomMetadata map[string]interface{} `json:"custom_metadata,omitempty"`
}

// UpdatePassportRequest is the request to update a passport
type UpdatePassportRequest struct {
	WalletAddress  *string                `json:"wallet_address,omitempty"`
	Level          *int                   `json:"level,omitempty"`
	Points         *int                   `json:"points,omitempty"`
	Experience     *int                   `json:"experience,omitempty"`
	Traits         map[string]interface{} `json:"traits,omitempty"`
	CustomMetadata map[string]interface{} `json:"custom_metadata,omitempty"`
}

// CreateTemplateRequest is the request to create a new template
type CreateTemplateRequest struct {
	Name        string                       `json:"name"`
	Description *string                      `json:"description,omitempty"`
	Slug        string                       `json:"slug"`
	IconURL     *string                      `json:"icon_url,omitempty"`
	IsDefault   bool                         `json:"is_default,omitempty"`
	Fields      []CreateTemplateFieldRequest `json:"fields,omitempty"`
}

// CreateTemplateFieldRequest is the request to create a template field
type CreateTemplateFieldRequest struct {
	Name          string                 `json:"name"`
	FieldKey      string                 `json:"field_key"`
	Description   *string                `json:"description,omitempty"`
	DataType      string                 `json:"data_type,omitempty"`
	DefaultValue  interface{}            `json:"default_value,omitempty"`
	Formula       *string                `json:"formula,omitempty"`
	FormulaType   *string                `json:"formula_type,omitempty"`
	Aggregation   *string                `json:"aggregation,omitempty"`
	EventFilter   map[string]interface{} `json:"event_filter,omitempty"`
	DisplayFormat *string                `json:"display_format,omitempty"`
	Icon          *string                `json:"icon,omitempty"`
	Color         *string                `json:"color,omitempty"`
	SortOrder     int                    `json:"sort_order,omitempty"`
	IsVisible     bool                   `json:"is_visible,omitempty"`
	IsComputed    bool                   `json:"is_computed,omitempty"`
}

// CreateBadgeRequest is the request to create a badge
type CreateBadgeRequest struct {
	BadgeID      string                 `json:"badge_id"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description,omitempty"`
	IconURL      *string                `json:"icon_url,omitempty"`
	Rarity       string                 `json:"rarity,omitempty"`
	Requirements map[string]interface{} `json:"requirements,omitempty"`
}

// CreateAchievementRequest is the request to create an achievement
type CreateAchievementRequest struct {
	AchievementID string                 `json:"achievement_id"`
	Name          string                 `json:"name"`
	Description   *string                `json:"description,omitempty"`
	Category      *string                `json:"category,omitempty"`
	PointsReward  int                    `json:"points_reward,omitempty"`
	Requirements  map[string]interface{} `json:"requirements,omitempty"`
}

// ListOptions for pagination
type PassportListOptions struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// =============================================================================
// Passport Client
// =============================================================================

// PassportClient provides access to passport operations
type PassportClient struct {
	http *HTTPClient
}

// NewPassportClient creates a new passport client
func NewPassportClient(http *HTTPClient) *PassportClient {
	return &PassportClient{http: http}
}

// ---------------------------------------------------------------------------
// Passports
// ---------------------------------------------------------------------------

// List returns all passports for the tenant
func (p *PassportClient) List(ctx context.Context, opts *PassportListOptions) ([]Passport, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	}

	var passports []Passport
	err := p.http.Get(ctx, "/passports", params, &passports)
	return passports, err
}

// Get returns a passport by user ID
func (p *PassportClient) Get(ctx context.Context, userID string) (*Passport, error) {
	var passport Passport
	err := p.http.Get(ctx, "/passports/"+url.PathEscape(userID), nil, &passport)
	if err != nil {
		return nil, err
	}
	return &passport, nil
}

// GetWithFields returns a passport with all field values
func (p *PassportClient) GetWithFields(ctx context.Context, userID string) (*PassportWithFields, error) {
	passport, err := p.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	fields, err := p.GetFieldValues(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &PassportWithFields{
		Passport:    *passport,
		FieldValues: fields,
	}, nil
}

// Create creates a new passport
func (p *PassportClient) Create(ctx context.Context, req *CreatePassportRequest) (*Passport, error) {
	var passport Passport
	err := p.http.Post(ctx, "/passports", req, &passport)
	if err != nil {
		return nil, err
	}
	return &passport, nil
}

// Update updates a passport
func (p *PassportClient) Update(ctx context.Context, userID string, req *UpdatePassportRequest) (*Passport, error) {
	var passport Passport
	err := p.http.Put(ctx, "/passports/"+url.PathEscape(userID), req, &passport)
	if err != nil {
		return nil, err
	}
	return &passport, nil
}

// Delete deletes a passport
func (p *PassportClient) Delete(ctx context.Context, userID string) error {
	return p.http.Delete(ctx, "/passports/"+url.PathEscape(userID))
}

// AddPoints adds points to a passport
func (p *PassportClient) AddPoints(ctx context.Context, userID string, points int, reason string) (*Passport, error) {
	var passport Passport
	err := p.http.Post(ctx, "/passports/"+url.PathEscape(userID)+"/add-points", map[string]interface{}{
		"points": points,
		"reason": reason,
	}, &passport)
	if err != nil {
		return nil, err
	}
	return &passport, nil
}

// LevelUp levels up a passport
func (p *PassportClient) LevelUp(ctx context.Context, userID string) (*Passport, error) {
	var passport Passport
	err := p.http.Post(ctx, "/passports/"+url.PathEscape(userID)+"/level-up", nil, &passport)
	if err != nil {
		return nil, err
	}
	return &passport, nil
}

// LinkWallet links a wallet address to a passport
func (p *PassportClient) LinkWallet(ctx context.Context, userID string, walletAddress string) (*Passport, error) {
	return p.Update(ctx, userID, &UpdatePassportRequest{WalletAddress: &walletAddress})
}

// ---------------------------------------------------------------------------
// Field Values
// ---------------------------------------------------------------------------

// GetFieldValues returns all field values for a passport
func (p *PassportClient) GetFieldValues(ctx context.Context, userID string) ([]FieldValue, error) {
	var fields []FieldValue
	err := p.http.Get(ctx, "/passports/"+url.PathEscape(userID)+"/fields", nil, &fields)
	return fields, err
}

// SetFieldValue sets a field value manually
func (p *PassportClient) SetFieldValue(ctx context.Context, userID string, fieldKey string, value interface{}) error {
	return p.http.Put(ctx, "/passports/"+url.PathEscape(userID)+"/fields/"+fieldKey, value, nil)
}

// RecomputeFields recomputes all computed field values
func (p *PassportClient) RecomputeFields(ctx context.Context, userID string) (map[string]interface{}, error) {
	var result struct {
		UpdatedFields map[string]interface{} `json:"updated_fields"`
	}
	err := p.http.Post(ctx, "/passports/"+url.PathEscape(userID)+"/recompute", nil, &result)
	if err != nil {
		return nil, err
	}
	return result.UpdatedFields, nil
}

// AssignTemplate assigns a template to a passport
func (p *PassportClient) AssignTemplate(ctx context.Context, userID string, templateID string) error {
	return p.http.Post(ctx, "/passports/"+url.PathEscape(userID)+"/assign-template/"+templateID, nil, nil)
}

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

// ListTemplates returns all passport templates
func (p *PassportClient) ListTemplates(ctx context.Context) ([]PassportTemplate, error) {
	var templates []PassportTemplate
	err := p.http.Get(ctx, "/passports/templates", nil, &templates)
	return templates, err
}

// GetTemplate returns a template by ID
func (p *PassportClient) GetTemplate(ctx context.Context, templateID string) (*PassportTemplate, error) {
	var template PassportTemplate
	err := p.http.Get(ctx, "/passports/templates/"+templateID, nil, &template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// CreateTemplate creates a new template
func (p *PassportClient) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*PassportTemplate, error) {
	var template PassportTemplate
	err := p.http.Post(ctx, "/passports/templates", req, &template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// AddTemplateField adds a field to a template
func (p *PassportClient) AddTemplateField(ctx context.Context, templateID string, req *CreateTemplateFieldRequest) (*TemplateField, error) {
	var field TemplateField
	err := p.http.Post(ctx, "/passports/templates/"+templateID+"/fields", req, &field)
	if err != nil {
		return nil, err
	}
	return &field, nil
}

// DeleteTemplate deletes a template
func (p *PassportClient) DeleteTemplate(ctx context.Context, templateID string) error {
	return p.http.Delete(ctx, "/passports/templates/"+templateID)
}

// ---------------------------------------------------------------------------
// Badges
// ---------------------------------------------------------------------------

// ListBadges returns all badges
func (p *PassportClient) ListBadges(ctx context.Context) ([]Badge, error) {
	var badges []Badge
	err := p.http.Get(ctx, "/passports/badges", nil, &badges)
	return badges, err
}

// CreateBadge creates a new badge
func (p *PassportClient) CreateBadge(ctx context.Context, req *CreateBadgeRequest) (*Badge, error) {
	var badge Badge
	err := p.http.Post(ctx, "/passports/badges", req, &badge)
	if err != nil {
		return nil, err
	}
	return &badge, nil
}

// AwardBadge awards a badge to a user
func (p *PassportClient) AwardBadge(ctx context.Context, userID string, badgeID string, metadata map[string]interface{}) (*UserBadge, error) {
	var userBadge UserBadge
	err := p.http.Post(ctx, "/passports/"+url.PathEscape(userID)+"/badges/"+badgeID, map[string]interface{}{
		"metadata": metadata,
	}, &userBadge)
	if err != nil {
		return nil, err
	}
	return &userBadge, nil
}

// GetUserBadges returns badges earned by a user
func (p *PassportClient) GetUserBadges(ctx context.Context, userID string) ([]UserBadge, error) {
	var badges []UserBadge
	err := p.http.Get(ctx, "/passports/"+url.PathEscape(userID)+"/badges", nil, &badges)
	return badges, err
}

// ---------------------------------------------------------------------------
// Achievements
// ---------------------------------------------------------------------------

// ListAchievements returns all achievements
func (p *PassportClient) ListAchievements(ctx context.Context) ([]Achievement, error) {
	var achievements []Achievement
	err := p.http.Get(ctx, "/passports/achievements", nil, &achievements)
	return achievements, err
}

// CreateAchievement creates a new achievement
func (p *PassportClient) CreateAchievement(ctx context.Context, req *CreateAchievementRequest) (*Achievement, error) {
	var achievement Achievement
	err := p.http.Post(ctx, "/passports/achievements", req, &achievement)
	if err != nil {
		return nil, err
	}
	return &achievement, nil
}

// GetUserAchievements returns achievements for a user
func (p *PassportClient) GetUserAchievements(ctx context.Context, userID string) ([]UserAchievement, error) {
	var achievements []UserAchievement
	err := p.http.Get(ctx, "/passports/"+url.PathEscape(userID)+"/achievements", nil, &achievements)
	return achievements, err
}

// UpdateAchievementProgress updates achievement progress
func (p *PassportClient) UpdateAchievementProgress(ctx context.Context, userID string, achievementID string, progress float64) (*UserAchievement, error) {
	var achievement UserAchievement
	err := p.http.Put(ctx, "/passports/"+url.PathEscape(userID)+"/achievements/"+achievementID, map[string]interface{}{
		"progress": progress,
	}, &achievement)
	if err != nil {
		return nil, err
	}
	return &achievement, nil
}

// ---------------------------------------------------------------------------
// History
// ---------------------------------------------------------------------------

// GetHistory returns passport history/activity log
func (p *PassportClient) GetHistory(ctx context.Context, userID string, opts *PassportListOptions) ([]PassportHistory, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	}

	var history []PassportHistory
	err := p.http.Get(ctx, "/passports/"+url.PathEscape(userID)+"/history?"+params.Encode(), nil, &history)
	return history, err
}
