package proofchain

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// RewardDefinition represents a reward configuration
type RewardDefinition struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Slug                 string                 `json:"slug"`
	Description          *string                `json:"description,omitempty"`
	RewardType           string                 `json:"reward_type"`
	Value                *float64               `json:"value,omitempty"`
	ValueCurrency        *string                `json:"value_currency,omitempty"`
	TokenContractAddress *string                `json:"token_contract_address,omitempty"`
	TokenSymbol          *string                `json:"token_symbol,omitempty"`
	TokenDecimals        int                    `json:"token_decimals"`
	TokenChain           string                 `json:"token_chain"`
	NFTMintingStrategy   *string                `json:"nft_minting_strategy,omitempty"`
	NFTPreMintCount      *int                   `json:"nft_pre_mint_count,omitempty"`
	NFTMintedPoolCount   int                    `json:"nft_minted_pool_count"`
	NFTIsSoulbound       bool                   `json:"nft_is_soulbound"`
	PassportThresholdID  *string                `json:"passport_threshold_id,omitempty"`
	TriggerType          string                 `json:"trigger_type"`
	TriggerConfig        map[string]interface{} `json:"trigger_config,omitempty"`
	MaxPerUser           *int                   `json:"max_per_user,omitempty"`
	MaxTotal             *int                   `json:"max_total,omitempty"`
	CurrentIssued        int                    `json:"current_issued"`
	IsActive             bool                   `json:"is_active"`
	IconURL              *string                `json:"icon_url,omitempty"`
	BadgeColor           *string                `json:"badge_color,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
}

// EarnedReward represents a reward earned by a user
type EarnedReward struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	UserExternalID string     `json:"user_external_id"`
	RewardName     string     `json:"reward_name"`
	RewardType     string     `json:"reward_type"`
	Status         string     `json:"status"`
	NFTTokenID     *int       `json:"nft_token_id,omitempty"`
	NFTTxHash      *string    `json:"nft_tx_hash,omitempty"`
	EarnedAt       time.Time  `json:"earned_at"`
	DistributedAt  *time.Time `json:"distributed_at,omitempty"`
}

// RewardAsset represents an asset for a reward
type RewardAsset struct {
	ID           string                 `json:"id"`
	DefinitionID string                 `json:"definition_id"`
	AssetType    string                 `json:"asset_type"`
	FilePath     string                 `json:"file_path"`
	MimeType     string                 `json:"mime_type"`
	FileSize     int64                  `json:"file_size"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// Request types
type CreateRewardDefinitionRequest struct {
	Name                 string                 `json:"name"`
	Slug                 string                 `json:"slug"`
	Description          *string                `json:"description,omitempty"`
	CampaignID           *string                `json:"campaign_id,omitempty"`
	RewardType           string                 `json:"reward_type,omitempty"`
	Value                *float64               `json:"value,omitempty"`
	ValueCurrency        *string                `json:"value_currency,omitempty"`
	TokenContractAddress *string                `json:"token_contract_address,omitempty"`
	TokenSymbol          *string                `json:"token_symbol,omitempty"`
	TokenDecimals        int                    `json:"token_decimals,omitempty"`
	TokenChain           string                 `json:"token_chain,omitempty"`
	NFTMetadataTemplate  map[string]interface{} `json:"nft_metadata_template,omitempty"`
	NFTIsSoulbound       bool                   `json:"nft_is_soulbound,omitempty"`
	NFTValidityDays      *int                   `json:"nft_validity_days,omitempty"`
	NFTMintingStrategy   string                 `json:"nft_minting_strategy,omitempty"`
	NFTPreMintCount      *int                   `json:"nft_pre_mint_count,omitempty"`
	PassportThresholdID  *string                `json:"passport_threshold_id,omitempty"`
	TriggerType          string                 `json:"trigger_type,omitempty"`
	TriggerConfig        map[string]interface{} `json:"trigger_config,omitempty"`
	MaxPerUser           *int                   `json:"max_per_user,omitempty"`
	MaxTotal             *int                   `json:"max_total,omitempty"`
	IsActive             bool                   `json:"is_active,omitempty"`
	IconURL              *string                `json:"icon_url,omitempty"`
	BadgeColor           *string                `json:"badge_color,omitempty"`
	IsPublic             bool                   `json:"is_public,omitempty"`
}

type ManualRewardRequest struct {
	DefinitionID          string                 `json:"definition_id"`
	UserIDs               []string               `json:"user_ids"`
	TriggerData           map[string]interface{} `json:"trigger_data,omitempty"`
	DistributeImmediately bool                   `json:"distribute_immediately,omitempty"`
}

type ListRewardsOptions struct {
	IsActive   *bool
	RewardType string
	Limit      int
	Offset     int
}

// RewardsClient provides reward operations
type RewardsClient struct {
	http *HTTPClient
}

// NewRewardsClient creates a new rewards client
func NewRewardsClient(http *HTTPClient) *RewardsClient {
	return &RewardsClient{http: http}
}

// ListDefinitions returns reward definitions
func (r *RewardsClient) ListDefinitions(ctx context.Context, opts *ListRewardsOptions) ([]RewardDefinition, error) {
	params := url.Values{}
	if opts != nil {
		if opts.IsActive != nil {
			params.Set("is_active", fmt.Sprintf("%t", *opts.IsActive))
		}
		if opts.RewardType != "" {
			params.Set("reward_type", opts.RewardType)
		}
	}

	var definitions []RewardDefinition
	err := r.http.Get(ctx, "/rewards/definitions", params, &definitions)
	return definitions, err
}

// GetDefinition returns a reward definition by ID
func (r *RewardsClient) GetDefinition(ctx context.Context, definitionID string) (*RewardDefinition, error) {
	var definition RewardDefinition
	err := r.http.Get(ctx, "/rewards/definitions/"+definitionID, nil, &definition)
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

// CreateDefinition creates a reward definition
func (r *RewardsClient) CreateDefinition(ctx context.Context, req *CreateRewardDefinitionRequest) (*RewardDefinition, error) {
	var definition RewardDefinition
	err := r.http.Post(ctx, "/rewards/definitions", req, &definition)
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

// UpdateDefinition updates a reward definition
func (r *RewardsClient) UpdateDefinition(ctx context.Context, definitionID string, req *CreateRewardDefinitionRequest) (*RewardDefinition, error) {
	var definition RewardDefinition
	err := r.http.Patch(ctx, "/rewards/definitions/"+definitionID, req, &definition)
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

// DeleteDefinition deletes a reward definition
func (r *RewardsClient) DeleteDefinition(ctx context.Context, definitionID string) error {
	return r.http.Delete(ctx, "/rewards/definitions/"+definitionID)
}

// ActivateDefinition activates a reward definition
func (r *RewardsClient) ActivateDefinition(ctx context.Context, definitionID string) (*RewardDefinition, error) {
	var definition RewardDefinition
	err := r.http.Post(ctx, "/rewards/definitions/"+definitionID+"/activate", nil, &definition)
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

// DeactivateDefinition deactivates a reward definition
func (r *RewardsClient) DeactivateDefinition(ctx context.Context, definitionID string) (*RewardDefinition, error) {
	var definition RewardDefinition
	err := r.http.Post(ctx, "/rewards/definitions/"+definitionID+"/deactivate", nil, &definition)
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

// ListEarned returns earned rewards
func (r *RewardsClient) ListEarned(ctx context.Context, userID, definitionID, status string, limit, offset int) ([]EarnedReward, error) {
	params := url.Values{}
	if userID != "" {
		params.Set("user_id", userID)
	}
	if definitionID != "" {
		params.Set("definition_id", definitionID)
	}
	if status != "" {
		params.Set("status", status)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", offset))
	}

	var rewards []EarnedReward
	err := r.http.Get(ctx, "/rewards/earned", params, &rewards)
	return rewards, err
}

// GetUserRewards returns earned rewards for a user
func (r *RewardsClient) GetUserRewards(ctx context.Context, userID string) ([]EarnedReward, error) {
	var rewards []EarnedReward
	err := r.http.Get(ctx, "/rewards/earned/user/"+userID, nil, &rewards)
	return rewards, err
}

// AwardManual manually awards rewards to users
func (r *RewardsClient) AwardManual(ctx context.Context, req *ManualRewardRequest) ([]EarnedReward, error) {
	var rewards []EarnedReward
	err := r.http.Post(ctx, "/rewards/award", req, &rewards)
	return rewards, err
}

// DistributePending distributes a pending reward
func (r *RewardsClient) DistributePending(ctx context.Context, earnedRewardID string) (*EarnedReward, error) {
	var reward EarnedReward
	err := r.http.Post(ctx, "/rewards/earned/"+earnedRewardID+"/distribute", nil, &reward)
	if err != nil {
		return nil, err
	}
	return &reward, nil
}

// ListAssets returns assets for a reward definition
func (r *RewardsClient) ListAssets(ctx context.Context, definitionID string) ([]RewardAsset, error) {
	var assets []RewardAsset
	err := r.http.Get(ctx, "/rewards/definitions/"+definitionID+"/assets", nil, &assets)
	return assets, err
}
