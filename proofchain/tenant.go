// Package proofchain provides a Go client for the ProofChain API.
package proofchain

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// APIKey represents an API key for the tenant (as listed; never contains the secret).
type APIKey struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	KeyPrefix          string     `json:"key_prefix"`
	Scopes             []string   `json:"scopes"`
	RateLimitPerMinute int        `json:"rate_limit_per_minute"`
	CreatedAt          Timestamp  `json:"created_at"`
	LastUsedAt         *Timestamp `json:"last_used_at,omitempty"`
	LastUsedIP         *string    `json:"last_used_ip,omitempty"`
	ExpiresAt          *Timestamp `json:"expires_at,omitempty"`
	IsActive           bool       `json:"is_active"`
	AllowedCIDRs       []string   `json:"allowed_cidrs"`
	NotifyNewIP        bool       `json:"notify_new_ip"`
	// IsLegacyAdmin marks a pre-hardening key carrying "admin"/"*"; rotate it.
	IsLegacyAdmin bool `json:"is_legacy_admin"`
}

// CreatedAPIKey is returned by CreateAPIKey. It is the only place the secret
// Key is available. The returned key is shown once; never embed it in a
// browser or mobile bundle.
type CreatedAPIKey struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	KeyPrefix          string     `json:"key_prefix"`
	Scopes             []string   `json:"scopes"`
	RateLimitPerMinute int        `json:"rate_limit_per_minute"`
	CreatedAt          Timestamp  `json:"created_at"`
	ExpiresAt          *Timestamp `json:"expires_at,omitempty"`
	AllowedCIDRs       []string   `json:"allowed_cidrs"`
	NotifyNewIP        bool       `json:"notify_new_ip"`
	IsLegacyAdmin      bool       `json:"is_legacy_admin"`
	Warning            string     `json:"warning"`
	// Key is the full secret. Read it once and store it in a secret manager.
	Key string `json:"key"`
}

// String implements fmt.Stringer and redacts the secret key.
func (k CreatedAPIKey) String() string {
	return fmt.Sprintf("proofchain.CreatedAPIKey{ID:%q Name:%q KeyPrefix:%q Key:%q}", k.ID, k.Name, k.KeyPrefix, redactSecret(k.Key))
}

// GoString implements fmt.GoStringer (%#v) and redacts the secret key.
func (k CreatedAPIKey) GoString() string { return k.String() }

// CreateAPIKeyRequest contains parameters for creating an API key.
//
// Scopes is required and must not contain "admin" or "*" — the server rejects
// such keys and the SDK validates this client-side. Keys expire after 90 days
// unless ExpiresInDays is set. AllowedCIDRs restricts the networks the key
// may be used from. NotifyNewIP (default true on the server) controls the
// first-use-from-new-IP owner email.
type CreateAPIKeyRequest struct {
	Name               string   `json:"name"`
	Scopes             []string `json:"scopes"`
	RateLimitPerMinute int      `json:"rate_limit_per_minute,omitempty"`
	ExpiresInDays      *int     `json:"expires_in_days,omitempty"`
	AllowedCIDRs       []string `json:"allowed_cidrs,omitempty"`
	NotifyNewIP        *bool    `json:"notify_new_ip,omitempty"`
}

// validate checks the request client-side so an invalid key is never
// requested from the server.
func (r *CreateAPIKeyRequest) validate() error {
	if r == nil {
		return NewValidationError("request is required", nil)
	}
	if strings.TrimSpace(r.Name) == "" {
		return NewValidationError("name is required", nil)
	}
	if len(r.Scopes) == 0 {
		return NewValidationError("scopes is required and must be non-empty", nil)
	}
	for _, s := range r.Scopes {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "", "admin", "*":
			return NewValidationError(fmt.Sprintf("scope %q is not allowed; request only the scopes the key needs", s), nil)
		}
	}
	if r.ExpiresInDays != nil && *r.ExpiresInDays <= 0 {
		return NewValidationError("expires_in_days must be a positive integer", nil)
	}
	return nil
}

// BlockchainStats contains blockchain statistics for the tenant.
type BlockchainStats struct {
	TotalTransactions   int     `json:"total_transactions"`
	TotalGasUsed        int64   `json:"total_gas_used"`
	TotalEventsAttested int     `json:"total_events_attested"`
	PendingEvents       int     `json:"pending_events"`
	LastTransaction     *string `json:"last_transaction,omitempty"`
	ContractAddress     *string `json:"contract_address,omitempty"`
	ChainID             int     `json:"chain_id"`
	ChainName           string  `json:"chain_name"`
}

// BlockchainProof contains blockchain proof for a certificate.
type BlockchainProof struct {
	CertificateID string   `json:"certificate_id"`
	Verified      bool     `json:"verified"`
	TxHash        *string  `json:"tx_hash,omitempty"`
	BlockNumber   *int64   `json:"block_number,omitempty"`
	MerkleRoot    *string  `json:"merkle_root,omitempty"`
	MerkleProof   []string `json:"merkle_proof,omitempty"`
	LeafIndex     *int     `json:"leaf_index,omitempty"`
	ChainName     string   `json:"chain_name"`
}

// TenantResource handles tenant management operations.
type TenantResource struct {
	http *HTTPClient
}

// ListAPIKeys lists all API keys for the tenant.
func (r *TenantResource) ListAPIKeys(ctx context.Context) ([]APIKey, error) {
	var result []APIKey
	err := r.http.Get(ctx, "/tenant/api-keys", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateAPIKey creates a new API key.
//
// The returned key is shown once; never embed it in a browser or mobile
// bundle. Scopes is required and must not contain "admin" or "*".
func (r *TenantResource) CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*CreatedAPIKey, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	var result CreatedAPIKey
	err := r.http.Post(ctx, "/tenant/api-keys", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteAPIKey deletes an API key.
func (r *TenantResource) DeleteAPIKey(ctx context.Context, keyID string) error {
	return r.http.Delete(ctx, "/tenant/api-keys/"+seg(keyID))
}

// UsageDetailed gets detailed usage statistics.
func (r *TenantResource) UsageDetailed(ctx context.Context, fromDate, toDate string) (map[string]interface{}, error) {
	params := url.Values{}
	if fromDate != "" {
		params.Set("from_date", fromDate)
	}
	if toDate != "" {
		params.Set("to_date", toDate)
	}

	var result map[string]interface{}
	err := r.http.Get(ctx, "/tenant/usage/detailed", params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Context gets tenant context information.
func (r *TenantResource) Context(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := r.http.Get(ctx, "/tenant/context", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// BlockchainStats gets blockchain statistics for the tenant.
func (r *TenantResource) BlockchainStats(ctx context.Context) (*BlockchainStats, error) {
	var result BlockchainStats
	err := r.http.Get(ctx, "/tenant/blockchain/stats", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// BlockchainVerify verifies a certificate on the blockchain.
func (r *TenantResource) BlockchainVerify(ctx context.Context, certificateID string) (*BlockchainProof, error) {
	var result BlockchainProof
	err := r.http.Get(ctx, "/tenant/blockchain/verify/"+seg(certificateID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// BlockchainCertificates lists blockchain-attested certificates.
func (r *TenantResource) BlockchainCertificates(ctx context.Context, limit, offset int) (map[string]interface{}, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", intToString(limit))
	}
	if offset > 0 {
		params.Set("offset", intToString(offset))
	}

	var result map[string]interface{}
	err := r.http.Get(ctx, "/tenant/blockchain/certificates", params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// BlockchainExport exports blockchain attestation data.
func (r *TenantResource) BlockchainExport(ctx context.Context, format string) (map[string]interface{}, error) {
	params := url.Values{}
	if format != "" {
		params.Set("format", format)
	}

	var result map[string]interface{}
	err := r.http.Get(ctx, "/tenant/blockchain/export", params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ForceBatch triggers immediate batch settlement.
func (r *TenantResource) ForceBatch(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := r.http.Post(ctx, "/tenant/events/force-batch", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SettleAll settles all pending events.
func (r *TenantResource) SettleAll(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := r.http.Post(ctx, "/tenant/events/settle-all", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SettleEvent settles a specific event immediately.
func (r *TenantResource) SettleEvent(ctx context.Context, eventID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := r.http.Post(ctx, "/tenant/events/"+seg(eventID)+"/settle", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
