package folders

import (
	"encoding/json"
	"time"
)

// FolderStatus represents the status filter for file and folder operations
// Possible values: EXISTS, TRASHED, DELETED, ALL
type FolderStatus string

const (
	StatusExists  FolderStatus = "EXISTS"
	StatusTrashed FolderStatus = "TRASHED"
	StatusDeleted FolderStatus = "DELETED"
	StatusAll     FolderStatus = "ALL"
)

// CreateFolderRequest is the payload for POST /drive/folders
type CreateFolderRequest struct {
	PlainName        string `json:"plainName"`
	ParentFolderUUID string `json:"parentFolderUuid"`
	ModificationTime string `json:"modificationTime"`
	CreationTime     string `json:"creationTime"`
}

// Folder represents the response from POST/GET /drive/folders
type Folder struct {
	Type             string      `json:"type"`
	ID               int64       `json:"id"`
	ParentID         int64       `json:"parentId"`
	ParentUUID       string      `json:"parentUuid"`
	Name             string      `json:"name"`
	Parent           interface{} `json:"parent"`
	Bucket           interface{} `json:"bucket"`
	UserID           int64       `json:"userId"`
	User             interface{} `json:"user"`
	EncryptVersion   string      `json:"encryptVersion"`
	Deleted          bool        `json:"deleted"`
	DeletedAt        *time.Time  `json:"deletedAt"`
	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
	UUID             string      `json:"uuid"`
	PlainName        string      `json:"plainName"`
	Size             int64       `json:"size"`
	Removed          bool        `json:"removed"`
	RemovedAt        *time.Time  `json:"removedAt"`
	CreationTime     time.Time   `json:"creationTime"`
	ModificationTime time.Time   `json:"modificationTime"`
	Status           string      `json:"status"`
}

// File represents the response object for files in a folder
// under GET /drive/folders/content/{uuid}/files
type File struct {
	ID               int64         `json:"id"`
	FileID           string        `json:"fileId"`
	UUID             string        `json:"uuid"`
	Name             string        `json:"name"`
	PlainName        string        `json:"plainName"`
	Type             string        `json:"type"`
	FolderID         json.Number   `json:"folderId"`
	FolderUUID       string        `json:"folderUuid"`
	Folder           interface{}   `json:"folder"`
	Bucket           string        `json:"bucket"`
	UserID           json.Number   `json:"userId"`
	User             interface{}   `json:"user"`
	EncryptVersion   string        `json:"encryptVersion"`
	Size             json.Number   `json:"size"`
	Deleted          bool          `json:"deleted"`
	DeletedAt        *time.Time    `json:"deletedAt"`
	Removed          bool          `json:"removed"`
	RemovedAt        *time.Time    `json:"removedAt"`
	Shares           []interface{} `json:"shares"`
	Sharings         []interface{} `json:"sharings"`
	Thumbnails       []interface{} `json:"thumbnails"`
	CreatedAt        time.Time     `json:"createdAt"`
	UpdatedAt        time.Time     `json:"updatedAt"`
	CreationTime     time.Time     `json:"creationTime"`
	ModificationTime time.Time     `json:"modificationTime"`
	Status           string        `json:"status"`
}

// ListOptions defines common pagination and sorting parameters
// for list endpoints.
type ListOptions struct {
	Limit  int
	Offset int
	Sort   string
	Order  string
}

// TreeNode is a recursive structure representing a folder, its files, and its child folders.
type TreeNode struct {
	Folder
	Files    []File     `json:"files"`
	Children []TreeNode `json:"children"`
}
