package proofchain

import (
	"context"
	"fmt"
	"net/url"
)

// Wallet represents a CDP wallet
type Wallet struct {
	WalletID       string                 `json:"wallet_id"`
	Address        string                 `json:"address"`
	UserID         string                 `json:"user_id"`
	WalletType     string                 `json:"wallet_type"`
	Network        string                 `json:"network"`
	Name           *string                `json:"name,omitempty"`
	Status         string                 `json:"status"`
	CreatedAt      string                 `json:"created_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	SupportsExport *bool                  `json:"supports_export,omitempty"`
	OwnerWalletID  *string                `json:"owner_wallet_id,omitempty"`
	IsDeployed     *bool                  `json:"is_deployed,omitempty"`
}

// DualWallets represents EOA + Smart Account pair
type DualWallets struct {
	UserID      string `json:"user_id"`
	AssetWallet Wallet `json:"asset_wallet"`
	SmartWallet Wallet `json:"smart_wallet"`
	Network     string `json:"network"`
}

// WalletBalance represents wallet token balances
type WalletBalance struct {
	WalletID string         `json:"wallet_id"`
	Address  string         `json:"address"`
	Network  string         `json:"network"`
	Balances []TokenBalance `json:"balances"`
}

// TokenBalance represents a single token balance
type TokenBalance struct {
	Token    string   `json:"token"`
	Symbol   string   `json:"symbol"`
	Balance  string   `json:"balance"`
	Decimals int      `json:"decimals"`
	USDValue *float64 `json:"usd_value,omitempty"`
}

// NFT represents an NFT in a wallet
type NFT struct {
	ID              string                 `json:"id"`
	WalletID        string                 `json:"wallet_id"`
	ContractAddress string                 `json:"contract_address"`
	TokenID         string                 `json:"token_id"`
	Network         string                 `json:"network"`
	Name            *string                `json:"name,omitempty"`
	Description     *string                `json:"description,omitempty"`
	ImageURL        *string                `json:"image_url,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Source          string                 `json:"source"`
	AcquiredAt      string                 `json:"acquired_at"`
}

// SwapQuote represents a token swap quote
type SwapQuote struct {
	FromToken          string `json:"from_token"`
	ToToken            string `json:"to_token"`
	FromAmount         string `json:"from_amount"`
	ToAmount           string `json:"to_amount"`
	MinToAmount        string `json:"min_to_amount"`
	ExchangeRate       string `json:"exchange_rate"`
	QuoteID            string `json:"quote_id"`
	SlippageBps        int    `json:"slippage_bps"`
	Network            string `json:"network"`
	LiquidityAvailable bool   `json:"liquidity_available"`
	ExpiresAt          string `json:"expires_at"`
}

// SwapResult represents a completed swap
type SwapResult struct {
	TransactionHash string `json:"transaction_hash"`
	Status          string `json:"status"`
	FromToken       string `json:"from_token"`
	ToToken         string `json:"to_token"`
	FromAmount      string `json:"from_amount"`
	ToAmount        string `json:"to_amount"`
	GasUsed         string `json:"gas_used"`
}

// WalletStats represents wallet statistics
type WalletStats struct {
	TotalWallets int            `json:"total_wallets"`
	ByType       map[string]int `json:"by_type"`
	ByNetwork    map[string]int `json:"by_network"`
}

// ComprehensiveWalletInfo contains all wallet data in one response
type ComprehensiveWalletInfo struct {
	WalletID       string                 `json:"wallet_id"`
	Address        string                 `json:"address"`
	UserID         string                 `json:"user_id"`
	WalletType     string                 `json:"wallet_type"`
	Network        string                 `json:"network"`
	Name           *string                `json:"name,omitempty"`
	Status         string                 `json:"status"`
	SupportsExport bool                   `json:"supports_export"`
	CreatedAt      string                 `json:"created_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	OwnerWalletID  *string                `json:"owner_wallet_id,omitempty"`
	OwnerAddress   *string                `json:"owner_address,omitempty"`
	IsDeployed     *bool                  `json:"is_deployed,omitempty"`
	Balances       *WalletBalanceInfo     `json:"balances,omitempty"`
	NFTs           *WalletNFTInfo         `json:"nfts,omitempty"`
	Activity       *WalletActivityInfo    `json:"activity,omitempty"`
}

// WalletBalanceInfo contains token balance data
type WalletBalanceInfo struct {
	Address   string         `json:"address"`
	Network   string         `json:"network"`
	Tokens    []TokenBalance `json:"tokens"`
	FetchedAt string         `json:"fetched_at"`
}

// WalletNFTInfo contains NFT holdings
type WalletNFTInfo struct {
	Total int   `json:"total"`
	Items []NFT `json:"items"`
}

// WalletActivityInfo contains recent activity
type WalletActivityInfo struct {
	RecentSwaps []SwapResult `json:"recent_swaps"`
	TotalSwaps  int          `json:"total_swaps"`
}

// UserWalletSummary contains summary of all user wallets
type UserWalletSummary struct {
	UserID       string              `json:"user_id"`
	TotalWallets int                 `json:"total_wallets"`
	TotalNFTs    int                 `json:"total_nfts"`
	TotalSwaps   int                 `json:"total_swaps"`
	Wallets      []WalletSummaryItem `json:"wallets"`
}

// WalletSummaryItem represents a wallet in the user summary
type WalletSummaryItem struct {
	WalletID       string             `json:"wallet_id"`
	Address        string             `json:"address"`
	WalletType     string             `json:"wallet_type"`
	Network        string             `json:"network"`
	Name           *string            `json:"name,omitempty"`
	SupportsExport bool               `json:"supports_export"`
	CreatedAt      string             `json:"created_at"`
	OwnerWalletID  *string            `json:"owner_wallet_id,omitempty"`
	IsDeployed     *bool              `json:"is_deployed,omitempty"`
	Balances       *WalletBalanceInfo `json:"balances,omitempty"`
}

// GetInfoOptions controls what data is included in GetInfo
type GetInfoOptions struct {
	IncludeBalances bool
	IncludeNFTs     bool
	IncludeActivity bool
}

// Request types
type CreateWalletRequest struct {
	UserID     string                 `json:"user_id"`
	WalletType string                 `json:"wallet_type,omitempty"`
	Network    string                 `json:"network,omitempty"`
	Name       *string                `json:"name,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type CreateDualWalletsRequest struct {
	UserID     string  `json:"user_id"`
	Network    string  `json:"network,omitempty"`
	NamePrefix *string `json:"name_prefix,omitempty"`
}

type SwapQuoteRequest struct {
	FromToken   string `json:"from_token"`
	ToToken     string `json:"to_token"`
	FromAmount  string `json:"from_amount"`
	Network     string `json:"network,omitempty"`
	SlippageBps int    `json:"slippage_bps,omitempty"`
}

type ExecuteSwapRequest struct {
	WalletID    string `json:"wallet_id"`
	FromToken   string `json:"from_token"`
	ToToken     string `json:"to_token"`
	FromAmount  string `json:"from_amount"`
	Network     string `json:"network,omitempty"`
	SlippageBps int    `json:"slippage_bps,omitempty"`
}

type AddNFTRequest struct {
	ContractAddress string                 `json:"contract_address"`
	TokenID         string                 `json:"token_id"`
	Network         string                 `json:"network"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Source          string                 `json:"source,omitempty"`
}

// TransferRequest represents a token transfer request
type TransferRequest struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	Token       string `json:"token,omitempty"`
	Network     string `json:"network,omitempty"`
}

// TransferResult represents the result of a token transfer
type TransferResult struct {
	TxHash  string `json:"tx_hash"`
	From    string `json:"from"`
	To      string `json:"to"`
	Amount  string `json:"amount"`
	Token   string `json:"token"`
	Network string `json:"network"`
	Status  string `json:"status"`
}

// WalletClient provides wallet operations
type WalletClient struct {
	http *HTTPClient
}

// NewWalletClient creates a new wallet client
func NewWalletClient(http *HTTPClient) *WalletClient {
	return &WalletClient{http: http}
}

// Create creates a single wallet
func (w *WalletClient) Create(ctx context.Context, req *CreateWalletRequest) (*Wallet, error) {
	var wallet Wallet
	err := w.http.Post(ctx, "/wallets", req, &wallet)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// Get returns a wallet by ID
func (w *WalletClient) Get(ctx context.Context, walletID string) (*Wallet, error) {
	var wallet Wallet
	err := w.http.Get(ctx, "/wallets/"+walletID, nil, &wallet)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// ListByUser returns wallets for a user
func (w *WalletClient) ListByUser(ctx context.Context, userID string) ([]Wallet, error) {
	var wallets []Wallet
	err := w.http.Get(ctx, "/wallets/user/"+url.PathEscape(userID), nil, &wallets)
	return wallets, err
}

// Stats returns wallet statistics
func (w *WalletClient) Stats(ctx context.Context) (*WalletStats, error) {
	var stats WalletStats
	err := w.http.Get(ctx, "/wallets/stats", nil, &stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// CreateDual creates dual wallets (EOA + Smart Account)
func (w *WalletClient) CreateDual(ctx context.Context, req *CreateDualWalletsRequest) (*DualWallets, error) {
	var dual DualWallets
	err := w.http.Post(ctx, "/wallets/dual", req, &dual)
	if err != nil {
		return nil, err
	}
	return &dual, nil
}

// CreateDualBulk creates dual wallets for multiple users
func (w *WalletClient) CreateDualBulk(ctx context.Context, userIDs []string, network string) ([]DualWallets, error) {
	var duals []DualWallets
	err := w.http.Post(ctx, "/wallets/dual/bulk", map[string]interface{}{
		"user_ids": userIDs,
		"network":  network,
	}, &duals)
	return duals, err
}

// GetBalance returns wallet balance
func (w *WalletClient) GetBalance(ctx context.Context, walletID string) (*WalletBalance, error) {
	var balance WalletBalance
	err := w.http.Get(ctx, "/wallets/"+walletID+"/balance", nil, &balance)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// GetInfo returns comprehensive wallet information in a single call.
// Returns everything about a wallet: details, balances, NFTs, and activity.
func (w *WalletClient) GetInfo(ctx context.Context, walletID string, opts *GetInfoOptions) (*ComprehensiveWalletInfo, error) {
	params := url.Values{}
	if opts != nil {
		params.Set("include_balances", fmt.Sprintf("%t", opts.IncludeBalances))
		params.Set("include_nfts", fmt.Sprintf("%t", opts.IncludeNFTs))
		params.Set("include_activity", fmt.Sprintf("%t", opts.IncludeActivity))
	}

	path := "/wallets/" + walletID + "/info"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var info ComprehensiveWalletInfo
	err := w.http.Get(ctx, path, nil, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetUserSummary returns comprehensive summary of all wallets for a user.
// Aggregates data across all user's wallets (EOA + Smart).
func (w *WalletClient) GetUserSummary(ctx context.Context, userID string, includeBalances bool) (*UserWalletSummary, error) {
	path := fmt.Sprintf("/wallets/user/%s/summary?include_balances=%t", url.PathEscape(userID), includeBalances)

	var summary UserWalletSummary
	err := w.http.Get(ctx, path, nil, &summary)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// ExportKey exports private key for an EOA wallet
func (w *WalletClient) ExportKey(ctx context.Context, walletID string) (string, error) {
	var result struct {
		PrivateKey string `json:"private_key"`
		Warning    string `json:"warning"`
	}
	err := w.http.Post(ctx, "/wallets/"+walletID+"/export-key", map[string]interface{}{
		"acknowledge_warning": true,
	}, &result)
	if err != nil {
		return "", err
	}
	return result.PrivateKey, nil
}

// Transfer sends tokens from one address to another.
// Returns the transaction result with hash and status.
func (w *WalletClient) Transfer(ctx context.Context, req *TransferRequest) (*TransferResult, error) {
	var result TransferResult
	err := w.http.Post(ctx, "/wallets/transfer", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSwapQuote gets a swap quote
func (w *WalletClient) GetSwapQuote(ctx context.Context, req *SwapQuoteRequest) (*SwapQuote, error) {
	var quote SwapQuote
	err := w.http.Post(ctx, "/wallets/swaps/quote", req, &quote)
	if err != nil {
		return nil, err
	}
	return &quote, nil
}

// ExecuteSwap executes a token swap
func (w *WalletClient) ExecuteSwap(ctx context.Context, req *ExecuteSwapRequest) (*SwapResult, error) {
	var result SwapResult
	err := w.http.Post(ctx, "/wallets/swaps/execute", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetNFTs returns NFTs for a wallet
func (w *WalletClient) GetNFTs(ctx context.Context, walletID string) ([]NFT, error) {
	var nfts []NFT
	err := w.http.Get(ctx, "/wallets/"+walletID+"/nfts", nil, &nfts)
	return nfts, err
}

// GetUserNFTs returns all NFTs for a user
func (w *WalletClient) GetUserNFTs(ctx context.Context, userID string) ([]NFT, error) {
	var nfts []NFT
	err := w.http.Get(ctx, "/wallets/user/"+url.PathEscape(userID)+"/nfts", nil, &nfts)
	return nfts, err
}

// AddNFT adds an NFT to wallet tracking
func (w *WalletClient) AddNFT(ctx context.Context, walletID string, req *AddNFTRequest) (*NFT, error) {
	var nft NFT
	err := w.http.Post(ctx, "/wallets/"+walletID+"/nfts", req, &nft)
	if err != nil {
		return nil, err
	}
	return &nft, nil
}

// ---------------------------------------------------------------------------
// Transaction History
// ---------------------------------------------------------------------------

// Transaction represents a single blockchain transaction
type Transaction struct {
	Hash      string `json:"hash"`
	Type      string `json:"type"` // "sent" or "received"
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	Asset     string `json:"asset"`
	Category  string `json:"category"`
	BlockNum  string `json:"block_num"`
	Timestamp string `json:"timestamp"`
}

// TransactionHistory contains transaction history for a wallet
type TransactionHistory struct {
	Address       string        `json:"address"`
	Network       string        `json:"network"`
	TotalSent     int           `json:"total_sent"`
	TotalReceived int           `json:"total_received"`
	Transactions  []Transaction `json:"transactions"`
	Error         *string       `json:"error,omitempty"`
}

// GetTransactions returns transaction history for a wallet.
func (w *WalletClient) GetTransactions(ctx context.Context, walletID string, limit, offset int) (*TransactionHistory, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", offset))
	}

	path := "/wallets/" + walletID + "/transactions"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var history TransactionHistory
	err := w.http.Get(ctx, path, nil, &history)
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// ---------------------------------------------------------------------------
// Users With Wallets
// ---------------------------------------------------------------------------

// UserWithWallets represents a user and their wallets
type UserWithWallets struct {
	UserID  string   `json:"user_id"`
	Wallets []Wallet `json:"wallets"`
}

// UsersWithWalletsResponse is a paginated list of users with wallets
type UsersWithWalletsResponse struct {
	Users  []UserWithWallets `json:"users"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

// ListUsersWithWallets returns all users who have wallets, grouped by user_id.
func (w *WalletClient) ListUsersWithWallets(ctx context.Context, limit, offset int) (*UsersWithWalletsResponse, error) {
	path := fmt.Sprintf("/wallets/users-with-wallets?limit=%d&offset=%d", limit, offset)

	var response UsersWithWalletsResponse
	err := w.http.Get(ctx, path, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// ---------------------------------------------------------------------------
// Token Management
// ---------------------------------------------------------------------------

// Token represents a registered token (native, ERC-20, or custom)
type Token struct {
	ID              string  `json:"id"`
	ContractAddress string  `json:"contract_address"`
	Network         string  `json:"network"`
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name"`
	Decimals        int     `json:"decimals"`
	LogoURL         *string `json:"logo_url,omitempty"`
	Color           *string `json:"color,omitempty"`
	CoingeckoID     *string `json:"coingecko_id,omitempty"`
	CoinmarketcapID *string `json:"coinmarketcap_id,omitempty"`
	TokenStandard   string  `json:"token_standard"`
	IsNativeWrapper bool    `json:"is_native_wrapper"`
	IsVerified      bool    `json:"is_verified"`
	IsActive        bool    `json:"is_active"`
	IsHidden        *bool   `json:"is_hidden,omitempty"`
	DisplayOrder    int     `json:"display_order"`
	IsGlobal        bool    `json:"is_global"`
}

// CreateTokenRequest registers a custom token for the tenant
type CreateTokenRequest struct {
	ContractAddress    string  `json:"contract_address"`
	Network            string  `json:"network"`
	Symbol             string  `json:"symbol"`
	Name               string  `json:"name"`
	Decimals           *int    `json:"decimals,omitempty"`
	LogoURL            *string `json:"logo_url,omitempty"`
	Color              *string `json:"color,omitempty"`
	CoingeckoID        *string `json:"coingecko_id,omitempty"`
	CoinmarketcapID    *string `json:"coinmarketcap_id,omitempty"`
	CustomPriceFeedURL *string `json:"custom_price_feed_url,omitempty"`
	IsNativeWrapper    *bool   `json:"is_native_wrapper,omitempty"`
	DisplayOrder       *int    `json:"display_order,omitempty"`
}

// UpdateTokenRequest updates a custom token
type UpdateTokenRequest struct {
	Symbol             *string `json:"symbol,omitempty"`
	Name               *string `json:"name,omitempty"`
	Decimals           *int    `json:"decimals,omitempty"`
	LogoURL            *string `json:"logo_url,omitempty"`
	Color              *string `json:"color,omitempty"`
	CoingeckoID        *string `json:"coingecko_id,omitempty"`
	CoinmarketcapID    *string `json:"coinmarketcap_id,omitempty"`
	CustomPriceFeedURL *string `json:"custom_price_feed_url,omitempty"`
	IsActive           *bool   `json:"is_active,omitempty"`
	IsHidden           *bool   `json:"is_hidden,omitempty"`
	DisplayOrder       *int    `json:"display_order,omitempty"`
}

// ListTokensOptions configures the ListTokens query
type ListTokensOptions struct {
	Network       string
	IncludeGlobal *bool
	IncludeHidden *bool
}

// CreateToken registers a custom token for the tenant.
func (w *WalletClient) CreateToken(ctx context.Context, req *CreateTokenRequest) (*Token, error) {
	var token Token
	err := w.http.Post(ctx, "/tokens", req, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// ListTokens returns all tokens available to the tenant.
func (w *WalletClient) ListTokens(ctx context.Context, opts *ListTokensOptions) ([]Token, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Network != "" {
			params.Set("network", opts.Network)
		}
		if opts.IncludeGlobal != nil {
			params.Set("include_global", fmt.Sprintf("%t", *opts.IncludeGlobal))
		}
		if opts.IncludeHidden != nil {
			params.Set("include_hidden", fmt.Sprintf("%t", *opts.IncludeHidden))
		}
	}

	path := "/tokens"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var tokens []Token
	err := w.http.Get(ctx, path, nil, &tokens)
	return tokens, err
}

// ListGlobalTokens returns well-known tokens available to all tenants.
func (w *WalletClient) ListGlobalTokens(ctx context.Context, network string) ([]Token, error) {
	path := "/tokens/global"
	if network != "" {
		path += "?network=" + url.QueryEscape(network)
	}

	var tokens []Token
	err := w.http.Get(ctx, path, nil, &tokens)
	return tokens, err
}

// GetToken returns a specific token by ID.
func (w *WalletClient) GetToken(ctx context.Context, tokenID string) (*Token, error) {
	var token Token
	err := w.http.Get(ctx, "/tokens/"+tokenID, nil, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetTokenByContract returns a token by contract address and network.
func (w *WalletClient) GetTokenByContract(ctx context.Context, contractAddress, network string) (*Token, error) {
	path := fmt.Sprintf("/tokens/by-contract/%s?network=%s", contractAddress, url.QueryEscape(network))

	var token Token
	err := w.http.Get(ctx, path, nil, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// UpdateToken updates a custom token.
func (w *WalletClient) UpdateToken(ctx context.Context, tokenID string, req *UpdateTokenRequest) (*Token, error) {
	var token Token
	err := w.http.Patch(ctx, "/tokens/"+tokenID, req, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteToken soft-deletes a custom token.
func (w *WalletClient) DeleteToken(ctx context.Context, tokenID string) error {
	var result struct {
		Message string `json:"message"`
		TokenID string `json:"token_id"`
	}
	return w.http.Request(ctx, "DELETE", "/tokens/"+tokenID, nil, &result)
}
