package proofchain

import "time"

// AttestRequest is the request for attesting a document file.
type AttestRequest struct {
	FilePath  string                 `json:"file_path"`
	UserID    string                 `json:"user_id"`
	EventType string                 `json:"event_type,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Encrypt   bool                   `json:"encrypt,omitempty"`
}

// AttestBytesRequest is the request for attesting raw bytes.
type AttestBytesRequest struct {
	Content   []byte                 `json:"content"`
	Filename  string                 `json:"filename"`
	UserID    string                 `json:"user_id"`
	EventType string                 `json:"event_type,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Encrypt   bool                   `json:"encrypt,omitempty"`
}

// CreateEventRequest is the request for creating an event.
type CreateEventRequest struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	Data      map[string]interface{} `json:"data"`
	Source    string                 `json:"event_source,omitempty"`
}

// ListEventsRequest is the request for listing events.
type ListEventsRequest struct {
	UserID    string `json:"user_id,omitempty"`
	EventType string `json:"event_type,omitempty"`
	Status    string `json:"status,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// SearchRequest is the request for searching events.
type SearchRequest struct {
	Query     string `json:"query"`
	UserID    string `json:"user_id,omitempty"`
	EventType string `json:"event_type,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Page      int    `json:"page,omitempty"`
}

// CreateChannelRequest is the request for creating a state channel.
type CreateChannelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// StreamEventRequest is the request for streaming an event to a channel.
type StreamEventRequest struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Source    string                 `json:"event_source,omitempty"`
}

// StreamBatchRequest is the request for streaming multiple events.
type StreamBatchRequest struct {
	Events []StreamEventRequest `json:"events"`
}

// IssueCertificateRequest is the request for issuing a certificate.
type IssueCertificateRequest struct {
	RecipientName  string                 `json:"recipient_name"`
	RecipientEmail string                 `json:"recipient_email,omitempty"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty"`
}

// ListCertificatesRequest is the request for listing certificates.
type ListCertificatesRequest struct {
	RecipientEmail string `json:"recipient_email,omitempty"`
	Limit          int    `json:"limit,omitempty"`
	Offset         int    `json:"offset,omitempty"`
}

// CreateWebhookRequest is the request for creating a webhook.
type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Secret string   `json:"secret,omitempty"`
}

// UpdateWebhookRequest is the request for updating a webhook.
type UpdateWebhookRequest struct {
	URL    *string   `json:"url,omitempty"`
	Events *[]string `json:"events,omitempty"`
	Active *bool     `json:"active,omitempty"`
}
