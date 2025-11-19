package buckets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/internxt/rclone/config"
)

type Shard struct {
	Hash string `json:"hash"`
	UUID string `json:"uuid"`
}

type FinishUploadResp struct {
	Bucket   string `json:"bucket"`
	Index    string `json:"index"`
	ID       string `json:"id"`
	Version  int    `json:"version"`
	Created  string `json:"created"`
	Renewal  string `json:"renewal"`
	Mimetype string `json:"mimetype"`
	Filename string `json:"filename"`
}

func FinishUpload(cfg *config.Config, bucketID, index string, shards []Shard) (*FinishUploadResp, error) {
	url := fmt.Sprintf("%s/v2/buckets/%s/files/finish", "https://api.internxt.com", bucketID)
	payload := map[string]interface{}{
		"index":  index,
		"shards": shards,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", cfg.BasicAuthHeader)
	req.Header.Set("internxt-version", "1.0")
	req.Header.Set("internxt-client", "drive-web")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == 500 && strings.Contains(bodyStr, "duplicate key error") {
			return nil, fmt.Errorf("file already exists on server (duplicate shard): %s", bodyStr)
		}
		return nil, fmt.Errorf("finish upload failed: status %d, %s", resp.StatusCode, bodyStr)
	}

	var result FinishUploadResp
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
