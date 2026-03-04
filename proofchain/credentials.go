package proofchain

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// CredentialType represents a tenant-defined credential template
type CredentialType struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Slug               string                 `json:"slug"`
	Description        *string                `json:"description,omitempty"`
	Category           *string                `json:"category,omitempty"`
	IconURL            *string                `json:"icon_url,omitempty"`
	BadgeColor         *string                `json:"badge_color,omitempty"`
	DisplayOrder       int                    `json:"display_order"`
	SchemaDefinition   map[string]interface{} `json:"schema_definition"`
	IsRevocable        bool                   `json:"is_revocable"`
	DefaultExpiryDays  *int                   `json:"default_expiry_days,omitempty"`
	MaxActivePerUser   int                    `json:"max_active_per_user"`
	AutoRenew          bool                   `json:"auto_renew"`
	DefaultVisibility  string                 `json:"default_visibility"`
	RequiresOptIn      bool                   `json:"requires_opt_in"`
	Status             string                 `json:"status"`
	TotalIssued        int                    `json:"total_issued"`
	TotalActive        int                    `json:"total_active"`
	CreatedAt          time.Time              `json:"created_at"`
}

// IssuedCredential represents a credential issued to a user
type IssuedCredential struct {
	ID                 string                 `json:"id"`
	CredentialTypeID   string                 `json:"credential_type_id"`
	CredentialTypeName string                 `json:"credential_type_name"`
	CredentialTypeSlug string                 `json:"credential_type_slug"`
	UserID             string                 `json:"user_id"`
	UserExternalID     *string                `json:"user_external_id,omitempty"`
	VerificationCode   string                 `json:"verification_code"`
	CredentialData     map[string]interface{} `json:"credential_data"`
	IssuedBy           *string                `json:"issued_by,omitempty"`
	IssueReason        *string                `json:"issue_reason,omitempty"`
	Visibility         string                 `json:"visibility"`
	Status             string                 `json:"status"`
	IsValid            bool                   `json:"is_valid"`
	ExpiresAt          *time.Time             `json:"expires_at,omitempty"`
	IssuedAt           time.Time              `json:"issued_at"`
	VerificationCount  int                    `json:"verification_count"`
	LastVerifiedAt     *time.Time             `json:"last_verified_at,omitempty"`
}

// CredentialVerifyResult is the public verification response
type CredentialVerifyResult struct {
	Valid              bool                   `json:"valid"`
	Status             string                 `json:"status"`
	CredentialType     string                 `json:"credential_type"`
	CredentialTypeSlug string                 `json:"credential_type_slug"`
	Category           *string                `json:"category,omitempty"`
	IconURL            *string                `json:"icon_url,omitempty"`
	BadgeColor         *string                `json:"badge_color,omitempty"`
	CredentialData     map[string]interface{} `json:"credential_data"`
	IssuedAt           time.Time              `json:"issued_at"`
	ExpiresAt          *time.Time             `json:"expires_at,omitempty"`
	Issuer             string                 `json:"issuer"`
	VerificationCount  int                    `json:"verification_count"`
}

// UserCredentialsSummary is the summary of a user's credentials
type UserCredentialsSummary struct {
	CredentialsEnabled bool               `json:"credentials_enabled"`
	TotalCredentials   int                `json:"total_credentials"`
	ActiveCredentials  int                `json:"active_credentials"`
	Credentials        []IssuedCredential `json:"credentials"`
}

// Request types

// CreateCredentialTypeRequest is the request to create a credential type
type CreateCredentialTypeRequest struct {
	Name               string                 `json:"name"`
	Slug               *string                `json:"slug,omitempty"`
	Description        *string                `json:"description,omitempty"`
	Category           *string                `json:"category,omitempty"`
	IconURL            *string                `json:"icon_url,omitempty"`
	BadgeColor         *string                `json:"badge_color,omitempty"`
	DisplayOrder       *int                   `json:"display_order,omitempty"`
	SchemaDefinition   map[string]interface{} `json:"schema_definition,omitempty"`
	IsRevocable        *bool                  `json:"is_revocable,omitempty"`
	DefaultExpiryDays  *int                   `json:"default_expiry_days,omitempty"`
	MaxActivePerUser   *int                   `json:"max_active_per_user,omitempty"`
	AutoRenew          *bool                  `json:"auto_renew,omitempty"`
	DefaultVisibility  *string                `json:"default_visibility,omitempty"`
	RequiresOptIn      *bool                  `json:"requires_opt_in,omitempty"`
	AutoIssueOnQuest   *string                `json:"auto_issue_on_quest,omitempty"`
	AutoIssueOnSegment *string                `json:"auto_issue_on_segment,omitempty"`
}

// IssueCredentialRequest is the request to issue a credential
type IssueCredentialRequest struct {
	CredentialTypeID string                 `json:"credential_type_id"`
	UserID           string                 `json:"user_id"`
	CredentialData   map[string]interface{} `json:"credential_data,omitempty"`
	IssueReason      *string                `json:"issue_reason,omitempty"`
	Visibility       *string                `json:"visibility,omitempty"`
	ExpiresAt        *string                `json:"expires_at,omitempty"`
}

// ListIssuedCredentialsOptions are options for listing issued credentials
type ListIssuedCredentialsOptions struct {
	UserID           string
	CredentialTypeID string
	Status           string
	Limit            int
	Offset           int
}

// CredentialsClient provides credential operations
type CredentialsClient struct {
	http *HTTPClient
}

// NewCredentialsClient creates a new credentials client
func NewCredentialsClient(http *HTTPClient) *CredentialsClient {
	return &CredentialsClient{http: http}
}

// ---------------------------------------------------------------------------
// Credential Types
// ---------------------------------------------------------------------------

// CreateType creates a new credential type
func (c *CredentialsClient) CreateType(ctx context.Context, req *CreateCredentialTypeRequest) (*CredentialType, error) {
	var ct CredentialType
	err := c.http.Post(ctx, "/credentials/types", req, &ct)
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// ListTypes returns credential types
func (c *CredentialsClient) ListTypes(ctx context.Context, status, category string) ([]CredentialType, error) {
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	if category != "" {
		params.Set("category", category)
	}

	var types []CredentialType
	err := c.http.Get(ctx, "/credentials/types", params, &types)
	return types, err
}

// GetType returns a credential type by ID
func (c *CredentialsClient) GetType(ctx context.Context, typeID string) (*CredentialType, error) {
	var ct CredentialType
	err := c.http.Get(ctx, "/credentials/types/"+typeID, nil, &ct)
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// UpdateType updates a credential type
func (c *CredentialsClient) UpdateType(ctx context.Context, typeID string, req *CreateCredentialTypeRequest) (*CredentialType, error) {
	var ct CredentialType
	err := c.http.Put(ctx, "/credentials/types/"+typeID, req, &ct)
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// ActivateType activates a credential type
func (c *CredentialsClient) ActivateType(ctx context.Context, typeID string) (*CredentialType, error) {
	var ct CredentialType
	err := c.http.Post(ctx, "/credentials/types/"+typeID+"/activate", nil, &ct)
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// ArchiveType archives a credential type
func (c *CredentialsClient) ArchiveType(ctx context.Context, typeID string) (*CredentialType, error) {
	var ct CredentialType
	err := c.http.Post(ctx, "/credentials/types/"+typeID+"/archive", nil, &ct)
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// ---------------------------------------------------------------------------
// Credential Issuance
// ---------------------------------------------------------------------------

// Issue issues a credential to a user
func (c *CredentialsClient) Issue(ctx context.Context, req *IssueCredentialRequest) (*IssuedCredential, error) {
	var cred IssuedCredential
	err := c.http.Post(ctx, "/credentials/issue", req, &cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// ListIssued returns issued credentials
func (c *CredentialsClient) ListIssued(ctx context.Context, opts *ListIssuedCredentialsOptions) ([]IssuedCredential, error) {
	params := url.Values{}
	if opts != nil {
		if opts.UserID != "" {
			params.Set("user_id", opts.UserID)
		}
		if opts.CredentialTypeID != "" {
			params.Set("credential_type_id", opts.CredentialTypeID)
		}
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	}

	var creds []IssuedCredential
	err := c.http.Get(ctx, "/credentials/issued", params, &creds)
	return creds, err
}

// Revoke revokes an issued credential
func (c *CredentialsClient) Revoke(ctx context.Context, credentialID string, reason string) (*IssuedCredential, error) {
	path := "/credentials/issued/" + credentialID + "/revoke"
	if reason != "" {
		path += "?reason=" + url.QueryEscape(reason)
	}
	var cred IssuedCredential
	err := c.http.Post(ctx, path, nil, &cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// Suspend suspends an issued credential
func (c *CredentialsClient) Suspend(ctx context.Context, credentialID string, reason string) (*IssuedCredential, error) {
	path := "/credentials/issued/" + credentialID + "/suspend"
	if reason != "" {
		path += "?reason=" + url.QueryEscape(reason)
	}
	var cred IssuedCredential
	err := c.http.Post(ctx, path, nil, &cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// Reinstate reinstates a suspended credential
func (c *CredentialsClient) Reinstate(ctx context.Context, credentialID string) (*IssuedCredential, error) {
	var cred IssuedCredential
	err := c.http.Post(ctx, "/credentials/issued/"+credentialID+"/reinstate", nil, &cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// ---------------------------------------------------------------------------
// User Management
// ---------------------------------------------------------------------------

// OptInUser opts a user in to identity credentials
func (c *CredentialsClient) OptInUser(ctx context.Context, userExternalID string) error {
	var result map[string]interface{}
	return c.http.Post(ctx, "/credentials/opt-in/"+userExternalID, nil, &result)
}

// OptOutUser opts a user out of identity credentials
func (c *CredentialsClient) OptOutUser(ctx context.Context, userExternalID string) error {
	var result map[string]interface{}
	return c.http.Post(ctx, "/credentials/opt-out/"+userExternalID, nil, &result)
}

// GetUserCredentials returns all credentials for a user
func (c *CredentialsClient) GetUserCredentials(ctx context.Context, userExternalID string) (*UserCredentialsSummary, error) {
	var summary UserCredentialsSummary
	err := c.http.Get(ctx, "/credentials/user/"+userExternalID, nil, &summary)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// ---------------------------------------------------------------------------
// Public Verification
// ---------------------------------------------------------------------------

// Verify verifies a credential by its verification code (public, no auth needed)
func (c *CredentialsClient) Verify(ctx context.Context, verificationCode string) (*CredentialVerifyResult, error) {
	var result CredentialVerifyResult
	err := c.http.Get(ctx, "/credentials/verify/"+verificationCode, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
