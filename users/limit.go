package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/internxt/rclone/config"
)

type LimitResponse struct {
	MaxSpaceBytes int64 `json:"maxSpaceBytes"`
}

// GetLimit calls {DRIVE_API_URL}/users/limit and returns the maximum available storage of the account.
func GetLimit(cfg *config.Config) (*LimitResponse, error) {
	url := fmt.Sprintf("%s/users/limit", cfg.DriveAPIURL)
	req, err := http.NewRequest("GET", url, nil)
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
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET %s returned %d: %s", url, resp.StatusCode, string(body))
	}

	var limit LimitResponse
	if err := json.NewDecoder(resp.Body).Decode(&limit); err != nil {
		return nil, err
	}

	return &limit, nil
}
