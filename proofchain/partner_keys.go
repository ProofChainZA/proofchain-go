package proofchain

import "context"

// OTTRequestResponse is the response from requesting a one-time token.
type OTTRequestResponse struct {
	OTT       string `json:"ott"`
	ExpiresIn int    `json:"expires_in"`
}

// OTTRedeemRequest is the request body for redeeming a one-time token.
type OTTRedeemRequest struct {
	OTT string `json:"ott"`
}

// OTTRedeemResponse is the response from redeeming a one-time token.
type OTTRedeemResponse struct {
	UserID         string                 `json:"user_id"`
	SessionTimeout int                    `json:"session_timeout"`
	SessionData    map[string]interface{} `json:"session_data"`
	JWT            string                 `json:"jwt,omitempty"`
}

// OTTConfigUpdate is the request body for updating OTT configuration.
type OTTConfigUpdate struct {
	OTTEnabled         *bool    `json:"ott_enabled,omitempty"`
	OTTTTLSeconds      *int     `json:"ott_ttl_seconds,omitempty"`
	OTTRedemptionMode  *string  `json:"ott_redemption_mode,omitempty"`
	OTTJWTTTLSeconds   *int     `json:"ott_jwt_ttl_seconds,omitempty"`
	OTTSessionDataKeys []string `json:"ott_session_data_keys,omitempty"`
}

// OTTConfigResponse is the response from getting/updating OTT configuration.
type OTTConfigResponse struct {
	PartnerKeyID       string   `json:"partner_key_id"`
	PartnerName        string   `json:"partner_name"`
	OTTEnabled         bool     `json:"ott_enabled"`
	OTTTTLSeconds      int      `json:"ott_ttl_seconds"`
	OTTRedemptionMode  string   `json:"ott_redemption_mode"`
	OTTJWTTTLSeconds   int      `json:"ott_jwt_ttl_seconds"`
	OTTSessionDataKeys []string `json:"ott_session_data_keys,omitempty"`
}

// PartnerKeysClient handles partner key OTT operations.
type PartnerKeysClient struct {
	http *HTTPClient
}

// NewPartnerKeysClient creates a new PartnerKeysClient.
func NewPartnerKeysClient(http *HTTPClient) *PartnerKeysClient {
	return &PartnerKeysClient{http: http}
}

// RequestOTT requests a one-time token for a partner key (end-user JWKS auth).
func (c *PartnerKeysClient) RequestOTT(ctx context.Context, keyID string) (*OTTRequestResponse, error) {
	var result OTTRequestResponse
	err := c.http.Post(ctx, "/partner-keys/"+keyID+"/ott/request", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RedeemOTT redeems a one-time token (partner key auth).
func (c *PartnerKeysClient) RedeemOTT(ctx context.Context, req *OTTRedeemRequest) (*OTTRedeemResponse, error) {
	var result OTTRedeemResponse
	err := c.http.Post(ctx, "/partner-keys/ott/redeem", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateOTTConfig updates OTT configuration for a partner key.
func (c *PartnerKeysClient) UpdateOTTConfig(ctx context.Context, keyID string, config *OTTConfigUpdate) (*OTTConfigResponse, error) {
	var result OTTConfigResponse
	err := c.http.Patch(ctx, "/partner-keys/"+keyID+"/ott-config", config, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
