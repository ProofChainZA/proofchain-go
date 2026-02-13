// Package proofchain provides a Go client for the ProofChain API.
//
// ProofChain is a blockchain-anchored document attestation platform.
// This SDK provides methods for attesting documents, managing events,
// streaming to state channels, issuing certificates, and more.
//
// Basic usage:
//
//	client := proofchain.NewClient("your-api-key")
//	result, err := client.Documents.Attest(ctx, &proofchain.AttestRequest{
//	    FilePath:  "contract.pdf",
//	    UserID:    "user@example.com",
//	})
package proofchain

import (
	"context"
	"time"
)

// Client is the main ProofChain API client.
type Client struct {
	http *HTTPClient

	// Resource managers
	Documents      *DocumentsResource
	Events         *EventsResource
	Channels       *ChannelsResource
	Certificates   *CertificatesResource
	Webhooks       *WebhooksResource
	Vault          *VaultResource
	Search         *SearchResource
	VerifyResource *VerifyResource
	Tenant         *TenantResource
	Passports      *PassportClient
	Wallets        *WalletClient
	Users          *EndUsersClient
	Rewards        *RewardsClient
	Quests         *QuestsClient
	Schemas        *SchemasClient
	DataViews      *DataViewsClient
	Cohorts        *CohortLeaderboardClient
	Fanpass        *FanpassLeaderboardClient
}

// NewClient creates a new ProofChain client.
func NewClient(apiKey string, opts ...HTTPClientOption) *Client {
	httpClient := NewHTTPClient(apiKey, opts...)
	return newClientFromHTTP(httpClient)
}

// NewClientFromEnv creates a client using the PROOFCHAIN_API_KEY environment variable.
func NewClientFromEnv(opts ...HTTPClientOption) (*Client, error) {
	httpClient, err := NewHTTPClientFromEnv(opts...)
	if err != nil {
		return nil, err
	}
	return newClientFromHTTP(httpClient), nil
}

func newClientFromHTTP(httpClient *HTTPClient) *Client {
	c := &Client{
		http: httpClient,
	}

	c.Documents = &DocumentsResource{http: httpClient}
	c.Events = &EventsResource{http: httpClient}
	c.Channels = &ChannelsResource{http: httpClient}
	c.Certificates = &CertificatesResource{http: httpClient}
	c.Webhooks = &WebhooksResource{http: httpClient}
	c.Vault = &VaultResource{http: httpClient}
	c.Search = &SearchResource{http: httpClient}
	c.VerifyResource = &VerifyResource{http: httpClient}
	c.Tenant = &TenantResource{http: httpClient}
	c.Passports = NewPassportClient(httpClient)
	c.Wallets = NewWalletClient(httpClient)
	c.Users = NewEndUsersClient(httpClient)
	c.Rewards = NewRewardsClient(httpClient)
	c.Quests = NewQuestsClient(httpClient)
	c.Schemas = NewSchemasClient(httpClient)
	c.DataViews = NewDataViewsClient(httpClient)
	c.Cohorts = NewCohortLeaderboardClient(httpClient)
	c.Fanpass = NewFanpassLeaderboardClient(httpClient)

	return c
}

// Verify verifies a document or event by its IPFS hash.
func (c *Client) Verify(ctx context.Context, ipfsHash string) (*VerificationResult, error) {
	var result VerificationResult
	err := c.http.Get(ctx, "/verify/"+ipfsHash, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// TenantInfo returns information about the current tenant.
func (c *Client) TenantInfo(ctx context.Context) (*TenantInfo, error) {
	var result TenantInfo
	err := c.http.Get(ctx, "/tenant/me", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Usage returns API usage statistics.
func (c *Client) Usage(ctx context.Context, period string) (*UsageStats, error) {
	if period == "" {
		period = "month"
	}
	var result UsageStats
	params := map[string][]string{"period": {period}}
	err := c.http.Get(ctx, "/tenant/usage", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DocumentsResource handles document attestation operations.
type DocumentsResource struct {
	http *HTTPClient
}

// Attest attests a document file.
func (r *DocumentsResource) Attest(ctx context.Context, req *AttestRequest) (*AttestationResult, error) {
	content, err := readFile(req.FilePath)
	if err != nil {
		return nil, err
	}

	filename := filepathBase(req.FilePath)
	eventType := req.EventType
	if eventType == "" {
		eventType = "document_uploaded"
	}

	fields := map[string]string{
		"user_id":    req.UserID,
		"event_type": eventType,
	}
	if req.Metadata != nil {
		metadataJSON, _ := jsonMarshal(req.Metadata)
		fields["metadata"] = string(metadataJSON)
	}
	if req.Encrypt {
		fields["encrypt"] = "1"
	}

	var result AttestationResult
	err = r.http.RequestMultipart(ctx, "/tenant/documents", fields, "file", filename, content, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// AttestBytes attests raw bytes content.
func (r *DocumentsResource) AttestBytes(ctx context.Context, req *AttestBytesRequest) (*AttestationResult, error) {
	eventType := req.EventType
	if eventType == "" {
		eventType = "document_uploaded"
	}

	fields := map[string]string{
		"user_id":    req.UserID,
		"event_type": eventType,
	}
	if req.Metadata != nil {
		metadataJSON, _ := jsonMarshal(req.Metadata)
		fields["metadata"] = string(metadataJSON)
	}
	if req.Encrypt {
		fields["encrypt"] = "1"
	}

	var result AttestationResult
	err := r.http.RequestMultipart(ctx, "/tenant/documents", fields, "file", req.Filename, req.Content, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a document by its IPFS hash.
func (r *DocumentsResource) Get(ctx context.Context, ipfsHash string) (*Event, error) {
	var result Event
	err := r.http.Get(ctx, "/tenant/events/by-hash/"+ipfsHash, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// EventsResource handles event operations.
type EventsResource struct {
	http *HTTPClient
}

// Create creates a new attestation event.
func (r *EventsResource) Create(ctx context.Context, req *CreateEventRequest) (*Event, error) {
	source := req.Source
	if source == "" {
		source = "api"
	}

	data := req.Data
	if data == nil {
		data = map[string]interface{}{}
	}

	payload := map[string]interface{}{
		"event_type":   req.EventType,
		"user_id":      req.UserID,
		"event_source": source,
		"data":         data,
	}

	var result Event
	err := r.http.Post(ctx, "/tenant/events", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves an event by ID.
func (r *EventsResource) Get(ctx context.Context, eventID string) (*Event, error) {
	var result Event
	err := r.http.Get(ctx, "/tenant/events/"+eventID, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List lists events with optional filters.
func (r *EventsResource) List(ctx context.Context, req *ListEventsRequest) ([]Event, error) {
	params := make(map[string][]string)
	if req.UserID != "" {
		params["user_id"] = []string{req.UserID}
	}
	if req.EventType != "" {
		params["event_type"] = []string{req.EventType}
	}
	if req.Status != "" {
		params["status"] = []string{req.Status}
	}
	if req.StartDate != "" {
		params["start_date"] = []string{req.StartDate}
	}
	if req.EndDate != "" {
		params["end_date"] = []string{req.EndDate}
	}
	if req.Limit > 0 {
		params["limit"] = []string{intToString(req.Limit)}
	}
	if req.Offset > 0 {
		params["offset"] = []string{intToString(req.Offset)}
	}

	var result struct {
		Events []Event `json:"events"`
	}
	err := r.http.Get(ctx, "/tenant/events", params, &result)
	if err != nil {
		return nil, err
	}
	return result.Events, nil
}

// Search searches events by query.
func (r *EventsResource) Search(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	payload := map[string]interface{}{
		"query": req.Query,
	}
	if req.UserID != "" {
		payload["user_id"] = req.UserID
	}
	if req.EventType != "" {
		payload["event_type"] = req.EventType
	}
	if req.StartDate != "" {
		payload["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		payload["end_date"] = req.EndDate
	}
	if req.Limit > 0 {
		payload["limit"] = req.Limit
	}
	if req.Page > 0 {
		payload["page"] = req.Page
	}

	var result SearchResult
	err := r.http.Post(ctx, "/search", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ByHash retrieves an event by its IPFS hash.
func (r *EventsResource) ByHash(ctx context.Context, ipfsHash string) (*Event, error) {
	var result Event
	err := r.http.Get(ctx, "/tenant/events/by-hash/"+ipfsHash, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ChannelsResource handles state channel operations.
type ChannelsResource struct {
	http *HTTPClient
}

// Create creates a new state channel.
func (r *ChannelsResource) Create(ctx context.Context, req *CreateChannelRequest) (*Channel, error) {
	payload := map[string]interface{}{
		"name": req.Name,
	}
	if req.Description != "" {
		payload["description"] = req.Description
	}

	var result Channel
	err := r.http.Post(ctx, "/channels", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a channel by ID.
func (r *ChannelsResource) Get(ctx context.Context, channelID string) (*Channel, error) {
	var result Channel
	err := r.http.Get(ctx, "/channels/"+channelID, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Status retrieves detailed status of a channel.
func (r *ChannelsResource) Status(ctx context.Context, channelID string) (*ChannelStatus, error) {
	var result ChannelStatus
	err := r.http.Get(ctx, "/channels/"+channelID+"/status", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List lists all channels.
func (r *ChannelsResource) List(ctx context.Context, limit, offset int) ([]Channel, error) {
	params := make(map[string][]string)
	if limit > 0 {
		params["limit"] = []string{intToString(limit)}
	}
	if offset > 0 {
		params["offset"] = []string{intToString(offset)}
	}

	// API returns array directly
	var result []Channel
	err := r.http.Get(ctx, "/channels", params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Stream streams an event to a channel.
func (r *ChannelsResource) Stream(ctx context.Context, channelID string, req *StreamEventRequest) (*StreamAck, error) {
	source := req.Source
	if source == "" {
		source = "sdk"
	}

	payload := map[string]interface{}{
		"event_type":   req.EventType,
		"user_id":      req.UserID,
		"event_source": source,
	}
	if req.Data != nil {
		payload["data"] = req.Data
	}

	var result StreamAck
	err := r.http.Post(ctx, "/channels/"+channelID+"/stream", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// StreamBatch streams multiple events in a single request.
func (r *ChannelsResource) StreamBatch(ctx context.Context, channelID string, events []StreamEventRequest) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"events": events,
	}

	var result map[string]interface{}
	err := r.http.Post(ctx, "/channels/"+channelID+"/stream/batch", payload, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Settle settles a channel on-chain.
func (r *ChannelsResource) Settle(ctx context.Context, channelID string) (*Settlement, error) {
	var result Settlement
	err := r.http.Post(ctx, "/channels/"+channelID+"/settle", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Close closes a channel.
func (r *ChannelsResource) Close(ctx context.Context, channelID string) (*Channel, error) {
	var result Channel
	err := r.http.Post(ctx, "/channels/"+channelID+"/close", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CertificatesResource handles certificate operations.
type CertificatesResource struct {
	http *HTTPClient
}

// Issue issues a new certificate.
func (r *CertificatesResource) Issue(ctx context.Context, req *IssueCertificateRequest) (*Certificate, error) {
	payload := map[string]interface{}{
		"recipient_name": req.RecipientName,
		"title":          req.Title,
	}
	if req.RecipientEmail != "" {
		payload["recipient_email"] = req.RecipientEmail
	}
	if req.Description != "" {
		payload["description"] = req.Description
	}
	if req.Metadata != nil {
		payload["metadata"] = req.Metadata
	}
	if req.ExpiresAt != nil {
		payload["expires_at"] = req.ExpiresAt.Format(time.RFC3339)
	}

	var result Certificate
	err := r.http.Post(ctx, "/certificates", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a certificate by ID.
func (r *CertificatesResource) Get(ctx context.Context, certificateID string) (*Certificate, error) {
	var result Certificate
	err := r.http.Get(ctx, "/certificates/"+certificateID, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List lists certificates.
func (r *CertificatesResource) List(ctx context.Context, req *ListCertificatesRequest) ([]Certificate, error) {
	params := make(map[string][]string)
	if req.RecipientEmail != "" {
		params["recipient_email"] = []string{req.RecipientEmail}
	}
	if req.Limit > 0 {
		params["limit"] = []string{intToString(req.Limit)}
	}
	if req.Offset > 0 {
		params["offset"] = []string{intToString(req.Offset)}
	}

	var result struct {
		Certificates []Certificate `json:"certificates"`
	}
	err := r.http.Get(ctx, "/certificates", params, &result)
	if err != nil {
		return nil, err
	}
	return result.Certificates, nil
}

// Revoke revokes a certificate.
func (r *CertificatesResource) Revoke(ctx context.Context, certificateID, reason string) (*Certificate, error) {
	payload := map[string]interface{}{}
	if reason != "" {
		payload["reason"] = reason
	}

	var result Certificate
	err := r.http.Post(ctx, "/certificates/"+certificateID+"/revoke", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Verify verifies a certificate.
func (r *CertificatesResource) Verify(ctx context.Context, certificateID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := r.http.Get(ctx, "/verify/certificate/"+certificateID, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// WebhooksResource handles webhook operations.
type WebhooksResource struct {
	http *HTTPClient
}

// Create creates a new webhook.
func (r *WebhooksResource) Create(ctx context.Context, req *CreateWebhookRequest) (*Webhook, error) {
	payload := map[string]interface{}{
		"url":    req.URL,
		"events": req.Events,
	}
	if req.Secret != "" {
		payload["secret"] = req.Secret
	}

	var result Webhook
	err := r.http.Post(ctx, "/webhooks", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a webhook by ID.
func (r *WebhooksResource) Get(ctx context.Context, webhookID string) (*Webhook, error) {
	var result Webhook
	err := r.http.Get(ctx, "/webhooks/"+webhookID, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List lists all webhooks.
func (r *WebhooksResource) List(ctx context.Context) ([]Webhook, error) {
	var result struct {
		Webhooks []Webhook `json:"webhooks"`
	}
	err := r.http.Get(ctx, "/webhooks", nil, &result)
	if err != nil {
		return nil, err
	}
	return result.Webhooks, nil
}

// Update updates a webhook.
func (r *WebhooksResource) Update(ctx context.Context, webhookID string, req *UpdateWebhookRequest) (*Webhook, error) {
	payload := map[string]interface{}{}
	if req.URL != nil {
		payload["url"] = *req.URL
	}
	if req.Events != nil {
		payload["events"] = *req.Events
	}
	if req.Active != nil {
		payload["active"] = *req.Active
	}

	var result Webhook
	err := r.http.Patch(ctx, "/webhooks/"+webhookID, payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a webhook.
func (r *WebhooksResource) Delete(ctx context.Context, webhookID string) error {
	return r.http.Delete(ctx, "/webhooks/"+webhookID)
}

// Test sends a test event to a webhook.
func (r *WebhooksResource) Test(ctx context.Context, webhookID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := r.http.Post(ctx, "/webhooks/"+webhookID+"/test", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
