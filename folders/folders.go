package folders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/internxt/rclone/config"
)

const foldersPath = "/folders"

// CreateFolder calls {DriveAPIURL}/folders with authorization.
// It auto‑fills CreationTime/ModificationTime if empty, checks status, and returns the newly created Folder.
func CreateFolder(cfg *config.Config, reqBody CreateFolderRequest) (*Folder, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if reqBody.CreationTime == "" {
		reqBody.CreationTime = now
	}
	if reqBody.ModificationTime == "" {
		reqBody.ModificationTime = now
	}

	endpoint := cfg.DriveAPIURL + foldersPath
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CreateFolder failed: %d %s", resp.StatusCode, string(body))
	}

	var folder Folder
	if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
		return nil, err
	}

	return &folder, nil
}

// DeleteFolders deletes a folder by UUID
func DeleteFolder(cfg *config.Config, uuid string) error {
	u, err := url.Parse(cfg.DriveAPIURL + foldersPath + "/" + uuid)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//Server returns 204 on success
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DeleteFolder failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetFolderSize retrieves the total size (in bytes) of a folder by UUID.
// Returns the size as int64, or an error.
func GetFolderSize(cfg *config.Config, uuid string) (int64, error) {
	endpoint := fmt.Sprintf("%s%s/%s/size", cfg.DriveAPIURL, foldersPath, uuid)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("GetFolderSize: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("GetFolderSize: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("GetFolderSize failed: %d %s", resp.StatusCode, string(body))
	}

	var result struct {
		Size int64 `json:"size"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("GetFolderSize: decode response: %w", err)
	}

	return result.Size, nil
}

// ListFolders lists child folders under the given parent UUID.
// Returns a slice of folders or error otherwise
func ListFolders(cfg *config.Config, parentUUID string, opts ListOptions) ([]Folder, error) {
	base := fmt.Sprintf("%s%s/content/%s/folders", cfg.DriveAPIURL, foldersPath, parentUUID)
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	q := u.Query()

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	sortField := opts.Sort
	if sortField == "" {
		sortField = "plainName"
	}
	order := opts.Order
	if order == "" {
		order = "ASC"
	}
	q.Set("offset", strconv.Itoa(offset))
	q.Set("limit", strconv.Itoa(limit))
	q.Set("sort", sortField)
	q.Set("order", order)

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ListFolders failed: %d %s", resp.StatusCode, string(body))
	}

	var wrapper struct {
		Folders []Folder `json:"folders"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, err
	}
	return wrapper.Folders, nil
}

// ListFiles lists files under the given parent folder UUID.
// Returns a slice of files or error otherwise
func ListFiles(cfg *config.Config, parentUUID string, opts ListOptions) ([]File, error) {
	base := fmt.Sprintf("%s%s/content/%s/files", cfg.DriveAPIURL, foldersPath, parentUUID)
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	q := u.Query()

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	sortField := opts.Sort
	if sortField == "" {
		sortField = "plainName"
	}
	order := opts.Order
	if order == "" {
		order = "ASC"
	}
	q.Set("offset", strconv.Itoa(offset))
	q.Set("limit", strconv.Itoa(limit))
	q.Set("sort", sortField)
	q.Set("order", order)

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ListFiles failed: %d %s", resp.StatusCode, string(body))
	}

	var wrapper struct {
		Files []File `json:"files"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&wrapper); err != nil {
		return nil, err
	}
	return wrapper.Files, nil
}

// This function will get all of the files in a folder, getting 50 at a time until completed
func ListAllFiles(cfg *config.Config, parentUUID string) ([]File, error) {
	var outFiles []File
	offset := 0
	loops := 0
	maxLoops := 10000 //Find sane number...
	for {
		files, err := ListFiles(cfg, parentUUID, ListOptions{Offset: offset})
		if err != nil {
			return nil, err
		}
		outFiles = append(outFiles, files...)
		offset += 50
		loops += 1
		if len(files) != 50 || loops >= maxLoops {
			break
		}
	}
	return outFiles, nil
}

// This function will get all of the folders in a folder, getting 50 at a time until completed
func ListAllFolders(cfg *config.Config, parentUUID string) ([]Folder, error) {
	var outFolders []Folder
	offset := 0
	loops := 0
	maxLoops := 10000 //Find sane number...
	for {
		files, err := ListFolders(cfg, parentUUID, ListOptions{Offset: offset})
		if err != nil {
			return nil, err
		}
		outFolders = append(outFolders, files...)
		offset += 50
		loops += 1
		if len(files) != 50 || loops >= maxLoops {
			break
		}
	}
	return outFolders, nil
}

// RenameFolder updates the plainName of an existing folder.
// Returns nil on HTTP 200, or an error otherwise.
func RenameFolder(cfg *config.Config, uuid, newName string) error {
	endpoint := fmt.Sprintf("%s%s/%s/meta", cfg.DriveAPIURL, foldersPath, uuid)

	payload := struct {
		PlainName string `json:"plainName"`
	}{PlainName: newName}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("RenameFolder: marshal payload: %w", err)
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("RenameFolder: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("RenameFolder: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RenameFolder failed: %d %s", resp.StatusCode, string(body))
	}
	return nil
}

// MoveFolder moves a folder into a new parent.
// Returns nil on HTTP 200, or an error otherwise.
func MoveFolder(cfg *config.Config, uuid, destUUID string) error {
	endpoint := fmt.Sprintf("%s%s/%s", cfg.DriveAPIURL, foldersPath, uuid)

	payload := struct {
		DestinationFolder string `json:"destinationFolder"`
	}{DestinationFolder: destUUID}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("MoveFolder: marshal payload: %w", err)
	}

	req, err := http.NewRequest("PATCH", endpoint, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("MoveFolder: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("MoveFolder: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MoveFolder failed: %d %s", resp.StatusCode, string(body))
	}
	return nil
}

// Tree retrieves a full recursive folder tree starting from the given UUID.
func Tree(cfg *config.Config, uuid string) (*TreeNode, error) {
	endpoint := fmt.Sprintf("%s%s/%s/tree", cfg.DriveAPIURL, foldersPath, uuid)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("Tree: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Tree: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Tree failed: %d %s", resp.StatusCode, string(body))
	}

	var response struct {
		Tree TreeNode `json:"tree"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&response); err != nil {
		return nil, fmt.Errorf("Tree: decode response: %w", err)
	}

	return &response.Tree, nil
}

// Ancestors retrieves the ancestor folders of the given folder UUID.
func Ancestors(cfg *config.Config, uuid string) ([]Folder, error) {
	endpoint := fmt.Sprintf("%s%s/%s/ancestors", cfg.DriveAPIURL, foldersPath, uuid)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("Ancestors: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ancestors: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Ancestors: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ancestors failed: %d %s", resp.StatusCode, string(body))
	}

	var folders []Folder
	if err := json.Unmarshal(body, &folders); err != nil {
		return nil, fmt.Errorf("Ancestors: decode JSON: %w", err)
	}

	return folders, nil
}

// GetFolderMetadataById retrieves metadata for a folder by numeric ID.
func GetFolderMetadataById(cfg *config.Config, id int64) (*Folder, error) {
	endpoint := fmt.Sprintf("%s%s/%d/metadata", cfg.DriveAPIURL, foldersPath, id)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("GetFolderMetadata: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetFolderMetadata: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GetFolderMetadata failed: %d %s", resp.StatusCode, string(body))
	}

	var folder Folder
	if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
		return nil, fmt.Errorf("GetFolderMetadata: decode response: %w", err)
	}

	return &folder, nil
}

// GetFolderMetadataByUUID retrieves metadata for a folder by its UUID.
func GetFolderMetadataByUUID(cfg *config.Config, uuid string) (*Folder, error) {
	endpoint := fmt.Sprintf("%s%s/%s/meta", cfg.DriveAPIURL, foldersPath, uuid)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("GetFolderMeta: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetFolderMeta: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GetFolderMeta failed: %d %s", resp.StatusCode, string(body))
	}

	var folder Folder
	if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
		return nil, fmt.Errorf("GetFolderMeta: decode response: %w", err)
	}

	return &folder, nil
}

// GetMetadataByPath retrieves folder metadata using a full path string.
func GetMetadataByPath(cfg *config.Config, path string) (*Folder, error) {
	endpoint := fmt.Sprintf("%s%s/meta?path=%s", cfg.DriveAPIURL, foldersPath, url.QueryEscape(path))

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("GetMetadataByPath: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetMetadataByPath: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GetMetadataByPath failed: %d %s", resp.StatusCode, string(body))
	}

	var folder Folder
	if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
		return nil, fmt.Errorf("GetMetadataByPath: decode response: %w", err)
	}

	return &folder, nil
}
