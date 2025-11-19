package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/internxt/rclone/config"
)

type UsageResponse struct {
	Drive int64 `json:"drive"`
}

// GetUsage calls GET {DRIVE_API_URL}/users/usage and returns the account's current usage in bytes.
func GetUsage(cfg *config.Config) (*UsageResponse, error) {
	url := fmt.Sprintf("%s/users/usage", cfg.DriveAPIURL)
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

	var usage UsageResponse
	if err := json.NewDecoder(resp.Body).Decode(&usage); err != nil {
		return nil, err
	}

	return &usage, nil
}
