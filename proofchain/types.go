// Package proofchain provides a Go client for the ProofChain API.
package proofchain

import (
	"strings"
	"time"
)

// Timestamp is a custom time type that handles various timestamp formats.
type Timestamp struct {
	time.Time
}

// UnmarshalJSON handles parsing timestamps with or without timezone.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}

	// Try various formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	var err error
	for _, format := range formats {
		t.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}
	return err
}

// AttestationMode represents how an attestation was processed.
type AttestationMode string

const (
	AttestationModeDirect  AttestationMode = "direct"
	AttestationModeBatch   AttestationMode = "batch"
	AttestationModeChannel AttestationMode = "channel"
)

// EventStatus represents the status of an event.
type EventStatus string

const (
	EventStatusPending   EventStatus = "pending"
	EventStatusQueued    EventStatus = "queued"
	EventStatusConfirmed EventStatus = "confirmed"
	EventStatusSettled   EventStatus = "settled"
	EventStatusFailed    EventStatus = "failed"
)

// ChannelState represents the state of a state channel.
type ChannelState string

const (
	ChannelStateOpen     ChannelState = "open"
	ChannelStateSyncing  ChannelState = "syncing"
	ChannelStateSettling ChannelState = "settling"
	ChannelStateSettled  ChannelState = "settled"
	ChannelStateClosed   ChannelState = "closed"
)

// AttestationResult is the result of a document attestation.
type AttestationResult struct {
	ID              string          `json:"id"`
	IPFSHash        string          `json:"ipfs_hash"`
	DocumentHash    string          `json:"document_hash"`
	GatewayURL      string          `json:"gateway_url"`
	VerifyURL       string          `json:"verify_url"`
	CertificateID   string          `json:"certificate_id"`
	Status          EventStatus     `json:"status"`
	AttestationMode AttestationMode `json:"attestation_mode"`
	Timestamp       Timestamp       `json:"timestamp"`
	BlockchainTx    *string         `json:"blockchain_tx,omitempty"`
	Proof           []string        `json:"proof,omitempty"`
}

// Event represents an attested event record.
type Event struct {
	ID              string                 `json:"id"`
	EventType       string                 `json:"event_type"`
	UserID          string                 `json:"user_id"`
	IPFSHash        string                 `json:"ipfs_hash"`
	DocumentHash    *string                `json:"document_hash,omitempty"`
	GatewayURL      string                 `json:"gateway_url"`
	CertificateID   string                 `json:"certificate_id"`
	Status          EventStatus            `json:"status"`
	AttestationMode AttestationMode        `json:"attestation_mode"`
	Timestamp       Timestamp              `json:"timestamp"`
	Data            map[string]interface{} `json:"data,omitempty"`
	BlockchainTx    *string                `json:"blockchain_tx,omitempty"`
	BatchID         *string                `json:"batch_id,omitempty"`
	ChannelID       *string                `json:"channel_id,omitempty"`
}

// Channel represents a state channel for high-volume streaming.
type Channel struct {
	ChannelID string       `json:"channel_id"`
	Name      string       `json:"name"`
	State     ChannelState `json:"state"`
	CreatedAt Timestamp    `json:"created_at"`
}

// ChannelStatus contains detailed status information for a channel.
type ChannelStatus struct {
	ChannelID      string       `json:"channel_id"`
	Name           string       `json:"name"`
	State          ChannelState `json:"state"`
	EventCount     int          `json:"event_count"`
	SyncedCount    int          `json:"synced_count"`
	PendingCount   int          `json:"pending_count"`
	CreatedAt      Timestamp    `json:"created_at"`
	LastActivity   *Timestamp   `json:"last_activity,omitempty"`
	LastSettlement *Timestamp   `json:"last_settlement,omitempty"`
	MerkleRoot     *string      `json:"merkle_root,omitempty"`
}

// Settlement is the result of settling a state channel on-chain.
type Settlement struct {
	ChannelID   string    `json:"channel_id"`
	TxHash      string    `json:"tx_hash"`
	MerkleRoot  string    `json:"merkle_root"`
	EventCount  int       `json:"event_count"`
	BlockNumber int64     `json:"block_number"`
	GasUsed     int64     `json:"gas_used"`
	SettledAt   Timestamp `json:"settled_at"`
}

// Certificate represents an issued certificate.
type Certificate struct {
	CertificateID    string                 `json:"certificate_id"`
	RecipientName    string                 `json:"recipient_name"`
	RecipientEmail   *string                `json:"recipient_email,omitempty"`
	Title            string                 `json:"title"`
	Description      *string                `json:"description,omitempty"`
	IPFSHash         string                 `json:"ipfs_hash"`
	VerifyURL        string                 `json:"verify_url"`
	QRCodeURL        string                 `json:"qr_code_url"`
	IssuedAt         Timestamp              `json:"issued_at"`
	ExpiresAt        *Timestamp             `json:"expires_at,omitempty"`
	Revoked          bool                   `json:"revoked"`
	RevokedAt        *Timestamp             `json:"revoked_at,omitempty"`
	RevocationReason *string                `json:"revocation_reason,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	BlockchainTx     *string                `json:"blockchain_tx,omitempty"`
}

// Webhook represents a registered webhook endpoint.
type Webhook struct {
	ID            string     `json:"id"`
	URL           string     `json:"url"`
	Events        []string   `json:"events"`
	Active        bool       `json:"active"`
	CreatedAt     Timestamp  `json:"created_at"`
	LastTriggered *Timestamp `json:"last_triggered,omitempty"`
	FailureCount  int        `json:"failure_count"`
}

// VerificationResult is the result of verifying a document or event.
type VerificationResult struct {
	Valid           bool             `json:"valid"`
	IPFSHash        string           `json:"ipfs_hash"`
	DocumentHash    *string          `json:"document_hash,omitempty"`
	Timestamp       *Timestamp       `json:"timestamp,omitempty"`
	CertificateID   *string          `json:"certificate_id,omitempty"`
	BlockchainTx    *string          `json:"blockchain_tx,omitempty"`
	BlockNumber     *int64           `json:"block_number,omitempty"`
	AttestationMode *AttestationMode `json:"attestation_mode,omitempty"`
	ProofVerified   bool             `json:"proof_verified"`
	Message         string           `json:"message"`
}

// SearchResult is the result of searching events.
type SearchResult struct {
	Total  int     `json:"total"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
	Events []Event `json:"events"`
}

// UsageStats contains API usage statistics.
type UsageStats struct {
	TenantID          string  `json:"tenant_id,omitempty"`
	EventsThisMonth   int     `json:"events_this_month,omitempty"`
	MaxEventsPerMonth int     `json:"max_events_per_month,omitempty"`
	UsagePercentage   float64 `json:"usage_percentage,omitempty"`
	StorageUsedBytes  int64   `json:"storage_used_bytes,omitempty"`
	MaxStorageGB      int     `json:"max_storage_gb,omitempty"`
	LastEventAt       string  `json:"last_event_at,omitempty"`
	// Legacy fields
	PeriodStart       *Timestamp `json:"period_start,omitempty"`
	PeriodEnd         *Timestamp `json:"period_end,omitempty"`
	EventsCreated     int        `json:"events_created,omitempty"`
	DocumentsAttested int        `json:"documents_attested,omitempty"`
	Verifications     int        `json:"verifications,omitempty"`
	APICalls          int        `json:"api_calls,omitempty"`
	StorageBytes      int64      `json:"storage_bytes,omitempty"`
	ChannelsCreated   int        `json:"channels_created,omitempty"`
	Settlements       int        `json:"settlements,omitempty"`
}

// TenantInfo contains tenant account information.
type TenantInfo struct {
	TenantID           string  `json:"tenant_id"`
	Name               string  `json:"name"`
	Slug               string  `json:"slug,omitempty"`
	ClientID           string  `json:"client_id,omitempty"`
	Tier               string  `json:"tier,omitempty"`
	Status             string  `json:"status,omitempty"`
	ContractAddress    *string `json:"contract_address,omitempty"`
	ChainID            int     `json:"chain_id,omitempty"`
	MaxEventsPerMonth  int     `json:"max_events_per_month,omitempty"`
	EventsThisMonth    int     `json:"events_this_month,omitempty"`
	EncryptionEnabled  bool    `json:"encryption_enabled,omitempty"`
	OnchainSyncEnabled bool    `json:"onchain_sync_enabled,omitempty"`
}

// StreamAck is the acknowledgment for a streamed event.
type StreamAck struct {
	Sequence  int64  `json:"sequence"`
	ChannelID string `json:"channel_id"`
	Received  bool   `json:"received"`
}
