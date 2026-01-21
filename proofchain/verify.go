// Package proofchain provides a Go client for the ProofChain API.
package proofchain

import (
	"context"
)

// CertificateVerifyResult is the result of verifying a certificate.
type CertificateVerifyResult struct {
	CertificateID string                 `json:"certificate_id"`
	Status        string                 `json:"status"`
	Type          string                 `json:"type"`
	Event         map[string]interface{} `json:"event"`
	Issuer        map[string]interface{} `json:"issuer"`
	Verification  map[string]interface{} `json:"verification"`
	Blockchain    map[string]interface{} `json:"blockchain"`
}

// IsValid returns true if the certificate is valid.
func (c *CertificateVerifyResult) IsValid() bool {
	return c.Status == "VALID"
}

// ProofVerifyRequest contains parameters for verifying a Merkle proof.
type ProofVerifyRequest struct {
	Leaf  string   `json:"leaf"`
	Proof []string `json:"proof"`
	Root  string   `json:"root"`
}

// ProofVerifyResult is the result of verifying a Merkle proof.
type ProofVerifyResult struct {
	Valid      bool   `json:"valid"`
	Message    string `json:"message"`
	VerifiedAt string `json:"verified_at"`
}

// BatchVerifyResult is the result of verifying a batch.
type BatchVerifyResult struct {
	BatchID      string                   `json:"batch_id"`
	MerkleRoot   string                   `json:"merkle_root"`
	TotalEvents  int                      `json:"total_events"`
	BlockchainTx *string                  `json:"blockchain_tx,omitempty"`
	BlockNumber  *int64                   `json:"block_number,omitempty"`
	Verified     bool                     `json:"verified"`
	Events       []map[string]interface{} `json:"events"`
}

// EventBatchProof contains the batch proof for an event.
type EventBatchProof struct {
	EventID       string   `json:"event_id"`
	CertificateID string   `json:"certificate_id"`
	BatchID       string   `json:"batch_id"`
	LeafIndex     int      `json:"leaf_index"`
	MerkleProof   []string `json:"merkle_proof"`
	MerkleRoot    string   `json:"merkle_root"`
	BlockchainTx  *string  `json:"blockchain_tx,omitempty"`
	Verified      bool     `json:"verified"`
}

// BatchVerifyItem is an item to verify in a batch request.
type BatchVerifyItem struct {
	Type string `json:"type"` // "certificate", "ipfs_hash", "event_id"
	ID   string `json:"id"`
}

// VerifyResource handles public verification operations.
type VerifyResource struct {
	http *HTTPClient
}

// Certificate verifies a certificate by ID.
func (r *VerifyResource) Certificate(ctx context.Context, certificateID string) (*CertificateVerifyResult, error) {
	var result CertificateVerifyResult
	err := r.http.Get(ctx, "/verify/cert/"+certificateID, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Event verifies an event by its IPFS hash.
func (r *VerifyResource) Event(ctx context.Context, ipfsHash string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := r.http.Get(ctx, "/verify/event/"+ipfsHash, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Proof verifies a Merkle proof cryptographically.
func (r *VerifyResource) Proof(ctx context.Context, req *ProofVerifyRequest) (*ProofVerifyResult, error) {
	var result ProofVerifyResult
	err := r.http.Post(ctx, "/verify/proof", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Batch verifies a batch by ID.
func (r *VerifyResource) Batch(ctx context.Context, batchID string) (*BatchVerifyResult, error) {
	var result BatchVerifyResult
	err := r.http.Get(ctx, "/verify/batch/"+batchID, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// EventBatchProof gets the batch proof for a specific event.
func (r *VerifyResource) EventBatchProof(ctx context.Context, eventID string) (*EventBatchProof, error) {
	var result EventBatchProof
	err := r.http.Get(ctx, "/verify/event/"+eventID+"/batch-proof", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Document verifies a document by uploading it.
func (r *VerifyResource) Document(ctx context.Context, filePath string, ipfsHash string) (map[string]interface{}, error) {
	content, err := readFile(filePath)
	if err != nil {
		return nil, err
	}

	filename := filepathBase(filePath)
	fields := map[string]string{}
	if ipfsHash != "" {
		fields["ipfs_hash"] = ipfsHash
	}

	var result map[string]interface{}
	err = r.http.RequestMultipart(ctx, "/verify/document", fields, "file", filename, content, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// BatchVerify verifies multiple items in a single request.
func (r *VerifyResource) BatchVerify(ctx context.Context, items []BatchVerifyItem) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"items": items,
	}

	var result map[string]interface{}
	err := r.http.Post(ctx, "/verify/batch", payload, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
