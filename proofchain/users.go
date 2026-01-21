package proofchain

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// EndUser represents an end-user discovered from events
type EndUser struct {
	ID             string                 `json:"id"`
	ExternalID     string                 `json:"external_id"`
	Email          *string                `json:"email,omitempty"`
	FirstName      *string                `json:"first_name,omitempty"`
	LastName       *string                `json:"last_name,omitempty"`
	DisplayName    *string                `json:"display_name,omitempty"`
	AvatarURL      *string                `json:"avatar_url,omitempty"`
	Phone          *string                `json:"phone,omitempty"`
	DateOfBirth    *time.Time             `json:"date_of_birth,omitempty"`
	Country        *string                `json:"country,omitempty"`
	City           *string                `json:"city,omitempty"`
	Timezone       *string                `json:"timezone,omitempty"`
	Language       *string                `json:"language,omitempty"`
	Bio            *string                `json:"bio,omitempty"`
	WalletAddress  *string                `json:"wallet_address,omitempty"`
	WalletSource   *string                `json:"wallet_source,omitempty"`
	Status         string                 `json:"status"`
	TotalEvents    int                    `json:"total_events"`
	FirstEventAt   *time.Time             `json:"first_event_at,omitempty"`
	LastEventAt    *time.Time             `json:"last_event_at,omitempty"`
	EventTypes     []string               `json:"event_types"`
	Segments       []string               `json:"segments"`
	Tags           map[string]interface{} `json:"tags"`
	PointsBalance  int                    `json:"points_balance"`
	LifetimePoints int                    `json:"lifetime_points"`
	Attributes     map[string]interface{} `json:"attributes"`
	CreatedAt      time.Time              `json:"created_at"`
	DiscoveredAt   *time.Time             `json:"discovered_at,omitempty"`
}

// EndUserListResponse is a paginated list of users
type EndUserListResponse struct {
	Users    []EndUser `json:"users"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	HasMore  bool      `json:"has_more"`
}

// UserActivity represents a user's activity event
type UserActivity struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// UserActivityResponse is a paginated activity response
type UserActivityResponse struct {
	UserID     string         `json:"user_id"`
	Activities []UserActivity `json:"activities"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

// Request types
type CreateEndUserRequest struct {
	ExternalID    string                 `json:"external_id"`
	Email         *string                `json:"email,omitempty"`
	FirstName     *string                `json:"first_name,omitempty"`
	LastName      *string                `json:"last_name,omitempty"`
	DisplayName   *string                `json:"display_name,omitempty"`
	Phone         *string                `json:"phone,omitempty"`
	DateOfBirth   *time.Time             `json:"date_of_birth,omitempty"`
	Country       *string                `json:"country,omitempty"`
	City          *string                `json:"city,omitempty"`
	Timezone      *string                `json:"timezone,omitempty"`
	Language      *string                `json:"language,omitempty"`
	Bio           *string                `json:"bio,omitempty"`
	WalletAddress *string                `json:"wallet_address,omitempty"`
	Segments      []string               `json:"segments,omitempty"`
	Tags          map[string]interface{} `json:"tags,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

type UpdateEndUserRequest struct {
	Email       *string                `json:"email,omitempty"`
	FirstName   *string                `json:"first_name,omitempty"`
	LastName    *string                `json:"last_name,omitempty"`
	DisplayName *string                `json:"display_name,omitempty"`
	AvatarURL   *string                `json:"avatar_url,omitempty"`
	Phone       *string                `json:"phone,omitempty"`
	DateOfBirth *time.Time             `json:"date_of_birth,omitempty"`
	Country     *string                `json:"country,omitempty"`
	City        *string                `json:"city,omitempty"`
	Timezone    *string                `json:"timezone,omitempty"`
	Language    *string                `json:"language,omitempty"`
	Bio         *string                `json:"bio,omitempty"`
	Segments    []string               `json:"segments,omitempty"`
	Tags        map[string]interface{} `json:"tags,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

type ListEndUsersOptions struct {
	Page      int
	PageSize  int
	Search    string
	Status    string
	Segment   string
	HasWallet *bool
	SortBy    string
	SortOrder string
}

type LinkWalletRequest struct {
	WalletAddress string  `json:"wallet_address"`
	WalletSource  *string `json:"wallet_source,omitempty"`
	Signature     *string `json:"signature,omitempty"`
}

type MergeUsersRequest struct {
	SourceUserID string `json:"source_user_id"`
	TargetUserID string `json:"target_user_id"`
}

// EndUsersClient provides end-user operations
type EndUsersClient struct {
	http *HTTPClient
}

// NewEndUsersClient creates a new end-users client
func NewEndUsersClient(http *HTTPClient) *EndUsersClient {
	return &EndUsersClient{http: http}
}

// List returns paginated end-users
func (u *EndUsersClient) List(ctx context.Context, opts *ListEndUsersOptions) (*EndUserListResponse, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			params.Set("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.PageSize > 0 {
			params.Set("page_size", fmt.Sprintf("%d", opts.PageSize))
		}
		if opts.Search != "" {
			params.Set("search", opts.Search)
		}
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.Segment != "" {
			params.Set("segment", opts.Segment)
		}
		if opts.HasWallet != nil {
			params.Set("has_wallet", fmt.Sprintf("%t", *opts.HasWallet))
		}
		if opts.SortBy != "" {
			params.Set("sort_by", opts.SortBy)
		}
		if opts.SortOrder != "" {
			params.Set("sort_order", opts.SortOrder)
		}
	}

	var response EndUserListResponse
	err := u.http.Get(ctx, "/end-users", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Get returns an end-user by ID
func (u *EndUsersClient) Get(ctx context.Context, userID string) (*EndUser, error) {
	var user EndUser
	err := u.http.Get(ctx, "/end-users/"+userID, nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByExternalID returns an end-user by external ID
func (u *EndUsersClient) GetByExternalID(ctx context.Context, externalID string) (*EndUser, error) {
	var user EndUser
	err := u.http.Get(ctx, "/end-users/external/"+url.PathEscape(externalID), nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates an end-user manually
func (u *EndUsersClient) Create(ctx context.Context, req *CreateEndUserRequest) (*EndUser, error) {
	var user EndUser
	err := u.http.Post(ctx, "/end-users", req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates an end-user profile
func (u *EndUsersClient) Update(ctx context.Context, userID string, req *UpdateEndUserRequest) (*EndUser, error) {
	var user EndUser
	err := u.http.Patch(ctx, "/end-users/"+userID, req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Delete deletes an end-user
func (u *EndUsersClient) Delete(ctx context.Context, userID string) error {
	return u.http.Delete(ctx, "/end-users/"+userID)
}

// LinkWallet links a wallet to an end-user
func (u *EndUsersClient) LinkWallet(ctx context.Context, userID string, req *LinkWalletRequest) (*EndUser, error) {
	var user EndUser
	err := u.http.Post(ctx, "/end-users/"+userID+"/wallet", req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UnlinkWallet unlinks wallet from an end-user
func (u *EndUsersClient) UnlinkWallet(ctx context.Context, userID string) (*EndUser, error) {
	var user EndUser
	err := u.http.Request(ctx, "DELETE", "/end-users/"+userID+"/wallet", nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetActivity returns user activity/events
func (u *EndUsersClient) GetActivity(ctx context.Context, userID string, page, pageSize int, eventType string) (*UserActivityResponse, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}
	if eventType != "" {
		params.Set("event_type", eventType)
	}

	var response UserActivityResponse
	err := u.http.Get(ctx, "/end-users/"+userID+"/activity", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// AddPoints adds points to a user
func (u *EndUsersClient) AddPoints(ctx context.Context, userID string, points int, reason string) (*EndUser, error) {
	var user EndUser
	err := u.http.Post(ctx, "/end-users/"+userID+"/points", map[string]interface{}{
		"points": points,
		"reason": reason,
	}, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// AddSegment adds a segment to a user
func (u *EndUsersClient) AddSegment(ctx context.Context, userID string, segment string) (*EndUser, error) {
	var user EndUser
	err := u.http.Post(ctx, "/end-users/"+userID+"/segments", map[string]interface{}{
		"segment": segment,
	}, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// RemoveSegment removes a segment from a user
func (u *EndUsersClient) RemoveSegment(ctx context.Context, userID string, segment string) (*EndUser, error) {
	var user EndUser
	err := u.http.Request(ctx, "DELETE", "/end-users/"+userID+"/segments/"+url.PathEscape(segment), nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Merge merges two users
func (u *EndUsersClient) Merge(ctx context.Context, req *MergeUsersRequest) (*EndUser, error) {
	var user EndUser
	err := u.http.Post(ctx, "/end-users/merge", req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
