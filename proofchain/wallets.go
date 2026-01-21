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
	FromToken    string `json:"from_token"`
	ToToken      string `json:"to_token"`
	FromAmount   string `json:"from_amount"`
	ToAmount     string `json:"to_amount"`
	ExchangeRate string `json:"exchange_rate"`
	PriceImpact  string `json:"price_impact"`
	GasEstimate  string `json:"gas_estimate"`
	ExpiresAt    string `json:"expires_at"`
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
