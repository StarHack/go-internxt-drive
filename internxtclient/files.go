package internxtclient

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"
	"time"
)

type FilesService struct {
	client *Client
}

// File represents the response object for files in a folder
// under GET /drive/folders/content/{uuid}/files
type File struct {
	ID               int64       `json:"id"`
	FileID           string      `json:"fileId"`
	UUID             string      `json:"uuid"`
	Name             string      `json:"name"`
	PlainName        string      `json:"plainName"`
	Type             string      `json:"type"`
	FolderID         json.Number `json:"folderId"`
	FolderUUID       string      `json:"folderUuid"`
	Folder           any         `json:"folder"`
	Bucket           string      `json:"bucket"`
	UserID           json.Number `json:"userId"`
	User             any         `json:"user"`
	EncryptVersion   string      `json:"encryptVersion"`
	Size             json.Number `json:"size"`
	Deleted          bool        `json:"deleted"`
	DeletedAt        *time.Time  `json:"deletedAt"`
	Removed          bool        `json:"removed"`
	RemovedAt        *time.Time  `json:"removedAt"`
	Shares           []any       `json:"shares"`
	Sharings         []any       `json:"sharings"`
	Thumbnails       []any       `json:"thumbnails"`
	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
	CreationTime     time.Time   `json:"creationTime"`
	ModificationTime time.Time   `json:"modificationTime"`
	Status           string      `json:"status"`
}

const filesPath = "/files"

// GetFileMeta gets file with metadata by UUID
func (f *FilesService) GetFileMeta(fileUUID string) (*File, error) {
	endpoint := path.Join(filesPath, fileUUID, "meta")

	var file File
	if resp, err := f.client.Get(APITypeDrive, endpoint, &file, nil); err != nil {
		return nil, f.client.GetError(endpoint, resp, err)
	}

	return &file, nil
}

// DeleteFile deletes a file by UUID
func (f *FilesService) DeleteFile(uuid string) error {
	endpoint := path.Join(filesPath, uuid)

	if resp, err := f.client.Delete(APITypeDrive, endpoint, nil, nil, nil); err != nil {
		return f.client.GetError(endpoint, resp, err)
	}

	return nil
}

// UpdateFileMeta updates the metadata of a file with the given UUID.
func (f *FilesService) UpdateFileMeta(fileUUID string, updated *File) (*File, error) {
	endpoint := path.Join(filesPath, fileUUID, "meta")
	var updatedFile File

	if resp, err := f.client.Put(APITypeDrive, endpoint, &updated, &updatedFile, nil); err != nil {
		return nil, f.client.GetError(endpoint, resp, err)
	}

	return &updatedFile, nil
}

// MoveFile moves the file with the given UUID to the destination folder.
func (f *FilesService) MoveFile(fileUUID, destinationFolderUUID string) (*File, error) {
	endpoint := path.Join(filesPath, fileUUID)
	var movedFile File
	body := map[string]string{
		"destinationFolder": destinationFolderUUID,
	}

	if resp, err := f.client.Patch(APITypeDrive, endpoint, &body, &movedFile, nil); err != nil {
		return nil, f.client.GetError(endpoint, resp, err)
	}

	return &movedFile, nil
}

// GetRecentFiles retrieves a list of recent files with the given limit.
func (f *FilesService) GetRecentFiles(limit int) ([]File, error) {
	endpoint := path.Join(filesPath, "recents")

	var files []File

	if resp, err := f.client.doRequestWithQuery(APITypeDrive, http.MethodGet, endpoint, map[string]string{"limit": strconv.Itoa(limit)}, nil, &files, nil); err != nil {
		return nil, f.client.GetError(endpoint, resp, err)
	}

	return files, nil
}
