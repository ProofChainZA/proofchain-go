// Package proofchain provides a Go client for the ProofChain API.
package proofchain

import (
	"context"
)

// VaultFile represents a file stored in the vault.
type VaultFile struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Size          int64      `json:"size"`
	MimeType      string     `json:"mime_type"`
	FolderID      *string    `json:"folder_id,omitempty"`
	IPFSHash      string     `json:"ipfs_hash"`
	CertificateID *string    `json:"certificate_id,omitempty"`
	TxHash        *string    `json:"tx_hash,omitempty"`
	Status        string     `json:"status"`
	AccessMode    string     `json:"access_mode"`
	CreatedAt     Timestamp  `json:"created_at"`
	UpdatedAt     *Timestamp `json:"updated_at,omitempty"`
}

// VaultFolder represents a folder in the vault.
type VaultFolder struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *string   `json:"parent_id,omitempty"`
	CreatedAt Timestamp `json:"created_at"`
}

// VaultListResponse is the response from listing vault contents.
type VaultListResponse struct {
	Files      []VaultFile   `json:"files"`
	Folders    []VaultFolder `json:"folders"`
	TotalFiles int           `json:"total_files"`
	TotalSize  int64         `json:"total_size"`
}

// VaultStats contains vault storage statistics.
type VaultStats struct {
	TotalFiles   int   `json:"total_files"`
	TotalFolders int   `json:"total_folders"`
	TotalSize    int64 `json:"total_size"`
	UsedQuota    int64 `json:"used_quota"`
	MaxQuota     int64 `json:"max_quota"`
}

// VaultUploadRequest contains parameters for uploading a file.
type VaultUploadRequest struct {
	FilePath   string
	UserID     string
	FolderID   string
	AccessMode string // "private" or "public"
	Encrypt    bool
}

// VaultUploadBytesRequest contains parameters for uploading raw bytes.
type VaultUploadBytesRequest struct {
	Content    []byte
	Filename   string
	MimeType   string
	UserID     string
	FolderID   string
	AccessMode string
	Encrypt    bool
}

// VaultResource handles file vault operations.
type VaultResource struct {
	http *HTTPClient
}

// List lists all files and folders in the vault.
func (r *VaultResource) List(ctx context.Context, folderID string) (*VaultListResponse, error) {
	params := make(map[string][]string)
	if folderID != "" {
		params["folder_id"] = []string{folderID}
	}

	var result VaultListResponse
	err := r.http.Get(ctx, "/tenant/vault", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Upload uploads a file from disk to the vault.
func (r *VaultResource) Upload(ctx context.Context, req *VaultUploadRequest) (*VaultFile, error) {
	content, err := readFile(req.FilePath)
	if err != nil {
		return nil, err
	}

	filename := filepathBase(req.FilePath)
	accessMode := req.AccessMode
	if accessMode == "" {
		accessMode = "private"
	}

	fields := map[string]string{
		"user_id":     req.UserID,
		"access_mode": accessMode,
	}
	if req.FolderID != "" {
		fields["folder_id"] = req.FolderID
	}
	if req.Encrypt {
		fields["encrypt"] = "true"
	}

	var result VaultFile
	err = r.http.RequestMultipart(ctx, "/tenant/vault/upload", fields, "file", filename, content, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UploadBytes uploads raw bytes to the vault.
func (r *VaultResource) UploadBytes(ctx context.Context, req *VaultUploadBytesRequest) (*VaultFile, error) {
	accessMode := req.AccessMode
	if accessMode == "" {
		accessMode = "private"
	}
	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	fields := map[string]string{
		"user_id":     req.UserID,
		"access_mode": accessMode,
	}
	if req.FolderID != "" {
		fields["folder_id"] = req.FolderID
	}
	if req.Encrypt {
		fields["encrypt"] = "true"
	}

	var result VaultFile
	err := r.http.RequestMultipart(ctx, "/tenant/vault/upload", fields, "file", req.Filename, req.Content, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves file details by ID.
func (r *VaultResource) Get(ctx context.Context, fileID string) (*VaultFile, error) {
	var result VaultFile
	err := r.http.Get(ctx, "/tenant/vault/files/"+fileID, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Download downloads a file's content.
func (r *VaultResource) Download(ctx context.Context, fileID string) ([]byte, error) {
	return r.http.GetRaw(ctx, "/tenant/vault/files/"+fileID+"/download")
}

// Delete deletes a file from the vault.
func (r *VaultResource) Delete(ctx context.Context, fileID string) error {
	return r.http.Delete(ctx, "/tenant/vault/files/"+fileID)
}

// Move moves a file to a different folder.
func (r *VaultResource) Move(ctx context.Context, fileID, folderID string) (*VaultFile, error) {
	payload := map[string]interface{}{
		"folder_id": folderID,
	}

	var result VaultFile
	err := r.http.Post(ctx, "/tenant/vault/files/"+fileID+"/move", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateFolder creates a new folder.
func (r *VaultResource) CreateFolder(ctx context.Context, name string, parentID string) (*VaultFolder, error) {
	payload := map[string]interface{}{
		"name": name,
	}
	if parentID != "" {
		payload["parent_id"] = parentID
	}

	var result VaultFolder
	err := r.http.Post(ctx, "/tenant/vault/folders", payload, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteFolder deletes a folder.
func (r *VaultResource) DeleteFolder(ctx context.Context, folderID string) error {
	return r.http.Delete(ctx, "/tenant/vault/folders/"+folderID)
}

// Stats returns vault storage statistics.
func (r *VaultResource) Stats(ctx context.Context) (*VaultStats, error) {
	var result VaultStats
	err := r.http.Get(ctx, "/tenant/vault/stats", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Share creates a shareable link for a file.
func (r *VaultResource) Share(ctx context.Context, fileID string, expiresInHours int) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"file_id": fileID,
	}
	if expiresInHours > 0 {
		payload["expires_in_hours"] = expiresInHours
	}

	var result map[string]interface{}
	err := r.http.Post(ctx, "/tenant/vault/share", payload, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
