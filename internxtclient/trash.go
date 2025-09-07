package internxtclient

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
)

type TrashService struct {
	client *Client
}

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
	trashPath = "/storage/trash"

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

// GetPaginatedTrashFolders gets folders in trash
func (t *TrashService) getPaginatedTrash(itemType ItemType, limit, offset int, sort SortField, order Order, root bool) ([]File, []Folder, error) {
	/*
		url := fmt.Sprintf("%s/storage/trash/paginated?limit=%d&offset=%d&type=%s&root=%t&sort=%s&order=%s",
			cfg.DriveAPIURL, limit, offset, ItemTypeFolders, root, sort, order)
	*/
	endpoint := path.Join(trashPath, "paginated")

	opts := struct {
		Limit  int       `url:"limit"`
		Offset int       `url:"offset"`
		Type   ItemType  `url:"type"`
		Root   bool      `url:"root"`
		Sort   SortField `url:"sort"`
		Order  Order     `url:"order"`
	}{
		Limit:  limit,
		Offset: offset,
		Type:   itemType,
		Root:   root,
		Sort:   sort,
		Order:  order,
	}

	var folderWrapper struct {
		Result []Folder `json:"result"`
	}

	var fileWrapper struct {
		Result []File `json:"result"`
	}

	if itemType == ItemTypeFolders {
		if resp, err := t.client.doRequestWithStruct(APITypeDrive, http.MethodGet, endpoint, opts, nil, &folderWrapper, nil); err != nil {
			return nil, nil, t.client.GetError(endpoint, resp, err)
		}
		return nil, folderWrapper.Result, nil
	}

	if itemType == ItemTypeFiles {
		if resp, err := t.client.doRequestWithStruct(APITypeDrive, http.MethodGet, endpoint, opts, nil, &fileWrapper, nil); err != nil {
			return nil, nil, t.client.GetError(endpoint, resp, err)
		}
		return fileWrapper.Result, nil, nil
	}

	return nil, nil, fmt.Errorf("invalid item type provided: %v", itemType)
}

// GetPaginatedTrashFiles gets files in trash
func (t *TrashService) GetPaginatedTrashFiles(limit, offset int, sort SortField, order Order, root bool) ([]File, error) {
	files, _, err := t.getPaginatedTrash(ItemTypeFiles, limit, offset, sort, order, root)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (t *TrashService) GetPaginatedTrashFolders(limit, offset int, sort SortField, order Order, root bool) ([]Folder, error) {
	_, folders, err := t.getPaginatedTrash(ItemTypeFolders, limit, offset, sort, order, root)
	if err != nil {
		return nil, err
	}
	return folders, nil
}

// AddToTrash adds an item to trash
func (t *TrashService) AddToTrash(items []TrashRef) error {
	endpoint := path.Join(trashPath, "add")

	if resp, err := t.client.Post(APITypeDrive, endpoint, &TrashItemsRequest{Items: items}, nil, nil); err != nil {
		return t.client.GetError(endpoint, resp, err)
	}
	return nil
}

// DeleteAllTrash deletes the entire trash
func (t *TrashService) DeleteAllTrash() error {
	endpoint := path.Join(trashPath, "all")
	if resp, err := t.client.Delete(APITypeDrive, endpoint, nil, nil, nil); err != nil {
		return t.client.GetError(endpoint, resp, err)
	}
	return nil
}

// RequestDeleteAllTrash deletes the entire trash
func (t *TrashService) RequestDeleteAllTrash() error {
	endpoint := path.Join(trashPath, "all", "request")
	if resp, err := t.client.Delete(APITypeDrive, endpoint, nil, nil, nil); err != nil {
		return t.client.GetError(endpoint, resp, err)
	}
	return nil
}

// DeleteSpecifiedTrashItems deletes items (either files or folders) identified by TrashRef from trash
func (t *TrashService) DeleteSpecifiedTrashItems(items []TrashRef) error {
	endpoint := path.Join(trashPath)
	if resp, err := t.client.Delete(APITypeDrive, endpoint, &TrashItemsRequest{Items: items}, nil, nil); err != nil {
		return t.client.GetError(endpoint, resp, err)
	}

	return nil
}

// DeleteTrashFile deletes a file from trash. This takes FileID as input, not UUID
func (t *TrashService) DeleteTrashFile(fileID string) error {
	endpoint := path.Join(trashPath, "file", fileID)

	if resp, err := t.client.Delete(APITypeDrive, endpoint, nil, nil, nil); err != nil {
		return t.client.GetError(endpoint, resp, err)
	}
	return nil
}

// DeleteTrashFolder deletes a folder from trash.  This takes FolderID as input, not UUID
func (t *TrashService) DeleteTrashFolder(folderID int64) error {
	endpoint := path.Join(trashPath, "folder", strconv.FormatInt(folderID, 10))

	if resp, err := t.client.Delete(APITypeDrive, endpoint, nil, nil, nil); err != nil {
		return t.client.GetError(endpoint, resp, err)
	}
	return nil
}

// NewTrashFile returns a new TrashRef of type file
func (t *TrashService) NewTrashFile(uuid string) TrashRef {
	return TrashRef{UUID: uuid, Type: TrashTypeFile}
}

// NewTrashFolder returns a new TrashRef of type folder
func (t *TrashService) NewTrashFolder(uuid string) TrashRef {
	return TrashRef{UUID: uuid, Type: TrashTypeFolder}
}

// FoldersToTrashRefs converts a slice of Folder to TrashRef
func (t *TrashService) FoldersToTrashRefs(folders []Folder) []TrashRef {
	refs := make([]TrashRef, 0, len(folders))
	for _, f := range folders {
		refs = append(refs, t.NewTrashFolder(f.UUID))
	}
	return refs
}

// FilesToTrashRefs converts a slice of File to TrashRef
func (t *TrashService) FilesToTrashRefs(files []File) []TrashRef {
	refs := make([]TrashRef, 0, len(files))
	for _, f := range files {
		refs = append(refs, t.NewTrashFile(f.UUID))
	}
	return refs
}
