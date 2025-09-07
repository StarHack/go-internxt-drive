package internxtclient

import (
	"net/http"
	"path"
	"time"
)

type FoldersService struct {
	client *Client
}

const foldersPath = "/folders"

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
	Type             string     `json:"type"`
	ID               int64      `json:"id"`
	ParentID         int64      `json:"parentId"`
	ParentUUID       string     `json:"parentUuid"`
	Name             string     `json:"name"`
	Parent           any        `json:"parent"`
	Bucket           any        `json:"bucket"`
	UserID           int64      `json:"userId"`
	User             *User      `json:"user"`
	EncryptVersion   string     `json:"encryptVersion"`
	Deleted          bool       `json:"deleted"`
	DeletedAt        *time.Time `json:"deletedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	UUID             string     `json:"uuid"`
	PlainName        string     `json:"plainName"`
	Size             int64      `json:"size"`
	Removed          bool       `json:"removed"`
	RemovedAt        *time.Time `json:"removedAt"`
	CreationTime     time.Time  `json:"creationTime"`
	ModificationTime time.Time  `json:"modificationTime"`
	Status           string     `json:"status"`
	Files            []File     `json:"files"`
	Children         []Folder   `json:"children"`
}

// ListOptions defines common pagination and sorting parameters
// for list endpoints.
type ListOptions struct {
	Limit  int    `url:"limit"`
	Offset int    `url:"offset"`
	Sort   string `url:"sort,omitempty"`
	Order  string `url:"order,omitempty"`
}

func (o *ListOptions) withDefaults() *ListOptions {
	if o == nil {
		o = &ListOptions{}
	}
	if o.Limit <= 0 {
		o.Limit = 50
	}
	if o.Offset < 0 {
		o.Offset = 0
	}
	if o.Sort == "" {
		o.Sort = "plainName"
	}
	if o.Order == "" {
		o.Order = "ASC"
	}
	return o
}

// CreateFolder calls {DriveAPIURL}/folders with authorization.
// It auto‑fills CreationTime/ModificationTime if empty, checks status, and returns the newly created Folder.
func (f *FoldersService) CreateFolder(reqBody CreateFolderRequest) (*Folder, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if reqBody.CreationTime == "" {
		reqBody.CreationTime = now
	}
	if reqBody.ModificationTime == "" {
		reqBody.ModificationTime = now
	}

	endpoint := foldersPath
	var folder Folder

	if resp, err := f.client.Post(APITypeDrive, endpoint, &reqBody, &folder, nil); err != nil {
		return nil, f.client.GetError(endpoint, resp, err)
	}

	return &folder, nil
}

// DeleteFolders deletes a folder by UUID
func (f *FoldersService) DeleteFolder(uuid string) error {
	endpoint := path.Join(foldersPath, uuid)

	//Server returns 204 on success
	resp, err := f.client.Delete(APITypeDrive, endpoint, nil, nil, nil)
	if err != nil {
		return f.client.GetError(endpoint, resp, err)
	}

	return nil
}

// GetFolderSize retrieves the total size (in bytes) of a folder by UUID.
// Returns the size as int64, or an error.
func (f *FoldersService) GetFolderSize(uuid string) (int64, error) {
	endpoint := path.Join(foldersPath, uuid, "size")
	var result struct {
		Size int64 `json:"size"`
	}

	if resp, err := f.client.Get(APITypeDrive, endpoint, &result, nil); err != nil {
		return -1, f.client.GetError(endpoint, resp, err)
	}

	return result.Size, nil
}

// ListFolders lists child folders under the given parent UUID.
// Returns a slice of folders or error
func (f *FoldersService) ListFolders(parentUUID string, opts *ListOptions) ([]Folder, error) {
	opts = opts.withDefaults()
	endpoint := path.Join(foldersPath, "content", parentUUID, "folders")
	var wrapper struct {
		Folders []Folder `json:"folders"`
	}

	if resp, err := f.client.doRequestWithStruct(APITypeDrive, http.MethodGet, endpoint, opts, nil, &wrapper, nil); err != nil {
		return nil, f.client.GetError(endpoint, resp, err)
	}

	return wrapper.Folders, nil
}

// ListFiles lists child files under the given parent UUID.
// Returns a slice of files or error otherwise
func (f *FoldersService) ListFiles(parentUUID string, opts *ListOptions) ([]File, error) {
	opts = opts.withDefaults()
	endpoint := path.Join(foldersPath, "content", parentUUID, "files")
	var wrapper struct {
		Files []File `json:"files"`
	}

	if resp, err := f.client.doRequestWithStruct(APITypeDrive, http.MethodGet, endpoint, opts, nil, &wrapper, nil); err != nil {
		return nil, f.client.GetError(endpoint, resp, err)
	}

	return wrapper.Files, nil
}

// This function will get all of the files in a folder, getting 50 at a time until completed
func (f *FoldersService) ListAllFiles(parentUUID string) ([]File, error) {
	var outFiles []File
	offset := 0
	loops := 0
	maxLoops := 10000 //Find sane number...
	for {
		files, err := f.ListFiles(parentUUID, &ListOptions{Offset: offset})
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
func (f *FoldersService) ListAllFolders(parentUUID string) ([]Folder, error) {
	var outFolders []Folder
	offset := 0
	loops := 0
	maxLoops := 10000 //Find sane number...
	for {
		files, err := f.ListFolders(parentUUID, &ListOptions{Offset: offset})
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
func (f *FoldersService) RenameFolder(uuid, newName string) error {
	endpoint := path.Join(foldersPath, uuid, "meta")

	payload := struct {
		PlainName string `json:"plainName"`
	}{
		PlainName: newName,
	}

	if resp, err := f.client.Put(APITypeDrive, endpoint, payload, nil, nil); err != nil {
		return f.client.GetError(endpoint, resp, err)
	}

	return nil
}

// MoveFolder moves a folder into a new parent.
func (f *FoldersService) MoveFolder(uuid, destUUID string) error {
	endpoint := path.Join(foldersPath, uuid)

	payload := struct {
		DestinationFolder string `json:"destinationFolder"`
	}{
		DestinationFolder: destUUID,
	}

	if resp, err := f.client.Patch(APITypeDrive, endpoint, payload, nil, nil); err != nil {
		return f.client.GetError(endpoint, resp, err)
	}

	return nil
}

// Gets the metadata for a folder by its UUID
func (f *FoldersService) GetFolderMeta(folderUUID string) (*Folder, error) {
	endpoint := path.Join(foldersPath, folderUUID, "meta")

	var folder Folder

	if resp, err := f.client.Get(APITypeDrive, endpoint, &folder, nil); err != nil {
		return nil, f.client.GetError(endpoint, resp, err)
	}

	return &folder, nil
}

// Tree lists child folders and files recursively under the given parent UUID.
func (f *FoldersService) Tree(parentUUID string) (*Folder, error) {
	endpoint := path.Join(foldersPath, parentUUID, "tree")

	var wrapper struct {
		Folder Folder `json:"tree"`
	}

	if resp, err := f.client.Get(APITypeDrive, endpoint, &wrapper, nil); err != nil {
		return nil, f.client.GetError(endpoint, resp, err)
	}

	return &wrapper.Folder, nil
}
