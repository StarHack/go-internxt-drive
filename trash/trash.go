package trash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/internxt/rclone/config"
	"github.com/internxt/rclone/folders"
)

type TrashType string
type SortField string
type Order string
type ItemType string

type TrashRef struct {
	UUID string    `json:"uuid"`
	Type TrashType `json:"type"`
}

type TrashItemsRequest struct {
	Items []TrashRef `json:"items"`
}

const (
	TrashTypeFile   TrashType = "file"
	TrashTypeFolder TrashType = "folder"

	OrderAsc  Order = "ASC"
	OrderDesc Order = "DESC"

	SortByPlainName SortField = "plainName"
	SortByUpdatedAt SortField = "updatedAt"
	SortBySize      SortField = "size"

	ItemTypeFiles   ItemType = "files"
	ItemTypeFolders ItemType = "folders"
)

// NewTrashFile returns a new TrashRef of type file
func NewTrashFile(uuid string) TrashRef {
	return TrashRef{UUID: uuid, Type: TrashTypeFile}
}

// NewTrashFile returns a new TrashRef of type folder
func NewTrashFolder(uuid string) TrashRef {
	return TrashRef{UUID: uuid, Type: TrashTypeFolder}
}

// GetPaginatedTrashFolders gets folders in trash
func GetPaginatedTrashFolders(cfg *config.Config, limit, offset int, sort SortField, order Order, root bool) ([]folders.Folder, error) {
	url := fmt.Sprintf("%s/storage/trash/paginated?limit=%d&offset=%d&type=%s&root=%t&sort=%s&order=%s",
		cfg.DriveAPIURL, limit, offset, ItemTypeFolders, root, sort, order)

	req, err := http.NewRequest(http.MethodGet, url, nil)
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
		return nil, fmt.Errorf("failed to fetch folders: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Result []json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	var result []folders.Folder
	for _, r := range raw.Result {
		var f folders.Folder
		if err := json.Unmarshal(r, &f); err == nil {
			result = append(result, f)
		}
	}
	return result, nil
}

// GetPaginatedTrashFiles gets files in trash
func GetPaginatedTrashFiles(cfg *config.Config, limit, offset int, sort SortField, order Order, root bool) ([]folders.File, error) {
	url := fmt.Sprintf("%s/storage/trash/paginated?limit=%d&offset=%d&type=%s&root=%t&sort=%s&order=%s",
		cfg.DriveAPIURL, limit, offset, ItemTypeFiles, root, sort, order)

	req, err := http.NewRequest(http.MethodGet, url, nil)
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
		return nil, fmt.Errorf("failed to fetch files: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Result []json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	var result []folders.File
	for _, r := range raw.Result {
		var f folders.File
		if err := json.Unmarshal(r, &f); err == nil {
			result = append(result, f)
		}
	}
	return result, nil
}

// AddToTrash adds an item to trash
func AddToTrash(cfg *config.Config, items []TrashRef) error {
	body, _ := json.Marshal(TrashItemsRequest{Items: items})

	req, err := http.NewRequest(http.MethodPost, cfg.DriveAPIURL+"/storage/trash/add", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add items to trash: %s", resp.Status)
	}
	return nil
}

// DeleteAllTrash deletes the entire trash
func DeleteAllTrash(cfg *config.Config) error {
	req, err := http.NewRequest(http.MethodDelete, cfg.DriveAPIURL+"/storage/trash/all", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete all trash: %s", resp.Status)
	}
	return nil
}

// RequestDeleteAllTrash deletes the entire trash
func RequestDeleteAllTrash(cfg *config.Config) error {
	req, err := http.NewRequest(http.MethodDelete, cfg.DriveAPIURL+"/storage/trash/all/request", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to request delete all trash: %s", resp.Status)
	}
	return nil
}

// DeleteSpecifiedTrashItems deletes items (either files or folders) identified by TrashRef from trash
func DeleteSpecifiedTrashItems(cfg *config.Config, items []TrashRef) error {
	body, _ := json.Marshal(TrashItemsRequest{Items: items})

	req, err := http.NewRequest(http.MethodDelete, cfg.DriveAPIURL+"/storage/trash", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete specified trash items: %s - %s", resp.Status, string(respBody))
	}
	return nil
}

// DeleteTrashFile deletes a file from trash
func DeleteTrashFile(cfg *config.Config, fileID string) error {
	url := fmt.Sprintf("%s/storage/trash/file/%s", cfg.DriveAPIURL, fileID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete trash file: %s", resp.Status)
	}
	return nil
}

// DeleteTrashFolder deletes a folder from trash
func DeleteTrashFolder(cfg *config.Config, folderID int64) error {
	url := fmt.Sprintf("%s/storage/trash/folder/%d", cfg.DriveAPIURL, folderID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete trash folder: %s", resp.Status)
	}
	return nil
}

// FoldersToTrashRefs converts a slice of Folder to TrashRef
func FoldersToTrashRefs(folders []folders.Folder) []TrashRef {
	refs := make([]TrashRef, 0, len(folders))
	for _, f := range folders {
		refs = append(refs, NewTrashFolder(f.UUID))
	}
	return refs
}

// FilesToTrashRefs converts a slice of File to TrashRef
func FilesToTrashRefs(files []folders.File) []TrashRef {
	refs := make([]TrashRef, 0, len(files))
	for _, f := range files {
		refs = append(refs, NewTrashFile(f.UUID))
	}
	return refs
}
