package proofchain

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// RewardDefinition represents a reward configuration
type RewardDefinition struct {
	ID                     string                 `json:"id"`
	Name                   string                 `json:"name"`
	Slug                   string                 `json:"slug"`
	Description            *string                `json:"description,omitempty"`
	RewardType             string                 `json:"reward_type"`
	Value                  *float64               `json:"value,omitempty"`
	ValueCurrency          *string                `json:"value_currency,omitempty"`
	TokenContractAddress   *string                `json:"token_contract_address,omitempty"`
	TokenSymbol            *string                `json:"token_symbol,omitempty"`
	TokenDecimals          int                    `json:"token_decimals"`
	TokenChain             string                 `json:"token_chain"`
	NFTMintingStrategy     *string                `json:"nft_minting_strategy,omitempty"`
	NFTPreMintCount        *int                   `json:"nft_pre_mint_count,omitempty"`
	NFTMintedPoolCount     int                    `json:"nft_minted_pool_count"`
	NFTIsSoulbound         bool                   `json:"nft_is_soulbound"`
	PassportThresholdID    *string                `json:"passport_threshold_id,omitempty"`
	VoucherProductIDs      []string               `json:"voucher_product_ids,omitempty"`
	VoucherCampaignID      *string                `json:"voucher_campaign_id,omitempty"`
	VoucherLanguage        *string                `json:"voucher_language,omitempty"`
	VoucherMessage         *string                `json:"voucher_message,omitempty"`
	VoucherDeliveryMethods []string               `json:"voucher_delivery_methods,omitempty"`
	TriggerType            string                 `json:"trigger_type"`
	TriggerConfig          map[string]interface{} `json:"trigger_config,omitempty"`
	MaxPerUser             *int                   `json:"max_per_user,omitempty"`
	MaxTotal               *int                   `json:"max_total,omitempty"`
	CurrentIssued          int                    `json:"current_issued"`
	IsActive               bool                   `json:"is_active"`
	IconURL                *string                `json:"icon_url,omitempty"`
	BadgeColor             *string                `json:"badge_color,omitempty"`
	CreatedAt              time.Time              `json:"created_at"`
}

// EarnedReward represents a reward earned by a user
type EarnedReward struct {
	ID             string  `json:"id"`
	UserID         string  `json:"user_id"`
	UserExternalID string  `json:"user_external_id"`
	RewardName     string  `json:"reward_name"`
	RewardType     string  `json:"reward_type"`
	Status         string  `json:"status"`
	NFTTokenID     *int    `json:"nft_token_id,omitempty"`
	NFTTxHash      *string `json:"nft_tx_hash,omitempty"`
	VoucherOrderID *string `json:"voucher_order_id,omitempty"`
	// Voucher delivery tracking (Tremendous does not report redemption).
	VoucherDeliveryMethod  *string    `json:"voucher_delivery_method,omitempty"`
	VoucherDeliveryStatus  *string    `json:"voucher_delivery_status,omitempty"`
	VoucherDeliveryFailure *string    `json:"voucher_delivery_failure,omitempty"`
	VoucherStatusUpdatedAt *Timestamp `json:"voucher_status_updated_at,omitempty"`
	EarnedAt               time.Time  `json:"earned_at"`
	DistributedAt          *time.Time `json:"distributed_at,omitempty"`
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
	Name                   string                 `json:"name"`
	Slug                   string                 `json:"slug"`
	Description            *string                `json:"description,omitempty"`
	CampaignID             *string                `json:"campaign_id,omitempty"`
	RewardType             string                 `json:"reward_type,omitempty"`
	Value                  *float64               `json:"value,omitempty"`
	ValueCurrency          *string                `json:"value_currency,omitempty"`
	TokenContractAddress   *string                `json:"token_contract_address,omitempty"`
	TokenSymbol            *string                `json:"token_symbol,omitempty"`
	TokenDecimals          int                    `json:"token_decimals,omitempty"`
	TokenChain             string                 `json:"token_chain,omitempty"`
	NFTMetadataTemplate    map[string]interface{} `json:"nft_metadata_template,omitempty"`
	NFTIsSoulbound         bool                   `json:"nft_is_soulbound,omitempty"`
	NFTValidityDays        *int                   `json:"nft_validity_days,omitempty"`
	NFTMintingStrategy     string                 `json:"nft_minting_strategy,omitempty"`
	NFTPreMintCount        *int                   `json:"nft_pre_mint_count,omitempty"`
	PassportThresholdID    *string                `json:"passport_threshold_id,omitempty"`
	VoucherProductIDs      []string               `json:"voucher_product_ids,omitempty"`
	VoucherCampaignID      *string                `json:"voucher_campaign_id,omitempty"`
	VoucherLanguage        *string                `json:"voucher_language,omitempty"`
	VoucherMessage         *string                `json:"voucher_message,omitempty"`
	VoucherDeliveryMethods []string               `json:"voucher_delivery_methods,omitempty"`
	TriggerType            string                 `json:"trigger_type,omitempty"`
	TriggerConfig          map[string]interface{} `json:"trigger_config,omitempty"`
	MaxPerUser             *int                   `json:"max_per_user,omitempty"`
	MaxTotal               *int                   `json:"max_total,omitempty"`
	IsActive               bool                   `json:"is_active,omitempty"`
	IconURL                *string                `json:"icon_url,omitempty"`
	BadgeColor             *string                `json:"badge_color,omitempty"`
	IsPublic               bool                   `json:"is_public,omitempty"`
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

// GetUserRewardsOptions are options for filtering a user's rewards
type GetUserRewardsOptions struct {
	Status     string
	RewardType string
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
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	}

	var definitions []RewardDefinition
	err := r.http.Get(ctx, "/rewards/definitions", params, &definitions)
	return definitions, err
}

// GetDefinition returns a reward definition by ID
func (r *RewardsClient) GetDefinition(ctx context.Context, definitionID string) (*RewardDefinition, error) {
	var definition RewardDefinition
	err := r.http.Get(ctx, "/rewards/definitions/"+seg(definitionID), nil, &definition)
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
	err := r.http.Patch(ctx, "/rewards/definitions/"+seg(definitionID), req, &definition)
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

// DeleteDefinition deletes a reward definition
func (r *RewardsClient) DeleteDefinition(ctx context.Context, definitionID string) error {
	return r.http.Delete(ctx, "/rewards/definitions/"+seg(definitionID))
}

// ActivateDefinition activates a reward definition
func (r *RewardsClient) ActivateDefinition(ctx context.Context, definitionID string) (*RewardDefinition, error) {
	var definition RewardDefinition
	err := r.http.Post(ctx, "/rewards/definitions/"+seg(definitionID)+"/activate", nil, &definition)
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

// DeactivateDefinition deactivates a reward definition
func (r *RewardsClient) DeactivateDefinition(ctx context.Context, definitionID string) (*RewardDefinition, error) {
	var definition RewardDefinition
	err := r.http.Post(ctx, "/rewards/definitions/"+seg(definitionID)+"/deactivate", nil, &definition)
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

// GetUserRewards returns earned rewards for a user with full display details.
// Works with both API key and end-user JWT authentication.
func (r *RewardsClient) GetUserRewards(ctx context.Context, userID string, opts *GetUserRewardsOptions) ([]UserReward, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.RewardType != "" {
			params.Set("reward_type", opts.RewardType)
		}
	}
	var rewards []UserReward
	err := r.http.Get(ctx, "/rewards/users/"+seg(userID)+"/rewards", params, &rewards)
	return rewards, err
}

// GetCatalog returns the public reward catalog (active, public definitions).
// Works with both API key and end-user JWT authentication.
func (r *RewardsClient) GetCatalog(ctx context.Context) ([]RewardDefinition, error) {
	var definitions []RewardDefinition
	err := r.http.Get(ctx, "/rewards/catalog", nil, &definitions)
	return definitions, err
}

// ClaimRewardResult represents the result of claiming a lazy-mint NFT reward.
type ClaimRewardResult struct {
	Success       bool   `json:"success"`
	RewardID      string `json:"reward_id"`
	TokenID       *int   `json:"token_id,omitempty"`
	TxHash        string `json:"tx_hash"`
	WalletAddress string `json:"wallet_address"`
}

// ClaimReward claims a lazy-minted NFT reward, minting it to the specified wallet.
// Works with both API key and end-user JWT authentication (scoped to own rewards).
func (r *RewardsClient) ClaimReward(ctx context.Context, earnedRewardID, walletAddress string) (*ClaimRewardResult, error) {
	var result ClaimRewardResult
	err := r.http.Post(ctx, fmt.Sprintf("/rewards/earned/%s/claim?wallet_address=%s", seg(earnedRewardID), url.QueryEscape(walletAddress)), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Voucher delivery methods for VoucherDelivery.Method and
// RewardDefinition.VoucherDeliveryMethods.
const (
	VoucherDeliveryLink  = "link"
	VoucherDeliveryEmail = "email"
	VoucherDeliverySMS   = "sms"
)

// VoucherDelivery selects how a voucher is delivered at claim time. Email is
// required for VoucherDeliveryEmail, Phone (E.164) for VoucherDeliverySMS; Name
// is optional. Contact details are sent to the provider and never stored by
// Proofchain; do not log or cache them.
type VoucherDelivery struct {
	Method string `json:"method"`
	Email  string `json:"email,omitempty"`
	Phone  string `json:"phone,omitempty"`
	Name   string `json:"name,omitempty"`
}

// ClaimVoucherResult carries the claim outcome. RedemptionURL is set only for
// link delivery (nil for email/SMS) and is a secret: open it directly and never
// persist, log, or cache it. DeliveryStatus is one of pending, delivered,
// failed, canceled, flagged or blocked.
type ClaimVoucherResult struct {
	Success        bool       `json:"success"`
	RewardID       string     `json:"reward_id"`
	Status         string     `json:"status"`
	RedemptionURL  *string    `json:"redemption_url,omitempty"`
	ExpiresAt      *Timestamp `json:"expires_at,omitempty"`
	DeliveryMethod string     `json:"delivery_method,omitempty"`
	DeliveryStatus string     `json:"delivery_status,omitempty"`
}

// VoucherProductSKU is a face-value band a voucher product supports in one currency.
type VoucherProductSKU struct {
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	CurrencyCode string  `json:"currency_code"`
}

// VoucherAllFeeFree is a sentinel for voucher_product_ids (case-insensitive):
// on create/update the server replaces it with every fee-free product offered in
// the definition's value currency whose range fits the value. Several product
// IDs let the fan choose on the redemption page.
const VoucherAllFeeFree = "ALL_FEE_FREE"

// VoucherProduct is a gift-card product selectable for voucher rewards.
type VoucherProduct struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Subcategory *string `json:"subcategory,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	Description *string `json:"description,omitempty"`
	// Disclosure is provider legal HTML; it is untrusted and must be sanitized
	// before rendering.
	Disclosure        *string             `json:"disclosure,omitempty"`
	UsageInstructions *string             `json:"usage_instructions,omitempty"`
	FeeFree           bool                `json:"fee_free"`
	CurrencyCodes     []string            `json:"currency_codes"`
	Countries         []string            `json:"countries"`
	SKUs              []VoucherProductSKU `json:"skus"`
}

// ListVoucherProductsParams filters the voucher catalog by one ISO-3166 alpha-2
// country (e.g. "GB") and/or one ISO-4217 currency (e.g. "GBP"). Category and
// Subcategory are comma-separated lists (e.g. "merchant_card,charity"). FeeFree
// nil returns all products; a pointer to false sends fee_free=false.
type ListVoucherProductsParams struct {
	Country     string
	Currency    string
	Category    string
	Subcategory string
	FeeFree     *bool
}

// ClaimVoucher claims an issued voucher reward. A nil delivery sends no body
// (link delivery: each call yields a fresh RedemptionURL; never persist, log,
// or cache it). Email/SMS delivery orders the gift card at claim time and the
// provider sends it to the fan; RedemptionURL is then nil.
func (r *RewardsClient) ClaimVoucher(ctx context.Context, earnedRewardID string, delivery *VoucherDelivery) (*ClaimVoucherResult, error) {
	var body interface{}
	if delivery != nil {
		body = map[string]*VoucherDelivery{"delivery": delivery}
	}
	var result ClaimVoucherResult
	if err := r.http.Post(ctx, "/rewards/earned/"+seg(earnedRewardID)+"/claim", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListVoucherProducts lists gift-card products available for voucher rewards.
func (r *RewardsClient) ListVoucherProducts(ctx context.Context, opts *ListVoucherProductsParams) ([]VoucherProduct, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Country != "" {
			params.Set("country", opts.Country)
		}
		if opts.Currency != "" {
			params.Set("currency", opts.Currency)
		}
		if opts.Category != "" {
			params.Set("category", opts.Category)
		}
		if opts.Subcategory != "" {
			params.Set("subcategory", opts.Subcategory)
		}
		if opts.FeeFree != nil {
			params.Set("fee_free", strconv.FormatBool(*opts.FeeFree))
		}
	}
	var products []VoucherProduct
	err := r.http.Get(ctx, "/rewards/voucher-products", params, &products)
	return products, err
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
	err := r.http.Post(ctx, "/rewards/earned/"+seg(earnedRewardID)+"/distribute", nil, &reward)
	if err != nil {
		return nil, err
	}
	return &reward, nil
}

// ListAssets returns assets for a reward definition
func (r *RewardsClient) ListAssets(ctx context.Context, definitionID string) ([]RewardAsset, error) {
	var assets []RewardAsset
	err := r.http.Get(ctx, "/rewards/definitions/"+seg(definitionID)+"/assets", nil, &assets)
	return assets, err
}
