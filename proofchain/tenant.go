// Package proofchain provides a Go client for the ProofChain API.
package proofchain

import (
	"context"
	"net/url"
)

// APIKey represents an API key for the tenant.
type APIKey struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	KeyPrefix   string     `json:"key_prefix"`
	Permissions []string   `json:"permissions"`
	CreatedAt   Timestamp  `json:"created_at"`
	LastUsedAt  *Timestamp `json:"last_used_at,omitempty"`
	ExpiresAt   *Timestamp `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
	// Key is only available when creating a new key
	Key string `json:"key,omitempty"`
}

// CreateAPIKeyRequest contains parameters for creating an API key.
type CreateAPIKeyRequest struct {
	Name          string   `json:"name"`
	Permissions   []string `json:"permissions,omitempty"`
	ExpiresInDays int      `json:"expires_in_days,omitempty"`
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
func (r *TenantResource) CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*APIKey, error) {
	payload := map[string]interface{}{
		"name": req.Name,
	}
	if len(req.Permissions) > 0 {
		payload["permissions"] = req.Permissions
	}
	if req.ExpiresInDays > 0 {
		payload["expires_in_days"] = req.ExpiresInDays
	}

	var result APIKey
	err := r.http.Post(ctx, "/tenant/api-keys", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteAPIKey deletes an API key.
func (r *TenantResource) DeleteAPIKey(ctx context.Context, keyID string) error {
	return r.http.Delete(ctx, "/tenant/api-keys/"+keyID)
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
	err := r.http.Get(ctx, "/tenant/blockchain/verify/"+certificateID, nil, &result)
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
	err := r.http.Post(ctx, "/tenant/events/"+eventID+"/settle", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
