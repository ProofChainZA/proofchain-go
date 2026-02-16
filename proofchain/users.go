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

// EndUser represents an end-user discovered from events.
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

// EndUserListResponse is a paginated list of users.
type EndUserListResponse struct {
	Users    []EndUser `json:"users"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	HasMore  bool      `json:"has_more"`
}

// UserActivityResponse is the activity summary for a user.
type UserActivityResponse struct {
	UserID       string                    `json:"user_id"`
	ExternalID   string                    `json:"external_id"`
	TotalEvents  int                       `json:"total_events"`
	EventsByType map[string]int            `json:"events_by_type"`
	EventsByDay  []map[string]interface{}   `json:"events_by_day"`
	RecentEvents []map[string]interface{}   `json:"recent_events"`
	RewardsEarned  int                     `json:"rewards_earned"`
	RewardsPending int                     `json:"rewards_pending"`
}

// UserReward represents a single earned reward.
type UserReward struct {
	ID            string  `json:"id"`
	RewardName    string  `json:"reward_name"`
	RewardType    string  `json:"reward_type"`
	Value         *float64 `json:"value,omitempty"`
	ValueCurrency *string `json:"value_currency,omitempty"`
	Status        string  `json:"status"`
	EarnedAt      *string `json:"earned_at,omitempty"`
	DistributedAt *string `json:"distributed_at,omitempty"`
	NFTTokenID    *int    `json:"nft_token_id,omitempty"`
	NFTTxHash     *string `json:"nft_tx_hash,omitempty"`
}

// UserRewardsResponse is a paginated list of user rewards.
type UserRewardsResponse struct {
	UserID   string       `json:"user_id"`
	Rewards  []UserReward `json:"rewards"`
	Total    int          `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	HasMore  bool         `json:"has_more"`
}

// WalletCreationResult is the response from wallet creation/registration.
type WalletCreationResult struct {
	Success       bool    `json:"success"`
	UserID        string  `json:"user_id"`
	WalletAddress string  `json:"wallet_address"`
	WalletType    *string `json:"wallet_type,omitempty"`
	Network       *string `json:"network,omitempty"`
	Source        string  `json:"source"`
}

// GDPRDeletionResponse is the response from a GDPR deletion.
type GDPRDeletionResponse struct {
	Success        bool           `json:"success"`
	UserID         string         `json:"user_id"`
	ExternalID     string         `json:"external_id"`
	DeletedRecords map[string]int `json:"deleted_records"`
	MerkleWarning  *string        `json:"merkle_warning,omitempty"`
	AuditID        *string        `json:"audit_id,omitempty"`
}

// GDPRPreviewResponse is the response from a GDPR deletion preview.
type GDPRPreviewResponse struct {
	User          map[string]interface{} `json:"user"`
	WouldDelete   map[string]int         `json:"would_delete"`
	MerkleWarning *string                `json:"merkle_warning,omitempty"`
}

// PointsResult is the response from adding/subtracting points.
type PointsResult struct {
	UserID         string `json:"user_id"`
	PointsAdded    int    `json:"points_added"`
	NewBalance     int    `json:"new_balance"`
	LifetimePoints int    `json:"lifetime_points"`
	Reason         string `json:"reason"`
}

// =============================================================================
// Request types
// =============================================================================

// CreateEndUserRequest creates a new end-user.
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
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// UpdateEndUserRequest updates an end-user profile.
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

// ListEndUsersOptions configures the List query.
type ListEndUsersOptions struct {
	Page      int
	PageSize  int
	Search    string
	Status    string
	Segment   string
	HasWallet *bool
	MinEvents *int
	SortBy    string
	SortOrder string
}

// LinkWalletRequest links an external wallet to a user.
type LinkWalletRequest struct {
	WalletAddress string  `json:"wallet_address"`
	WalletSource  *string `json:"wallet_source,omitempty"`
	Signature     *string `json:"signature,omitempty"`
}

// CreateUserWalletRequest creates a CDP wallet for an end-user.
type CreateUserWalletRequest struct {
	WalletType string `json:"wallet_type,omitempty"`
	Network    string `json:"network,omitempty"`
}

// RegisterWalletRequest registers an external wallet.
type RegisterWalletRequest struct {
	WalletAddress string  `json:"wallet_address"`
	Signature     *string `json:"signature,omitempty"`
}

// MergeUsersRequest merges source users into a target user.
type MergeUsersRequest struct {
	SourceUserIDs []string `json:"source_user_ids"`
	TargetUserID  string   `json:"target_user_id"`
}

// GDPRDeletionRequest requests permanent deletion of user data.
type GDPRDeletionRequest struct {
	Confirm      bool    `json:"confirm"`
	DeleteEvents *bool   `json:"delete_events,omitempty"`
	DeleteWallets *bool  `json:"delete_wallets,omitempty"`
	Reason       *string `json:"reason,omitempty"`
}

// =============================================================================
// Client
// =============================================================================

// EndUsersClient provides end-user operations.
type EndUsersClient struct {
	http *HTTPClient
}

// NewEndUsersClient creates a new end-users client.
func NewEndUsersClient(http *HTTPClient) *EndUsersClient {
	return &EndUsersClient{http: http}
}

// List returns paginated end-users.
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
		if opts.MinEvents != nil {
			params.Set("min_events", fmt.Sprintf("%d", *opts.MinEvents))
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

// Get returns an end-user by internal UUID.
func (u *EndUsersClient) Get(ctx context.Context, userID string) (*EndUser, error) {
	var user EndUser
	err := u.http.Get(ctx, "/end-users/"+userID, nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByExternalID returns an end-user by external ID.
func (u *EndUsersClient) GetByExternalID(ctx context.Context, externalID string) (*EndUser, error) {
	var user EndUser
	err := u.http.Get(ctx, "/end-users/by-external/"+url.PathEscape(externalID), nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates an end-user manually.
func (u *EndUsersClient) Create(ctx context.Context, req *CreateEndUserRequest) (*EndUser, error) {
	var user EndUser
	err := u.http.Post(ctx, "/end-users", req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates an end-user profile by internal UUID.
func (u *EndUsersClient) Update(ctx context.Context, userID string, req *UpdateEndUserRequest) (*EndUser, error) {
	var user EndUser
	err := u.http.Patch(ctx, "/end-users/"+userID, req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateByExternalID updates an end-user profile by external ID.
func (u *EndUsersClient) UpdateByExternalID(ctx context.Context, externalID string, req *UpdateEndUserRequest) (*EndUser, error) {
	var user EndUser
	err := u.http.Patch(ctx, "/end-users/by-external/"+url.PathEscape(externalID), req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// LinkWallet links a wallet to an end-user by external ID.
func (u *EndUsersClient) LinkWallet(ctx context.Context, externalID string, req *LinkWalletRequest) (*EndUser, error) {
	var user EndUser
	err := u.http.Post(ctx, "/end-users/by-external/"+url.PathEscape(externalID)+"/link-wallet", req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateWallet creates a CDP (custodial) wallet for an end-user by external ID.
func (u *EndUsersClient) CreateWallet(ctx context.Context, externalID string, req *CreateUserWalletRequest) (*WalletCreationResult, error) {
	var result WalletCreationResult
	err := u.http.Post(ctx, "/end-users/by-external/"+url.PathEscape(externalID)+"/create-wallet", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RegisterWallet registers an external wallet for an end-user by external ID.
func (u *EndUsersClient) RegisterWallet(ctx context.Context, externalID string, req *RegisterWalletRequest) (*WalletCreationResult, error) {
	var result WalletCreationResult
	err := u.http.Post(ctx, "/end-users/by-external/"+url.PathEscape(externalID)+"/register-wallet", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetActivity returns user activity summary by external ID.
func (u *EndUsersClient) GetActivity(ctx context.Context, externalID string, days int) (*UserActivityResponse, error) {
	params := url.Values{}
	if days > 0 {
		params.Set("days", fmt.Sprintf("%d", days))
	}

	var response UserActivityResponse
	err := u.http.Get(ctx, "/end-users/by-external/"+url.PathEscape(externalID)+"/activity", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// AddPoints adds or subtracts points from a user by external ID.
func (u *EndUsersClient) AddPoints(ctx context.Context, externalID string, points int, reason string) (*PointsResult, error) {
	params := url.Values{}
	params.Set("points", fmt.Sprintf("%d", points))
	if reason != "" {
		params.Set("reason", reason)
	}

	var result PointsResult
	path := "/end-users/by-external/" + url.PathEscape(externalID) + "/add-points?" + params.Encode()
	err := u.http.Post(ctx, path, map[string]interface{}{}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRewards returns rewards earned by a user by external ID.
func (u *EndUsersClient) GetRewards(ctx context.Context, externalID string, status string, page, pageSize int) (*UserRewardsResponse, error) {
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	var response UserRewardsResponse
	err := u.http.Get(ctx, "/end-users/by-external/"+url.PathEscape(externalID)+"/rewards", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetRewardsByInternalID returns rewards earned by a user by internal UUID.
func (u *EndUsersClient) GetRewardsByInternalID(ctx context.Context, userID string, status string, page, pageSize int) (*UserRewardsResponse, error) {
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	var response UserRewardsResponse
	err := u.http.Get(ctx, "/end-users/"+userID+"/rewards", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Merge merges source users into a target user.
func (u *EndUsersClient) Merge(ctx context.Context, req *MergeUsersRequest) (*EndUser, error) {
	var user EndUser
	err := u.http.Post(ctx, "/end-users/merge", req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// =============================================================================
// Convenience methods
// =============================================================================

// UpdateAttributes merges the provided attributes into the user's existing attributes by external ID.
func (u *EndUsersClient) UpdateAttributes(ctx context.Context, externalID string, attributes map[string]interface{}) (*EndUser, error) {
	return u.UpdateByExternalID(ctx, externalID, &UpdateEndUserRequest{
		Attributes: attributes,
	})
}

// RemoveAttributes removes specific attribute keys from a user by external ID.
// Fetches the current user, removes the keys, and saves.
func (u *EndUsersClient) RemoveAttributes(ctx context.Context, externalID string, keys []string) (*EndUser, error) {
	user, err := u.GetByExternalID(ctx, externalID)
	if err != nil {
		return nil, err
	}

	attributes := make(map[string]interface{})
	for k, v := range user.Attributes {
		attributes[k] = v
	}
	for _, key := range keys {
		delete(attributes, key)
	}

	return u.UpdateByExternalID(ctx, externalID, &UpdateEndUserRequest{
		Attributes: attributes,
	})
}

// SetProfile sets profile fields on a user by external ID.
// Only the non-nil fields in the request are updated.
func (u *EndUsersClient) SetProfile(ctx context.Context, externalID string, req *UpdateEndUserRequest) (*EndUser, error) {
	return u.UpdateByExternalID(ctx, externalID, req)
}

// EnsureWallet guarantees a user has a wallet, creating one if they don't.
//
// This is the recommended way to guarantee a wallet exists before performing
// any wallet-dependent operation (e.g. attestation, on-chain claims, token
// transfers). If the user already has a wallet, returns immediately with the
// existing address. If not, creates a CDP wallet with the specified options.
func (u *EndUsersClient) EnsureWallet(ctx context.Context, externalID string, req *CreateUserWalletRequest) (*WalletCreationResult, error) {
	user, err := u.GetByExternalID(ctx, externalID)
	if err != nil {
		return nil, err
	}

	if user.WalletAddress != nil && *user.WalletAddress != "" {
		result := &WalletCreationResult{
			Success:       true,
			UserID:        externalID,
			WalletAddress: *user.WalletAddress,
			Source:        "existing",
		}
		if user.WalletSource != nil {
			result.WalletType = user.WalletSource
		}
		return result, nil
	}

	if req == nil {
		req = &CreateUserWalletRequest{}
	}
	return u.CreateWallet(ctx, externalID, req)
}

// =============================================================================
// GDPR
// =============================================================================

// GDPRPreview previews what would be deleted for a GDPR request.
func (u *EndUsersClient) GDPRPreview(ctx context.Context, userID string) (*GDPRPreviewResponse, error) {
	var result GDPRPreviewResponse
	err := u.http.Get(ctx, "/end-users/"+userID+"/gdpr/preview", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GDPRDelete permanently deletes all user data (Right to Be Forgotten).
func (u *EndUsersClient) GDPRDelete(ctx context.Context, userID string, req *GDPRDeletionRequest) (*GDPRDeletionResponse, error) {
	var result GDPRDeletionResponse
	err := u.http.Request(ctx, "DELETE", "/end-users/"+userID+"/gdpr", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
