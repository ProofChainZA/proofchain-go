// Package proofchain provides a Go client for the ProofChain API.
package proofchain

import (
	"context"
	"net/url"
	"time"
)

// SearchFilters contains search filter criteria.
type SearchFilters struct {
	Query              string                 `json:"query,omitempty"`
	EventTypes         []string               `json:"event_types,omitempty"`
	EventSources       []string               `json:"event_sources,omitempty"`
	UserIDs            []string               `json:"user_ids,omitempty"`
	CertificateIDs     []string               `json:"certificate_ids,omitempty"`
	Status             string                 `json:"status,omitempty"`
	HasDocument        *bool                  `json:"has_document,omitempty"`
	HasBlockchainProof *bool                  `json:"has_blockchain_proof,omitempty"`
	FromDate           *Timestamp             `json:"from_date,omitempty"`
	ToDate             *Timestamp             `json:"to_date,omitempty"`
	DataFilters        map[string]interface{} `json:"data_filters,omitempty"`
}

// SearchRequest contains parameters for searching events.
type SearchQueryRequest struct {
	Filters     *SearchFilters `json:"filters,omitempty"`
	Offset      int            `json:"offset,omitempty"`
	Limit       int            `json:"limit,omitempty"`
	IncludeData bool           `json:"include_data,omitempty"`
}

// SearchEventResult is a single event in search results.
type SearchEventResult struct {
	ID                 string                 `json:"id"`
	CertificateID      *string                `json:"certificate_id,omitempty"`
	EventType          string                 `json:"event_type"`
	EventSource        string                 `json:"event_source"`
	UserID             string                 `json:"user_id"`
	Status             string                 `json:"status"`
	Timestamp          Timestamp              `json:"timestamp"`
	IPFSHash           *string                `json:"ipfs_hash,omitempty"`
	DocumentName       *string                `json:"document_name,omitempty"`
	DocumentType       *string                `json:"document_type,omitempty"`
	DocumentSize       *int64                 `json:"document_size,omitempty"`
	HasBlockchainProof bool                   `json:"has_blockchain_proof"`
	BlockchainTxHash   *string                `json:"blockchain_tx_hash,omitempty"`
	Data               map[string]interface{} `json:"data,omitempty"`
}

// SearchResponse is the response from a search query.
type SearchResponse struct {
	Results     []SearchEventResult    `json:"results"`
	Total       int                    `json:"total"`
	Offset      int                    `json:"offset"`
	Limit       int                    `json:"limit"`
	QueryTimeMs int                    `json:"query_time_ms"`
	Facets      map[string]interface{} `json:"facets,omitempty"`
}

// Facet is an aggregation bucket.
type Facet struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

// FacetsResponse contains faceted aggregations.
type FacetsResponse struct {
	EventTypes   []Facet `json:"event_types"`
	EventSources []Facet `json:"event_sources"`
	Statuses     []Facet `json:"statuses"`
	Users        []Facet `json:"users"`
}

// SearchStats contains search statistics.
type SearchStats struct {
	PeriodDays        int                      `json:"period_days"`
	TotalEvents       int                      `json:"total_events"`
	UniqueUsers       int                      `json:"unique_users"`
	EventsPerDay      float64                  `json:"events_per_day"`
	TopUsers          []map[string]interface{} `json:"top_users"`
	TopEventTypes     []map[string]interface{} `json:"top_event_types"`
	TotalStorageBytes int64                    `json:"total_storage_bytes"`
}

// SearchResource handles search operations.
type SearchResource struct {
	http *HTTPClient
}

// Query searches events with filters.
func (r *SearchResource) Query(ctx context.Context, req *SearchQueryRequest) (*SearchResponse, error) {
	payload := map[string]interface{}{
		"offset": req.Offset,
		"limit":  req.Limit,
	}
	if req.Limit == 0 {
		payload["limit"] = 50
	}
	if req.IncludeData {
		payload["include_data"] = true
	}
	if req.Filters != nil {
		filters := make(map[string]interface{})
		if req.Filters.Query != "" {
			filters["query"] = req.Filters.Query
		}
		if len(req.Filters.EventTypes) > 0 {
			filters["event_types"] = req.Filters.EventTypes
		}
		if len(req.Filters.EventSources) > 0 {
			filters["event_sources"] = req.Filters.EventSources
		}
		if len(req.Filters.UserIDs) > 0 {
			filters["user_ids"] = req.Filters.UserIDs
		}
		if req.Filters.Status != "" {
			filters["status"] = req.Filters.Status
		}
		if req.Filters.HasDocument != nil {
			filters["has_document"] = *req.Filters.HasDocument
		}
		if req.Filters.HasBlockchainProof != nil {
			filters["has_blockchain_proof"] = *req.Filters.HasBlockchainProof
		}
		if req.Filters.FromDate != nil {
			filters["from_date"] = req.Filters.FromDate.Format(time.RFC3339)
		}
		if req.Filters.ToDate != nil {
			filters["to_date"] = req.Filters.ToDate.Format(time.RFC3339)
		}
		if len(filters) > 0 {
			payload["filters"] = filters
		}
	}

	var result SearchResponse
	err := r.http.Post(ctx, "/search", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Quick performs a quick search across all fields.
func (r *SearchResource) Quick(ctx context.Context, query string, limit int) (*SearchResponse, error) {
	if limit == 0 {
		limit = 20
	}
	params := url.Values{
		"q":     {query},
		"limit": {intToString(limit)},
	}

	var result SearchResponse
	err := r.http.Get(ctx, "/search/quick", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ByUser gets all events for a specific user.
func (r *SearchResource) ByUser(ctx context.Context, userID string, limit, offset int) (*SearchResponse, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", intToString(limit))
	}
	if offset > 0 {
		params.Set("offset", intToString(offset))
	}

	var result SearchResponse
	err := r.http.Get(ctx, "/search/by-user/"+userID, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ByCertificate gets an event by certificate ID.
func (r *SearchResource) ByCertificate(ctx context.Context, certificateID string) (*SearchEventResult, error) {
	var result SearchEventResult
	err := r.http.Get(ctx, "/search/by-certificate/"+certificateID, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Facets gets faceted aggregations for building filter UIs.
func (r *SearchResource) Facets(ctx context.Context, fromDate, toDate *Timestamp) (*FacetsResponse, error) {
	params := url.Values{}
	if fromDate != nil {
		params.Set("from_date", fromDate.Format(time.RFC3339))
	}
	if toDate != nil {
		params.Set("to_date", toDate.Format(time.RFC3339))
	}

	var result FacetsResponse
	err := r.http.Get(ctx, "/search/facets", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Stats gets search statistics.
func (r *SearchResource) Stats(ctx context.Context) (*SearchStats, error) {
	var result SearchStats
	err := r.http.Get(ctx, "/search/stats", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
