package buckets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/internxt/rclone/config"
)

type CreateMetaRequest struct {
	Name             string    `json:"name"`
	Bucket           string    `json:"bucket"`
	FileID           string    `json:"fileId"`
	EncryptVersion   string    `json:"encryptVersion"`
	FolderUuid       string    `json:"folderUuid"`
	Size             int64     `json:"size"`
	PlainName        string    `json:"plainName"`
	Type             string    `json:"type"`
	CreationTime     time.Time `json:"creationTime"`
	Date             time.Time `json:"date"`
	ModificationTime time.Time `json:"modificationTime"`
}

type CreateMetaResponse struct {
	UUID           string      `json:"uuid"`
	Name           string      `json:"name"`
	Bucket         string      `json:"bucket"`
	FileID         string      `json:"fileId"`
	EncryptVersion string      `json:"encryptVersion"`
	FolderUuid     string      `json:"folderUuid"`
	Size           json.Number `json:"size"`
	PlainName      string      `json:"plainName"`
	Type           string      `json:"type"`
	Created        string      `json:"created"`
}

func CreateMetaFile(cfg *config.Config, name, bucketID, fileID, encryptVersion, folderUuid, plainName, fileType string, size int64, modTime time.Time) (*CreateMetaResponse, error) {
	url := "https://api.internxt.com/drive/files"
	reqBody := CreateMetaRequest{
		Name:             name,
		Bucket:           bucketID,
		FileID:           fileID,
		EncryptVersion:   encryptVersion,
		FolderUuid:       folderUuid,
		Size:             size,
		PlainName:        plainName,
		Type:             fileType,
		CreationTime:     modTime,
		Date:             modTime,
		ModificationTime: modTime,
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("internxt-version", "v1.0.436")
	req.Header.Set("internxt-client", "drive-web")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("create meta failed: status %d, %s", resp.StatusCode, string(body))
	}
	var result CreateMetaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
