package proofchain

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Schema represents an event schema
type Schema struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	DisplayName *string    `json:"display_name,omitempty"`
	Description *string    `json:"description,omitempty"`
	Status      string     `json:"status"`
	IsDefault   bool       `json:"is_default"`
	UsageCount  int        `json:"usage_count"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// SchemaDetail includes the full schema definition
type SchemaDetail struct {
	Schema
	SchemaDefinition map[string]interface{} `json:"schema_definition"`
	YAMLContent      string                 `json:"yaml_content"`
}

// SchemaField represents a field in a schema
type SchemaField struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required,omitempty"`
	Description *string  `json:"description,omitempty"`
	Default     any      `json:"default,omitempty"`
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
	Pattern     *string  `json:"pattern,omitempty"`
	Values      []string `json:"values,omitempty"`
}

// SchemaValidationResult represents validation results
type SchemaValidationResult struct {
	Valid         bool                        `json:"valid"`
	SchemaName    string                      `json:"schema_name"`
	SchemaVersion *string                     `json:"schema_version,omitempty"`
	Errors        []SchemaValidationErrorItem `json:"errors"`
}

// SchemaValidationErrorItem represents a single validation error
type SchemaValidationErrorItem struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// SchemaListResponse is a paginated list of schemas
type SchemaListResponse struct {
	Schemas []Schema `json:"schemas"`
	Total   int      `json:"total"`
}

// SchemaUsageStats represents schema usage statistics
type SchemaUsageStats struct {
	TotalValidations      int        `json:"total_validations"`
	SuccessfulValidations int        `json:"successful_validations"`
	FailedValidations     int        `json:"failed_validations"`
	LastUsedAt            *time.Time `json:"last_used_at,omitempty"`
}

// Request types
type CreateSchemaRequest struct {
	YAMLContent string `json:"yaml_content"`
}

type ValidateDataRequest struct {
	SchemaName    string                 `json:"schema_name"`
	SchemaVersion *string                `json:"schema_version,omitempty"`
	Data          map[string]interface{} `json:"data"`
}

type ListSchemasOptions struct {
	Status string
	Search string
	Limit  int
	Offset int
}

// SchemasClient provides schema operations
type SchemasClient struct {
	http *HTTPClient
}

// NewSchemasClient creates a new schemas client
func NewSchemasClient(http *HTTPClient) *SchemasClient {
	return &SchemasClient{http: http}
}

// List returns schemas
func (s *SchemasClient) List(ctx context.Context, opts *ListSchemasOptions) (*SchemaListResponse, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.Search != "" {
			params.Set("search", opts.Search)
		}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	}

	var response SchemaListResponse
	err := s.http.Get(ctx, "/schemas", params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Get returns a schema by name and optional version
func (s *SchemasClient) Get(ctx context.Context, name string, version *string) (*SchemaDetail, error) {
	path := "/schemas/" + url.PathEscape(name)
	if version != nil {
		path += "/" + url.PathEscape(*version)
	}

	var schema SchemaDetail
	err := s.http.Get(ctx, path, nil, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

// Create creates a schema from YAML content
func (s *SchemasClient) Create(ctx context.Context, yamlContent string) (*SchemaDetail, error) {
	var schema SchemaDetail
	err := s.http.Post(ctx, "/schemas", &CreateSchemaRequest{YAMLContent: yamlContent}, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

// Update updates a schema (creates new version)
func (s *SchemasClient) Update(ctx context.Context, name string, yamlContent string) (*SchemaDetail, error) {
	var schema SchemaDetail
	err := s.http.Put(ctx, "/schemas/"+url.PathEscape(name), &CreateSchemaRequest{YAMLContent: yamlContent}, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

// Delete deletes a schema
func (s *SchemasClient) Delete(ctx context.Context, name string, version *string) error {
	path := "/schemas/" + url.PathEscape(name)
	if version != nil {
		path += "/" + url.PathEscape(*version)
	}
	return s.http.Delete(ctx, path)
}

// Activate activates a schema
func (s *SchemasClient) Activate(ctx context.Context, name string, version *string) (*Schema, error) {
	path := "/schemas/" + url.PathEscape(name)
	if version != nil {
		path += "/" + url.PathEscape(*version)
	}
	path += "/activate"

	var schema Schema
	err := s.http.Post(ctx, path, nil, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

// Deprecate deprecates a schema
func (s *SchemasClient) Deprecate(ctx context.Context, name string, version *string) (*Schema, error) {
	path := "/schemas/" + url.PathEscape(name)
	if version != nil {
		path += "/" + url.PathEscape(*version)
	}
	path += "/deprecate"

	var schema Schema
	err := s.http.Post(ctx, path, nil, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

// SetDefault sets a schema as the default for its name
func (s *SchemasClient) SetDefault(ctx context.Context, name, version string) (*Schema, error) {
	var schema Schema
	err := s.http.Post(ctx, "/schemas/"+url.PathEscape(name)+"/"+url.PathEscape(version)+"/set-default", nil, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

// Validate validates data against a schema
func (s *SchemasClient) Validate(ctx context.Context, req *ValidateDataRequest) (*SchemaValidationResult, error) {
	var result SchemaValidationResult
	err := s.http.Post(ctx, "/schemas/validate", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ValidateMultiple validates data against multiple schemas
func (s *SchemasClient) ValidateMultiple(ctx context.Context, schemaNames []string, data map[string]interface{}) ([]SchemaValidationResult, error) {
	var results []SchemaValidationResult
	err := s.http.Post(ctx, "/schemas/validate/batch", map[string]interface{}{
		"schema_names": schemaNames,
		"data":         data,
	}, &results)
	return results, err
}

// GetUsageStats returns schema usage statistics
func (s *SchemasClient) GetUsageStats(ctx context.Context, name string) (*SchemaUsageStats, error) {
	var stats SchemaUsageStats
	err := s.http.Get(ctx, "/schemas/"+url.PathEscape(name)+"/stats", nil, &stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// Clone clones a schema with a new name
func (s *SchemasClient) Clone(ctx context.Context, sourceName, newName string, newVersion *string) (*SchemaDetail, error) {
	body := map[string]interface{}{
		"new_name": newName,
	}
	if newVersion != nil {
		body["new_version"] = *newVersion
	}

	var schema SchemaDetail
	err := s.http.Post(ctx, "/schemas/"+url.PathEscape(sourceName)+"/clone", body, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}
